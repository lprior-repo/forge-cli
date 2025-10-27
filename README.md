# Forge

**Forge** is a developer-friendly tool for building and deploying serverless applications on AWS Lambda. It combines the power of Terraform with streamlined Lambda deployment workflows, inspired by battle-tested tools like Terramate and terraform-exec.

## Features

- **Multi-Runtime Support**: Go, Python, Node.js, and custom runtimes
- **Dependency Management**: Automatic topological sorting of stack dependencies
- **Incremental Deploys**: Build and deploy only what changed
- **Type-Safe**: Leverages terraform-exec for reliable Terraform operations
- **Testable**: Clean abstractions make testing easy
- **Production-Ready**: Based on patterns from Terramate, Atlantis, and terraform-exec

## Quick Start

### Installation

```bash
go install github.com/lewis/forge/cmd/forge@latest
```

### Create a New Project

```bash
forge new my-serverless-app
cd my-serverless-app
```

This creates:
```
my-serverless-app/
├── forge.hcl       # Project configuration
├── .gitignore
└── README.md
```

### Add a Lambda Function

```bash
forge new --stack api --runtime go1.x --description "REST API handler"
```

Creates:
```
api/
├── stack.forge.hcl  # Stack metadata
├── main.go          # Lambda handler
├── go.mod
└── main.tf          # Terraform configuration
```

### Deploy

```bash
# Build and deploy all stacks
forge deploy

# Deploy specific stack
forge deploy api

# Auto-approve (for CI/CD)
forge deploy --auto-approve
```

## Project Structure

### forge.hcl

Project-wide configuration:

```hcl
project {
  name   = "my-serverless-app"
  region = "us-east-1"
}

defaults {
  runtime = "go1.x"
  timeout = 30
  memory  = 256
}
```

### stack.forge.hcl

Per-stack configuration:

```hcl
stack {
  name        = "api"
  description = "REST API Lambda"
  runtime     = "go1.x"
  handler     = "."

  # Dependencies (deployed first)
  after = ["../shared-layer"]
}
```

## Supported Runtimes

### Go

```bash
forge new --stack my-func --runtime go1.x
```

Automatically:
- Compiles with `GOOS=linux GOARCH=amd64`
- Creates `bootstrap` binary for `provided.al2023` runtime
- Packages into deployment ZIP

### Python

```bash
forge new --stack my-func --runtime python3.11
```

Automatically:
- Installs dependencies from `requirements.txt`
- Packages code and dependencies into ZIP

### Node.js

```bash
forge new --stack my-func --runtime nodejs20.x
```

Automatically:
- Runs `npm install --production`
- Packages code and `node_modules` into ZIP

## Dependency Management

Forge automatically handles deployment order based on stack dependencies:

```hcl
# stacks/shared/stack.forge.hcl
stack {
  name = "shared"
  # No dependencies
}

# stacks/api/stack.forge.hcl
stack {
  name  = "api"
  after = ["../shared"]  # Deploys after shared
}

# stacks/worker/stack.forge.hcl
stack {
  name  = "worker"
  after = ["../shared"]  # Also depends on shared
}
```

Deployment order: `shared` → `api` & `worker` (parallel)

## Commands

### `forge new [project-name]`

Create a new Forge project.

```bash
forge new my-app
```

### `forge new --stack <name>`

Add a new stack to existing project.

```bash
forge new --stack worker --runtime python3.11
```

Options:
- `--runtime`: Runtime (go1.x, python3.11, nodejs20.x, etc.)
- `--description`: Stack description

### `forge init`

Initialize Terraform for all stacks.

```bash
forge init
```

### `forge deploy [stack-name]`

Build and deploy stacks.

```bash
# Deploy all
forge deploy

# Deploy specific stack
forge deploy api

# Auto-approve (no prompts)
forge deploy --auto-approve

# Parallel deployment (experimental)
forge deploy --parallel
```

### `forge destroy [stack-name]`

Destroy infrastructure.

```bash
# Destroy all
forge destroy

# Destroy specific stack
forge destroy api

# Auto-approve
forge destroy --auto-approve
```

### `forge version`

Print version information.

```bash
forge version
```

## Architecture

Forge is built on production-tested patterns:

### Terraform Execution

Uses HashiCorp's [terraform-exec](https://github.com/hashicorp/terraform-exec) for reliable Terraform operations:

```go
tf, _ := terraform.New(workdir, "terraform")
tf.Init(ctx)
hasChanges, _ := tf.Plan(ctx)
tf.Apply(ctx, terraform.AutoApprove(true))
```

### Stack Detection

Inspired by [Terramate](https://github.com/terramate-io/terramate):

```go
detector := stack.NewDetector(projectRoot)
stacks, _ := detector.FindStacks()
```

### Dependency Graph

Topological sorting ensures correct deployment order:

```go
graph, _ := stack.NewGraph(stacks)
ordered, _ := graph.TopologicalSort()
```

### Build System

Pluggable builders for different runtimes:

```go
builder := build.Get("go1.x")
artifact, _ := builder.Build(ctx, &build.Config{
    SourceDir:  "./api",
    OutputPath: "./api/bootstrap",
})
```

## Testing

Forge uses a three-level testing strategy from terraform-exec:

### Unit Tests (Fast)

No external dependencies:

```bash
make test
# or
go test -short ./...
```

### Integration Tests

Requires Terraform binary:

```bash
make test-integration
# or
go test -tags=integration ./...
```

### E2E Tests

Requires AWS credentials:

```bash
make test-e2e
# or
go test -tags=e2e -timeout=30m ./...
```

## CI/CD Integration

Forge auto-detects CI environments and provides native integrations:

### GitHub Actions

```yaml
- name: Deploy with Forge
  run: forge deploy --auto-approve
  env:
    AWS_ACCESS_KEY_ID: ${{ secrets.AWS_ACCESS_KEY_ID }}
    AWS_SECRET_ACCESS_KEY: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
```

Forge outputs GitHub Actions annotations for errors.

### GitLab CI

```yaml
deploy:
  script:
    - forge deploy --auto-approve
```

### Local Development

```bash
forge deploy  # Interactive prompts
```

## Configuration

### Environment Variables

- `FORGE_REGION`: Override AWS region from forge.hcl
- `AWS_REGION`: AWS region for Terraform
- `AWS_PROFILE`: AWS profile to use

### Region Override

```bash
# Command-line flag
forge deploy --region us-west-2

# Environment variable
FORGE_REGION=us-west-2 forge deploy
```

## Examples

### Multi-Function Project

```
my-app/
├── forge.hcl
├── shared-layer/
│   ├── stack.forge.hcl
│   └── main.tf
├── api/
│   ├── stack.forge.hcl
│   ├── main.go
│   └── main.tf
└── worker/
    ├── stack.forge.hcl
    ├── handler.py
    ├── requirements.txt
    └── main.tf
```

### Different Runtimes

```bash
forge new my-app
cd my-app

# Go function
forge new --stack api --runtime go1.x

# Python function
forge new --stack processor --runtime python3.11

# Node.js function
forge new --stack webhook --runtime nodejs20.x

forge deploy
```

## Development

### Building from Source

```bash
git clone https://github.com/lewis/forge
cd forge
make build
./bin/forge version
```

### Running Tests

```bash
make test          # Unit tests
make test-all      # Unit + integration
make verify        # Tests + linting
```

### Project Layout

```
forge/
├── cmd/forge/           # CLI entry point
├── internal/
│   ├── terraform/       # Terraform wrapper
│   ├── stack/           # Stack detection & graphing
│   ├── build/           # Build system
│   ├── scaffold/        # Code generation
│   ├── config/          # Configuration loading
│   ├── cli/             # Cobra commands
│   └── ci/              # CI/CD integration
├── testdata/            # Test fixtures
└── Makefile
```

## Inspiration

Forge builds on proven patterns from:

- **[terraform-exec](https://github.com/hashicorp/terraform-exec)**: Reliable Terraform execution
- **[Terramate](https://github.com/terramate-io/terramate)**: Stack-based architecture
- **[Atlantis](https://github.com/runatlantis/atlantis)**: Terraform automation
- **[Lingon](https://github.com/volvo-cars/lingon)**: Type-safe infrastructure

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR.

## Roadmap

- [ ] Shared layers and dependencies
- [ ] State management (S3 backend auto-config)
- [ ] Plan preview in PR comments
- [ ] Drift detection
- [ ] Cost estimation
- [ ] CloudFormation outputs import
- [ ] API Gateway integration templates
- [ ] EventBridge patterns
- [ ] Step Functions orchestration

---

Built with ❤️ for developers who want Terraform's power with better DX.
