# internal/cli

**Command-line interface for Forge - the imperative shell around the functional core**

## Overview

The `cli` package provides all user-facing commands for Forge using the [Cobra](https://github.com/spf13/cobra) framework. It serves as the **imperative shell** that wraps Forge's **pure functional core**, handling I/O, user interaction, and orchestration of internal packages.

## Architecture Pattern: Functional Core, Imperative Shell

```
┌─────────────────────────────────────────┐
│         CLI Commands (Imperative)       │  ← User interaction, I/O
│  - Parse flags                          │
│  - Print output                         │
│  - Handle errors                        │
│  - Exit codes                           │
└─────────────────────────────────────────┘
                  ↓
┌─────────────────────────────────────────┐
│     Internal Packages (Functional)      │  ← Pure logic, testable
│  - build: Build functions               │
│  - pipeline: Orchestrate stages         │
│  - terraform: Execute terraform         │
│  - scaffold: Generate code              │
└─────────────────────────────────────────┘
```

**Why this separation?**
- ✅ **CLI logic is thin** - just command parsing and output formatting
- ✅ **Core logic is pure** - easily testable without CLI
- ✅ **Better error handling** - core returns Either, CLI converts to exit codes
- ✅ **Reusable logic** - core packages can be used by other tools (SDK, API, etc.)

## Commands

### Root Command (`root.go`)

**Entry point** for all Forge commands.

```go
func Execute() {
    cmd := NewRootCmd()
    if err := cmd.Execute(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

**Global flags:**
- `--verbose, -v` - Enable verbose output (debug logging)
- `--region, -r` - AWS region override (overrides `forge.hcl`)

**Subcommands:**
- `forge new` - Create new project
- `forge build` - Build Lambda functions
- `forge deploy` - Deploy infrastructure
- `forge destroy` - Tear down infrastructure
- `forge version` - Show version information

### `forge new` (`new.go`)

**Purpose:** Scaffold a new Forge project with convention-based structure.

**Usage:**
```bash
forge new my-app --runtime=go --auto-state
forge new my-app --runtime=python --region=us-west-2
```

**Flags:**
- `--runtime` - Runtime for initial function (go, python, nodejs, java)
- `--region` - AWS region (default: `us-east-1`)
- `--auto-state` - Auto-provision S3 bucket + DynamoDB for Terraform state

**What it does:**
1. **Validates** project name (must be lowercase, alphanumeric + hyphens)
2. **Generates** project structure:
   ```
   my-app/
   ├── infra/
   │   ├── main.tf       # Lambda + IAM resources
   │   ├── variables.tf  # namespace variable
   │   ├── backend.tf    # S3 state backend (if --auto-state)
   │   └── outputs.tf
   ├── src/functions/
   │   └── api/          # Hello-world function
   │       └── main.go   # (or handler.py, index.js, etc.)
   ├── .github/workflows/
   │   ├── deploy.yml    # Production deployment
   │   ├── pr-preview.yml  # PR preview environments
   │   └── pr-cleanup.yml  # PR cleanup
   ├── forge.hcl         # Project configuration
   ├── .gitignore
   └── README.md
   ```
3. **Provisions** state backend (if `--auto-state`):
   - Creates S3 bucket: `forge-state-{project-name}`
   - Enables versioning and encryption
   - Creates DynamoDB table: `forge-locks-{project-name}`
   - Generates `backend.tf` with namespace-aware state keys

**Implementation:**
```go
func runNew(cmd *cobra.Command, args []string) error {
    projectName := args[0]

    // Call pure function from internal/scaffold
    opts := scaffold.ProjectOptions{
        Name:   projectName,
        Region: region,
    }

    err := scaffold.GenerateProject(projectName, &opts)
    if err != nil {
        return err
    }

    // If --auto-state, provision backend (I/O action)
    if autoState {
        // Call internal/state package
        state.ProvisionBackend(projectName, region)
    }

    fmt.Printf("✓ Created project: %s\n", projectName)
    return nil
}
```

### `forge build` (`build.go`)

**Purpose:** Build all Lambda functions using convention-based discovery.

**Usage:**
```bash
forge build                # Build all functions in src/functions/*
forge build --verbose      # Show detailed build logs
```

**What it does:**
1. **Discovers** functions via `internal/discovery` (scans `src/functions/*`)
2. **Detects** runtime from entry files (`main.go`, `index.js`, `app.py`)
3. **Builds** each function using `internal/build` registry
4. **Caches** artifacts based on SHA256 checksums
5. **Outputs** build results:
   ```
   Building 3 functions...
   ✓ api (go1.x) - 2.1s
   ✓ worker (python3.13) - 1.8s
   ✓ notifier (nodejs20.x) - 1.2s

   Built 3 functions in 5.1s
   Artifacts in .forge/build/
   ```

**Implementation:**
```go
func runBuild(cmd *cobra.Command, args []string) error {
    // 1. Discover functions (pure function call)
    functions, err := discovery.ScanFunctions(".")
    if err != nil {
        return err
    }

    // 2. Convert to build configs (pure transformation)
    configs := lo.Map(functions, func(f discovery.Function, _ int) build.Config {
        return discovery.ToBuildConfig(f, ".forge/build")
    })

    // 3. Build all (Either monad)
    registry := build.NewRegistry()
    result := build.BuildAll(cmd.Context(), configs, registry)

    // 4. Handle result (imperative shell)
    return E.Fold(
        func(err error) error {
            return fmt.Errorf("build failed: %w", err)
        },
        func(artifacts []build.Artifact) error {
            fmt.Printf("Built %d functions\n", len(artifacts))
            return nil
        },
    )(result)
}
```

### `forge deploy` (`deploy.go`)

**Purpose:** Deploy infrastructure to AWS via Terraform (pipeline-first).

**Usage:**
```bash
forge deploy                      # Deploy to production
forge deploy --namespace=pr-123   # Deploy to ephemeral environment
forge deploy --auto-approve       # Skip plan confirmation
```

**Flags:**
- `--namespace` - Resource namespace (for PR previews)
- `--auto-approve` - Skip interactive approval (for CI/CD)
- `--var` - Pass Terraform variables (e.g., `--var="region=us-west-2"`)

**What it does:**
1. **Builds** functions (runs `forge build` internally)
2. **Initializes** Terraform (`terraform init`)
3. **Plans** changes (`terraform plan`)
4. **Prompts** user for approval (unless `--auto-approve`)
5. **Applies** changes (`terraform apply`)
6. **Outputs** deployment results (Function URLs, ARNs, etc.)

**Namespace behavior:**
```bash
forge deploy --namespace=pr-123
```
- Sets `TF_VAR_namespace=pr-123`
- All resources prefixed: `my-app-pr-123-api`
- State file: `forge/pr-123-terraform.tfstate`
- Isolated AWS resources (no conflicts with production)

**Implementation:**
```go
func runDeploy(cmd *cobra.Command, args []string) error {
    ctx := cmd.Context()

    // 1. Build stage (pure function)
    buildStage := pipeline.NewBuildStage()

    // 2. Terraform stages (imperative actions)
    tfStages := pipeline.NewTerraformStages(terraform.NewExecutor("terraform"))

    // 3. Compose pipeline (functional composition)
    deployPipeline := pipeline.Chain(
        pipeline.New(buildStage),
        tfStages,
    )

    // 4. Run pipeline (Either monad)
    initialState := pipeline.State{
        ProjectDir: ".",
        Namespace:  namespace,
    }

    result := pipeline.Run(deployPipeline, ctx, initialState)

    // 5. Handle result (imperative shell)
    return E.Fold(
        func(err error) error {
            return fmt.Errorf("deployment failed: %w", err)
        },
        func(state pipeline.State) error {
            fmt.Println("✓ Deployment successful")
            printOutputs(state.Outputs)
            return nil
        },
    )(result)
}
```

### `forge destroy` (`destroy.go`)

**Purpose:** Tear down infrastructure (for ephemeral environments).

**Usage:**
```bash
forge destroy --namespace=pr-123        # Destroy PR environment
forge destroy --namespace=pr-123 --auto-approve  # Skip confirmation
```

**Flags:**
- `--namespace` - **Required** - Which namespace to destroy
- `--auto-approve` - Skip interactive confirmation

**Safety checks:**
- ❌ **Cannot destroy production** without namespace (prevents accidents)
- ✅ **Requires explicit namespace** - no default, no guessing
- ✅ **Shows plan first** - preview what will be destroyed

**What it does:**
1. **Validates** namespace is provided (error if missing)
2. **Runs** `terraform plan -destroy` to preview
3. **Prompts** for confirmation (unless `--auto-approve`)
4. **Destroys** all resources via `terraform destroy`
5. **Deletes** state file for that namespace

**Implementation:**
```go
func runDestroy(cmd *cobra.Command, args []string) error {
    if namespace == "" {
        return errors.New("--namespace is required for destroy (safety check)")
    }

    fmt.Printf("WARNING: This will destroy all resources in namespace: %s\n", namespace)

    if !autoApprove {
        if !confirmDestroy(namespace) {
            fmt.Println("Destroy cancelled")
            return nil
        }
    }

    // Execute terraform destroy
    executor := terraform.NewExecutor("terraform")
    err := executor.Destroy(ctx, "infra",
        terraform.WithVar("namespace", namespace),
    )

    if err != nil {
        return fmt.Errorf("destroy failed: %w", err)
    }

    fmt.Printf("✓ Destroyed namespace: %s\n", namespace)
    return nil
}
```

### `forge version` (`version.go`)

**Purpose:** Show version information (for debugging and support).

**Usage:**
```bash
forge version
```

**Output:**
```
Forge v0.1.0
Go version: go1.21.5
Commit: a9e82a5
Built: 2024-01-15T10:30:00Z
```

**Implementation:**
```go
var (
    version   = "dev"
    commit    = "unknown"
    buildDate = "unknown"
    goVersion = runtime.Version()
)

func runVersion(cmd *cobra.Command, args []string) error {
    fmt.Printf("Forge %s\n", version)
    fmt.Printf("Go version: %s\n", goVersion)
    fmt.Printf("Commit: %s\n", commit)
    fmt.Printf("Built: %s\n", buildDate)
    return nil
}
```

These variables are set at build time via `-ldflags`:
```bash
go build -ldflags="-X 'main.version=v0.1.0' -X 'main.commit=$(git rev-parse HEAD)'"
```

## Error Handling

The CLI converts Either monad results into appropriate exit codes:

```go
func Execute() {
    cmd := NewRootCmd()
    if err := cmd.Execute(); err != nil {
        // Print error to stderr
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)

        // Exit with code 1 (failure)
        os.Exit(1)
    }
    // Implicit os.Exit(0) on success
}
```

**Exit codes:**
- `0` - Success
- `1` - Generic error
- Future: specific codes for different error types (e.g., 2 = validation error, 3 = AWS error)

## Testing

### Unit Tests

CLI commands are tested using Cobra's test helpers:

```go
func TestNewCommand(t *testing.T) {
    cmd := NewNewCmd()
    cmd.SetArgs([]string{"my-app", "--runtime=go"})

    err := cmd.Execute()
    assert.NoError(t, err)
    assert.DirExists(t, "my-app")
}
```

### Integration Tests

Full end-to-end tests in `integration_commands_test.go`:

```go
func TestDeployWorkflow(t *testing.T) {
    // 1. Create project
    runCommand(t, "new", "test-app", "--runtime=go")

    // 2. Build functions
    runCommand(t, "build")

    // 3. Deploy (with mock terraform executor)
    runCommand(t, "deploy", "--auto-approve")

    // 4. Verify state
    assert.FileExists(t, ".forge/build/api.zip")
}
```

## Files

- **`root.go`** - Root command and global flags
- **`new.go`** - `forge new` command (project scaffolding)
- **`build.go`** - `forge build` command (function builds)
- **`deploy.go`** - `forge deploy` command (deployment pipeline)
- **`destroy.go`** - `forge destroy` command (teardown)
- **`version.go`** - `forge version` command (version info)
- **`*_test.go`** - Unit and integration tests

## Dependencies

```go
import (
    "github.com/spf13/cobra"          // CLI framework
    "github.com/lewis/forge/internal/build"
    "github.com/lewis/forge/internal/pipeline"
    "github.com/lewis/forge/internal/scaffold"
    "github.com/lewis/forge/internal/terraform"
    // ... other internal packages
)
```

## Design Principles

1. **Thin CLI layer** - logic lives in internal packages, not CLI
2. **Functional core** - CLI calls pure functions, handles I/O at boundary
3. **Either monad** - core returns Either, CLI converts to exit codes
4. **User-friendly errors** - clear, actionable error messages
5. **Pipeline orchestration** - CLI composes pipelines, doesn't implement logic

## Related Packages

- **`internal/pipeline`** - Orchestrates multi-stage deployment workflows
- **`internal/build`** - Builds Lambda functions across runtimes
- **`internal/scaffold`** - Generates project boilerplate code
- **`internal/terraform`** - Executes Terraform operations
- **`internal/discovery`** - Convention-based function discovery

## Future Enhancements

- [ ] `forge logs` - Tail CloudWatch logs by namespace
- [ ] `forge list` - Show all deployed namespaces
- [ ] `forge plan` - Show Terraform plan without deploying
- [ ] `forge init` - Initialize existing directory as Forge project
- [ ] `forge validate` - Validate project structure and config
- [ ] Interactive TUI (bubbletea) for `forge new` with visual prompts
- [ ] `forge rollback` - Rollback to previous deployment
- [ ] `forge status` - Show deployment status across namespaces
