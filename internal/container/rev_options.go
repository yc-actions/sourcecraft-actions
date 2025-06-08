package container

import (
	"fmt"

	"github.com/yc-actions/sourcecraft-actions/pkg/env"
	"github.com/yc-actions/sourcecraft-actions/pkg/loglevel"
	"github.com/yc-actions/sourcecraft-actions/pkg/memory"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// CreateRevisionOptions contains all the options for creating a revision.
type CreateRevisionOptions struct {
	ContainerID      string
	ImageURL         string
	MemoryValue      int64
	Cores            int64
	CoreFraction     int64
	Concurrency      int64
	ExecutionTimeout int64
	WorkingDir       string
	Commands         []string
	Args             []string
	Env              map[string]string
	Secrets          []*Secret
	Provisioned      *int64
	NetworkID        string
	ServiceAccountID string
	LogOptions       *LogOptions
	StorageMounts    []*StorageMount
}

const (
	inputRevisionServiceAccountID     = "REVISION_SERVICE_ACCOUNT_ID"
	inputRevisionCores                = "REVISION_CORES"
	inputRevisionMemory               = "REVISION_MEMORY"
	inputRevisionCoreFraction         = "REVISION_CORE_FRACTION"
	inputRevisionConcurrency          = "REVISION_CONCURRENCY"
	inputRevisionImageURL             = "REVISION_IMAGE_URL"
	inputRevisionExecutionTimeout     = "REVISION_EXECUTION_TIMEOUT"
	inputRevisionWorkingDir           = "REVISION_WORKING_DIR"
	inputRevisionCommands             = "REVISION_COMMANDS"
	inputRevisionArgs                 = "REVISION_ARGS"
	inputRevisionEnv                  = "REVISION_ENV"
	inputRevisionSecrets              = "REVISION_SECRETS"
	inputRevisionProvisioned          = "REVISION_PROVISIONED"
	inputRevisionNetworkID            = "REVISION_NETWORK_ID"
	inputRevisionLogOptionsDisabled   = "REVISION_LOG_OPTIONS_DISABLED"
	inputRevisionLogOptionsLogGroupID = "REVISION_LOG_OPTIONS_LOG_GROUP_ID"
	inputRevisionLogOptionsFolderID   = "REVISION_LOG_OPTIONS_FOLDER_ID"
	inputRevisionLogOptionsMinLevel   = "REVISION_LOG_OPTIONS_MIN_LEVEL"
	inputRevisionStorageMounts        = "REVISION_STORAGE_MOUNTS"
)

func (r CreateRevisionOptions) Log() {
	// Log the inputs that will be used to create the revision
	sourcecraft.Info(fmt.Sprintf("Creating revision with image URL: %s", r.ImageURL))
	sourcecraft.Info(fmt.Sprintf("Memory: %d bytes", r.MemoryValue))
	sourcecraft.Info(fmt.Sprintf("Cores: %d", r.Cores))
	sourcecraft.Info(fmt.Sprintf("Core fraction: %d", r.CoreFraction))
	sourcecraft.Info(fmt.Sprintf("Concurrency: %d", r.Concurrency))
	sourcecraft.Info(fmt.Sprintf("Execution timeout: %d seconds", r.ExecutionTimeout))

	if r.WorkingDir != "" {
		sourcecraft.Info(fmt.Sprintf("Working directory: %s", r.WorkingDir))
	}

	if len(r.Commands) > 0 {
		sourcecraft.Info(fmt.Sprintf("Commands: %v", r.Commands))
	}

	if len(r.Args) > 0 {
		sourcecraft.Info(fmt.Sprintf("Args: %v", r.Args))
	}

	if r.Env != nil {
		sourcecraft.Info(fmt.Sprintf("Environment variables: %v", r.Env))
	}

	if r.Secrets != nil {
		sourcecraft.Info(fmt.Sprintf("Secrets: %d", len(r.Secrets)))
	}

	if r.Provisioned != nil {
		sourcecraft.Info(fmt.Sprintf("Provisioned instances: %d", *r.Provisioned))
	}

	if r.NetworkID != "" {
		sourcecraft.Info(fmt.Sprintf("Network ID: %s", r.NetworkID))
	}

	if r.ServiceAccountID != "" {
		sourcecraft.Info(fmt.Sprintf("Service account ID: %s", r.ServiceAccountID))
	}

	if r.LogOptions.Disabled {
		sourcecraft.Info("Logging is disabled")
	} else {
		if r.LogOptions.LogGroupID != "" {
			sourcecraft.Info(fmt.Sprintf("Log group ID: %s", r.LogOptions.LogGroupID))
		}

		if r.LogOptions.FolderID != "" {
			sourcecraft.Info(fmt.Sprintf("Log folder ID: %s", r.LogOptions.FolderID))
		}

		sourcecraft.Info(fmt.Sprintf("Log level: %v", r.LogOptions.MinLevel))
	}

	if r.StorageMounts != nil {
		sourcecraft.Info(fmt.Sprintf("Storage mounts: %d", len(r.StorageMounts)))

		for i, mount := range r.StorageMounts {
			sourcecraft.Info(
				fmt.Sprintf(
					"  Mount %d: %s -> %s (read-only: %v)",
					i+1,
					mount.BucketID,
					mount.MountPointPath,
					mount.ReadOnly,
				),
			)
		}
	}
}

func ParseRevOptions() (*CreateRevisionOptions, error) {
	imageURL := sourcecraft.GetInput(inputRevisionImageURL)
	if imageURL == "" {
		return nil, fmt.Errorf("revision-image-url is required")
	}

	res := &CreateRevisionOptions{
		ImageURL: imageURL,
		// Parse cores
		Cores: sourcecraft.GetInt64Input(inputRevisionCores, 1),

		// Parse core fraction
		CoreFraction: sourcecraft.GetInt64Input(inputRevisionCoreFraction, 100),

		// Parse concurrency
		Concurrency: sourcecraft.GetInt64Input(inputRevisionConcurrency, 1),

		// Parse execution timeout
		ExecutionTimeout: sourcecraft.GetInt64Input(inputRevisionExecutionTimeout, 3),

		// Parse provisioned instances
		Provisioned: sourcecraft.GetInt64InputOpt(inputRevisionProvisioned),

		// Parse environment variables
		Env: env.ParseEnvironmentVariables(sourcecraft.GetMultilineInput(inputRevisionEnv)),

		// Parse working directory
		WorkingDir: sourcecraft.GetInput(inputRevisionWorkingDir),

		// Parse commands
		Commands: sourcecraft.GetMultilineInput(inputRevisionCommands),

		// Parse args
		Args: sourcecraft.GetMultilineInput(inputRevisionArgs),

		// Parse network ID
		NetworkID: sourcecraft.GetInput(inputRevisionNetworkID),

		// Parse service account ID
		ServiceAccountID: sourcecraft.GetInput(inputRevisionServiceAccountID),
	}
	// Parse memory

	var err error
	res.MemoryValue, err = memory.ParseMemory(sourcecraft.GetInput(inputRevisionMemory))
	if err != nil {
		return nil, fmt.Errorf("failed to parse revision-memory: %w", err)
	}

	// Parse secrets
	envSecrets := env.ParseSecrets(sourcecraft.GetMultilineInput(inputRevisionSecrets))
	res.Secrets = make([]*Secret, 0, len(envSecrets))
	for _, envSecret := range envSecrets {
		res.Secrets = append(res.Secrets, &Secret{
			EnvironmentVariable: envSecret.EnvironmentVariable,
			SecretID:            envSecret.SecretID,
			VersionID:           envSecret.VersionID,
			Key:                 envSecret.Key,
		})
	}

	// Parse log options
	logOptionsDisabled := sourcecraft.GetBooleanInput(inputRevisionLogOptionsDisabled)
	logOptionsLogGroupID := sourcecraft.GetInput(inputRevisionLogOptionsLogGroupID)
	logOptionsFolderID := sourcecraft.GetInput(inputRevisionLogOptionsFolderID)

	// Check if both log group ID and folder ID are provided
	if logOptionsLogGroupID != "" && logOptionsFolderID != "" {
		return nil, fmt.Errorf("both log group ID and folder ID are provided, please set only one of them")
	}

	logLevel, err := loglevel.ParseLogLevel(sourcecraft.GetInput(inputRevisionLogOptionsMinLevel))
	if err != nil {
		return nil, fmt.Errorf("failed to parse revision-log-options-min-level: %w", err)
	}

	// Create log options
	res.LogOptions = &LogOptions{
		Disabled:   logOptionsDisabled,
		LogGroupID: logOptionsLogGroupID,
		FolderID:   logOptionsFolderID,
		MinLevel:   logLevel,
	}

	// Log the log options for debugging
	sourcecraft.Info(
		fmt.Sprintf("Log options: disabled=%v, logGroupID=%s, folderID=%s, minLevel=%v",
			res.LogOptions.Disabled, res.LogOptions.LogGroupID, res.LogOptions.FolderID, res.LogOptions.MinLevel),
	)

	// Parse storage mounts
	res.StorageMounts, err = ParseStorageMounts(
		sourcecraft.GetMultilineInput(inputRevisionStorageMounts),
	)
	if err != nil {

		return nil, fmt.Errorf("failed to parse revision-storage-mounts: %w", err)
	}

	return res, nil
}
