package mocks_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage/mocks"
)

func TestStorageServiceMock(t *testing.T) {
	// Create a new mock instance
	mockService := mocks.NewMockStorageService(t)

	// Set up expectations
	mockObject := storage.NewStorageObjectFromString("test-bucket", "test-object", "test-content")
	mockService.EXPECT().
		GetObject(mock.Anything, "test-bucket", "test-object").
		Return(mockObject, nil)
	mockService.EXPECT().PutObject(mock.Anything, mockObject).Return(nil)
	mockService.EXPECT().
		ListObjects(mock.Anything, "test-bucket", int32(10), "").
		Return([]string{"test-object"}, "", false, nil)
	mockService.EXPECT().
		DeleteObjects(mock.Anything, "test-bucket", []string{"test-object"}).
		Return(1, nil)
	mockService.EXPECT().ClearBucket(mock.Anything, "test-bucket").Return(nil)

	// Test the mock
	ctx := context.Background()

	// Test GetObject
	obj, err := mockService.GetObject(ctx, "test-bucket", "test-object")
	assert.NoError(t, err)
	assert.Equal(t, "test-bucket", obj.BucketName)
	assert.Equal(t, "test-object", obj.ObjectName)
	assert.Equal(t, "test-content", obj.GetData())

	// Test PutObject
	err = mockService.PutObject(ctx, mockObject)
	assert.NoError(t, err)

	// Test ListObjects
	keys, token, truncated, err := mockService.ListObjects(ctx, "test-bucket", 10, "")
	assert.NoError(t, err)
	assert.Equal(t, []string{"test-object"}, keys)
	assert.Equal(t, "", token)
	assert.False(t, truncated)

	// Test DeleteObjects
	deleted, err := mockService.DeleteObjects(ctx, "test-bucket", []string{"test-object"})
	assert.NoError(t, err)
	assert.Equal(t, 1, deleted)

	// Test ClearBucket
	err = mockService.ClearBucket(ctx, "test-bucket")
	assert.NoError(t, err)
}
