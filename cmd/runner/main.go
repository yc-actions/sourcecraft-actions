package main

import (
	"context"
	"fmt"

	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/internal/runner"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// Action inputs.
const (
	inputYcSaJsonCredentials = "YC_SA_JSON_CREDENTIALS"
	inputYcIamToken          = "YC_IAM_TOKEN"
)

// start creates a VM with a Sourcecraft runner.
func start(ctx context.Context, sdk *ycsdk.SDK, config *runner.Config) error {
	// Get Sourcecraft registration token
	token := sourcecraft.GetInput("PAT_TOKEN")

	// Generate a unique label for the runner
	label := config.GenerateUniqueLabel()

	// Create VM
	instanceID, err := runner.CreateVM(ctx, sdk, config, token, label)
	if err != nil {
		return fmt.Errorf("failed to create VM: %w", err)
	}

	// Set outputs
	sourcecraft.SetOutput("LABEL", label)
	sourcecraft.SetOutput("INSTANCE_ID", instanceID)

	// Wait for runner to be registered
	// TODO: Wait until this is implemented in the Sourcecraft API
	//err = runner.WaitForRunnerRegistered(ctx, config, label)
	//if err != nil {
	//	return fmt.Errorf("failed to wait for runner registration: %w", err)
	//}

	return nil
}

// stop destroys a VM and removes the Sourcecraft runner.
func stop(ctx context.Context, sdk *ycsdk.SDK, config *runner.Config) error {
	// Destroy VM
	err := runner.DestroyVM(ctx, sdk, config.Input.InstanceID)
	if err != nil {
		return fmt.Errorf("failed to destroy VM: %w", err)
	}

	return nil
}

func main() {
	ctx := context.Background()

	// Parse inputs
	ycSaJsonCredentials := sourcecraft.GetInput(inputYcSaJsonCredentials)
	ycIamToken := sourcecraft.GetInput(inputYcIamToken)

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

		sourcecraft.Info("Using service account JSON credentials")
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
	} else {
		sourcecraft.SetFailed("No credentials provided")

		return
	}

	// Create config
	config, err := runner.NewConfig()
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to create config: %v", err))

		return
	}

	// Run the appropriate mode
	switch config.Input.Mode {
	case "start":
		err = start(ctx, sdk, config)
	case "stop":
		err = stop(ctx, sdk, config)
	default:
		err = fmt.Errorf("unknown mode: %s", config.Input.Mode)
	}

	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to run %s mode: %v", config.Input.Mode, err))

		return
	}

	sourcecraft.Info(fmt.Sprintf("Successfully ran %s mode", config.Input.Mode))
}
