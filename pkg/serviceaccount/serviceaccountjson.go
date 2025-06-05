package serviceaccount

import (
	"encoding/json"
	"fmt"
)

// ServiceAccountKey represents the structure of a service account key JSON file
type ServiceAccountKey struct {
	ID               string `json:"id"`
	CreatedAt        string `json:"created_at"`
	KeyAlgorithm     string `json:"key_algorithm"`
	ServiceAccountID string `json:"service_account_id"`
	PrivateKey       string `json:"private_key"`
	PublicKey        string `json:"public_key"`
}

// FromServiceAccountJSONFile parses a service account key JSON file and returns the credentials
func FromServiceAccountJSONFile(data []byte) (*ServiceAccountKey, error) {
	var key ServiceAccountKey
	if err := json.Unmarshal(data, &key); err != nil {
		return nil, fmt.Errorf("failed to parse service account key JSON: %w", err)
	}

	// Check required fields
	requiredFields := []struct {
		name  string
		value string
	}{
		{"id", key.ID},
		{"private_key", key.PrivateKey},
		{"service_account_id", key.ServiceAccountID},
	}

	var missingFields []string
	for _, field := range requiredFields {
		if field.value == "" {
			missingFields = append(missingFields, field.name)
		}
	}

	if len(missingFields) > 0 {
		return nil, fmt.Errorf("service account key is missing required fields: %v", missingFields)
	}

	return &key, nil
}
