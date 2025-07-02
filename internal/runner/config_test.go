package runner

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Save original environment variables
	origEnv := make(map[string]string)

	for _, env := range os.Environ() {
		for i := 0; i < len(env); i++ {
			if env[i] == '=' {
				origEnv[env[:i]] = env[i+1:]
				break
			}
		}
	}

	// Restore environment variables after test
	defer func() {
		os.Clearenv()

		for k, v := range origEnv {
			_ = os.Setenv(k, v)
		}
	}()

	t.Run("basic config", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "start")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")
		_ = os.Setenv("IMAGE_ID", "imageId")
		_ = os.Setenv("SUBNET_ID", "subnetId")
		_ = os.Setenv("FOLDER_ID", "folderId")

		config, err := NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}

		// Verify config values
		if config.Input.Mode != "start" {
			t.Errorf("config.Input.Mode = %v, want %v", config.Input.Mode, "start")
		}

		if config.Input.GithubToken != "githubToken" {
			t.Errorf("config.Input.GithubToken = %v, want %v", config.Input.GithubToken, "githubToken")
		}

		if config.Input.ImageID != "imageId" {
			t.Errorf("config.Input.ImageID = %v, want %v", config.Input.ImageID, "imageId")
		}

		if config.Input.SubnetID != "subnetId" {
			t.Errorf("config.Input.SubnetID = %v, want %v", config.Input.SubnetID, "subnetId")
		}

		if config.Input.FolderID != "folderId" {
			t.Errorf("config.Input.FolderID = %v, want %v", config.Input.FolderID, "folderId")
		}
	})

	t.Run("with secondary disk", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "start")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")
		_ = os.Setenv("IMAGE_ID", "imageId")
		_ = os.Setenv("SUBNET_ID", "subnetId")
		_ = os.Setenv("FOLDER_ID", "folderId")
		_ = os.Setenv("IMAGE2_ID", "secondDiskImageId")
		_ = os.Setenv("DISK2_SIZE", "30Gb")

		config, err := NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}

		// Verify secondary disk config
		if config.Input.SecondDiskImageID != "secondDiskImageId" {
			t.Errorf("config.Input.SecondDiskImageID = %v, want %v", config.Input.SecondDiskImageID, "secondDiskImageId")
		}

		if config.Input.SecondDiskSize != 30*1024*1024*1024 {
			t.Errorf("config.Input.SecondDiskSize = %v, want %v", config.Input.SecondDiskSize, 30*1024*1024*1024)
		}
	})

	t.Run("secondary disk without image ID", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "start")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")
		_ = os.Setenv("IMAGE_ID", "imageId")
		_ = os.Setenv("SUBNET_ID", "subnetId")
		_ = os.Setenv("FOLDER_ID", "folderId")
		_ = os.Setenv("DISK2_SIZE", "30Gb")

		_, err := NewConfig()
		if err == nil {
			t.Errorf("NewConfig() error = nil, want error about missing secondary disk image ID")
		}
	})

	t.Run("with TTL", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "start")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")
		_ = os.Setenv("IMAGE_ID", "imageId")
		_ = os.Setenv("SUBNET_ID", "subnetId")
		_ = os.Setenv("FOLDER_ID", "folderId")
		_ = os.Setenv("TTL", "1h")

		config, err := NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}

		// Verify TTL
		if config.Input.TTL == nil {
			t.Errorf("config.Input.TTL is nil, want 1h")
		} else if *config.Input.TTL != time.Hour {
			t.Errorf("config.Input.TTL = %v, want %v", *config.Input.TTL, time.Hour)
		}
	})

	t.Run("stop mode", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "stop")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")
		_ = os.Setenv("LABEL", "label")
		_ = os.Setenv("INSTANCE_ID", "instanceId")

		config, err := NewConfig()
		if err != nil {
			t.Fatalf("NewConfig() error = %v", err)
		}

		// Verify config values
		if config.Input.Mode != "stop" {
			t.Errorf("config.Input.Mode = %v, want %v", config.Input.Mode, "stop")
		}
		if config.Input.Label != "label" {
			t.Errorf("config.Input.Label = %v, want %v", config.Input.Label, "label")
		}
		if config.Input.InstanceID != "instanceId" {
			t.Errorf("config.Input.InstanceID = %v, want %v", config.Input.InstanceID, "instanceId")
		}
	})

	t.Run("invalid mode", func(t *testing.T) {
		// Set up environment variables for test
		os.Clearenv()
		_ = os.Setenv("MODE", "invalid")
		_ = os.Setenv("GITHUB_TOKEN", "githubToken")

		_, err := NewConfig()
		if err == nil {
			t.Errorf("NewConfig() error = nil, want error about invalid mode")
		}
	})
}

func TestGenerateUniqueLabel(t *testing.T) {
	config := &Config{}

	label1 := config.GenerateUniqueLabel()
	// Add a small sleep to ensure we get a different timestamp
	time.Sleep(1 * time.Millisecond)

	label2 := config.GenerateUniqueLabel()

	if label1 == label2 {
		t.Errorf("GenerateUniqueLabel() generated the same label twice: %v", label1)
	}

	if len(label1) != 5 {
		t.Errorf("GenerateUniqueLabel() label length = %v, want 5", len(label1))
	}
}
