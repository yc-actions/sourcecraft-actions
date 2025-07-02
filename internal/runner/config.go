package runner

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/yc-actions/sourcecraft-actions/pkg/memory"
	"github.com/yc-actions/sourcecraft-actions/pkg/sourcecraft"
)

// ResourcesSpec defines the resources for the VM.
type ResourcesSpec struct {
	Memory       int64
	Cores        int
	CoreFraction int
}

// ActionConfig defines the configuration for the runner action.
type ActionConfig struct {
	ImageID          string
	Mode             string
	GithubToken      string
	RunnerHomeDir    string
	Label            string
	SubnetID         string
	PublicIP         bool
	ServiceAccountID string
	DiskType         string
	DiskSize         int64
	FolderID         string
	ZoneID           string
	PlatformID       string
	ResourcesSpec    ResourcesSpec

	SecondDiskImageID string
	SecondDiskType    string
	SecondDiskSize    int64

	User         string
	SSHPublicKey string

	InstanceID string

	RunnerVersion string
	TTL           *time.Duration
	DisableUpdate bool
}

// GithubRepo defines a GitHub repository.
type GithubRepo struct {
	Owner string
	Repo  string
}

// Config holds the configuration for the runner.
type Config struct {
	Input         ActionConfig
	GithubContext GithubRepo
}

// NewConfig creates a new Config instance.
func NewConfig() (*Config, error) {
	input, err := parseVMInputs()
	if err != nil {
		return nil, err
	}

	config := &Config{
		Input: input,
		GithubContext: GithubRepo{
			Owner: sourcecraft.GetSourcecraftRepositoryOwner(),
			Repo:  sourcecraft.GetSourcecraftRepository(),
		},
	}

	// Validate input
	if config.Input.Mode == "" {
		return nil, fmt.Errorf("the 'mode' input is not specified")
	}

	if config.Input.GithubToken == "" {
		return nil, fmt.Errorf("the 'github-token' input is not specified")
	}

	switch config.Input.Mode {
	case "start":
		if config.Input.ImageID == "" || config.Input.SubnetID == "" ||
			config.Input.FolderID == "" {
			return nil, fmt.Errorf("not all the required inputs are provided for the 'start' mode")
		}

		if config.Input.SecondDiskSize > 0 && config.Input.SecondDiskImageID == "" {
			return nil, fmt.Errorf("secondary disk image id is missing")
		}
	case "stop":
		if config.Input.Label == "" || config.Input.InstanceID == "" {
			return nil, fmt.Errorf("not all the required inputs are provided for the 'stop' mode")
		}
	default:
		return nil, fmt.Errorf("wrong mode. Allowed values: start, stop")
	}

	return config, nil
}

// GenerateUniqueLabel generates a unique label for the runner.
func (c *Config) GenerateUniqueLabel() string {
	// Initialize random number generator with current time
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Generate a random 5-character string
	return fmt.Sprintf("%x", r.Int31())[:5]
}

// parseVMInputs parses the VM inputs from environment variables.
func parseVMInputs() (ActionConfig, error) {
	sourcecraft.StartGroup("Parsing Action Inputs")
	defer sourcecraft.EndGroup()

	folderID := sourcecraft.GetInput("FOLDER_ID")

	mode := sourcecraft.GetInput("MODE")
	githubToken := sourcecraft.GetInput("GITHUB_TOKEN")
	runnerHomeDir := sourcecraft.GetInput("RUNNER_HOME_DIR")
	label := sourcecraft.GetInput("LABEL")

	serviceAccountID := sourcecraft.GetInput("SERVICE_ACCOUNT_ID")

	imageID := sourcecraft.GetInput("IMAGE_ID")

	zoneID := sourcecraft.GetInput("ZONE_ID")
	if zoneID == "" {
		zoneID = "ru-central1-a"
	}

	subnetID := sourcecraft.GetInput("SUBNET_ID")
	publicIP := sourcecraft.GetBooleanInput("PUBLIC_IP")

	platformID := sourcecraft.GetInput("PLATFORM_ID")
	if platformID == "" {
		platformID = "standard-v3"
	}

	const coresDefaultValue = 2
	cores := sourcecraft.GetIntInput("CORES", coresDefaultValue)

	memoryStr := sourcecraft.GetInput("MEMORY")
	if memoryStr == "" {
		memoryStr = "1Gb"
	}

	memoryValue, err := memory.ParseMemory(memoryStr)
	if err != nil {
		return ActionConfig{}, fmt.Errorf("failed to parse memory: %w", err)
	}

	diskType := sourcecraft.GetInput("DISK_TYPE")
	if diskType == "" {
		diskType = "network-ssd"
	}

	diskSizeStr := sourcecraft.GetInput("DISK_SIZE")
	if diskSizeStr == "" {
		diskSizeStr = "30Gb"
	}

	diskSize, err := memory.ParseMemory(diskSizeStr)
	if err != nil {
		return ActionConfig{}, fmt.Errorf("failed to parse disk size: %w", err)
	}

	const coreFracDefaultValue = 100
	coreFraction := sourcecraft.GetIntInput("CORE_FRACTION", coreFracDefaultValue)

	secondDiskImageID := sourcecraft.GetInput("IMAGE2_ID")

	secondDiskType := sourcecraft.GetInput("DISK2_TYPE")
	if secondDiskType == "" {
		secondDiskType = "network-ssd"
	}

	secondDiskSizeStr := sourcecraft.GetInput("DISK2_SIZE")
	if secondDiskSizeStr == "" {
		secondDiskSizeStr = "0Gb"
	}

	secondDiskSize, err := memory.ParseMemory(secondDiskSizeStr)
	if err != nil {
		return ActionConfig{}, fmt.Errorf("failed to parse second disk size: %w", err)
	}

	user := sourcecraft.GetInput("USER")
	sshPublicKey := sourcecraft.GetInput("SSH_PUBLIC_KEY")

	instanceID := sourcecraft.GetInput("INSTANCE_ID")

	runnerVersion := sourcecraft.GetInput("RUNNER_VERSION")
	disableUpdate := sourcecraft.GetBooleanInput("DISABLE_UPDATE")

	var ttl *time.Duration

	ttlInput := sourcecraft.GetInput("TTL")
	if ttlInput != "" {
		duration, err := time.ParseDuration(ttlInput)
		if err != nil {
			return ActionConfig{}, fmt.Errorf("failed to parse TTL: %w", err)
		}

		ttl = &duration
	}

	return ActionConfig{
		InstanceID:        instanceID,
		ImageID:           imageID,
		DiskType:          diskType,
		DiskSize:          diskSize,
		SubnetID:          subnetID,
		PublicIP:          publicIP,
		ZoneID:            zoneID,
		PlatformID:        platformID,
		FolderID:          folderID,
		Mode:              mode,
		GithubToken:       githubToken,
		RunnerHomeDir:     runnerHomeDir,
		Label:             label,
		ServiceAccountID:  serviceAccountID,
		SecondDiskImageID: secondDiskImageID,
		SecondDiskType:    secondDiskType,
		SecondDiskSize:    secondDiskSize,
		User:              user,
		SSHPublicKey:      sshPublicKey,
		RunnerVersion:     runnerVersion,
		TTL:               ttl,
		DisableUpdate:     disableUpdate,
		ResourcesSpec: ResourcesSpec{
			Cores:        cores,
			Memory:       memoryValue,
			CoreFraction: coreFraction,
		},
	}, nil
}
