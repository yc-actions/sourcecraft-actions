package function

import "github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"

// ActionInputs represents the input parameters for the GitHub Action
type ActionInputs struct {
	FolderID           string
	FunctionName       string
	Runtime            string
	Entrypoint         string
	Memory             int64
	Include            []string
	ExcludePattern     []string
	SourceRoot         string
	ExecutionTimeout   int
	Environment        []string
	ServiceAccount     string
	ServiceAccountName string
	Bucket             string
	Description        string
	Secrets            []string
	NetworkID          string
	Tags               []string
	LogsDisabled       bool
	LogsGroupID        string
	LogLevel           logging.LogLevel_Level

	Async              bool
	AsyncSaID          string
	AsyncSaName        string
	AsyncRetriesCount  int
	AsyncSuccessYmqArn string
	AsyncSuccessSaID   string
	AsyncFailureYmqArn string
	AsyncFailureSaID   string
	AsyncSuccessSaName string
	AsyncFailureSaName string
}
