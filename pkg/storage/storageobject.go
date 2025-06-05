package storage

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// StorageObject represents an object in Yandex Cloud Object Storage
type StorageObject struct {
	BucketName string
	ObjectName string
	Data       []byte
}

// NewStorageObjectFromFile creates a new StorageObject from a file
func NewStorageObjectFromFile(bucketName, objectName, fileName string) (*StorageObject, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &StorageObject{
		BucketName: bucketName,
		ObjectName: objectName,
		Data:       data,
	}, nil
}

// NewStorageObjectFromString creates a new StorageObject from a string
func NewStorageObjectFromString(bucketName, objectName, content string) *StorageObject {
	return &StorageObject{
		BucketName: bucketName,
		ObjectName: objectName,
		Data:       []byte(content),
	}
}

// NewStorageObjectFromBytes creates a new StorageObject from a byte slice
func NewStorageObjectFromBytes(bucketName, objectName string, data []byte) *StorageObject {
	return &StorageObject{
		BucketName: bucketName,
		ObjectName: objectName,
		Data:       data,
	}
}

// GetData returns the data as a string with the specified encoding
func (o *StorageObject) GetData() string {
	return string(o.Data)
}

// GetReader returns a reader for the object's data
func (o *StorageObject) GetReader() io.Reader {
	return bytes.NewReader(o.Data)
}
