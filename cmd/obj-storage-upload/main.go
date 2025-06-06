package main

import (
	"context"
	"fmt"

	"github.com/spf13/afero"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/internal/objstore"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage"
)

// Action inputs.
const (
	inputBucket              = "BUCKET"
	inputPrefix              = "PREFIX"
	inputRoot                = "ROOT"
	inputInclude             = "INCLUDE"
	inputExclude             = "EXCLUDE"
	inputClear               = "CLEAR"
	inputCacheControl        = "CACHE_CONTROL"
	inputYcSaJsonCredentials = "YC_SA_JSON_CREDENTIALS"
	inputYcIamToken          = "YC_IAM_TOKEN"
)

// clearBucket clears all objects from a bucket.
func clearBucket(ctx context.Context, storageService storage.StorageService, bucket string) error {
	sourcecraft.Info(fmt.Sprintf("Clearing bucket %s", bucket))

	// Clear bucket
	err := storageService.ClearBucket(ctx, bucket)
	if err != nil {
		return fmt.Errorf("failed to clear bucket: %w", err)
	}

	sourcecraft.Info(fmt.Sprintf("Bucket %s cleared successfully", bucket))

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
	} else {
		sourcecraft.SetFailed("No credentials provided")

		return
	}

	// Create inputs
	inputs := &objstore.ActionInputs{
		Bucket:  sourcecraft.GetInput(inputBucket),
		Prefix:  sourcecraft.GetInput(inputPrefix),
		Root:    sourcecraft.GetInput(inputRoot),
		Include: sourcecraft.GetMultilineInput(inputInclude),
		Exclude: sourcecraft.GetMultilineInput(inputExclude),
		Clear:   sourcecraft.GetBooleanInput(inputClear),
		CacheControl: objstore.ParseCacheControlFormats(
			sourcecraft.GetMultilineInput(inputCacheControl),
		),
	}

	// Validate inputs
	if inputs.Bucket == "" {
		sourcecraft.SetFailed("bucket is required")

		return
	}

	if inputs.Root == "" {
		sourcecraft.SetFailed("root is required")

		return
	}

	if len(inputs.Include) == 0 {
		// Default to include everything
		inputs.Include = []string{"."}
	}

	// Create storage service
	storageService := storage.NewStorageService(sdk)

	fs := afero.NewOsFs()

	// Clear bucket if requested
	if inputs.Clear {
		err = clearBucket(ctx, storageService, inputs.Bucket)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to clear bucket: %v", err))

			return
		}
	}

	// Upload files
	err = objstore.Upload(ctx, fs, storageService, inputs)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to Upload files: %v", err))

		return
	}

	sourcecraft.Info("Upload complete")
}
