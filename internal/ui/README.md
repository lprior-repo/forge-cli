# internal/ui

**Terminal UI components for colored output, progress bars, and interactive prompts.**

## Overview

The `ui` package provides a clean terminal user interface for Forge CLI commands. It includes colored output formatting, progress indicators, and interactive prompts - all designed to be testable and composable. The package wraps `fatih/color` for colored output and provides a consistent UX across all CLI commands.

## Architecture

```
┌──────────────────────────────────────────────────┐
│              UI Components                       │
│   - Output: Colored messages                     │
│   - Progress: Build/deployment progress          │
│   - Prompt: Interactive user input               │
└──────────────────────────────────────────────────┘
                      ↓
    ┌─────────────────┴─────────────────┐
    ↓                                   ↓
┌─────────────────┐           ┌─────────────────┐
│   Color Themes  │           │  User Input     │
│  (Pure Config)  │           │  (I/O Actions)  │
└─────────────────┘           └─────────────────┘
```

## Key Types

```go
// Output provides colored terminal output
type Output struct {
    writer  io.Writer
    success *color.Color  // Green
    error   *color.Color  // Red
    warning *color.Color  // Yellow
    info    *color.Color  // Cyan
    dim     *color.Color  // Faint
}

// Progress tracks build/deployment progress
type Progress struct {
    total   int
    current int
    output  *Output
}

// Prompt provides interactive user input
type Prompt struct {
    reader io.Reader
    output *Output
}
```

## Output Formatting

### Success Messages

```go
out := ui.DefaultOutput()

// ✓ Build succeeded
out.Success("Build succeeded")

// ✓ Deployed 3 functions
out.Success("Deployed %d functions", 3)
```

### Error Messages

```go
// ✗ Build failed: compilation error
out.Error("Build failed: %v", err)

// ✗ Terraform not found
out.Error("Terraform not found")
```

### Warnings

```go
// ⚠ No tests found
out.Warning("No tests found")

// ⚠ Using cached build
out.Warning("Using cached build")
```

### Info Messages

```go
// ℹ Initializing Terraform...
out.Info("Initializing Terraform...")

// ℹ Found 5 functions
out.Info("Found %d functions", 5)
```

### Structured Output

```go
// === Deploying Lambda Functions ===
out.Header("Deploying Lambda Functions")

// [1/3] Building api...
// [2/3] Building worker...
// [3/3] Building notifier...
out.Step(1, 3, "Building api...")
out.Step(2, 3, "Building worker...")
out.Step(3, 3, "Building notifier...")

// Dimmed text for secondary info
out.Dim("Output directory: .forge/build")
```

## Progress Tracking

### Build Progress

```go
// Create progress tracker
progress := ui.NewProgress(3, out)

// Start first task
progress.Start("Building api...")
// ... do work ...
progress.Complete("api: bootstrap (2.5 MB)")

// Start next task
progress.Start("Building worker...")
// ... do work ...
progress.Complete("worker: lambda.zip (1.2 MB)")

progress.Start("Building notifier...")
// ... do work ...
progress.Complete("notifier: lambda.zip (0.8 MB)")

// All done
progress.Finish()
```

**Output:**
```
[1/3] Building api...
      ✓ api: bootstrap (2.5 MB)
[2/3] Building worker...
      ✓ worker: lambda.zip (1.2 MB)
[3/3] Building notifier...
      ✓ notifier: lambda.zip (0.8 MB)
```

### Deployment Progress

```go
progress := ui.NewProgress(4, out)

progress.Start("terraform init...")
progress.Complete("Terraform initialized")

progress.Start("terraform plan...")
progress.Complete("Plan: 3 to add, 0 to change, 0 to destroy")

progress.Start("terraform apply...")
progress.Complete("Applied successfully")

progress.Start("Capturing outputs...")
progress.Complete("Outputs captured")

progress.Finish()
```

## Interactive Prompts

### Yes/No Confirmation

```go
prompt := ui.NewPrompt(os.Stdin, out)

// Do you want to continue? (yes/no):
confirmed, err := prompt.Confirm("Do you want to continue?")
if err != nil {
    log.Fatal(err)
}

if confirmed {
    // User said yes
} else {
    // User said no
}
```

### Text Input

```go
// Enter project name:
name, err := prompt.Ask("Enter project name")
if err != nil {
    log.Fatal(err)
}

// Enter AWS region (default: us-east-1):
region, err := prompt.AskWithDefault("Enter AWS region", "us-east-1")
if err != nil {
    log.Fatal(err)
}
```

### Selection

```go
options := []string{"us-east-1", "us-west-2", "eu-west-1"}

// Select AWS region:
//   1) us-east-1
//   2) us-west-2
//   3) eu-west-1
selected, err := prompt.Select("Select AWS region", options)
if err != nil {
    log.Fatal(err)
}
```

## Testing

### Output Testing

```go
func TestBuildCommand(t *testing.T) {
    // Use buffer instead of stdout
    buf := &bytes.Buffer{}
    out := ui.NewOutput(buf)

    // Run command
    out.Success("Build complete")

    // Verify output
    assert.Contains(t, buf.String(), "Build complete")
    assert.Contains(t, buf.String(), "✓")
}
```

### Progress Testing

```go
func TestProgressTracking(t *testing.T) {
    buf := &bytes.Buffer{}
    out := ui.NewOutput(buf)
    progress := ui.NewProgress(2, out)

    progress.Start("Task 1")
    progress.Complete("Done")
    progress.Start("Task 2")
    progress.Complete("Done")

    assert.Contains(t, buf.String(), "[1/2]")
    assert.Contains(t, buf.String(), "[2/2]")
}
```

### Prompt Testing

```go
func TestPromptInput(t *testing.T) {
    // Simulate user input
    input := strings.NewReader("yes\n")
    out := ui.NewOutput(io.Discard)
    prompt := ui.NewPrompt(input, out)

    confirmed, err := prompt.Confirm("Continue?")
    assert.NoError(t, err)
    assert.True(t, confirmed)
}
```

## Color Theme

### Default Colors

- **Success** - Green (`✓`)
- **Error** - Red (`✗`)
- **Warning** - Yellow (`⚠`)
- **Info** - Cyan (`ℹ`)
- **Dim** - Faint text for secondary info
- **Regular** - Normal terminal color

### Disabling Colors

Colors automatically disable when:
- Output is not a TTY (pipes, redirects)
- `NO_COLOR` environment variable is set
- Running in CI/CD environments

```bash
# Disable colors
NO_COLOR=1 forge build

# Pipe output (colors auto-disabled)
forge build | tee build.log
```

## Usage Patterns

### CLI Command Structure

```go
func buildCommand(ctx context.Context) error {
    // 1. Create output
    out := ui.DefaultOutput()

    // 2. Show header
    out.Header("Building Lambda Functions")

    // 3. Show info
    out.Info("Found %d functions", len(functions))

    // 4. Track progress
    progress := ui.NewProgress(len(functions), out)

    for _, fn := range functions {
        progress.Start(fmt.Sprintf("Building %s...", fn.Name))

        // ... build logic ...

        progress.Complete(fmt.Sprintf("%s: %s", fn.Name, artifactPath))
    }

    progress.Finish()

    // 5. Show success
    out.Success("All functions built successfully")
    return nil
}
```

### Error Handling with Output

```go
func deploy(ctx context.Context) error {
    out := ui.DefaultOutput()
    out.Header("Deploying Infrastructure")

    // Build
    out.Info("Building functions...")
    if err := build(); err != nil {
        out.Error("Build failed: %v", err)
        out.Warning("Debug tips:")
        out.Print("  • Check function source code")
        out.Print("  • Verify dependencies")
        return err
    }

    // Deploy
    out.Info("Running terraform apply...")
    if err := terraformApply(); err != nil {
        out.Error("Deployment failed: %v", err)
        return err
    }

    out.Success("Deployment complete")
    return nil
}
```

## Design Principles

1. **Testable** - All output can be captured for testing
2. **Consistent** - Same symbols and colors across commands
3. **Informative** - Progress tracking shows what's happening
4. **Non-intrusive** - Colors disable automatically when not on TTY
5. **Composable** - Output, Progress, and Prompt work together

## Related Packages

- **internal/cli** - Uses UI components for all commands
- **internal/pipeline** - Uses Output for pipeline event display
- **fatih/color** - Underlying color library

## Coverage

- ✅ **97.5%** test coverage
- ✅ All output types tested
- ✅ Progress tracking tested
- ✅ Interactive prompts tested
- ✅ Color handling tested
