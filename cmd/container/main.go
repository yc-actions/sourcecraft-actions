package main

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/access"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/containers/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/internal/container"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Input constants.
const (
	inputFolderID            = "folder-id"
	inputContainerName       = "container-name"
	inputPublic              = "public"
	inputYcSaJsonCredentials = "yc-sa-json-credentials"
	inputYcIamToken          = "yc-iam-token"
)

// createRevision creates a new revision for a container.
func createRevision(
	ctx context.Context,
	sdk *ycsdk.SDK,
	options *container.CreateRevisionOptions,
) (string, error) {
	// Create the request
	req := &containers.DeployContainerRevisionRequest{
		ContainerId: options.ContainerID,
		Resources: &containers.Resources{
			Memory:       options.MemoryValue,
			Cores:        options.Cores,
			CoreFraction: options.CoreFraction,
		},
		ExecutionTimeout: &durationpb.Duration{Seconds: options.ExecutionTimeout},
		Concurrency:      options.Concurrency,
		ServiceAccountId: options.ServiceAccountID,
	}

	// Set image spec
	req.ImageSpec = &containers.ImageSpec{
		ImageUrl:   options.ImageURL,
		WorkingDir: options.WorkingDir,
	}

	// Set commands if provided
	if len(options.Commands) > 0 {
		// Create a Command struct with the commands
		req.ImageSpec.Command = &containers.Command{
			Command: options.Commands,
		}
	}

	// Set args if provided
	if len(options.Args) > 0 {
		// Create an Args struct with the args
		req.ImageSpec.Args = &containers.Args{
			Args: options.Args,
		}
	}

	// Set environment variables if provided
	if len(options.Env) > 0 {
		req.ImageSpec.Environment = options.Env
	}

	// Set network ID if provided
	if options.NetworkID != "" {
		req.Connectivity = &containers.Connectivity{
			NetworkId: options.NetworkID,
		}
	}

	// Set provisioned instances if provided
	if options.Provisioned != nil {
		req.ProvisionPolicy = &containers.ProvisionPolicy{
			MinInstances: *options.Provisioned,
		}
	}

	// Set secrets if provided
	if len(options.Secrets) > 0 {
		req.Secrets = make([]*containers.Secret, 0, len(options.Secrets))
		for _, secret := range options.Secrets {
			req.Secrets = append(req.Secrets, &containers.Secret{
				Id:        secret.SecretID,
				VersionId: secret.VersionID,
				Key:       secret.Key,
				Reference: &containers.Secret_EnvironmentVariable{
					EnvironmentVariable: secret.EnvironmentVariable,
				},
			})
		}
	}

	// Set log options if provided
	if options.LogOptions != nil {
		req.LogOptions = &containers.LogOptions{
			Disabled: options.LogOptions.Disabled,
			MinLevel: options.LogOptions.MinLevel,
		}

		if options.LogOptions.LogGroupID != "" {
			req.LogOptions.Destination = &containers.LogOptions_LogGroupId{
				LogGroupId: options.LogOptions.LogGroupID,
			}
		} else if options.LogOptions.FolderID != "" {
			req.LogOptions.Destination = &containers.LogOptions_FolderId{
				FolderId: options.LogOptions.FolderID,
			}
		}
	}

	// Set storage mounts if provided
	if len(options.StorageMounts) > 0 {
		req.StorageMounts = make([]*containers.StorageMount, 0, len(options.StorageMounts))
		for _, mount := range options.StorageMounts {
			req.StorageMounts = append(req.StorageMounts, &containers.StorageMount{
				BucketId:       mount.BucketID,
				Prefix:         mount.Prefix,
				MountPointPath: mount.MountPointPath,
				ReadOnly:       mount.ReadOnly,
			})
		}
	}

	// Create the revision and wrap the operation
	op, err := sdk.WrapOperation(
		sdk.Serverless().Containers().Container().DeployRevision(ctx, req),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create revision: %w", err)
	}

	// Wait for the operation to complete
	err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for operation: %w", err)
	}

	// Get the created revision
	result, err := op.Response()
	if err != nil {
		return "", fmt.Errorf("failed to get operation response: %w", err)
	}

	// Extract the revision ID
	revision, ok := result.(*containers.Revision)
	if !ok {
		return "", fmt.Errorf("unexpected response type: %T", result)
	}

	return revision.Id, nil
}

// createContainer creates a new container.
func createContainer(ctx context.Context, sdk *ycsdk.SDK, folderID, name string) (string, error) {
	// Get repository info for description
	repoOwner := sourcecraft.GetSourcecraftRepositoryOwner()
	repoName := sourcecraft.GetSourcecraftRepository()

	// Create container
	req := &containers.CreateContainerRequest{
		FolderId:    folderID,
		Name:        name,
		Description: fmt.Sprintf("Created from: %s/%s", repoOwner, repoName),
	}

	// Create the container and wrap the operation
	op, err := sdk.WrapOperation(
		sdk.Serverless().Containers().Container().Create(ctx, req),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Wait for the operation to complete
	err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for operation: %w", err)
	}

	// Get the created container
	result, err := op.Response()
	if err != nil {
		return "", fmt.Errorf("failed to get operation response: %w", err)
	}

	// Extract the container ID
	c, ok := result.(*containers.Container)
	if !ok {
		return "", fmt.Errorf("unexpected response type: %T", result)
	}

	return c.Id, nil
}

// findContainerByName finds a container by name.
func findContainerByName(
	ctx context.Context,
	sdk *ycsdk.SDK,
	folderID, name string,
) (string, error) {
	// Create a filter to find the container by name
	filter := fmt.Sprintf("name = \"%s\"", name)

	// List containers with the filter
	resp, err := sdk.Serverless().
		Containers().
		Container().
		List(ctx, &containers.ListContainersRequest{
			FolderId: folderID,
			Filter:   filter,
		})
	if err != nil {
		return "", fmt.Errorf("failed to list containers: %w", err)
	}

	// Check if any containers were found
	if len(resp.Containers) == 0 {
		return "", nil
	}

	// Return the ID of the first matching container
	return resp.Containers[0].Id, nil
}

// makeContainerPublic sets the access bindings for a container to make it publicly accessible.
func makeContainerPublic(ctx context.Context, sdk *ycsdk.SDK, containerID string) error {
	// Create the access binding for allUsers
	binding := &access.AccessBinding{
		RoleId: "serverless.containers.invoker",
		Subject: &access.Subject{
			Id:   "allUsers",
			Type: "system",
		},
	}

	// Create the request to set access bindings
	req := &access.SetAccessBindingsRequest{
		ResourceId: containerID,
		AccessBindings: []*access.AccessBinding{
			binding,
		},
	}

	// Call the API to set access bindings
	op, err := sdk.WrapOperation(
		sdk.Serverless().Containers().Container().SetAccessBindings(ctx, req),
	)
	if err != nil {
		return fmt.Errorf("failed to set access bindings: %w", err)
	}

	// Wait for the operation to complete
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for operation: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	sourcecraft.Info("Starting serverless container deployment")

	// Get required inputs
	folderID := sourcecraft.GetInput(inputFolderID)
	if folderID == "" {
		sourcecraft.SetFailed("folder-id is required")

		return
	}

	containerName := sourcecraft.GetInput(inputContainerName)
	if containerName == "" {
		sourcecraft.SetFailed("container-name is required")

		return
	}

	// Get authentication credentials
	ycSaJsonCredentials := sourcecraft.GetInput(inputYcSaJsonCredentials)
	ycIamToken := sourcecraft.GetInput(inputYcIamToken)

	// Check if we have valid credentials
	if ycSaJsonCredentials == "" && ycIamToken == "" {
		sourcecraft.SetFailed("No credentials provided")

		return
	}

	// Log which credentials we're using
	if ycSaJsonCredentials != "" {
		sourcecraft.Info("Using service account JSON credentials")
	} else {
		sourcecraft.Info("Using IAM token")
	}

	// Create SDK
	var sdk *ycsdk.SDK

	var err error

	if ycSaJsonCredentials != "" {
		// Create credentials from service account JSON
		key, err := iamkey.ReadFromJSONBytes([]byte(ycSaJsonCredentials))
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to read service account JSON: %v", err))

			return
		}

		credentials, err := ycsdk.ServiceAccountKey(key)
		if err != nil {
			sourcecraft.SetFailed(
				fmt.Sprintf("Failed to create credentials from service account JSON: %v", err),
			)

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
	} else {
		// Create SDK with IAM token
		sdk, err = ycsdk.Build(ctx, ycsdk.Config{
			Credentials: ycsdk.NewIAMTokenCredentials(ycIamToken),
		})
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create SDK: %v", err))

			return
		}
	}

	revOptions, err := container.ParseRevOptions()
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to parse revision options: %v", err))

		return
	}

	// Parse public flag
	isPublic := sourcecraft.GetBooleanInput(inputPublic)

	// Create a container object to store the results
	var containerID string

	var revisionID string

	// Find the container by name
	containerID, err = findContainerByName(ctx, sdk, folderID, containerName)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to find container: %v", err))

		return
	}

	revOptions.Log()

	if containerID == "" {
		// Container does not exist, create a new one
		sourcecraft.Info(
			fmt.Sprintf("There is no container with name: %s. Creating a new one.", containerName),
		)

		containerID, err = createContainer(ctx, sdk, folderID, containerName)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create container: %v", err))

			return
		}

		sourcecraft.Info(fmt.Sprintf("Container successfully created. Id: %s", containerID))
	} else {
		// Container exists, update it
		sourcecraft.Info(
			fmt.Sprintf(
				"Container with name: %s already exists and has id: %s",
				containerName,
				containerID,
			),
		)
	}

	// Create a new revision
	revisionID, err = createRevision(
		ctx,
		sdk,
		revOptions,
	)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to create revision: %v", err))

		return
	}

	sourcecraft.Info(fmt.Sprintf("Revision created successfully. Id: %s", revisionID))

	// Set outputs
	sourcecraft.SetOutput("CONTAINER_ID", containerID)
	sourcecraft.SetOutput("REVISION_ID", revisionID)

	// Make the container public if requested
	if isPublic {
		sourcecraft.Info(fmt.Sprintf("Making container %s public", containerID))

		// Call the API to make the container public
		err = makeContainerPublic(ctx, sdk, containerID)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to make container public: %v", err))

			return
		}

		sourcecraft.Info("Container is public now")
	}
}
