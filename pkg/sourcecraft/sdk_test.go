package sourcecraft_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

func TestParseRepoNameFromURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "Default repository",
			url:  "https://git.o.cloud.yandex.net/test/sourcecraft-actions.git",
			want: "sourcecraft-actions",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.ParseRepoNameFromURL(tt.url); got != tt.want {
				t.Errorf("GetSourcecraftRepository() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRepoOwnerFromURL(t *testing.T) {
	type args struct {
		repoURL string
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Default repository",
			args: args{repoURL: "https://git.o.cloud.yandex.net/test/sourcecraft-actions.git"},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.ParseRepoOwnerFromURL(tt.args.repoURL); got != tt.want {
				t.Errorf("parseRepoOwnerFromURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetInt64InputOpt(t *testing.T) {
	type args struct {
		name string
	}

	// Set up test environment variables
	os.Setenv("EMPTY_VAR", "")
	os.Setenv("VALID_INT", "42")

	// Create expected values
	validInt := int64(42)

	tests := []struct {
		name string
		args args
		want *int64
	}{
		{
			name: "Empty input",
			args: args{name: "EMPTY_VAR"},
			want: nil,
		},
		{
			name: "Valid integer",
			args: args{name: "VALID_INT"},
			want: &validInt,
		},
		{
			name: "Non-existent variable",
			args: args{name: "NON_EXISTENT_VAR"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.GetInt64InputOpt(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInt64InputOpt() = %v, want %v", got, tt.want)
			}
		})
	}

	// Clean up environment variables
	os.Unsetenv("EMPTY_VAR")
	os.Unsetenv("VALID_INT")
}

func TestGetBooleanInput(t *testing.T) {
	type args struct {
		name string
	}

	// Set up test environment variables
	os.Setenv("TRUE_VAR", "true")
	os.Setenv("YES_VAR", "yes")
	os.Setenv("ONE_VAR", "1")
	os.Setenv("FALSE_VAR", "false")
	os.Setenv("EMPTY_VAR", "")

	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "True value",
			args: args{name: "TRUE_VAR"},
			want: true,
		},
		{
			name: "Yes value",
			args: args{name: "YES_VAR"},
			want: true,
		},
		{
			name: "One value",
			args: args{name: "ONE_VAR"},
			want: true,
		},
		{
			name: "False value",
			args: args{name: "FALSE_VAR"},
			want: false,
		},
		{
			name: "Empty value",
			args: args{name: "EMPTY_VAR"},
			want: false,
		},
		{
			name: "Non-existent variable",
			args: args{name: "NON_EXISTENT_VAR"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.GetBooleanInput(tt.args.name); got != tt.want {
				t.Errorf("GetBooleanInput() = %v, want %v", got, tt.want)
			}
		})
	}

	// Clean up environment variables
	os.Unsetenv("TRUE_VAR")
	os.Unsetenv("YES_VAR")
	os.Unsetenv("ONE_VAR")
	os.Unsetenv("FALSE_VAR")
	os.Unsetenv("EMPTY_VAR")
}

func TestGetMultilineInput(t *testing.T) {
	type args struct {
		name string
	}

	// Set up test environment variables
	os.Setenv("MULTILINE_VAR", "line1\nline2\nline3")
	os.Setenv("EMPTY_VAR", "")

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Multiline value",
			args: args{name: "MULTILINE_VAR"},
			want: []string{"line1", "line2", "line3"},
		},
		{
			name: "Empty value",
			args: args{name: "EMPTY_VAR"},
			want: nil,
		},
		{
			name: "Non-existent variable",
			args: args{name: "NON_EXISTENT_VAR"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.GetMultilineInput(tt.args.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMultilineInput() = %v, want %v", got, tt.want)
			}
		})
	}

	// Clean up environment variables
	os.Unsetenv("MULTILINE_VAR")
	os.Unsetenv("EMPTY_VAR")
}

func TestGetMultilineInputDefault(t *testing.T) {
	type args struct {
		name         string
		defaultValue string
	}

	// Set up test environment variables
	os.Setenv("MULTILINE_VAR", "line1\nline2\nline3")
	os.Setenv("EMPTY_VAR", "")

	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Multiline value",
			args: args{name: "MULTILINE_VAR", defaultValue: "default"},
			want: []string{"line1", "line2", "line3"},
		},
		{
			name: "Empty value with default",
			args: args{name: "EMPTY_VAR", defaultValue: "default"},
			want: []string{"default"},
		},
		{
			name: "Non-existent variable with default",
			args: args{name: "NON_EXISTENT_VAR", defaultValue: "default"},
			want: []string{"default"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sourcecraft.GetMultilineInputDefault(tt.args.name, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMultilineInputDefault() = %v, want %v", got, tt.want)
			}
		})
	}

	// Clean up environment variables
	os.Unsetenv("MULTILINE_VAR")
	os.Unsetenv("EMPTY_VAR")
}

func TestSetEnv(t *testing.T) {
	// Create a temporary file to use as the SOURCECRAFT_ENV file
	tmpFile, err := os.CreateTemp("", "sourcecraft_env_test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the file when the test is done

	// Set the SOURCECRAFT_ENV environment variable to point to the temporary file
	os.Setenv("SOURCECRAFT_ENV", tmpFile.Name())
	defer os.Unsetenv("SOURCECRAFT_ENV") // Clean up the environment variable

	// Test case 1: Setting a variable when the file is empty
	sourcecraft.SetOutput("TEST_KEY1", "TEST_VALUE1")

	// Read the file content
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected1 := "TEST_KEY1=TEST_VALUE1\n"
	if string(content) != expected1 {
		t.Errorf("SetOutput() with empty file = %v, want %v", string(content), expected1)
	}

	// Test case 2: Setting a variable when the file already has content
	sourcecraft.SetOutput("TEST_KEY2", "TEST_VALUE2")

	// Read the file content again
	content, err = os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected2 := "TEST_KEY1=TEST_VALUE1\nTEST_KEY2=TEST_VALUE2\n"
	if string(content) != expected2 {
		t.Errorf("SetOutput() with existing content = %v, want %v", string(content), expected2)
	}

	// Test case 3: Setting another variable
	sourcecraft.SetOutput("TEST_KEY3", "TEST_VALUE3")

	// Read the file content again
	content, err = os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected3 := "TEST_KEY1=TEST_VALUE1\nTEST_KEY2=TEST_VALUE2\nTEST_KEY3=TEST_VALUE3\n"
	if string(content) != expected3 {
		t.Errorf("SetOutput() with multiple values = %v, want %v", string(content), expected3)
	}
}
