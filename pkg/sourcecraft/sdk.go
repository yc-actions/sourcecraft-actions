package sourcecraft

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// Environment variables.
const (
	EnvSourcecraftWorkspace = "SOURCECRAFT_WORKSPACE"
	EnvSourcecraftSHA       = "SOURCECRAFT_COMMIT_SHA"
)

// GetInput gets an input value from environment variables.
func GetInput(name string) string {
	return os.Getenv(name)
}

// GetMultilineInput gets a multiline input value from environment variables.
func GetMultilineInput(name string) []string {
	value := GetInput(name)
	if value == "" {
		return nil
	}

	return strings.Split(value, "\n")
}

func GetMultilineInputDefault(name, defaultValue string) []string {
	value := GetInput(name)
	if value == "" {
		value = defaultValue
	}

	return strings.Split(value, "\n")
}

// GetBooleanInput gets a boolean input value from environment variables.
func GetBooleanInput(name string) bool {
	value := strings.ToLower(GetInput(name))

	return value == "true" || value == "yes" || value == "1"
}

// getIntInput gets an integer input value from environment variables.
// If the input is empty or not a valid integer, it returns the default value.
// If the input is not a valid integer, it sets a failure message.
func getIntInput(name string, defaultValue int64, base int) (int64, error) {
	value := GetInput(name)
	if value == "" {
		return defaultValue, nil
	}

	intValue, err := strconv.ParseInt(value, base, 64)
	if err != nil {
		return defaultValue, fmt.Errorf("failed to parse %s: %w", name, err)
	}

	return intValue, nil
}

// GetInt64Input gets an int64 input value from environment variables.
// If the input is empty or not a valid integer, it returns the default value.
// If the input is not a valid integer, it sets a failure message.
func GetInt64Input(name string, defaultValue int64) int64 {
	intValue, err := getIntInput(name, defaultValue, 10)
	if err != nil {
		SetFailed(err.Error())
		return defaultValue
	}
	return intValue
}

// GetIntInput gets an int input value from environment variables.
// If the input is empty or not a valid integer, it returns the default value.
// If the input is not a valid integer, it sets a failure message.
func GetIntInput(name string, defaultValue int) int {
	intValue, err := getIntInput(name, int64(defaultValue), 10)
	if err != nil {
		SetFailed(err.Error())
		return defaultValue
	}
	return int(intValue)
}

// GetInt64InputOpt gets an int64 input value from environment variables.
// It returns nil if the input is empty.
// If the input is not a valid integer, it sets a failure message.
func GetInt64InputOpt(name string) *int64 {
	value := GetInput(name)
	if value == "" {
		return nil
	}

	intValue, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		SetFailed(fmt.Sprintf("failed to parse %s: %v", name, err))
		return nil
	}
	return &intValue
}

// SetOutput sets an environment variable that will be available to subsequent cubes.
// It appends a KEY=VALUE pair to the file specified by the SOURCECRAFT_ENV environment variable.
func SetOutput(name, value string) {
	cubeName := os.Getenv("SOURCECRAFT_CUBE")

	// Get the file path from the SOURCECRAFT_ENV environment variable
	filePath := os.Getenv("SOURCECRAFT_ENV")
	if filePath == "" {
		// If SOURCECRAFT_ENV is not set, log an error and return
		ErrorLog("SOURCECRAFT_ENV environment variable is not set")
		return
	}

	// Open the file in append mode, create it if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		ErrorLog(fmt.Sprintf("Failed to open file %s: %v", filePath, err))
		return
	}
	defer file.Close() // Ensure the file is closed when the function returns

	// Create the KEY=VALUE pair
	data := fmt.Sprintf("%s_%s=%s\n", UpperSnakeCase(cubeName), UpperSnakeCase(name), value)

	// Write the data to the file
	if _, err := file.WriteString(data); err != nil {
		ErrorLog(fmt.Sprintf("Failed to write to file %s: %v", filePath, err))
	}
}

func UpperSnakeCase(s string) string {
	if s == "" {
		return ""
	}

	// Replace hyphens with underscores
	s = strings.ReplaceAll(s, "-", "_")

	var result strings.Builder

	for i, char := range s {
		// If it's an uppercase letter and not the first character
		// and the previous character is not an underscore or uppercase letter, add an underscore
		if i > 0 && char >= 'A' && char <= 'Z' && s[i-1] != '_' && (s[i-1] < 'A' || s[i-1] > 'Z') {
			result.WriteRune('_')
		}
		// Convert to uppercase and add to result
		result.WriteRune(unicode.ToUpper(char))
	}

	return result.String()
}

// SetFailed sets the action as failed.
func SetFailed(message string) {
	fmt.Printf("::error::%s\n", message)
	os.Exit(1)
}

// Info logs an Info message.
func Info(message string) {
	fmt.Println(message)
}

// Debug logs a Debug message.
func Debug(message string) {
	fmt.Println(message)
}

// error logs an error message.
func ErrorLog(message string) {
	fmt.Printf("::error::%s\n", message)
}

// StartGroup starts a log group.
func StartGroup(name string) {
	fmt.Printf("::group::%s\n", name)
}

// EndGroup ends a log group.
func EndGroup() {
	fmt.Println("::endgroup::")
}

// GetSourcecraftWorkspace gets the Sourcecraft workspace directory.
func GetSourcecraftWorkspace() string {
	workspace := os.Getenv(EnvSourcecraftWorkspace)
	if workspace == "" {
		workspace = "."
	}

	return workspace
}

// GetSourcecraftSHA gets the Sourcecraft commit SHA.
func GetSourcecraftSHA() string {
	return os.Getenv(EnvSourcecraftSHA)
}

// ParseRepoOwnerFromURL extracts the repository owner from a URL string.
func ParseRepoOwnerFromURL(repoURL string) string {
	if repoURL == "" {
		return ""
	}

	// Remove protocol part if exists
	if idx := strings.Index(repoURL, "://"); idx != -1 {
		repoURL = repoURL[idx+3:]
	}

	// Split by '/' and get the owner part (after hostname)
	parts := strings.Split(repoURL, "/")
	if len(parts) >= 2 {
		return parts[1]
	}

	return ""
}

// GetSourcecraftRepositoryOwner extracts the repository owner from SOURCECRAFT_REPO_URL.
func GetSourcecraftRepositoryOwner() string {
	repoURL := os.Getenv("SOURCECRAFT_REPO_URL")

	return ParseRepoOwnerFromURL(repoURL)
}

// GetSourcecraftRepository extracts the repository name from SOURCECRAFT_REPO_URL
// ParseRepoNameFromURL extracts the repository name from a URL string.
func ParseRepoNameFromURL(repoURL string) string {
	if repoURL == "" {
		return ""
	}

	// Remove protocol part if exists
	if idx := strings.Index(repoURL, "://"); idx != -1 {
		repoURL = repoURL[idx+3:]
	}

	// Split by '/' and get the last part
	parts := strings.Split(repoURL, "/")
	if len(parts) < 2 {
		return ""
	}

	// Remove .git suffix if present
	repoName := parts[len(parts)-1]
	if strings.HasSuffix(repoName, ".git") {
		repoName = repoName[:len(repoName)-4]
	}

	return repoName
}

// GetSourcecraftRepository extracts the repository name from SOURCECRAFT_REPO_URL.
func GetSourcecraftRepository() string {
	repoURL := os.Getenv("SOURCECRAFT_REPO_URL")

	return ParseRepoNameFromURL(repoURL)
}
