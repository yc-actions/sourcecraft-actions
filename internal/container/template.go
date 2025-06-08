package container

import (
	"bytes"
	"fmt"
	"text/template"
)

func ReplaceVariablesInSpec(
	specContent []byte,
	variables map[string]string,
) ([]byte, error) {
	if len(variables) == 0 {
		return specContent, nil
	}

	// Parse the template
	tmpl, err := template.New("spec").Parse(string(specContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute the template with the variables
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return nil, fmt.Errorf("failed to execute template: %v", err)
	}

	return buf.Bytes(), nil
}
