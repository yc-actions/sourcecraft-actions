package objstore_test

import (
	"testing"

	"github.com/yc-actions/sourcecraft-actions/internal/objstore"
)

func TestParseCacheControlFormats(t *testing.T) {
	t.Run("should parse cache control formats", func(t *testing.T) {
		formats := []string{
			"*.html:public, max-age=3600",
			"*.css:public, max-age=3600",
			"*:public, max-age=3600",
		}
		result := objstore.ParseCacheControlFormats(formats)

		// Check mapping
		if len(result.Mapping) != 2 {
			t.Errorf("Expected mapping to have 2 entries, got %d", len(result.Mapping))
		}

		if result.Mapping["*.html"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.html] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.html"],
			)
		}

		if result.Mapping["*.css"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.css] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.css"],
			)
		}

		// Check default
		if result.Default != "public, max-age=3600" {
			t.Errorf("Expected default to be 'public, max-age=3600', got '%s'", result.Default)
		}
	})

	t.Run("should handle empty value", func(t *testing.T) {
		formats := []string{"*.html:public, max-age=3600", "*.css:public, max-age=3600", "*:"}
		result := objstore.ParseCacheControlFormats(formats)

		// Check mapping
		if len(result.Mapping) != 2 {
			t.Errorf("Expected mapping to have 2 entries, got %d", len(result.Mapping))
		}

		if result.Mapping["*.html"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.html] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.html"],
			)
		}

		if result.Mapping["*.css"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.css] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.css"],
			)
		}

		// Check default
		if result.Default != "" {
			t.Errorf("Expected default to be empty, got '%s'", result.Default)
		}
	})

	t.Run("should handle empty string", func(t *testing.T) {
		formats := []string{"*.html:public, max-age=3600", "*.css:public, max-age=3600", "*: "}
		result := objstore.ParseCacheControlFormats(formats)

		// Check mapping
		if len(result.Mapping) != 2 {
			t.Errorf("Expected mapping to have 2 entries, got %d", len(result.Mapping))
		}

		if result.Mapping["*.html"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.html] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.html"],
			)
		}

		if result.Mapping["*.css"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.css] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.css"],
			)
		}

		// Check default
		if result.Default != " " {
			t.Errorf("Expected default to be ' ', got '%s'", result.Default)
		}
	})

	t.Run("should handle empty input", func(t *testing.T) {
		var formats []string
		result := objstore.ParseCacheControlFormats(formats)

		// Check mapping
		if len(result.Mapping) != 0 {
			t.Errorf("Expected mapping to be empty, got %d entries", len(result.Mapping))
		}

		// Check default
		if result.Default != "" {
			t.Errorf("Expected default to be empty, got '%s'", result.Default)
		}
	})

	t.Run("should handle multiple keys separated by comma", func(t *testing.T) {
		formats := []string{
			"*.html, *.htm:public, max-age=3600",
			"*.css:public, max-age=3600",
			"*:public, max-age=3600",
		}
		result := objstore.ParseCacheControlFormats(formats)

		// Check mapping
		if len(result.Mapping) != 3 {
			t.Errorf("Expected mapping to have 3 entries, got %d", len(result.Mapping))
		}

		if result.Mapping["*.html"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.html] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.html"],
			)
		}

		if result.Mapping["*.htm"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.htm] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.htm"],
			)
		}

		if result.Mapping["*.css"] != "public, max-age=3600" {
			t.Errorf(
				"Expected mapping[*.css] to be 'public, max-age=3600', got '%s'",
				result.Mapping["*.css"],
			)
		}

		// Check default
		if result.Default != "public, max-age=3600" {
			t.Errorf("Expected default to be 'public, max-age=3600', got '%s'", result.Default)
		}
	})
}

func TestGetCacheControlValue(t *testing.T) {
	t.Run("should return value for key", func(t *testing.T) {
		mapping := map[string]string{
			"*.html": "html-value",
			"*.css":  "css-value",
		}
		cacheControl := objstore.CacheControlConfig{
			Mapping: mapping,
			Default: "default-value",
		}
		key := "file.html"
		result := objstore.GetCacheControlValue(cacheControl, key)

		if result != "html-value" {
			t.Errorf("Expected result to be 'html-value', got '%s'", result)
		}
	})

	t.Run("should return default value", func(t *testing.T) {
		mapping := map[string]string{
			"*.html": "html-value",
			"*.css":  "css-value",
		}
		cacheControl := objstore.CacheControlConfig{
			Mapping: mapping,
			Default: "default-value",
		}
		key := "file.js"
		result := objstore.GetCacheControlValue(cacheControl, key)

		if result != "default-value" {
			t.Errorf("Expected result to be 'default-value', got '%s'", result)
		}
	})

	t.Run("should match long path", func(t *testing.T) {
		mapping := map[string]string{
			"*.html": "html-value",
			"*.css":  "css-value",
		}
		cacheControl := objstore.CacheControlConfig{
			Mapping: mapping,
			Default: "default-value",
		}
		key := "path/to/file.html"
		result := objstore.GetCacheControlValue(cacheControl, key)

		if result != "html-value" {
			t.Errorf("Expected result to be 'html-value', got '%s'", result)
		}
	})
}
