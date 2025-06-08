package main

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/internal/function"
	"github.com/yc-actions/sourcecraft-actions/pkg/env"
	"github.com/yc-actions/sourcecraft-actions/pkg/loglevel"
	"github.com/yc-actions/sourcecraft-actions/pkg/memory"
	"github.com/yc-actions/sourcecraft-actions/pkg/serviceaccount"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage"
	"google.golang.org/protobuf/types/known/durationpb"
)

// Action inputs.
const (
	inputFolderID            = "FOLDER_ID"
	inputFunctionName        = "FUNCTION_NAME"
	inputRuntime             = "RUNTIME"
	inputEntrypoint          = "ENTRYPOINT"
	inputMemory              = "MEMORY"
	inputInclude             = "INCLUDE"
	inputExclude             = "EXCLUDE"
	inputSourceRoot          = "SOURCE_ROOT"
	inputExecutionTimeout    = "EXECUTION_TIMEOUT"
	inputEnvironment         = "ENVIRONMENT"
	inputServiceAccount      = "SERVICE_ACCOUNT"
	inputServiceAccountName  = "SERVICE_ACCOUNT_NAME"
	inputBucket              = "BUCKET"
	inputDescription         = "DESCRIPTION"
	inputSecrets             = "SECRETS"
	inputNetworkID           = "NETWORK_ID"
	inputTags                = "TAGS"
	inputLogsDisabled        = "LOGS_DISABLED"
	inputLogsGroupID         = "LOGS_GROUP_ID"
	inputLogLevel            = "LOG_LEVEL"
	inputAsync               = "ASYNC"
	inputAsyncSaID           = "ASYNC_SA_ID"
	inputAsyncSaName         = "ASYNC_SA_NAME"
	inputAsyncRetriesCount   = "ASYNC_RETRIES_COUNT"
	inputAsyncSuccessYmqArn  = "ASYNC_SUCCESS_YMQ_ARN"
	inputAsyncSuccessSaID    = "ASYNC_SUCCESS_SA_ID"
	inputAsyncFailureYmqArn  = "ASYNC_FAILURE_YMQ_ARN"
	inputAsyncFailureSaID    = "ASYNC_FAILURE_SA_ID"
	inputAsyncSuccessSaName  = "ASYNC_SUCCESS_SA_NAME"
	inputAsyncFailureSaName  = "ASYNC_FAILURE_SA_NAME"
	inputYcSaJsonCredentials = "YC_SA_JSON_CREDENTIALS"
	inputYcIamToken          = "YC_IAM_TOKEN"
)

// parseIgnoreGlobPatterns parses ignore glob patterns from a string slice.
func parseIgnoreGlobPatterns(patterns []string) []string {
	var result []string

	for _, pattern := range patterns {
		if pattern != "" {
			result = append(result, pattern)
		}
	}

	sourcecraft.Info(fmt.Sprintf("Source ignore pattern: %q", result))

	return result
}

// zipSources zips source files.
func zipSources(inputs *function.ActionInputs) ([]byte, error) {
	sourcecraft.StartGroup("ZipDirectory")
	defer sourcecraft.EndGroup()

	// Create a buffer to write the zip file to
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// Get workspace directory
	workspace := sourcecraft.GetSourcecraftWorkspace()

	// Get source root
	root := filepath.Join(workspace, inputs.SourceRoot)

	// Parse ignore patterns
	patterns := parseIgnoreGlobPatterns(inputs.ExcludePattern)

	// Add files to zip
	for _, include := range inputs.Include {
		if include == "" {
			continue
		}

		pathFromSourceRoot := filepath.Join(root, include)

		matches, err := filepath.Glob(pathFromSourceRoot)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern: %w", err)
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				return nil, fmt.Errorf("failed to stat file: %w", err)
			}

			if info.IsDir() {
				sourcecraft.Debug(fmt.Sprintf("match: dir %s", match))

				err = filepath.Walk(match, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// Skip directories
					if info.IsDir() {
						return nil
					}

					// Check if file matches any ignore pattern
					relPath, err := filepath.Rel(root, path)
					if err != nil {
						return fmt.Errorf("failed to get relative path: %w", err)
					}

					for _, pattern := range patterns {
						matched, err := filepath.Match(pattern, relPath)
						if err != nil {
							return fmt.Errorf("failed to match pattern: %w", err)
						}

						if matched {
							return nil
						}
					}

					// Add file to zip
					return addFileToZip(zipWriter, path, relPath)
				})
				if err != nil {
					return nil, fmt.Errorf("failed to walk directory: %w", err)
				}
			} else {
				sourcecraft.Debug(fmt.Sprintf("match: file %s", match))

				relPath, err := filepath.Rel(root, match)
				if err != nil {
					return nil, fmt.Errorf("failed to get relative path: %w", err)
				}

				err = addFileToZip(zipWriter, match, relPath)
				if err != nil {
					return nil, fmt.Errorf("failed to add file to zip: %w", err)
				}
			}
		}
	}

	// Close the zip writer
	err := zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close zip writer: %w", err)
	}

	sourcecraft.Info("Archive finalized")
	sourcecraft.Info(fmt.Sprintf("Buffer size: %d bytes", buf.Len()))

	return buf.Bytes(), nil
}

// addFileToZip adds a file to a zip writer.
func addFileToZip(zipWriter *zip.Writer, filePath, zipPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	sourcecraft.Info(fmt.Sprintf("add: %s", zipPath))

	writer, err := zipWriter.Create(zipPath)
	if err != nil {
		return fmt.Errorf("failed to create entry in zip: %w", err)
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("failed to write file to zip: %w", err)
	}

	return nil
}

// uploadToS3 uploads a file to S3.
func uploadToS3(
	ctx context.Context,
	bucket, functionID string,
	sdk *ycsdk.SDK,
	fileContents []byte,
) (string, error) {
	// Get Sourcecraft SHA
	sourcecraftSHA := sourcecraft.GetSourcecraftSHA()
	if sourcecraftSHA == "" {
		return "", fmt.Errorf("missing SOURCECRAFT_COMMIT_SHA")
	}

	// Set object name
	objectName := fmt.Sprintf("%s/%s.zip", functionID, sourcecraftSHA)
	sourcecraft.Info(fmt.Sprintf("Upload to bucket: %q", bucket+"/"+objectName))

	// Create storage service
	storageService := storage.NewStorageService(sdk)

	// Create storage object
	storageObject := storage.NewStorageObject(bucket, objectName, io.NopCloser(bytes.NewReader(fileContents)))

	// Upload object
	err := storageService.PutObject(ctx, storageObject)
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return objectName, nil
}

// getOrCreateFunctionID gets or creates a function ID.
func getOrCreateFunctionID(
	ctx context.Context,
	sdk *ycsdk.SDK,
	inputs *function.ActionInputs,
) (string, error) {
	sourcecraft.StartGroup("Find function id")
	defer sourcecraft.EndGroup()

	// List functions
	functionService := sdk.Serverless().Functions()

	resp, err := functionService.Function().List(ctx, &functions.ListFunctionsRequest{
		FolderId: inputs.FolderID,
		Filter:   fmt.Sprintf("name = '%s'", inputs.FunctionName),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list functions: %w", err)
	}

	// If function exists, return its ID
	if len(resp.Functions) > 0 {
		functionID := resp.Functions[0].Id
		sourcecraft.Info(
			fmt.Sprintf(
				"There is the function named '%s' in the folder already. Its id is '%s'",
				inputs.FunctionName,
				functionID,
			),
		)
		sourcecraft.SetOutput("FUNCTION_ID", functionID)

		return functionID, nil
	}

	// Otherwise create a new function
	op, err := sdk.WrapOperation(
		functionService.Function().Create(ctx, &functions.CreateFunctionRequest{
			FolderId:    inputs.FolderID,
			Name:        inputs.FunctionName,
			Description: inputs.Description,
		}),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create function: %w", err)
	}

	// Wait for operation to complete
	err = op.Wait(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to wait for operation: %w", err)
	}

	meta, err := op.Metadata()
	if err != nil {
		return "", fmt.Errorf("failed to get operation metadata: %w", err)
	}
	// Get function ID from metadata
	var createFunctionMetadata *functions.CreateFunctionMetadata

	if ok := meta.(*functions.CreateFunctionMetadata) != nil; !ok {
		return "", fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	createFunctionMetadata = meta.(*functions.CreateFunctionMetadata)

	functionID := createFunctionMetadata.FunctionId
	sourcecraft.Info(
		fmt.Sprintf(
			"There was no function named '%s' in the folder. So it was created. Id is '%s'",
			inputs.FunctionName,
			functionID,
		),
	)
	sourcecraft.SetOutput("FUNCTION_ID", functionID)

	return functionID, nil
}

// createFunctionVersion creates a function version.
func createFunctionVersion(
	ctx context.Context,
	sdk *ycsdk.SDK,
	functionID string,
	fileContents []byte,
	bucketObjectName string,
	inputs *function.ActionInputs,
) error {
	sourcecraft.StartGroup("Create function version")
	defer sourcecraft.EndGroup()

	sourcecraft.Info(fmt.Sprintf("Function '%s' %s", inputs.FunctionName, functionID))
	sourcecraft.Info(fmt.Sprintf("Parsed memory: %d", inputs.Memory))
	sourcecraft.Info(fmt.Sprintf("Parsed timeout: %d", inputs.ExecutionTimeout))

	// Resolve service account ID
	serviceAccountID, err := serviceaccount.ResolveID(
		ctx,
		sdk,
		inputs.FolderID,
		inputs.ServiceAccount,
		inputs.ServiceAccountName,
	)
	if err != nil {
		return fmt.Errorf("failed to resolve service account: %w", err)
	}

	// Create request
	request := &functions.CreateFunctionVersionRequest{
		FunctionId:       functionID,
		Runtime:          inputs.Runtime,
		Entrypoint:       inputs.Entrypoint,
		Resources:        &functions.Resources{Memory: inputs.Memory},
		ServiceAccountId: serviceAccountID,
		Description:      inputs.Description,
		Environment:      env.ParseEnvironmentVariables(inputs.Environment),
		ExecutionTimeout: &durationpb.Duration{Seconds: int64(inputs.ExecutionTimeout)},
		Tag:              inputs.Tags,
		Connectivity: &functions.Connectivity{
			NetworkId: inputs.NetworkID,
		},
		LogOptions: &functions.LogOptions{
			Disabled: inputs.LogsDisabled,
			Destination: &functions.LogOptions_LogGroupId{
				LogGroupId: inputs.LogsGroupID,
			},
			MinLevel: inputs.LogLevel,
		},
	}

	if function.IsAsync(inputs) {
		// Set up async invocation config
		asyncConfig, err := function.CreateAsyncInvocationConfig(ctx, sdk, inputs)
		if err != nil {
			return fmt.Errorf("failed to create async invocation config: %w", err)
		}

		request.AsyncInvocationConfig = asyncConfig
	}

	// Set up secrets
	secrets := env.ParseSecrets(inputs.Secrets)
	for _, secret := range secrets {
		request.Secrets = append(request.Secrets, &functions.Secret{
			Id:        secret.SecretID,
			VersionId: secret.VersionID,
			Key:       secret.Key,
			Reference: &functions.Secret_EnvironmentVariable{
				EnvironmentVariable: secret.EnvironmentVariable,
			},
		})
	}

	// Set package or content
	if inputs.Bucket != "" {
		sourcecraft.Info(fmt.Sprintf("From bucket: %q", inputs.Bucket))
		request.PackageSource = &functions.CreateFunctionVersionRequest_Package{
			Package: &functions.Package{
				BucketName: inputs.Bucket,
				ObjectName: bucketObjectName,
			},
		}
	} else {
		const limit = 3670016 // 3.5 MB

		if len(fileContents) > limit {
			return fmt.Errorf("zip file is too big: %d bytes. Provide bucket name", len(fileContents))
		}

		request.PackageSource = &functions.CreateFunctionVersionRequest_Content{
			Content: fileContents,
		}
	}

	// Create function version
	functionService := sdk.Serverless().Functions()

	op, err := sdk.WrapOperation(functionService.Function().CreateVersion(ctx, request))
	if err != nil {
		return fmt.Errorf("failed to create function version: %w", err)
	}

	// Wait for operation to complete
	err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for operation: %w", err)
	}

	sourcecraft.Info("Operation complete")

	meta, err := op.Metadata()
	if err != nil {
		return fmt.Errorf("failed to get operation metadata: %w", err)
	}

	// Get version ID from metadata
	var createFunctionVersionMetadata *functions.CreateFunctionVersionMetadata
	createFunctionVersionMetadata, ok := meta.(*functions.CreateFunctionVersionMetadata)

	if !ok {
		return fmt.Errorf("failed to unmarshal metadata")
	}

	sourcecraft.SetOutput("VERSION_ID", createFunctionVersionMetadata.FunctionVersionId)

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
		sourcecraft.SetFailed("No credentials")

		return
	}

	// Parse memory
	memoryStr := sourcecraft.GetInput(inputMemory)
	if memoryStr == "" {
		memoryStr = "128Mb"
	}

	memoryValue, err := memory.ParseMemory(memoryStr)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to parse memory: %v", err))

		return
	}

	// Parse execution timeout
	executionTimeout := sourcecraft.GetIntInput(inputExecutionTimeout, 5)

	// Parse log level
	logLevelStr := sourcecraft.GetInput(inputLogLevel)

	logLevel, err := loglevel.ParseLogLevel(logLevelStr)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to parse log level: %v", err))

		return
	}

	// Parse async retries count
	asyncRetriesCount := sourcecraft.GetIntInput(inputAsyncRetriesCount, 3)

	// Create inputs
	inputs := &function.ActionInputs{
		FolderID:           sourcecraft.GetInput(inputFolderID),
		FunctionName:       sourcecraft.GetInput(inputFunctionName),
		Runtime:            sourcecraft.GetInput(inputRuntime),
		Entrypoint:         sourcecraft.GetInput(inputEntrypoint),
		Memory:             memoryValue,
		Include:            sourcecraft.GetMultilineInputDefault(inputInclude, "."),
		ExcludePattern:     sourcecraft.GetMultilineInput(inputExclude),
		SourceRoot:         sourcecraft.GetInput(inputSourceRoot),
		ExecutionTimeout:   executionTimeout,
		Environment:        sourcecraft.GetMultilineInput(inputEnvironment),
		ServiceAccount:     sourcecraft.GetInput(inputServiceAccount),
		ServiceAccountName: sourcecraft.GetInput(inputServiceAccountName),
		Bucket:             sourcecraft.GetInput(inputBucket),
		Description:        sourcecraft.GetInput(inputDescription),
		Secrets:            sourcecraft.GetMultilineInput(inputSecrets),
		NetworkID:          sourcecraft.GetInput(inputNetworkID),
		Tags:               sourcecraft.GetMultilineInput(inputTags),
		LogsDisabled:       sourcecraft.GetBooleanInput(inputLogsDisabled),
		LogsGroupID:        sourcecraft.GetInput(inputLogsGroupID),
		LogLevel:           logLevel,
		Async:              sourcecraft.GetBooleanInput(inputAsync),
		AsyncSaID:          sourcecraft.GetInput(inputAsyncSaID),
		AsyncSaName:        sourcecraft.GetInput(inputAsyncSaName),
		AsyncRetriesCount:  asyncRetriesCount,
		AsyncSuccessYmqArn: sourcecraft.GetInput(inputAsyncSuccessYmqArn),
		AsyncSuccessSaID:   sourcecraft.GetInput(inputAsyncSuccessSaID),
		AsyncFailureYmqArn: sourcecraft.GetInput(inputAsyncFailureYmqArn),
		AsyncFailureSaID:   sourcecraft.GetInput(inputAsyncFailureSaID),
		AsyncSuccessSaName: sourcecraft.GetInput(inputAsyncSuccessSaName),
		AsyncFailureSaName: sourcecraft.GetInput(inputAsyncFailureSaName),
	}

	// Set default source root if not provided
	if inputs.SourceRoot == "" {
		inputs.SourceRoot = "."
	}

	sourcecraft.Info("Function inputs set")

	// Validate inputs
	if inputs.FolderID == "" {
		sourcecraft.SetFailed("folder-id is required")

		return
	}

	if inputs.FunctionName == "" {
		sourcecraft.SetFailed("function-name is required")

		return
	}

	if inputs.Runtime == "" {
		sourcecraft.SetFailed("runtime is required")

		return
	}

	if inputs.Entrypoint == "" {
		sourcecraft.SetFailed("entrypoint is required")

		return
	}

	// Validate async configuration
	err = function.ValidateAsync(inputs)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Invalid async configuration: %v", err))

		return
	}

	// Zip sources
	fileContents, err := zipSources(inputs)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to zip sources: %v", err))

		return
	}

	sourcecraft.Info(fmt.Sprintf("Buffer size: %d bytes", len(fileContents)))

	// Get or create function ID
	functionID, err := getOrCreateFunctionID(ctx, sdk, inputs)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to get or create function: %v", err))

		return
	}

	// Upload to S3 if bucket is provided
	var bucketObjectName string
	if inputs.Bucket != "" {
		bucketObjectName, err = uploadToS3(ctx, inputs.Bucket, functionID, sdk, fileContents)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to upload to S3: %v", err))

			return
		}
	}

	// Create function version
	err = createFunctionVersion(ctx, sdk, functionID, fileContents, bucketObjectName, inputs)
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to create function version: %v", err))

		return
	}
}
