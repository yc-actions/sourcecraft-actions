package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/pkg/memory"
	"github.com/yc-actions/sourcecraft-actions/pkg/serviceaccount"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// Constants for metadata keys.
const (
	DockerContainerDeclarationKey = "docker-container-declaration"
	DockerComposeKey              = "docker-compose"
)

// Action inputs.
const (
	inputFolderID             = "FOLDER_ID"
	inputUserDataPath         = "USER_DATA_PATH"
	inputDockerComposePath    = "DOCKER_COMPOSE_PATH"
	inputVMName               = "VM_NAME"
	inputVMServiceAccountID   = "VM_SERVICE_ACCOUNT_ID"
	inputVMServiceAccountName = "VM_SERVICE_ACCOUNT_NAME"
	inputVMZoneID             = "VM_ZONE_ID"
	inputVMSubnetID           = "VM_SUBNET_ID"
	inputVMPublicIP           = "VM_PUBLIC_IP"
	inputVMPlatformID         = "VM_PLATFORM_ID"
	inputVMCores              = "VM_CORES"
	inputVMMemory             = "VM_MEMORY"
	inputVMDiskType           = "VM_DISK_TYPE"
	inputVMDiskSize           = "VM_DISK_SIZE"
	inputVMCoreFraction       = "VM_CORE_FRACTION"
	inputYcSaJsonCredentials  = "YC_SA_JSON_CREDENTIALS"
	inputYcIamToken           = "YC_IAM_TOKEN"
	inputYcSaID               = "YC_SA_ID"
)

// No local environment variable constants needed as they are defined in pkg/sourcecraft/sdk.go

// VMParams represents the parameters for a VM.
type VMParams struct {
	UserDataPath       string
	DockerComposePath  string
	SubnetID           string
	IPAddress          string
	ServiceAccountID   string
	ServiceAccountName string
	DiskType           string
	DiskSize           int64
	FolderID           string
	Name               string
	ZoneID             string
	PlatformID         string
	ResourcesSpec      *compute.ResourcesSpec
}

// findCoiImageID finds the latest Container Optimized Image ID.
func findCoiImageID(ctx context.Context, sdk *ycsdk.SDK) (string, error) {
	sourcecraft.StartGroup("Find COI image id")
	defer sourcecraft.EndGroup()

	req := &compute.GetImageLatestByFamilyRequest{
		FolderId: "standard-images",
		Family:   "container-optimized-image",
	}

	imageService := sdk.Compute().Image()

	image, err := imageService.GetLatestByFamily(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get latest COI image: %w", err)
	}

	sourcecraft.Info(fmt.Sprintf("COI image id: %s", image.Id))

	return image.Id, nil
}

// findVM finds a VM by name.
func findVM(ctx context.Context, sdk *ycsdk.SDK, folderID, name string) (string, error) {
	sourcecraft.StartGroup("Find VM by name")
	defer sourcecraft.EndGroup()

	instanceService := sdk.Compute().Instance()

	resp, err := instanceService.List(ctx, &compute.ListInstancesRequest{
		FolderId: folderID,
		Filter:   fmt.Sprintf("name = '%s'", name),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list instances: %w", err)
	}

	if len(resp.Instances) > 0 {
		return resp.Instances[0].Id, nil
	}

	return "", nil
}

// prepareConfig prepares a configuration file by rendering templates.
func prepareConfig(filePath string) (string, error) {
	workspace := sourcecraft.GetSourcecraftWorkspace()

	fullPath := filepath.Join(workspace, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %w", fullPath, err)
	}

	// Simple template replacement for environment variables
	// This is a basic implementation; a more sophisticated template engine might be needed
	result := string(content)

	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			result = strings.ReplaceAll(result, fmt.Sprintf("{{env.%s}}", key), value)
		}
	}

	return result, nil
}

// setOutputs sets the outputs for the Sourcecraft Action.
func setOutputs(instance *compute.Instance) {
	sourcecraft.SetOutput("INSTANCE_ID", instance.Id)

	if instance.BootDisk != nil {
		sourcecraft.SetOutput("DISK_ID", instance.BootDisk.DiskId)
	}

	if len(instance.NetworkInterfaces) > 0 &&
		instance.NetworkInterfaces[0].PrimaryV4Address != nil &&
		instance.NetworkInterfaces[0].PrimaryV4Address.OneToOneNat != nil {
		sourcecraft.SetOutput(
			"PUBLIC_IP",
			instance.NetworkInterfaces[0].PrimaryV4Address.OneToOneNat.Address,
		)
	}
}

// createVM creates a new VM.
func createVM(
	ctx context.Context,
	sdk *ycsdk.SDK,
	vmParams *VMParams,
	repoOwner, repoName string,
) error {
	coiImageID, err := findCoiImageID(ctx, sdk)
	if err != nil {
		return err
	}

	sourcecraft.StartGroup("Create new VM")
	defer sourcecraft.EndGroup()

	sourcecraft.SetOutput("VM_CREATED", "true")

	userData, err := prepareConfig(vmParams.UserDataPath)
	if err != nil {
		return fmt.Errorf("failed to prepare user data: %w", err)
	}

	dockerCompose, err := prepareConfig(vmParams.DockerComposePath)
	if err != nil {
		return fmt.Errorf("failed to prepare docker compose: %w", err)
	}

	// Get Sourcecraft SHA
	sourcecraftSHA := sourcecraft.GetSourcecraftSHA()

	metadata := map[string]string{
		"user-data":       userData,
		"docker-compose":  dockerCompose,
		"sourcecraft-sha": sourcecraftSHA,
	}

	req := &compute.CreateInstanceRequest{
		FolderId:      vmParams.FolderID,
		Name:          vmParams.Name,
		Description:   fmt.Sprintf("Created from: %s/%s", repoOwner, repoName),
		ZoneId:        vmParams.ZoneID,
		PlatformId:    vmParams.PlatformID,
		ResourcesSpec: vmParams.ResourcesSpec,
		Metadata:      metadata,
		BootDiskSpec: &compute.AttachedDiskSpec{
			Mode:       compute.AttachedDiskSpec_READ_WRITE,
			AutoDelete: true,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					TypeId: vmParams.DiskType,
					Size:   vmParams.DiskSize,
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: coiImageID,
					},
				},
			},
		},
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{
			{
				SubnetId: vmParams.SubnetID,
				PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
					OneToOneNatSpec: &compute.OneToOneNatSpec{
						IpVersion: compute.IpVersion_IPV4,
					},
				},
			},
		},
		ServiceAccountId: vmParams.ServiceAccountID,
	}

	// Set IP address if provided
	if vmParams.IPAddress != "" {
		req.NetworkInterfaceSpecs[0].PrimaryV4AddressSpec.OneToOneNatSpec.Address = vmParams.IPAddress
	}

	instanceService := sdk.Compute().Instance()

	op, err := instanceService.Create(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create instance: %w", err)
	}

	// Since we can't use sdk.WrapOperation due to SDK version differences,
	// we'll just check if the operation is done
	if !op.Done {
		return fmt.Errorf("operation is not done")
	}

	// Get instance ID directly from the operation
	instanceID := op.GetId()

	// Get instance
	instance, err := instanceService.Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	sourcecraft.Info(fmt.Sprintf("Created instance with id '%s'", instance.Id))
	setOutputs(instance)

	return nil
}

// updateMetadata updates the metadata of an existing VM.
func updateMetadata(
	ctx context.Context,
	sdk *ycsdk.SDK,
	instanceID string,
	vmParams *VMParams,
) error {
	sourcecraft.StartGroup("Update metadata")
	defer sourcecraft.EndGroup()

	sourcecraft.SetOutput("created", "false")

	userData, err := prepareConfig(vmParams.UserDataPath)
	if err != nil {
		return fmt.Errorf("failed to prepare user data: %w", err)
	}

	dockerCompose, err := prepareConfig(vmParams.DockerComposePath)
	if err != nil {
		return fmt.Errorf("failed to prepare docker compose: %w", err)
	}

	// Get Sourcecraft SHA
	sourcecraftSHA := sourcecraft.GetSourcecraftSHA()

	req := &compute.UpdateInstanceMetadataRequest{
		InstanceId: instanceID,
		Upsert: map[string]string{
			"user-data":       userData,
			"docker-compose":  dockerCompose,
			"sourcecraft-sha": sourcecraftSHA,
		},
	}

	instanceService := sdk.Compute().Instance()

	op, err := instanceService.UpdateMetadata(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update instance metadata: %w", err)
	}

	// Since we can't use sdk.WrapOperation due to SDK version differences,
	// we'll just check if the operation is done
	if !op.Done {
		return fmt.Errorf("operation is not done")
	}

	// Get instance
	instance, err := instanceService.Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
	})
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	sourcecraft.Info(fmt.Sprintf("Updated instance with id '%s'", instanceID))
	setOutputs(instance)

	return nil
}

// parseVMInputs parses the VM inputs from environment variables.
func parseVMInputs() (*VMParams, error) {
	sourcecraft.StartGroup("Parsing Action Inputs")
	defer sourcecraft.EndGroup()

	folderID := sourcecraft.GetInput(inputFolderID)
	if folderID == "" {
		return nil, fmt.Errorf("folder-id is required")
	}

	userDataPath := sourcecraft.GetInput(inputUserDataPath)
	if userDataPath == "" {
		return nil, fmt.Errorf("user-data-path is required")
	}

	dockerComposePath := sourcecraft.GetInput(inputDockerComposePath)
	if dockerComposePath == "" {
		return nil, fmt.Errorf("docker-compose-path is required")
	}

	name := sourcecraft.GetInput(inputVMName)
	if name == "" {
		return nil, fmt.Errorf("vm-name is required")
	}

	serviceAccountID := sourcecraft.GetInput(inputVMServiceAccountID)
	serviceAccountName := sourcecraft.GetInput(inputVMServiceAccountName)

	if serviceAccountID == "" && serviceAccountName == "" {
		return nil, fmt.Errorf(
			"either vm-service-account-id or vm-service-account-name should be provided",
		)
	}

	zoneID := sourcecraft.GetInput(inputVMZoneID)
	if zoneID == "" {
		zoneID = "ru-central1-a"
	}

	subnetID := sourcecraft.GetInput(inputVMSubnetID)
	if subnetID == "" {
		return nil, fmt.Errorf("vm-subnet-id is required")
	}

	ipAddress := sourcecraft.GetInput(inputVMPublicIP)

	platformID := sourcecraft.GetInput(inputVMPlatformID)
	if platformID == "" {
		platformID = "standard-v3"
	}

	cores := sourcecraft.GetInt64Input(inputVMCores, 2)

	memoryStr := sourcecraft.GetInput(inputVMMemory)
	if memoryStr == "" {
		memoryStr = "2Gb"
	}

	memoryValue, err := memory.ParseMemory(memoryStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vm-memory: %w", err)
	}

	diskTypeStr := sourcecraft.GetInput(inputVMDiskType)
	if diskTypeStr == "" {
		diskTypeStr = "network-ssd"
	}

	diskSizeStr := sourcecraft.GetInput(inputVMDiskSize)
	if diskSizeStr == "" {
		diskSizeStr = "30Gb"
	}

	diskSize, err := memory.ParseMemory(diskSizeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse vm-disk-size: %w", err)
	}

	coreFraction := sourcecraft.GetInt64Input(inputVMCoreFraction, 100)

	return &VMParams{
		UserDataPath:       userDataPath,
		DockerComposePath:  dockerComposePath,
		SubnetID:           subnetID,
		IPAddress:          ipAddress,
		ServiceAccountID:   serviceAccountID,
		ServiceAccountName: serviceAccountName,
		DiskType:           diskTypeStr,
		DiskSize:           diskSize,
		FolderID:           folderID,
		Name:               name,
		ZoneID:             zoneID,
		PlatformID:         platformID,
		ResourcesSpec: &compute.ResourcesSpec{
			Memory:       memoryValue,
			Cores:        cores,
			CoreFraction: coreFraction,
		},
	}, nil
}

// detectMetadataConflict checks if there's a metadata conflict.
func detectMetadataConflict(ctx context.Context, sdk *ycsdk.SDK, instanceID string) error {
	sourcecraft.StartGroup("Check metadata")
	defer sourcecraft.EndGroup()

	instanceService := sdk.Compute().Instance()

	instance, err := instanceService.Get(ctx, &compute.GetInstanceRequest{
		InstanceId: instanceID,
		View:       compute.InstanceView_FULL,
	})
	if err != nil {
		return fmt.Errorf("failed to get instance: %w", err)
	}

	if _, ok := instance.Metadata[DockerContainerDeclarationKey]; ok {
		return fmt.Errorf(
			"provided VM was created with '%s' metadata key. "+
				"It will conflict with '%s' key this action using. "+
				"Either recreate VM using docker-compose as container definition "+
				"or let the action create the new one by dropping 'name' parameter",
			DockerContainerDeclarationKey, DockerComposeKey)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Parse inputs
	ycSaJsonCredentials := sourcecraft.GetInput(inputYcSaJsonCredentials)
	ycIamToken := sourcecraft.GetInput(inputYcIamToken)
	ycSaID := sourcecraft.GetInput(inputYcSaID)

	// Create SDK
	var sdk *ycsdk.SDK

	var err error

	if ycSaJsonCredentials != "" {
		// Create credentials
		var key *iamkey.Key

		key, err = iamkey.ReadFromJSONBytes([]byte(ycSaJsonCredentials))
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to read service account JSON: %v", err))

			return
		}

		var credentials ycsdk.Credentials

		credentials, err = ycsdk.ServiceAccountKey(key)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create credentials: %v", err))

			return
		}

		// Create SDK
		sdk, err = ycsdk.Build(ctx, ycsdk.Config{
			Credentials: credentials,
		})
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create SDK: %v", err))

			return
		}

		sourcecraft.Info("Parsed Service account JSON")
	} else if ycIamToken != "" {
		// Create SDK with IAM token
		sdk, err = ycsdk.Build(ctx, ycsdk.Config{
			Credentials: ycsdk.NewIAMTokenCredentials(ycIamToken),
		})
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create SDK: %v", err))

			return
		}

		sourcecraft.Info("Using IAM token")
	} else if ycSaID != "" {
		// In Sourcecraft, we would use getIDToken() to get a Sourcecraft token
		// Since there's no direct equivalent in Go, we'll use a different approach
		// This is a placeholder for now
		sourcecraft.SetFailed("Token exchange not implemented in Go version yet")

		return
	} else {
		sourcecraft.SetFailed("No credentials provided")

		return
	}

	// Parse VM inputs
	vmParams, err := parseVMInputs()
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to parse VM inputs: %v", err))

		return
	}

	sourcecraft.Info(fmt.Sprintf("Folder ID: %s, name: %s", vmParams.FolderID, vmParams.Name))

	// Resolve service account ID if name is provided
	if vmParams.ServiceAccountID == "" && vmParams.ServiceAccountName != "" {
		serviceAccountID, err := serviceaccount.ResolveID(
			ctx,
			sdk,
			vmParams.FolderID,
			"",
			vmParams.ServiceAccountName,
		)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to resolve service account: %v", err))

			return
		}

		if serviceAccountID == "" {
			sourcecraft.SetFailed(
				fmt.Sprintf(
					"There is no service account '%s' in folder %s",
					vmParams.ServiceAccountName,
					vmParams.FolderID,
				),
			)

			return
		}

		vmParams.ServiceAccountID = serviceAccountID
	}

	// Find VM by name
	vmID, err := findVM(ctx, sdk, vmParams.FolderID, vmParams.Name)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to find VM: %v", err))

		return
	}

	// Create or update VM
	if vmID == "" {
		// Get repository owner and name from environment variables
		repoOwner := sourcecraft.GetSourcecraftRepositoryOwner()
		repoName := sourcecraft.GetSourcecraftRepository()

		err = createVM(ctx, sdk, vmParams, repoOwner, repoName)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create VM: %v", err))

			return
		}
	} else {
		// Check for metadata conflict
		err = detectMetadataConflict(ctx, sdk, vmID)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Metadata conflict detected: %v", err))

			return
		}

		// Update VM metadata
		err = updateMetadata(ctx, sdk, vmID, vmParams)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to update VM metadata: %v", err))

			return
		}
	}
}
