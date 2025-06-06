package storage

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

// StorageService defines the interface for interacting with Yandex Cloud Object Storage.
type StorageService interface {
	GetObject(ctx context.Context, bucketName, objectName string) (*StorageObject, error)
	PutObject(ctx context.Context, object *StorageObject) error
	ListObjects(
		ctx context.Context,
		bucketName string,
		maxKeys int32,
		continuationToken string,
	) ([]string, string, bool, error)
	DeleteObjects(ctx context.Context, bucketName string, objectKeys []string) (int, error)
	ClearBucket(ctx context.Context, bucketName string) error
}

// StorageServiceImpl implements the StorageService interface using direct HTTP requests.
type StorageServiceImpl struct {
	s3Client *s3.Client
}

// NewStorageService creates a new StorageService with default options.
func NewStorageService(sdk *ycsdk.SDK) *StorageServiceImpl {
	return NewStorageServiceWithOptions(sdk)
}

// NewStorageServiceWithOptions creates a new S3StorageService with the specified options.
func NewStorageServiceWithOptions(sdk *ycsdk.SDK) *StorageServiceImpl {
	// Create S3 client
	s3Client := s3.New(s3.Options{
		Region:             "ru-central1",
		EndpointResolverV2: &resolverV2{},
	},
		swapAuth(sdk),
	)

	service := &StorageServiceImpl{
		s3Client: s3Client,
	}

	return service
}

// GetObject retrieves an object from Yandex Cloud Object Storage.
func (s *StorageServiceImpl) GetObject(
	ctx context.Context,
	bucketName, objectName string,
) (*StorageObject, error) {
	object, err := s.s3Client.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectName),
		})
	if err != nil {
		return nil, err
	}

	return NewStorageObject(bucketName, objectName, object.Body), nil
}

// PutObject uploads an object to Yandex Cloud Object Storage.
func (s *StorageServiceImpl) PutObject(ctx context.Context, object *StorageObject) error {
	metadata := make(map[string]string)

	if object.ContentType != "" {
		metadata["Content-Type"] = object.ContentType
	}

	if object.CacheControl != "" {
		metadata["Cache-Control"] = object.CacheControl
	}

	// Get IAM token from cache or create a new one
	_, err := s.s3Client.PutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:   aws.String(object.BucketName),
			Key:      aws.String(object.ObjectName),
			Body:     object.GetReader(),
			Metadata: metadata,
		})

	return err
}

// ListObjects lists objects in a bucket.
func (s *StorageServiceImpl) ListObjects(
	ctx context.Context,
	bucketName string,
	maxKeys int32,
	continuationToken string,
) ([]string, string, bool, error) {
	// Create list objects input
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	// Set max keys if specified
	if maxKeys > 0 {
		input.MaxKeys = aws.Int32(maxKeys)
	}

	// Set continuation token if specified
	if continuationToken != "" {
		input.ContinuationToken = aws.String(continuationToken)
	}

	// List objects
	output, err := s.s3Client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to list objects: %w", err)
	}

	// Extract object keys
	var keys []string

	for _, obj := range output.Contents {
		if obj.Key != nil {
			keys = append(keys, *obj.Key)
		}
	}

	// Get continuation token and truncation status
	var nextContinuationToken string
	if output.NextContinuationToken != nil {
		nextContinuationToken = *output.NextContinuationToken
	}

	// Check if the result is truncated
	isTruncated := false
	if output.IsTruncated != nil {
		isTruncated = *output.IsTruncated
	}

	return keys, nextContinuationToken, isTruncated, nil
}

// DeleteObjects deletes objects from a bucket.
func (s *StorageServiceImpl) DeleteObjects(
	ctx context.Context,
	bucketName string,
	objectKeys []string,
) (int, error) {
	// Check if there are any objects to delete
	if len(objectKeys) == 0 {
		return 0, nil
	}

	// Create delete objects input
	var objects []types.ObjectIdentifier
	for _, key := range objectKeys {
		objects = append(objects, types.ObjectIdentifier{
			Key: aws.String(key),
		})
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(bucketName),
		Delete: &types.Delete{
			Objects: objects,
			Quiet:   aws.Bool(false),
		},
	}

	// Delete objects
	output, err := s.s3Client.DeleteObjects(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to delete objects: %w", err)
	}

	// Return the number of deleted objects
	if output.Deleted != nil {
		return len(output.Deleted), nil
	}

	return 0, nil
}

// ClearBucket clears all objects from a bucket.
func (s *StorageServiceImpl) ClearBucket(ctx context.Context, bucketName string) error {
	// Set max keys to 1000 (default and maximum)
	maxKeys := int32(1000)

	// Initialize variables
	var continuationToken string

	isTruncated := true
	totalDeleted := 0

	// Loop until all objects are deleted
	for isTruncated {
		// List objects
		keys, nextContinuationToken, truncated, err := s.ListObjects(
			ctx,
			bucketName,
			maxKeys,
			continuationToken,
		)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		// Check if there are any objects to delete
		if len(keys) == 0 {
			break
		}

		// Update variables for next iteration
		isTruncated = truncated
		continuationToken = nextContinuationToken

		// Delete objects
		deleted, err := s.DeleteObjects(ctx, bucketName, keys)
		if err != nil {
			return fmt.Errorf("failed to delete objects: %w", err)
		}

		// Update total deleted count
		totalDeleted += deleted
	}

	return nil
}

type resolverV2 struct {
	// you could inject additional application context here as well
}

func (*resolverV2) ResolveEndpoint(ctx context.Context, params s3.EndpointParameters) (
	smithyendpoints.Endpoint, error,
) {
	u, err := url.Parse("https://storage.yandexcloud.net")
	if err != nil {
		return smithyendpoints.Endpoint{}, err
	}

	u.Path += "/" + *params.Bucket

	return smithyendpoints.Endpoint{
		URI: *u,
	}, nil
}
