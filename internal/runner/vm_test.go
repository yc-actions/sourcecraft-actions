package runner

import (
	"strings"
	"testing"
)

func TestBuildUserDataScript(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "",
			User:                    "",
			SSHPublicKey:            "",
		}

		script := BuildUserDataScript(params)
		lines := strings.Split(script, "\n")

		// Check first line is shebang
		if !strings.HasPrefix(lines[0], "#!/bin/bash") {
			t.Errorf("First line is not shebang: %s", lines[0])
		}

		// Check script has expected number of lines
		if len(lines) < 8 {
			t.Errorf("Script has %d lines, want at least 8", len(lines))
		}

		// Check script doesn't contain --disableupdate
		if strings.Contains(script, "--disableupdate") {
			t.Errorf("Script contains --disableupdate when it shouldn't")
		}
	})

	t.Run("with home dir", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "foo",
			User:                    "",
			SSHPublicKey:            "",
		}

		script := BuildUserDataScript(params)
		lines := strings.Split(script, "\n")

		// Check first line is shebang
		if !strings.HasPrefix(lines[0], "#!/bin/bash") {
			t.Errorf("First line is not shebang: %s", lines[0])
		}

		// Check script has expected number of lines
		if len(lines) < 5 {
			t.Errorf("Script has %d lines, want at least 5", len(lines))
		}

		// Check script contains cd to home dir
		if !strings.Contains(script, `cd "foo"`) {
			t.Errorf("Script doesn't contain cd to home dir")
		}
	})

	t.Run("with user and ssh key", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "",
			User:                    "user",
			SSHPublicKey:            "key",
		}

		script := BuildUserDataScript(params)
		lines := strings.Split(script, "\n")

		// Check first line is cloud-config
		if !strings.HasPrefix(lines[0], "#cloud-config") {
			t.Errorf("First line is not cloud-config: %s", lines[0])
		}

		// Check script contains user and ssh key
		if !strings.Contains(script, "name: user") {
			t.Errorf("Script doesn't contain user name")
		}

		if !strings.Contains(script, `"key"`) {
			t.Errorf("Script doesn't contain ssh key")
		}
	})

	t.Run("with home dir and user and ssh key", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "foo",
			User:                    "user",
			SSHPublicKey:            "key",
		}

		script := BuildUserDataScript(params)
		lines := strings.Split(script, "\n")

		// Check first line is cloud-config
		if !strings.HasPrefix(lines[0], "#cloud-config") {
			t.Errorf("First line is not cloud-config: %s", lines[0])
		}

		// Check script contains user and ssh key
		if !strings.Contains(script, "name: user") {
			t.Errorf("Script doesn't contain user name")
		}

		if !strings.Contains(script, `"key"`) {
			t.Errorf("Script doesn't contain ssh key")
		}

		// Check script contains cd to home dir
		if !strings.Contains(script, `cd "foo"`) {
			t.Errorf("Script doesn't contain cd to home dir")
		}
	})

	t.Run("with disable update", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "",
			User:                    "",
			SSHPublicKey:            "",
		}

		script := BuildUserDataScript(params)

		// Check script contains --disableupdate
		if !strings.Contains(script, "--disableupdate") {
			t.Errorf("Script doesn't contain --disableupdate when it should")
		}
	})

	t.Run("without disable update", func(t *testing.T) {
		params := UserDataScriptParams{
			GithubRegistrationToken: "githubRegistrationToken",
			Label:                   "label",
			RunnerHomeDir:           "",
			User:                    "",
			SSHPublicKey:            "",
		}

		script := BuildUserDataScript(params)

		// Check script doesn't contain --disableupdate
		if strings.Contains(script, "--disableupdate") {
			t.Errorf("Script contains --disableupdate when it shouldn't")
		}
	})
}
