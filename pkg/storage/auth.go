package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/middleware"
	"github.com/aws/smithy-go/tracing"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type iamRequestMiddleware struct {
	sdk         *ycsdk.SDK
	cachedToken string
	mutex       sync.Mutex
}

func (*iamRequestMiddleware) ID() string {
	return "IamToken"
}

func (m *iamRequestMiddleware) HandleFinalize(
	ctx context.Context,
	in middleware.FinalizeInput,
	next middleware.FinalizeHandler,
) (
	out middleware.FinalizeOutput, metadata middleware.Metadata, err error,
) {
	_, span := tracing.StartSpan(ctx, "IamToken")
	defer span.End()

	req, ok := in.Request.(*smithyhttp.Request)
	if !ok {
		return out, metadata, fmt.Errorf("unexpected transport type %T", in.Request)
	}

	token, err := m.getIAMToken(ctx)
	if err != nil {
		return middleware.FinalizeOutput{}, middleware.Metadata{}, fmt.Errorf(
			"failed to create IAM token: %w",
			err,
		)
	}

	req.Header.Set("X-YaCloud-SubjectToken", token)

	span.End()

	return next.HandleFinalize(ctx, in)
}

// getIAMToken gets the IAM token from cache or creates a new one.
func (s *iamRequestMiddleware) getIAMToken(ctx context.Context) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// If token is cached, return it
	if s.cachedToken != "" {
		return s.cachedToken, nil
	}

	iamToken, err := s.sdk.CreateIAMToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get IAM token: %w", err)
	}

	// Cache the token
	s.cachedToken = iamToken.IamToken

	return s.cachedToken, nil
}

func swapAuth(sdk *ycsdk.SDK) func(options *s3.Options) {
	return func(options *s3.Options) {
		options.APIOptions = append(options.APIOptions, func(stack *middleware.Stack) error {
			_, err := stack.Finalize.Swap("Signing", &iamRequestMiddleware{
				sdk: sdk,
			})
			if err != nil {
				return err
			}

			return nil
		})
	}
}
