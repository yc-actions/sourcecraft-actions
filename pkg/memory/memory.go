package memory

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	// KB represents 1 kilobyte in bytes
	KB = 1024
	// MB represents 1 megabyte in bytes
	MB = 1024 * KB
	// GB represents 1 gigabyte in bytes
	GB = 1024 * MB
)

// ParseMemory parses a memory string (e.g., "128Mb" or "1Gb") into bytes
func ParseMemory(input string) (int64, error) {
	re := regexp.MustCompile(`^(\d+)\s?(mb|gb)$`)
	match := re.FindStringSubmatch(strings.ToLower(input))
	if match == nil {
		return 0, fmt.Errorf("memory has unknown format")
	}

	digits, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse memory value: %w", err)
	}

	var multiplier int64
	if match[2] == "mb" {
		multiplier = MB
	} else {
		multiplier = GB
	}

	return digits * multiplier, nil
}
