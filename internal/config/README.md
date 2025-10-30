# internal/config

**HCL configuration loading and validation for Forge projects**

## Overview

The `config` package handles loading and validating `forge.hcl` configuration files. It uses HashiCorp's HCL (HashiCorp Configuration Language) for human-readable, declarative configuration with type safety and validation.

## Purpose

Provides **minimal, optional configuration** for projects that need to override conventions:

- Project name and AWS region
- Default runtime, timeout, memory settings
- Environment-specific overrides

**Philosophy:** Configuration is **optional** - conventions handle 90% of cases. Config file only needed for overrides.

## Configuration File: `forge.hcl`

```hcl
project {
  name   = "my-app"
  region = "us-east-1"
}

defaults {
  runtime = "go1.x"
  timeout = 30
  memory  = 256
}
```

## Data Structures

```go
// Config represents the complete forge.hcl file
type Config struct {
    Project  *ProjectBlock  `hcl:"project,block"`
    Defaults *DefaultsBlock `hcl:"defaults,block"`
}

// ProjectBlock contains project-wide settings
type ProjectBlock struct {
    Name   string `hcl:"name"`     // Project name (required)
    Region string `hcl:"region"`   // AWS region (required)
}

// DefaultsBlock contains default values for stacks
type DefaultsBlock struct {
    Runtime string `hcl:"runtime,optional"`  // Default runtime
    Timeout int    `hcl:"timeout,optional"`  // Default timeout (seconds)
    Memory  int    `hcl:"memory,optional"`   // Default memory (MB)
}
```

## Usage

### Load Configuration

```go
import "github.com/lewis/forge/internal/config"

// Load from project root
cfg, err := config.Load("/path/to/project")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Project: %s\n", cfg.Project.Name)
fmt.Printf("Region: %s\n", cfg.Project.Region)
```

### Environment Variable Overrides

Config values can be overridden by environment variables:

```bash
export FORGE_REGION=us-west-2
forge deploy  # Uses us-west-2 instead of forge.hcl value
```

**Precedence** (highest to lowest):
1. CLI flags (`--region=us-west-2`)
2. Environment variables (`FORGE_REGION=us-west-2`)
3. `forge.hcl` file
4. Default values

### Validation

The package provides **pure validation functions**:

```go
// ValidateConfig ensures the configuration is valid (PURE)
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
```

**Why pure functions?**
- ✅ Easier to test (no I/O dependencies)
- ✅ Composable (can combine validators)
- ✅ Predictable (same input = same output)

### Get Stack Defaults

```go
// GetStackDefaults returns default values for a stack (PURE)
func GetStackDefaults(c *Config) *DefaultsBlock {
    if c.Defaults == nil {
        // Sensible defaults
        return &DefaultsBlock{
            Runtime: "go1.x",
            Timeout: 30,
            Memory:  256,
        }
    }
    return c.Defaults
}
```

## Default Values

If `defaults` block is missing or incomplete, the package applies these defaults:

| Setting | Default | Description |
|---------|---------|-------------|
| `runtime` | `"go1.x"` | Lambda runtime |
| `timeout` | `30` | Function timeout (seconds) |
| `memory` | `256` | Function memory (MB) |

## HCL Benefits

**Why HCL over YAML/JSON?**

1. **Native Terraform syntax** - same language as `main.tf`
2. **Type safety** - struct tags enforce types at parse time
3. **Comments** - supports `//` and `#` comments
4. **Validation** - schema validation built-in
5. **Expressions** - supports interpolation and functions

**Example HCL features:**

```hcl
project {
  name   = "my-app"
  region = "us-east-1"  # Comment explaining region choice
}

defaults {
  # Go is our primary language
  runtime = "go1.x"

  # Conservative timeout for cost control
  timeout = 30

  # Minimal memory for small functions
  memory  = 256
}
```

## Error Handling

All errors are wrapped with context:

```go
func Load(projectRoot string) (*Config, error) {
    configPath := filepath.Join(projectRoot, "forge.hcl")

    // Check file exists
    if _, err := os.Stat(configPath); os.IsNotExist(err) {
        return nil, fmt.Errorf("forge.hcl not found in %s", projectRoot)
    }

    // Parse HCL
    var cfg Config
    err := hclsimple.DecodeFile(configPath, nil, &cfg)
    if err != nil {
        return nil, fmt.Errorf("failed to parse forge.hcl: %w", err)
    }

    // Validate
    if err := ValidateConfig(&cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

## Testing

### Unit Tests

```go
func TestLoadConfig(t *testing.T) {
    cfg, err := config.Load("./testdata/valid-project")

    assert.NoError(t, err)
    assert.Equal(t, "test-app", cfg.Project.Name)
    assert.Equal(t, "us-east-1", cfg.Project.Region)
}

func TestValidateConfig_MissingProject(t *testing.T) {
    cfg := &config.Config{Project: nil}

    err := config.ValidateConfig(cfg)

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "project block is required")
}
```

### Test Data

`testdata/forge.hcl`:
```hcl
project {
  name   = "test-app"
  region = "us-east-1"
}

defaults {
  runtime = "python3.13"
  timeout = 60
  memory  = 512
}
```

## Files

- **`config.go`** - Core types, Load function, validation

## Dependencies

```go
import "github.com/hashicorp/hcl/v2/hclsimple"  // HCL parsing
```

## Design Principles

1. **Pure functions** - `ValidateConfig`, `GetStackDefaults` have no side effects
2. **Immutable data** - Config structs are never mutated after loading
3. **Fail fast** - Validation happens immediately after parsing
4. **Sensible defaults** - Missing config values use production-ready defaults
5. **Environment overrides** - Allow runtime configuration without editing files

## Future Enhancements

- [ ] Support for multiple environments (dev, staging, prod)
- [ ] Config inheritance (extend base config)
- [ ] Secret references (AWS Secrets Manager, SSM Parameter Store)
- [ ] Variable interpolation (`${var.region}`)
- [ ] Schema validation with detailed error messages
- [ ] Config generation command (`forge config init`)
