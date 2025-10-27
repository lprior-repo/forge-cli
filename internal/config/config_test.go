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
			err := os.WriteFile(configPath, []byte(tt.content), 0644)
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
			err := tt.cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
