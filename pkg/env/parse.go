package env

import (
	"fmt"
	"strings"

	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// Secret represents a Lockbox secret.
type Secret struct {
	EnvironmentVariable string
	SecretID            string
	VersionID           string
	Key                 string
}

// ParseEnvironmentVariables parses environment variables from a string slice.
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

// ParseSecret parses a secret string in the format "ENV_VAR=secretID/versionID/key"
func ParseSecret(input string) (*Secret, error) {
	parts := strings.Split(input, "=")
	if len(parts) != 2 {
		return nil, fmt.Errorf("secret has wrong format: %s", input)
	}

	envVar := strings.TrimSpace(parts[0])
	secretPath := strings.TrimSpace(parts[1])

	if envVar == "" {
		return nil, fmt.Errorf("secret has empty environment variable: %s", input)
	}

	if secretPath == "" {
		return nil, fmt.Errorf("secret has empty path: %s", input)
	}

	pathParts := strings.Split(secretPath, "/")
	if len(pathParts) != 3 {
		return nil, fmt.Errorf("secret path has wrong format (should be secretID/versionID/key): %s", secretPath)
	}

	secretID := pathParts[0]
	versionID := pathParts[1]
	key := pathParts[2]

	if secretID == "" || versionID == "" || key == "" {
		return nil, fmt.Errorf("secret path has empty parts: %s", secretPath)
	}

	return &Secret{
		EnvironmentVariable: envVar,
		SecretID:            secretID,
		VersionID:           versionID,
		Key:                 key,
	}, nil
}

// ParseSecrets parses multiple secret strings and logs errors instead of returning them
func ParseSecrets(inputs []string) []*Secret {
	sourcecraft.Info(fmt.Sprintf("Secrets string: %q", inputs))

	var secrets []*Secret

	for _, input := range inputs {
		if input == "" {
			continue
		}

		secret, err := ParseSecret(input)
		if err != nil {
			sourcecraft.SetFailed(fmt.Sprintf("Broken reference to Lockbox Secret: %s", input))
			continue
		}

		secrets = append(secrets, secret)
	}

	sourcecraft.Info(fmt.Sprintf("SecretsObject: %q", secrets))

	return secrets
}
