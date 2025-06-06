package objstore

// ActionInputs represents the inputs for the action.
type ActionInputs struct {
	Bucket       string
	Prefix       string
	Root         string
	Include      []string
	Exclude      []string
	Clear        bool
	CacheControl CacheControlConfig
}

// CacheControlConfig represents the cache control configuration.
type CacheControlConfig struct {
	Mapping map[string]string
	Default string
}
