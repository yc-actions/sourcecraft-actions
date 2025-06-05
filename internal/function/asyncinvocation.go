package function

import (
	"context"
	"fmt"

	"sourcecraft-actions/pkg/serviceaccount"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/functions/v1"
	"github.com/yandex-cloud/go-sdk"
)

// IsAsync checks if async invocation is enabled
func IsAsync(inputs *ActionInputs) bool {
	return inputs.Async
}

// ValidateAsync validates the async invocation configuration
func ValidateAsync(inputs *ActionInputs) error {
	if !IsAsync(inputs) {
		return nil
	}

	// Either AsyncSuccessSaID or AsyncSuccessSaName must be set if AsyncSuccessYmqArn is set
	if inputs.AsyncSuccessYmqArn != "" {
		if inputs.AsyncSuccessSaID == "" && inputs.AsyncSuccessSaName == "" {
			return fmt.Errorf("either AsyncSuccessSaID or AsyncSuccessSaName must be set if AsyncSuccessYmqArn is set")
		}
		// But not both
		if inputs.AsyncSuccessSaID != "" && inputs.AsyncSuccessSaName != "" {
			return fmt.Errorf("either AsyncSuccessSaID or AsyncSuccessSaName must be set, but not both")
		}
	}

	// Either AsyncFailureSaID or AsyncFailureSaName must be set if AsyncFailureYmqArn is set
	if inputs.AsyncFailureYmqArn != "" {
		if inputs.AsyncFailureSaID == "" && inputs.AsyncFailureSaName == "" {
			return fmt.Errorf("either AsyncFailureSaID or AsyncFailureSaName must be set if AsyncFailureYmqArn is set")
		}
		// But not both
		if inputs.AsyncFailureSaID != "" && inputs.AsyncFailureSaName != "" {
			return fmt.Errorf("either AsyncFailureSaID or AsyncFailureSaName must be set, but not both")
		}
	}

	return nil
}

// CreateAsyncInvocationConfig creates an async invocation configuration
func CreateAsyncInvocationConfig(ctx context.Context, sdk *ycsdk.SDK, inputs *ActionInputs) (*functions.AsyncInvocationConfig, error) {
	if !IsAsync(inputs) {
		return nil, nil
	}

	var successTarget *functions.AsyncInvocationConfig_ResponseTarget
	var failureTarget *functions.AsyncInvocationConfig_ResponseTarget

	if inputs.AsyncSuccessYmqArn != "" {
		serviceAccountID, err := serviceaccount.ResolveServiceAccountID(
			ctx,
			sdk,
			inputs.FolderID,
			inputs.AsyncSuccessSaID,
			inputs.AsyncSuccessSaName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve success service account: %w", err)
		}

		successTarget = &functions.AsyncInvocationConfig_ResponseTarget{
			Target: &functions.AsyncInvocationConfig_ResponseTarget_YmqTarget{
				YmqTarget: &functions.YMQTarget{
					QueueArn:         inputs.AsyncSuccessYmqArn,
					ServiceAccountId: serviceAccountID,
				},
			},
		}
	} else {
		successTarget = &functions.AsyncInvocationConfig_ResponseTarget{
			Target: &functions.AsyncInvocationConfig_ResponseTarget_EmptyTarget{
				EmptyTarget: &functions.EmptyTarget{},
			},
		}
	}

	if inputs.AsyncFailureYmqArn != "" {
		serviceAccountID, err := serviceaccount.ResolveServiceAccountID(
			ctx,
			sdk,
			inputs.FolderID,
			inputs.AsyncFailureSaID,
			inputs.AsyncFailureSaName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve failure service account: %w", err)
		}

		failureTarget = &functions.AsyncInvocationConfig_ResponseTarget{
			Target: &functions.AsyncInvocationConfig_ResponseTarget_YmqTarget{
				YmqTarget: &functions.YMQTarget{
					QueueArn:         inputs.AsyncFailureYmqArn,
					ServiceAccountId: serviceAccountID,
				},
			},
		}
	} else {
		failureTarget = &functions.AsyncInvocationConfig_ResponseTarget{
			Target: &functions.AsyncInvocationConfig_ResponseTarget_EmptyTarget{
				EmptyTarget: &functions.EmptyTarget{},
			},
		}
	}

	// Resolve service account ID for async invocation
	serviceAccountID, err := serviceaccount.ResolveServiceAccountID(
		ctx,
		sdk,
		inputs.FolderID,
		inputs.AsyncSaID,
		inputs.AsyncSaName,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve async service account: %w", err)
	}

	// If no specific async service account is provided, use the function's service account
	if serviceAccountID == "" {
		serviceAccountID, err = serviceaccount.ResolveServiceAccountID(
			ctx,
			sdk,
			inputs.FolderID,
			inputs.ServiceAccount,
			inputs.ServiceAccountName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve function service account: %w", err)
		}
	}

	return &functions.AsyncInvocationConfig{
		RetriesCount:     int64(inputs.AsyncRetriesCount),
		SuccessTarget:    successTarget,
		FailureTarget:    failureTarget,
		ServiceAccountId: serviceAccountID,
	}, nil
}
