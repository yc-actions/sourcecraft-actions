package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// UserDataScriptParams contains parameters for building a user data script.
type UserDataScriptParams struct {
	GithubRegistrationToken string
	Label                   string
	RunnerHomeDir           string
	User                    string
	SSHPublicKey            string
}

// BuildUserDataScript builds a user data script for the VM.
func BuildUserDataScript(params UserDataScriptParams) string {
	var script []string

	if params.RunnerHomeDir != "" {
		// If runner home directory is specified, we expect the actions-runner software (and dependencies)
		// to be pre-installed in the image, so we simply cd into that directory and then start the runner
		script = []string{
			"#!/bin/bash",
			fmt.Sprintf(`cd "%s"`, params.RunnerHomeDir),
			"export RUNNER_ALLOW_RUNASROOT=1",
			fmt.Sprintf("/self-hosted-processor init  --root-dir /tmp --token %s --tags %s > config.yaml",
				params.GithubRegistrationToken, params.Label),
			"./self-hosted-processor run --config-path config.yaml",
		}
	} else {
		// If runner home directory is not specified, we need to download and install the runner
		script = []string{
			"#!/bin/bash",
			"mkdir actions-runner && cd actions-runner",
			"case $(uname -m) in aarch64) ARCH=\"arm64\" ;; amd64|x86_64) ARCH=\"x64\" ;; esac && export RUNNER_ARCH=${ARCH}",
			`curl -O -L https://storage.yandexcloud.net/src-processor-downloads/self-hosted-processor-latest/linux/amd64/self-hosted-processor`,
			`chmod +x self-hosted-processor`,
			fmt.Sprintf("./self-hosted-processor init --root-dir /tmp --token %s --tags %s > config.yaml",
				params.GithubRegistrationToken, params.Label),
			"./self-hosted-processor run --config-path config.yaml",
		}
	}

	if params.User != "" && params.SSHPublicKey != "" {
		// If user and SSH public key are specified, we need to create a cloud-init configuration
		cloudInit := fmt.Sprintf(`#cloud-config
ssh_pwauth: no
users:
  - name: %s
    sudo: ALL=(ALL) NOPASSWD:ALL
    shell: /bin/bash
    ssh_authorized_keys:
      - "%s"
runcmd:
`, params.User, params.SSHPublicKey)

		// Add the script commands to the cloud-init configuration
		for i, cmd := range script {
			if i > 0 { // Skip the shebang line
				cloudInit += fmt.Sprintf("  - %s\n", cmd)
			}
		}

		return cloudInit
	}

	// If user and SSH public key are not specified, return the script as is
	result := script[0] + "\n"
	for _, cmd := range script[1:] {
		result += cmd + "\n"
	}

	return result
}

// CreateVM creates a VM in Yandex Cloud.
func CreateVM(ctx context.Context, sdk *ycsdk.SDK, config *Config, githubRegistrationToken, label string) (string, error) {
	sourcecraft.StartGroup("Create VM")
	defer sourcecraft.EndGroup()

	// Create instance service client
	instanceService := sdk.Compute().Instance()

	// Create secondary disk specs if needed
	var secondaryDiskSpecs []*compute.AttachedDiskSpec
	if config.Input.SecondDiskSize > 0 {
		secondaryDiskSpecs = append(secondaryDiskSpecs, &compute.AttachedDiskSpec{
			AutoDelete: true,
			Mode:       compute.AttachedDiskSpec_READ_WRITE,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					TypeId: config.Input.SecondDiskType,
					Size:   config.Input.SecondDiskSize,
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: config.Input.SecondDiskImageID,
					},
				},
			},
		})
	}

	// Create network interface spec
	primaryV4AddressSpec := &compute.PrimaryAddressSpec{}
	if config.Input.PublicIP {
		primaryV4AddressSpec.OneToOneNatSpec = &compute.OneToOneNatSpec{
			IpVersion: compute.IpVersion_IPV4,
		}
	}

	networkInterfaceSpec := &compute.NetworkInterfaceSpec{
		SubnetId:             config.Input.SubnetID,
		PrimaryV4AddressSpec: primaryV4AddressSpec,
	}

	// Create labels
	labels := make(map[string]string)
	if config.Input.TTL != nil {
		// Set `expires` label to the current time + TTL Duration
		// Instance won't automatically be destroyed by Yandex.Cloud, you should handle it yourself
		// For example, by using Cron trigger that will call Cloud Function to destroy the instance.
		labels["expires"] = fmt.Sprintf("%d", time.Now().Add(*config.Input.TTL).Unix())
	}

	// Build user data script
	userDataParams := UserDataScriptParams{
		GithubRegistrationToken: githubRegistrationToken,
		Label:                   label,
		RunnerHomeDir:           config.Input.RunnerHomeDir,
		User:                    config.Input.User,
		SSHPublicKey:            config.Input.SSHPublicKey,
	}
	userData := BuildUserDataScript(userDataParams)

	// Create VM
	createRequest := &compute.CreateInstanceRequest{
		FolderId:    config.Input.FolderID,
		Name:        fmt.Sprintf("runner-%s", label),
		Description: fmt.Sprintf("Runner for: %s/%s", config.GithubContext.Owner, config.GithubContext.Repo),
		ZoneId:      config.Input.ZoneID,
		PlatformId:  config.Input.PlatformID,
		ResourcesSpec: &compute.ResourcesSpec{
			Memory:       config.Input.ResourcesSpec.Memory,
			Cores:        int64(config.Input.ResourcesSpec.Cores),
			CoreFraction: int64(config.Input.ResourcesSpec.CoreFraction),
		},
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			Mode:       compute.AttachedDiskSpec_READ_WRITE,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					TypeId: config.Input.DiskType,
					Size:   config.Input.DiskSize,
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: config.Input.ImageID,
					},
				},
			},
		},
		SecondaryDiskSpecs:    secondaryDiskSpecs,
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{networkInterfaceSpec},
		Metadata: map[string]string{
			"user-data": userData,
		},
		Labels:           labels,
		ServiceAccountId: config.Input.ServiceAccountID,
	}

	// Create instance
	op, err := instanceService.Create(ctx, createRequest)
	if err != nil {
		return "", fmt.Errorf("failed to create instance: %w", err)
	}

	// Since we can't use sdk.WrapOperation due to SDK version differences,
	// we'll just check if the operation is done
	if !op.Done {
		return "", fmt.Errorf("operation is not done")
	}

	// Get instance ID directly from the operation
	instanceID := op.GetId()
	sourcecraft.Info(fmt.Sprintf("Created instance with id '%s'", instanceID))

	return instanceID, nil
}

// DestroyVM destroys a VM in Yandex Cloud.
func DestroyVM(ctx context.Context, sdk *ycsdk.SDK, instanceID string) error {
	sourcecraft.StartGroup("Destroy VM")
	defer sourcecraft.EndGroup()

	// Create instance service client
	instanceService := sdk.Compute().Instance()

	// Delete instance
	op, err := instanceService.Delete(ctx, &compute.DeleteInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	// Since we can't use sdk.WrapOperation due to SDK version differences,
	// we'll just check if the operation is done
	if !op.Done {
		return fmt.Errorf("operation is not done")
	}

	sourcecraft.Info(fmt.Sprintf("Destroyed instance with id '%s'", instanceID))

	return nil
}
