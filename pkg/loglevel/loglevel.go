package loglevel

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

var logLevelValues = []string{"TRACE", "DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// ParseLogLevel parses a log level string into the corresponding enum value.
func ParseLogLevel(levelKey string) (logging.LogLevel_Level, error) {
	if levelKey == "" {
		return logging.LogLevel_LEVEL_UNSPECIFIED, nil
	}

	upperLevelKey := strings.ToUpper(levelKey)

	// Check if the level is valid
	isValid := false

	for _, level := range logLevelValues {
		if level == upperLevelKey {
			isValid = true

			break
		}
	}

	if !isValid {
		return logging.LogLevel_LEVEL_UNSPECIFIED, fmt.Errorf("log level has unknown value")
	}

	// Map the string to the enum value
	switch upperLevelKey {
	case "TRACE":
		return logging.LogLevel_TRACE, nil
	case "DEBUG":
		return logging.LogLevel_DEBUG, nil
	case "INFO":
		return logging.LogLevel_INFO, nil
	case "WARN":
		return logging.LogLevel_WARN, nil
	case "ERROR":
		return logging.LogLevel_ERROR, nil
	case "FATAL":
		return logging.LogLevel_FATAL, nil
	default:
		return logging.LogLevel_LEVEL_UNSPECIFIED, nil
	}
}
