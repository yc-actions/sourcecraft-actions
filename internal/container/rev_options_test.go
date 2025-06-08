package container

import (
	"testing"
)

func TestParseRevOptions(t *testing.T) {
	// Set up environment variables for testing
	t.Setenv("revision-image-url", "test-image-url")
	t.Setenv("revision-cores", "2")
	t.Setenv("revision-core-fraction", "50")
	t.Setenv("revision-concurrency", "5")
	t.Setenv("revision-execution-timeout", "10")
	t.Setenv("revision-memory", "256Mb")
	t.Setenv("revision-working-dir", "/app")
	t.Setenv("revision-commands", "cmd1\ncmd2")
	t.Setenv("revision-args", "arg1\narg2")
	t.Setenv("revision-env", "KEY1=value1\nKEY2=value2")
	t.Setenv("revision-secrets", "ENV_VAR1=secret1/version1/key1\nENV_VAR2=secret2/version2/key2")
	t.Setenv("revision-network-id", "network-id")
	t.Setenv("revision-service-account-id", "sa-id")
	t.Setenv("revision-log-options-disabled", "false")
	t.Setenv("revision-log-options-log-group-id", "log-group-id")
	t.Setenv("revision-log-options-min-level", "INFO")
	t.Setenv("revision-storage-mounts", "bucket1/folder:mountpoint1\nbucket2:/mountpoint2:read-only")

	// Test the function
	got, err := ParseRevOptions()
	if err != nil {
		t.Errorf("ParseRevOptions() error = %v", err)
		return
	}

	// Verify the results
	if got.ImageURL != "test-image-url" {
		t.Errorf("ParseRevOptions() ImageURL = %v, want %v", got.ImageURL, "test-image-url")
	}
	if got.Cores != 2 {
		t.Errorf("ParseRevOptions() Cores = %v, want %v", got.Cores, 2)
	}
	if got.CoreFraction != 50 {
		t.Errorf("ParseRevOptions() CoreFraction = %v, want %v", got.CoreFraction, 50)
	}
	if got.Concurrency != 5 {
		t.Errorf("ParseRevOptions() Concurrency = %v, want %v", got.Concurrency, 5)
	}
	if got.ExecutionTimeout != 10 {
		t.Errorf("ParseRevOptions() ExecutionTimeout = %v, want %v", got.ExecutionTimeout, 10)
	}
	if got.MemoryValue != 256*1024*1024 {
		t.Errorf("ParseRevOptions() MemoryValue = %v, want %v", got.MemoryValue, 256*1024*1024)
	}
	if got.WorkingDir != "/app" {
		t.Errorf("ParseRevOptions() WorkingDir = %v, want %v", got.WorkingDir, "/app")
	}
	if len(got.Commands) != 2 || got.Commands[0] != "cmd1" || got.Commands[1] != "cmd2" {
		t.Errorf("ParseRevOptions() Commands = %v, want %v", got.Commands, []string{"cmd1", "cmd2"})
	}
	if len(got.Args) != 2 || got.Args[0] != "arg1" || got.Args[1] != "arg2" {
		t.Errorf("ParseRevOptions() Args = %v, want %v", got.Args, []string{"arg1", "arg2"})
	}
	if len(got.Env) != 2 || got.Env["KEY1"] != "value1" || got.Env["KEY2"] != "value2" {
		t.Errorf("ParseRevOptions() Env = %v, want %v", got.Env, map[string]string{"KEY1": "value1", "KEY2": "value2"})
	}
	if len(got.Secrets) != 2 {
		t.Errorf("ParseRevOptions() Secrets length = %v, want %v", len(got.Secrets), 2)
	} else {
		if got.Secrets[0].EnvironmentVariable != "ENV_VAR1" || got.Secrets[0].SecretID != "secret1" || got.Secrets[0].VersionID != "version1" || got.Secrets[0].Key != "key1" {
			t.Errorf("ParseRevOptions() Secrets[0] = %v, want EnvironmentVariable=ENV_VAR1, SecretID=secret1, VersionID=version1, Key=key1", got.Secrets[0])
		}
		if got.Secrets[1].EnvironmentVariable != "ENV_VAR2" || got.Secrets[1].SecretID != "secret2" || got.Secrets[1].VersionID != "version2" || got.Secrets[1].Key != "key2" {
			t.Errorf("ParseRevOptions() Secrets[1] = %v, want EnvironmentVariable=ENV_VAR2, SecretID=secret2, VersionID=version2, Key=key2", got.Secrets[1])
		}
	}
	if got.NetworkID != "network-id" {
		t.Errorf("ParseRevOptions() NetworkID = %v, want %v", got.NetworkID, "network-id")
	}
	if got.ServiceAccountID != "sa-id" {
		t.Errorf("ParseRevOptions() ServiceAccountID = %v, want %v", got.ServiceAccountID, "sa-id")
	}
	if got.LogOptions.Disabled {
		t.Errorf("ParseRevOptions() LogOptions.Disabled = %v, want %v", got.LogOptions.Disabled, false)
	}
	if got.LogOptions.LogGroupID != "log-group-id" {
		t.Errorf("ParseRevOptions() LogOptions.LogGroupID = %v, want %v", got.LogOptions.LogGroupID, "log-group-id")
	}
	if len(got.StorageMounts) != 2 {
		t.Errorf("ParseRevOptions() StorageMounts length = %v, want %v", len(got.StorageMounts), 2)
	} else {
		if got.StorageMounts[0].BucketID != "bucket1" || got.StorageMounts[0].Prefix != "folder" || got.StorageMounts[0].MountPointPath != "mountpoint1" || !got.StorageMounts[0].ReadOnly {
			t.Errorf("ParseRevOptions() StorageMounts[0] = %v, want BucketID=bucket1, Prefix=folder, MountPointPath=mountpoint1, ReadOnly=true", got.StorageMounts[0])
		}
		if got.StorageMounts[1].BucketID != "bucket2" || got.StorageMounts[1].Prefix != "" || got.StorageMounts[1].MountPointPath != "/mountpoint2" || !got.StorageMounts[1].ReadOnly {
			t.Errorf("ParseRevOptions() StorageMounts[1] = %v, want BucketID=bucket2, Prefix=, MountPointPath=/mountpoint2, ReadOnly=true", got.StorageMounts[1])
		}
	}
}
