package container

import (
	"fmt"
	"strings"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/logging/v1"
)

// Secret represents a secret for a container
type Secret struct {
	EnvironmentVariable string
	SecretID            string
	VersionID           string
	Key                 string
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

// ParseSecrets parses multiple secret strings
func ParseSecrets(inputs []string) ([]*Secret, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	secrets := make([]*Secret, 0, len(inputs))

	for _, input := range inputs {
		if input == "" {
			continue
		}

		secret, err := ParseSecret(input)
		if err != nil {
			return nil, err
		}

		secrets = append(secrets, secret)
	}

	return secrets, nil
}

// ParseEnvironmentVariables parses environment variables in the format "KEY=VALUE"
func ParseEnvironmentVariables(inputs []string) (map[string]string, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	env := make(map[string]string)

	for _, input := range inputs {
		if input == "" {
			continue
		}

		parts := strings.SplitN(input, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("environment variable has wrong format: %s", input)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if key == "" {
			return nil, fmt.Errorf("environment variable has empty key: %s", input)
		}

		env[key] = value
	}

	return env, nil
}

// Container represents a serverless container
type Container struct {
	ID     string
	RevID  string
	Domain string
}

// LogOptions represents log options for a container
type LogOptions struct {
	Disabled   bool
	LogGroupID string
	FolderID   string
	MinLevel   logging.LogLevel_Level
}
