package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid config",
			content: `
project {
  name   = "test-project"
  region = "us-east-1"
}

defaults {
  runtime = "python3.11"
  timeout = 60
  memory  = 512
}
`,
			want: &Config{
				Project: &ProjectBlock{
					Name:   "test-project",
					Region: "us-east-1",
				},
				Defaults: &DefaultsBlock{
					Runtime: "python3.11",
					Timeout: 60,
					Memory:  512,
				},
			},
			wantErr: false,
		},
		{
			name: "minimal config",
			content: `
project {
  name   = "minimal"
  region = "eu-west-1"
}
`,
			want: &Config{
				Project: &ProjectBlock{
					Name:   "minimal",
					Region: "eu-west-1",
				},
				Defaults: &DefaultsBlock{
					Runtime: "go1.x",
					Timeout: 30,
					Memory:  256,
				},
			},
			wantErr: false,
		},
		{
			name: "missing project name",
			content: `
project {
  region = "us-east-1"
}
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "forge.hcl")

			// Write config file
			err := os.WriteFile(configPath, []byte(tt.content), 0o644)
			require.NoError(t, err)

			// Load config
			got, err := Load(tmpDir)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.Project.Name, got.Project.Name)
			assert.Equal(t, tt.want.Project.Region, got.Project.Region)
			assert.Equal(t, tt.want.Defaults.Runtime, got.Defaults.Runtime)
			assert.Equal(t, tt.want.Defaults.Timeout, got.Defaults.Timeout)
			assert.Equal(t, tt.want.Defaults.Memory, got.Defaults.Memory)
		})
	}
}

func TestLoadNonexistentConfig(t *testing.T) {
	// Test that missing forge.hcl returns specific error
	tmpDir := t.TempDir()
	_, err := Load(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forge.hcl not found")
	assert.Contains(t, err.Error(), tmpDir)
}

func TestLoadInvalidHCL(t *testing.T) {
	// Test that invalid HCL returns parse error
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "forge.hcl")
	err := os.WriteFile(configPath, []byte("invalid {{{"), 0o644)
	require.NoError(t, err)

	_, err = Load(tmpDir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse forge.hcl")
}

func TestLoadWithInvalidConfig(t *testing.T) {
	// Test that validation errors are returned
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "forge.hcl")
	content := `
project {
  # Missing name and region
}
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	_, err = Load(tmpDir)
	require.Error(t, err)
	// Validation error should be returned
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Test that FORGE_REGION env var overrides config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "forge.hcl")
	content := `
project {
  name   = "test-project"
  region = "us-east-1"
}
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	// Set environment variable
	t.Setenv("FORGE_REGION", "eu-west-2")

	cfg, err := Load(tmpDir)
	require.NoError(t, err)
	assert.Equal(t, "eu-west-2", cfg.Project.Region, "FORGE_REGION should override config")
}

func TestEnvironmentVariableWithNilProject(t *testing.T) {
	// Test that env var doesn't panic when Project is nil
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "forge.hcl")
	content := `
defaults {
  runtime = "go1.x"
}
`
	err := os.WriteFile(configPath, []byte(content), 0o644)
	require.NoError(t, err)

	t.Setenv("FORGE_REGION", "eu-west-2")

	_, err = Load(tmpDir)
	// Should fail validation, but shouldn't panic
	require.Error(t, err)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Project: &ProjectBlock{
					Name:   "test",
					Region: "us-east-1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing project block",
			cfg: &Config{
				Project: nil,
			},
			wantErr: true,
		},
		{
			name: "missing project name",
			cfg: &Config{
				Project: &ProjectBlock{
					Region: "us-east-1",
				},
			},
			wantErr: true,
		},
		{
			name: "missing region",
			cfg: &Config{
				Project: &ProjectBlock{
					Name: "test",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetStackDefaults(t *testing.T) {
	t.Run("returns explicit defaults when set", func(t *testing.T) {
		cfg := &Config{
			Project: &ProjectBlock{
				Name:   "test",
				Region: "us-east-1",
			},
			Defaults: &DefaultsBlock{
				Runtime: "python3.11",
				Timeout: 60,
				Memory:  512,
			},
		}

		defaults := GetStackDefaults(cfg)
		assert.Equal(t, "python3.11", defaults.Runtime)
		assert.Equal(t, 60, defaults.Timeout)
		assert.Equal(t, 512, defaults.Memory)
	})

	t.Run("returns hardcoded defaults when Defaults is nil", func(t *testing.T) {
		cfg := &Config{
			Project: &ProjectBlock{
				Name:   "test",
				Region: "us-east-1",
			},
			Defaults: nil,
		}

		defaults := GetStackDefaults(cfg)
		assert.Equal(t, "go1.x", defaults.Runtime)
		assert.Equal(t, 30, defaults.Timeout, "Default timeout must be exactly 30")
		assert.Equal(t, 256, defaults.Memory, "Default memory must be exactly 256")
	})

	t.Run("default values are precisely defined", func(t *testing.T) {
		cfg := &Config{Defaults: nil}
		defaults := GetStackDefaults(cfg)

		// These exact values are critical - mutation testing ensures we catch changes
		assert.NotEqual(t, 29, defaults.Timeout, "Timeout should not be 29")
		assert.NotEqual(t, 31, defaults.Timeout, "Timeout should not be 31")
		assert.Equal(t, 30, defaults.Timeout, "Timeout must be exactly 30")

		assert.NotEqual(t, 255, defaults.Memory, "Memory should not be 255")
		assert.NotEqual(t, 257, defaults.Memory, "Memory should not be 257")
		assert.Equal(t, 256, defaults.Memory, "Memory must be exactly 256")
	})
}
