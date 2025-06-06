package storage

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
)

// StorageObject represents an object in Yandex Cloud Object Storage.
type StorageObject struct {
	BucketName   string
	ObjectName   string
	reader       io.ReadCloser
	CacheControl string
	ContentType  string
}

// NewStorageObject creates a new StorageObject
func NewStorageObject(bucketName, objectName string, reader io.ReadCloser) *StorageObject {
	return &StorageObject{
		BucketName: bucketName,
		ObjectName: objectName,
		reader:     reader,
	}
}

// NewStorageObjectFromString creates a new StorageObject from a string
func NewStorageObjectFromString(bucketName, objectName, content string) *StorageObject {
	return &StorageObject{
		BucketName: bucketName,
		ObjectName: objectName,
		reader:     io.NopCloser(strings.NewReader(content)),
	}
}

// GetReader returns a reader for the object's data.
func (o *StorageObject) GetReader() io.Reader {
	return o.reader
}

// File returns the reader as an io.ReadCloser.
func (o *StorageObject) File() io.ReadCloser {
	return o.reader
}

// GetData reads all data from the reader and returns it as a string.
func (o *StorageObject) GetData() string {
	if o.reader == nil {
		return ""
	}

	// Create a new reader from the original reader
	data, err := ioutil.ReadAll(o.reader)
	if err != nil {
		return ""
	}

	// Reset the reader for future use
	o.reader = io.NopCloser(bytes.NewReader(data))

	return string(data)
}

func (o *StorageObject) Close() error {
	if o.reader != nil {
		return o.reader.Close()
	}

	return nil
}
