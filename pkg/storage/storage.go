package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	"github.com/yandex-cloud/go-sdk"
)

// StorageService defines the interface for interacting with Yandex Cloud Object Storage
type StorageService interface {
	GetObject(ctx context.Context, bucketName, objectName string) (*StorageObject, error)
	PutObject(ctx context.Context, object *StorageObject) error
}

// StorageServiceImpl implements the StorageService interface
type StorageServiceImpl struct {
	sdk        *ycsdk.SDK
	endpoint   string
	httpClient *http.Client
}

// NewStorageService creates a new StorageService
func NewStorageService(sdk *ycsdk.SDK) *StorageServiceImpl {
	return &StorageServiceImpl{
		sdk:        sdk,
		endpoint:   "storage.yandexcloud.net:443",
		httpClient: &http.Client{},
	}
}

// GetObject retrieves an object from Yandex Cloud Object Storage
func (s *StorageServiceImpl) GetObject(ctx context.Context, bucketName, objectName string) (*StorageObject, error) {
	// Get IAM token
	iamToken, err := s.sdk.IAM().IamToken().Create(ctx, &iam.CreateIamTokenRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to get IAM token: %w", err)
	}

	// Create request
	u := s.buildURL(bucketName, objectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("X-YaCloud-SubjectToken", iamToken.IamToken)

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get object, status code: %d", resp.StatusCode)
	}

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return NewStorageObjectFromBytes(bucketName, objectName, data), nil
}

// PutObject uploads an object to Yandex Cloud Object Storage
func (s *StorageServiceImpl) PutObject(ctx context.Context, object *StorageObject) error {
	// Get IAM token
	iamToken, err := s.sdk.IAM().IamToken().Create(ctx, &iam.CreateIamTokenRequest{})
	if err != nil {
		return fmt.Errorf("failed to get IAM token: %w", err)
	}

	// Create request
	u := s.buildURL(object.BucketName, object.ObjectName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, object.GetReader())
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("X-YaCloud-SubjectToken", iamToken.IamToken)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(object.Data)))

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to put object, status code: %d", resp.StatusCode)
	}

	return nil
}

// buildURL builds the URL for a storage object
func (s *StorageServiceImpl) buildURL(bucketName, objectName string) string {
	return fmt.Sprintf("https://%s/%s/%s", s.endpoint, url.PathEscape(bucketName), path.Join(url.PathEscape(objectName)))
}
