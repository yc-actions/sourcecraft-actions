package objstore

import (
	"context"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage/mocks"
)

// setupTest creates a test file system with the given files.
func setupTest(t *testing.T) afero.Fs {
	// Create a memory file system
	appFS := afero.NewMemMapFs()

	// Create test files and directories
	err := appFS.MkdirAll("src/a", 0o755)
	assert.NoError(t, err)

	err = afero.WriteFile(appFS, "src/a/index.js", []byte("function foo(){}"), 0o644)
	assert.NoError(t, err)

	err = afero.WriteFile(appFS, "src/main.css", []byte("body"), 0o644)
	assert.NoError(t, err)
	err = afero.WriteFile(appFS, "src/index.html", []byte("<html></html>"), 0o644)
	assert.NoError(t, err)

	return appFS
}

// TestExist verifies that the test file system is set up correctly.
func TestExist(t *testing.T) {
	appFS := setupTest(t)
	name := "src/index.html"
	_, err := appFS.Stat(name)
	if err != nil {
		t.Errorf("file \"%s\" does not exist.\n", name)
	}
}

// TestUploadFile tests the Upload function with various configurations.
func TestUploadFile(t *testing.T) {
	// Define test cases
	testCases := []struct {
		name            string
		inputs          ActionInputs
		expectedFiles   []string
		unexpectedFiles []string
	}{
		{
			name: "Basic upload with exclude patterns",
			inputs: ActionInputs{
				CacheControl: CacheControlConfig{
					Mapping: map[string]string{
						"*.html": "public, max-age=3600",
						"*.css":  "public, max-age=7200",
					},
					Default: "public, max-age=86400",
				},
				Bucket:  "test-bucket",
				Root:    "src",
				Include: []string{"."},
				Exclude: []string{"*.js"},
				Clear:   false,
				Prefix:  "",
			},
			expectedFiles:   []string{"index.html", "main.css"},
			unexpectedFiles: []string{"a/index.js"},
		},
		{
			name: "Upload with no exclude patterns",
			inputs: ActionInputs{
				CacheControl: CacheControlConfig{
					Mapping: map[string]string{
						"*.html": "public, max-age=3600",
						"*.css":  "public, max-age=7200",
					},
					Default: "public, max-age=86400",
				},
				Bucket:  "test-bucket",
				Root:    "src",
				Include: []string{"."},
				Exclude: []string{},
				Clear:   false,
				Prefix:  "",
			},
			expectedFiles:   []string{"index.html", "main.css", "a/index.js"},
			unexpectedFiles: []string{},
		},
		{
			name: "Upload with specific include pattern",
			inputs: ActionInputs{
				CacheControl: CacheControlConfig{
					Mapping: map[string]string{
						"*.html": "public, max-age=3600",
						"*.css":  "public, max-age=7200",
					},
					Default: "public, max-age=86400",
				},
				Bucket:  "test-bucket",
				Root:    "src",
				Include: []string{"*.html"},
				Exclude: []string{},
				Clear:   false,
				Prefix:  "",
			},
			expectedFiles:   []string{"index.html"},
			unexpectedFiles: []string{"main.css", "a/index.js"},
		},
		{
			name: "Upload with prefix",
			inputs: ActionInputs{
				CacheControl: CacheControlConfig{
					Mapping: map[string]string{
						"*.html": "public, max-age=3600",
						"*.css":  "public, max-age=7200",
					},
					Default: "public, max-age=86400",
				},
				Bucket:  "test-bucket",
				Root:    "src",
				Include: []string{"."},
				Exclude: []string{"*.js"},
				Clear:   false,
				Prefix:  "assets",
			},
			expectedFiles:   []string{"assets/index.html", "assets/main.css"},
			unexpectedFiles: []string{"assets/a/index.js", "index.html", "main.css", "a/index.js"},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a context
			ctx := context.Background()

			// Create a mock storage service
			mockStorage := mocks.NewMockStorageService(t)

			// Set up the file system
			f := setupTest(t)

			// Set up expectations for the mock storage service
			mockStorage.EXPECT().
				PutObject(mock.Anything, mock.MatchedBy(func(obj *storage.StorageObject) bool {
					// Check if the object name is in the expected files list
					for _, expectedFile := range tc.expectedFiles {
						if obj.ObjectName == expectedFile && obj.BucketName == tc.inputs.Bucket {
							return true
						}
					}

					return false
				})).
				Return(nil).
				Times(len(tc.expectedFiles))

			// Call the Upload function
			err := Upload(ctx, f, mockStorage, &tc.inputs)
			assert.NoError(t, err, "Upload should not return an error")
		})
	}
}
