package objstore

import (
	"context"
	"fmt"
	"io/fs"
	"mime"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
	"github.com/yc-actions/sourcecraft-actions/pkg/storage"
)

// uploadFile uploads a file to object storage.
func uploadFile(
	ctx context.Context,
	f afero.Fs,
	storageService storage.StorageService,
	filePath, root, bucket, prefix string,
	cacheControl CacheControlConfig,
) error {
	// Check if file is a directory
	info, err := f.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil
	}

	// Get relative path
	relPath, err := filepath.Rel(root, filePath)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Build object key
	key := relPath
	if prefix != "" {
		key = filepath.Join(prefix, key)
	}

	// Create storage object
	sourcecraft.Info(fmt.Sprintf("Uploading %s to %s/%s", filePath, bucket, key))

	file, err := f.Open(filePath)
	if err != nil {
		return err
	}

	storageObject := storage.NewStorageObject(bucket, key, file)
	defer storageObject.Close()

	// Set cache control header if specified
	storageObject.CacheControl = GetCacheControlValue(cacheControl, key)
	storageObject.ContentType = mime.TypeByExtension(filepath.Ext(key))

	// Upload object
	err = storageService.PutObject(ctx, storageObject)
	if err != nil {
		return fmt.Errorf("failed to Upload object: %w", err)
	}

	return nil
}

// Upload uploads files to object storage.
func Upload(
	ctx context.Context,
	f afero.Fs,
	storageService storage.StorageService,
	inputs *ActionInputs,
) error {
	sourcecraft.StartGroup("Upload")
	defer sourcecraft.EndGroup()

	sourcecraft.Info("Upload start")

	// Get workspace directory
	workspace := sourcecraft.GetSourcecraftWorkspace()

	// Get source root
	root := filepath.Join(workspace, inputs.Root)

	// Parse ignore
	ignore := parseIgnoreGlobPatterns(inputs.Exclude)

	// Process include ignore
	for _, include := range inputs.Include {
		if include == "" {
			continue
		}

		pathFromSourceRoot := filepath.Join(root, include)

		matches, err := afero.Glob(f, pathFromSourceRoot)
		if err != nil {
			return fmt.Errorf("failed to glob pattern: %w", err)
		}

		for _, match := range matches {
			info, err := f.Stat(match)
			if err != nil {
				return fmt.Errorf("failed to stat file: %w", err)
			}

			if info.IsDir() {
				// Walk directory
				err = afero.Walk(f, match, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}

					// Skip directories
					if info.IsDir() {
						return nil
					}

					// Check if file matches any ignore pattern
					relPath, err := filepath.Rel(root, path)
					if err != nil {
						return fmt.Errorf("failed to get relative path: %w", err)
					}

					skip := false

					for _, pattern := range ignore {
						// Try to match the full path
						matched, err := filepath.Match(pattern, relPath)
						if err != nil {
							return fmt.Errorf("failed to match pattern: %w", err)
						}

						if matched {
							skip = true

							break
						}

						// Try to match just the base name
						baseName := filepath.Base(relPath)

						matched, err = filepath.Match(pattern, baseName)
						if err != nil {
							return fmt.Errorf("failed to match pattern: %w", err)
						}

						if matched {
							skip = true

							break
						}
					}

					if skip {
						return nil
					}

					// Upload file
					return uploadFile(
						ctx,
						f,
						storageService,
						path,
						root,
						inputs.Bucket,
						inputs.Prefix,
						inputs.CacheControl,
					)
				})
				if err != nil {
					return fmt.Errorf("failed to walk directory: %w", err)
				}
			} else {
				// Check if file matches any ignore pattern
				relPath, err := filepath.Rel(root, match)
				if err != nil {
					return fmt.Errorf("failed to get relative path: %w", err)
				}

				skip := false

				for _, pattern := range ignore {
					// Try to match the full path
					matched, err := filepath.Match(pattern, relPath)
					if err != nil {
						return fmt.Errorf("failed to match pattern: %w", err)
					}

					if matched {
						skip = true

						break
					}

					// Try to match just the base name
					baseName := filepath.Base(relPath)

					matched, err = filepath.Match(pattern, baseName)
					if err != nil {
						return fmt.Errorf("failed to match pattern: %w", err)
					}

					if matched {
						skip = true

						break
					}
				}

				if !skip {
					// Upload file
					err = uploadFile(ctx, f, storageService, match, root, inputs.Bucket, inputs.Prefix, inputs.CacheControl)
					if err != nil {
						return fmt.Errorf("failed to Upload file: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// Note: The getCacheControlValue function has been moved to cache-control.go

// parseIgnoreGlobPatterns parses ignore glob patterns from a string slice.
func parseIgnoreGlobPatterns(patterns []string) []string {
	var result []string

	for _, pattern := range patterns {
		if pattern != "" {
			result = append(result, pattern)
		}
	}

	sourcecraft.Info(fmt.Sprintf("Source ignore pattern: %q", result))

	return result
}
