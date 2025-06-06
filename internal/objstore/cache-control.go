package objstore

import (
	"path/filepath"
	"strings"
)

// ParseCacheControlFormats parses cache control formats from a string slice.
func ParseCacheControlFormats(formats []string) CacheControlConfig {
	result := make(map[string]string)

	for _, format := range formats {
		parts := strings.Split(format, ":")
		if len(parts) != 2 {
			continue
		}

		keysPart := parts[0]
		valuePart := parts[1]

		// Special handling for default value to preserve spaces
		if strings.TrimSpace(keysPart) == "*" {
			// Don't trim spaces for default value
			result["*"] = valuePart
		} else {
			// For regular keys, trim spaces
			valuePart = strings.TrimSpace(valuePart)
			keys := strings.Split(keysPart, ",")

			for _, key := range keys {
				result[strings.TrimSpace(key)] = valuePart
			}
		}
	}

	// Handle default value
	defaultValue, hasDefault := result["*"]
	if hasDefault {
		delete(result, "*")
	}

	// In TypeScript, we only set default to undefined if it's an empty string
	// In Go, we'll keep the value as is, even if it's just a space

	return CacheControlConfig{
		Mapping: result,
		Default: defaultValue,
	}
}

// GetCacheControlValue returns the cache control value for the given key.
func GetCacheControlValue(config CacheControlConfig, key string) string {
	for pattern, value := range config.Mapping {
		matched, err := filepath.Match(pattern, key)
		if err != nil {
			continue
		}

		if matched {
			return value
		}

		// Try to match with base name for paths
		if strings.Contains(key, "/") {
			base := filepath.Base(key)

			matched, err := filepath.Match(pattern, base)
			if err != nil {
				continue
			}

			if matched {
				return value
			}
		}
	}

	return config.Default
}
