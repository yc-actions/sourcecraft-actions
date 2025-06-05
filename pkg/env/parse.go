package env

import (
	"fmt"
	"strings"

	"sourcecraft-actions/pkg/sourcecraft"
)

// ParseEnvironmentVariables parses environment variables from a string slice
func ParseEnvironmentVariables(env []string) map[string]string {
	sourcecraft.Info(fmt.Sprintf("Environment string: %q", env))

	environment := make(map[string]string)
	for _, line := range env {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			environment[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	sourcecraft.Info(fmt.Sprintf("EnvObject: %q", environment))
	return environment
}
