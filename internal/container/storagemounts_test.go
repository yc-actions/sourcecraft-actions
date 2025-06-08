package container_test

import (
	"reflect"
	"testing"

	"github.com/yc-actions/sourcecraft-actions/internal/container"
)

func TestParseStorageMounts(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		var inputs []string
		mounts, err := container.ParseStorageMounts(inputs)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if mounts != nil {
			t.Errorf("Expected nil mounts, got %v", mounts)
		}
	})

	t.Run("valid inputs", func(t *testing.T) {
		inputs := []string{
			"bucket1/folder:mountpoint1",
			"bucket2:/mountpoint2:read-only",
			"bucket3:/mountpoint3:read-write",
		}

		expected := []*container.StorageMount{
			{
				BucketID:       "bucket1",
				Prefix:         "folder",
				MountPointPath: "mountpoint1",
				ReadOnly:       true,
			},
			{
				BucketID:       "bucket2",
				Prefix:         "",
				MountPointPath: "/mountpoint2",
				ReadOnly:       true,
			},
			{
				BucketID:       "bucket3",
				Prefix:         "",
				MountPointPath: "/mountpoint3",
				ReadOnly:       false,
			},
		}

		mounts, err := container.ParseStorageMounts(inputs)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !reflect.DeepEqual(mounts, expected) {
			t.Errorf("Expected %v, got %v", expected, mounts)
		}
	})

	t.Run("invalid inputs", func(t *testing.T) {
		testCases := []struct {
			name   string
			inputs []string
		}{
			{
				name:   "wrong format",
				inputs: []string{"bucket1:mountpoint1:read-only:extra"},
			},
			{
				name:   "empty S3 path",
				inputs: []string{":mountpoint1"},
			},
			{
				name:   "empty mount path",
				inputs: []string{"bucket1:"},
			},
			{
				name:   "invalid access mode",
				inputs: []string{"bucket1:mountpoint1:invalid-mode"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mounts, err := container.ParseStorageMounts(tc.inputs)

				if err == nil {
					t.Errorf("Expected error, got nil")
				}

				if mounts != nil {
					t.Errorf("Expected nil mounts, got %v", mounts)
				}
			})
		}
	})

	t.Run("mixed valid and empty inputs", func(t *testing.T) {
		inputs := []string{
			"",
			"bucket1/folder:mountpoint1",
			"",
			"bucket2:/mountpoint2:read-only",
		}

		expected := []*container.StorageMount{
			{
				BucketID:       "bucket1",
				Prefix:         "folder",
				MountPointPath: "mountpoint1",
				ReadOnly:       true,
			},
			{
				BucketID:       "bucket2",
				Prefix:         "",
				MountPointPath: "/mountpoint2",
				ReadOnly:       true,
			},
		}

		mounts, err := container.ParseStorageMounts(inputs)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !reflect.DeepEqual(mounts, expected) {
			t.Errorf("Expected %v, got %v", expected, mounts)
		}
	})

	t.Run("all access mode variations", func(t *testing.T) {
		readOnlyModes := []string{
			"read-only", "ro", "readonly", "read_only",
		}

		readWriteModes := []string{
			"read-write", "rw", "readwrite", "read_write",
		}

		// Test read-only modes
		for _, mode := range readOnlyModes {
			t.Run("read-only mode: "+mode, func(t *testing.T) {
				inputs := []string{"bucket1:mountpoint1:" + mode}

				mounts, err := container.ParseStorageMounts(inputs)

				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if len(mounts) != 1 {
					t.Errorf("Expected 1 mount, got %d", len(mounts))
				} else if !mounts[0].ReadOnly {
					t.Errorf("Expected ReadOnly to be true, got false")
				}
			})
		}

		// Test read-write modes
		for _, mode := range readWriteModes {
			t.Run("read-write mode: "+mode, func(t *testing.T) {
				inputs := []string{"bucket1:mountpoint1:" + mode}

				mounts, err := container.ParseStorageMounts(inputs)

				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if len(mounts) != 1 {
					t.Errorf("Expected 1 mount, got %d", len(mounts))
				} else if mounts[0].ReadOnly {
					t.Errorf("Expected ReadOnly to be false, got true")
				}
			})
		}
	})
}
