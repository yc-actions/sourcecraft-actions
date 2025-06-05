package sourcecraft

import (
	"fmt"
	"os"
	"strings"
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
