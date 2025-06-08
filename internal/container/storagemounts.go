package container

import (
	"fmt"
	"strings"
)

const (
	// Delimiters
	storageDelimiter = ":"
	pathDelimiter    = "/"
)

// AccessMode represents the access mode for a storage mount
type AccessMode int

const (
	// AccessModeReadOnly represents read-only access
	AccessModeReadOnly AccessMode = iota
	// AccessModeReadWrite represents read-write access
	AccessModeReadWrite
)

// StorageMount represents a storage mount for a container
type StorageMount struct {
	BucketID       string
	Prefix         string
	MountPointPath string
	ReadOnly       bool
}

// ParseStorageMount parses a storage mount string in the format "S3_PATH:MOUNT_PATH:ACCESS_MODE"
// or "S3_PATH:MOUNT_PATH" (defaults to read-only)
func ParseStorageMount(input string) (*StorageMount, error) {
	parts := strings.Split(input, storageDelimiter)

	if len(parts) < 2 || len(parts) > 3 {
		return nil, fmt.Errorf("storage mount has wrong format: %s", input)
	}

	s3Path := strings.TrimSpace(parts[0])
	mountPointPath := strings.TrimSpace(parts[1])

	// Validate required fields
	if s3Path == "" {
		return nil, fmt.Errorf("storage mount has empty S3 path: %s", input)
	}

	if mountPointPath == "" {
		return nil, fmt.Errorf("storage mount has empty mount path: %s", input)
	}

	// Parse S3 path
	s3Parts := strings.Split(s3Path, pathDelimiter)
	bucketID := s3Parts[0]

	var prefix string
	if len(s3Parts) > 1 {
		prefix = strings.Join(s3Parts[1:], pathDelimiter)
	}

	// Default to read-only
	readOnly := true

	// Parse access mode if provided
	if len(parts) == 3 {
		accessMode := strings.TrimSpace(parts[2])
		if accessMode != "" {
			switch strings.ToLower(accessMode) {
			case "read-only", "ro", "readonly", "read_only":
				readOnly = true
			case "read-write", "rw", "readwrite", "read_write":
				readOnly = false
			default:
				return nil, fmt.Errorf("invalid access mode: %s", accessMode)
			}
		}
	}

	return &StorageMount{
		BucketID:       bucketID,
		Prefix:         prefix,
		MountPointPath: mountPointPath,
		ReadOnly:       readOnly,
	}, nil
}

// ParseStorageMounts parses multiple storage mount strings
func ParseStorageMounts(inputs []string) ([]*StorageMount, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	mounts := make([]*StorageMount, 0, len(inputs))

	for _, input := range inputs {
		if input == "" {
			continue
		}

		mount, err := ParseStorageMount(input)
		if err != nil {
			return nil, err
		}

		mounts = append(mounts, mount)
	}

	return mounts, nil
}
