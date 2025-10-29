package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Config represents the forge.hcl configuration file
type Config struct {
	Project  *ProjectBlock  `hcl:"project,block"`
	Defaults *DefaultsBlock `hcl:"defaults,block"`
}

// ProjectBlock contains project-wide configuration
type ProjectBlock struct {
	Name   string `hcl:"name"`
	Region string `hcl:"region"`
}

// DefaultsBlock contains default values for stacks
type DefaultsBlock struct {
	Runtime string `hcl:"runtime,optional"`
	Timeout int    `hcl:"timeout,optional"`
	Memory  int    `hcl:"memory,optional"`
}

// Load loads configuration from forge.hcl
func Load(projectRoot string) (*Config, error) {
	configPath := filepath.Join(projectRoot, "forge.hcl")

	// Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("forge.hcl not found in %s", projectRoot)
	}

	// Parse HCL
	var cfg Config
	err := hclsimple.DecodeFile(configPath, nil, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse forge.hcl: %w", err)
	}

	// Apply environment variable overrides
	if region := os.Getenv("FORGE_REGION"); region != "" && cfg.Project != nil {
		cfg.Project.Region = region
	}

	// Set defaults
	if cfg.Defaults == nil {
		cfg.Defaults = &DefaultsBlock{}
	}
	if cfg.Defaults.Runtime == "" {
		cfg.Defaults.Runtime = "go1.x"
	}
	if cfg.Defaults.Timeout == 0 {
		cfg.Defaults.Timeout = 30
	}
	if cfg.Defaults.Memory == 0 {
		cfg.Defaults.Memory = 256
	}

	// Validate (pure function - no method)
	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ValidateConfig ensures the configuration is valid
// Pure function - no methods, takes Config as parameter
func ValidateConfig(c *Config) error {
	if c.Project == nil {
		return fmt.Errorf("project block is required")
	}
	if c.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if c.Project.Region == "" {
		return fmt.Errorf("project region is required")
	}
	return nil
}

// GetStackDefaults returns default values for a stack
// Pure function - no methods, takes Config as parameter
func GetStackDefaults(c *Config) *DefaultsBlock {
	if c.Defaults == nil {
		return &DefaultsBlock{
			Runtime: "go1.x",
			Timeout: 30,
			Memory:  256,
		}
	}
	return c.Defaults
}
