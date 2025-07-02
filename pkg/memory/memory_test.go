package memory

import (
	"testing"
)

func TestParseMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
		wantErr  bool
	}{
		// Test MB values
		{"128mb", "128mb", 128 * MB, false},
		{"256MB", "256MB", 256 * MB, false},
		{"512 mb", "512 mb", 512 * MB, false},
		{"1024mb", "1024mb", 1024 * MB, false},
		{"2048MB", "2048MB", 2048 * MB, false},

		// Test GB values
		{"0gb", "0gb", 0 * GB, false},
		{"1GB", "1GB", 1 * GB, false},
		{"2Gb", "2Gb", 2 * GB, false},
		{"4 gb", "4 gb", 4 * GB, false},
		{"8GB", "8GB", 8 * GB, false},

		// Test error cases
		{"invalid", "invalid", 0, true},
		{"123", "123", 0, true},
		{"123kb", "123kb", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMemory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMemory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.expected {
				t.Errorf("ParseMemory() = %v, want %v", got, tt.expected)
			}
		})
	}
}
