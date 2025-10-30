package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type (
	// Config represents the forge.hcl configuration file.
	Config struct {
		Project  *ProjectBlock  `hcl:"project,block"`
		Defaults *DefaultsBlock `hcl:"defaults,block"`
	}

	// ProjectBlock contains project-wide configuration.
	ProjectBlock struct {
		Name   string `hcl:"name"`
		Region string `hcl:"region"`
	}

	// DefaultsBlock contains default values for stacks.
	DefaultsBlock struct {
		Runtime string `hcl:"runtime,optional"`
		Timeout int    `hcl:"timeout,optional"`
		Memory  int    `hcl:"memory,optional"`
	}
)

// ACTION: I/O operation that reads file and applies pure transformations.
func Load(projectRoot string) (*Config, error) {
	configPath := filepath.Join(projectRoot, "forge.hcl")

	// I/O: Check if config exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("forge.hcl not found in %s", projectRoot)
	}

	// I/O: Parse HCL
	var cfg Config
	err := hclsimple.DecodeFile(configPath, nil, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse forge.hcl: %w", err)
	}

	// PURE: Apply defaults and overrides immutably
	regionOverride := os.Getenv("FORGE_REGION")
	cfgWithDefaults := applyDefaults(cfg, regionOverride)

	// PURE: Validate
	if err := ValidateConfig(&cfgWithDefaults); err != nil {
		return nil, err
	}

	return &cfgWithDefaults, nil
}

// PURE: Calculation - no mutations, returns new Config.
func applyDefaults(cfg Config, regionOverride string) Config {
	newCfg := cfg

	// Apply region override immutably
	if regionOverride != "" && newCfg.Project != nil {
		// Create new ProjectBlock instead of mutating
		proj := *newCfg.Project
		proj.Region = regionOverride
		newCfg.Project = &proj
	}

	// Set defaults immutably
	if newCfg.Defaults == nil {
		newCfg.Defaults = &DefaultsBlock{
			Runtime: "go1.x",
			Timeout: 30,
			Memory:  256,
		}
	} else {
		// Create new DefaultsBlock instead of mutating
		defaults := *newCfg.Defaults
		if defaults.Runtime == "" {
			defaults.Runtime = "go1.x"
		}
		if defaults.Timeout == 0 {
			defaults.Timeout = 30
		}
		if defaults.Memory == 0 {
			defaults.Memory = 256
		}
		newCfg.Defaults = &defaults
	}

	return newCfg
}

// Pure function - no methods, takes Config as parameter.
func ValidateConfig(c *Config) error {
	if c.Project == nil {
		return errors.New("project block is required")
	}
	if c.Project.Name == "" {
		return errors.New("project name is required")
	}
	if c.Project.Region == "" {
		return errors.New("project region is required")
	}
	return nil
}

// Pure function - no methods, takes Config as parameter.
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
