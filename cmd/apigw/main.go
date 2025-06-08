package main

import (
	"context"
	"fmt"
	"os"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/serverless/apigateway/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
	"github.com/yandex-cloud/go-sdk/iamkey"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// Input constants.
const (
	inputFolderID            = "folder-id"
	inputGatewayName         = "gateway-name"
	inputSpecFile            = "spec-file"
	inputSpec                = "spec"
	inputYcSaJsonCredentials = "yc-sa-json-credentials"
	inputYcIamToken          = "yc-iam-token"
)

// Gateway represents an API Gateway.
type Gateway struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
}

func main() {
	ctx := context.Background()

	sourcecraft.Info("start")

	// Get required inputs
	folderID := sourcecraft.GetInput(inputFolderID)
	if folderID == "" {
		sourcecraft.SetFailed("folder-id is required")

		return
	}

	gatewayName := sourcecraft.GetInput(inputGatewayName)
	if gatewayName == "" {
		sourcecraft.SetFailed("gateway-name is required")

		return
	}

	gatewaySpecFile := sourcecraft.GetInput(inputSpecFile)
	gatewaySpec := sourcecraft.GetInput(inputSpec)

	if gatewaySpec == "" && gatewaySpecFile == "" {
		sourcecraft.SetFailed("Either spec or spec-file input must be provided")

		return
	}

	if gatewaySpec != "" && gatewaySpecFile != "" {
		sourcecraft.SetFailed("Only one of spec or spec-file input must be provided, not both")

		return
	}

	// Get credentials
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

	sourcecraft.Info(fmt.Sprintf("Folder ID: %s, gateway name: %s", folderID, gatewayName))

	// Get the spec content
	var specContent []byte

	if gatewaySpecFile != "" {
		var err error

		specContent, err = os.ReadFile(gatewaySpecFile)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to read spec file: %v", err))

			return
		}
	} else {
		specContent = []byte(gatewaySpec)
	}

	// Check if the gateway exists
	listResp, err := sdk.Serverless().
		APIGateway().
		ApiGateway().
		List(ctx, &apigateway.ListApiGatewayRequest{
			FolderId: folderID,
			Filter:   fmt.Sprintf("name=\"%s\"", gatewayName),
		})
	if err != nil {
		sourcecraft.SetFailed(fmt.Sprintf("Failed to list API gateways: %v", err))

		return
	}

	var gateway Gateway

	if len(listResp.ApiGateways) > 0 {
		// Gateway exists, update it
		existingGateway := listResp.ApiGateways[0]
		gateway.ID = existingGateway.Id
		gateway.Domain = existingGateway.Domain

		sourcecraft.Info(
			fmt.Sprintf(
				"Gateway with name: %s already exists and has id: %s",
				gatewayName,
				gateway.ID,
			),
		)

		if err = updateGateway(ctx, sdk, &gateway, specContent); err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to update API gateway: %v", err))

			return
		}

		sourcecraft.Info("Gateway updated successfully")
	} else {
		// Gateway does not exist, create a new one
		sourcecraft.Info(fmt.Sprintf("There is no gateway with name: %s. Creating a new one.", gatewayName))

		if err = createGateway(ctx, sdk, &gateway, folderID, gatewayName, specContent); err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Failed to create API gateway: %v", err))

			return
		}

		sourcecraft.Info(fmt.Sprintf("Gateway successfully created. Id: %s", gateway.ID))
	}
	// Set outputs
	sourcecraft.SetOutput("id", gateway.ID)
	sourcecraft.SetOutput("domain", gateway.Domain)
}

func createGateway(
	ctx context.Context,
	sdk *ycsdk.SDK,
	gateway *Gateway,
	folderID string,
	gatewayName string,
	specContent []byte,
) error {
	// Get repository info for description
	repoOwner := sourcecraft.GetSourcecraftRepositoryOwner()
	repoName := sourcecraft.GetSourcecraftRepository()

	// Create gateway
	createReq := &apigateway.CreateApiGatewayRequest{
		FolderId:    folderID,
		Name:        gatewayName,
		Description: fmt.Sprintf("Created from: %s/%s", repoOwner, repoName),
		Spec: &apigateway.CreateApiGatewayRequest_OpenapiSpec{
			OpenapiSpec: string(specContent),
		},
	}

	// Create the gateway and wrap the operation
	metaOp, err := sdk.WrapOperation(
		sdk.Serverless().APIGateway().ApiGateway().Create(ctx, createReq),
	)
	if err != nil {
		return err
	}

	err = metaOp.Wait(ctx)
	if err != nil {
		return err
	}

	// Get the created gateway
	result, err := metaOp.Response()
	if err != nil {
		return err
	}

	createdGateway, ok := result.(*apigateway.ApiGateway)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", result)
	}

	gateway.ID = createdGateway.Id
	gateway.Domain = createdGateway.Domain

	return nil
}

func updateGateway(
	ctx context.Context,
	sdk *ycsdk.SDK,
	gateway *Gateway,
	specContent []byte,
) error {
	// Update gateway
	updateReq := &apigateway.UpdateApiGatewayRequest{
		ApiGatewayId: gateway.ID,
		Spec: &apigateway.UpdateApiGatewayRequest_OpenapiSpec{
			OpenapiSpec: string(specContent),
		},
	}

	// Update the gateway and wrap the operation
	metaOp, err := sdk.WrapOperation(
		sdk.Serverless().APIGateway().ApiGateway().Update(ctx, updateReq),
	)
	if err != nil {
		return err
	}

	err = metaOp.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}
