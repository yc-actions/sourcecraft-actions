package sourcecraft

import (
	"fmt"
	"os"
	"strings"
)

// Environment variables
const (
	EnvSourcecraftWorkspace       = "SOURCECRAFT_WORKSPACE"
	EnvSourcecraftSHA             = "SOURCECRAFT_COMMIT_SHA"
	EnvSourcecraftRepositoryOwner = "SOURCECRAFT_REPOSITORY_OWNER"
	EnvSourcecraftRepository      = "SOURCECRAFT_REPOSITORY"
)

// GetInput gets an input value from environment variables
func GetInput(name string) string {
	return os.Getenv(name)
}

// GetMultilineInput gets a multiline input value from environment variables
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

// GetBooleanInput gets a boolean input value from environment variables
func GetBooleanInput(name string) bool {
	value := strings.ToLower(GetInput(name))
	return value == "true" || value == "yes" || value == "1"
}

// SetOutput sets an output value
func SetOutput(name, value string) {
	fmt.Printf("::set-output name=%s::%s\n", name, value)
}

// SetFailed sets the action as failed
func SetFailed(message string) {
	fmt.Printf("::error::%s\n", message)
	os.Exit(1)
}

// Info logs an Info message
func Info(message string) {
	fmt.Println(message)
}

// Debug logs a Debug message
func Debug(message string) {
	fmt.Println(message)
}

// error logs an error message
func ErrorLog(message string) {
	fmt.Printf("::error::%s\n", message)
}

// StartGroup starts a log group
func StartGroup(name string) {
	fmt.Printf("::group::%s\n", name)
}

// EndGroup ends a log group
func EndGroup() {
	fmt.Println("::endgroup::")
}

// GetSourcecraftWorkspace gets the Sourcecraft workspace directory
func GetSourcecraftWorkspace() string {
	workspace := os.Getenv(EnvSourcecraftWorkspace)
	if workspace == "" {
		workspace = "."
	}
	return workspace
}

// GetSourcecraftSHA gets the Sourcecraft commit SHA
func GetSourcecraftSHA() string {
	return os.Getenv(EnvSourcecraftSHA)
}

// parseRepoOwnerFromURL extracts the repository owner from a URL string
func parseRepoOwnerFromURL(repoURL string) string {
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

// GetSourcecraftRepositoryOwner extracts the repository owner from SOURCECRAFT_REPO_URL
func GetSourcecraftRepositoryOwner() string {
	repoURL := os.Getenv("SOURCECRAFT_REPO_URL")
	return parseRepoOwnerFromURL(repoURL)
}

// GetSourcecraftRepository extracts the repository name from SOURCECRAFT_REPO_URL
// parseRepoNameFromURL extracts the repository name from a URL string
func parseRepoNameFromURL(repoURL string) string {
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

// GetSourcecraftRepository extracts the repository name from SOURCECRAFT_REPO_URL
func GetSourcecraftRepository() string {
	repoURL := os.Getenv("SOURCECRAFT_REPO_URL")
	return parseRepoNameFromURL(repoURL)
}
