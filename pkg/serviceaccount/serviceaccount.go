package serviceaccount

import (
	"context"
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

// ResolveID resolves a service account ID from either a direct ID or a service account name.
func ResolveID(
	ctx context.Context,
	sdk *ycsdk.SDK,
	folderID, serviceAccountID, serviceAccountName string,
) (string, error) {
	if serviceAccountID != "" {
		return serviceAccountID, nil
	}

	if serviceAccountName != "" {
		// Create a filter to find the service account by name
		filter := fmt.Sprintf("name = \"%s\"", serviceAccountName)

		// List service accounts with the filter
		req := &iam.ListServiceAccountsRequest{
			FolderId: folderID,
			Filter:   filter,
		}

		resp, err := sdk.IAM().ServiceAccount().List(ctx, req)
		if err != nil {
			return "", fmt.Errorf("failed to list service accounts: %w", err)
		}

		// Check if any service accounts were found
		if len(resp.ServiceAccounts) == 0 {
			return "", fmt.Errorf("service account with name %s not found", serviceAccountName)
		}

		// Return the ID of the first matching service account
		return resp.ServiceAccounts[0].Id, nil
	}

	return "", nil
}
