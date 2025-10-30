# internal/scaffold

**Project and stack scaffolding generator - imperative code generation for boilerplate**

## Overview

The `scaffold` package generates boilerplate code for new Forge projects and stacks. It uses **imperative code generation** (not templates) to create working Terraform configurations, function code, and CI/CD workflows.

## Philosophy

**Generate approved, working code** - not empty templates:

- ✅ Generated code is ready to deploy
- ✅ Includes IAM roles, CloudWatch logging, Terraform state setup
- ✅ Production-ready defaults (encryption, versioning, least-privilege IAM)
- ❌ No placeholder comments like "TODO: Add your code here"

## Functions

### GenerateProject

Creates a complete Forge project structure:

```go
func GenerateProject(projectRoot string, opts *ProjectOptions) error
```

**ProjectOptions:**
```go
type ProjectOptions struct {
    Name   string  // Project name
    Region string  // AWS region
}
```

**Generated structure:**
```
my-app/
├── forge.hcl              # Project configuration
├── .gitignore             # Terraform artifacts
├── README.md              # Quick start guide
└── (ready for stacks)
```

### GenerateStack

Creates a runtime-specific stack (Lambda function + Terraform):

```go
func GenerateStack(projectRoot string, opts *StackOptions) error
```

**StackOptions:**
```go
type StackOptions struct {
    Name        string  // Stack name
    Runtime     string  // go1.x, python3.13, nodejs20.x, java21
    Description string  // Optional description
}
```

**Generated structure (Go example):**
```
my-app/api/
├── main.go              # Lambda handler (working code)
├── go.mod               # Module definition
├── main.tf              # Lambda + IAM + outputs
└── stack.forge.hcl      # Stack metadata
```

## Runtime Support

### Go (`generateGoStack`)

**Generated files:**
- `main.go` - API Gateway handler with JSON response
- `go.mod` - Module with `aws-lambda-go` dependency
- `main.tf` - Lambda function + IAM role + outputs

**Handler code:**
```go
// main.go (auto-generated)
package main

import (
    "context"
    "encoding/json"
    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
    response := map[string]interface{}{
        "message": "Hello from Forge!",
        "path":    event.Path,
        "method":  event.HTTPMethod,
    }

    body, _ := json.Marshal(response)

    return events.APIGatewayProxyResponse{
        StatusCode: 200,
        Headers:    map[string]string{"Content-Type": "application/json"},
        Body:       string(body),
    }, nil
}

func main() {
    lambda.Start(handler)
}
```

**Terraform code:**
```hcl
# main.tf (auto-generated)
resource "aws_lambda_function" "api" {
  function_name = "api"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  filename      = "bootstrap.zip"

  environment {
    variables = {
      LOG_LEVEL = "info"
    }
  }
}

resource "aws_iam_role" "lambda" {
  name = "api-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

output "function_name" {
  value = aws_lambda_function.api.function_name
}

output "function_arn" {
  value = aws_lambda_function.api.arn
}
```

### Python (`generatePythonStack`)

**Generated files:**
- `handler.py` - Lambda handler with logging
- `requirements.txt` - Empty (ready for dependencies)
- `main.tf` - Lambda + IAM + outputs

### Node.js (`generateNodeStack`)

**Generated files:**
- `index.js` - Async/await handler
- `package.json` - Package metadata
- `main.tf` - Lambda + IAM + outputs

### Java (`generateJavaStack`)

**Generated files:**
- `src/main/java/com/example/Handler.java` - RequestHandler implementation
- `pom.xml` - Maven configuration with Lambda dependencies
- `main.tf` - Lambda + IAM + outputs

## Code Generation Approach

### Imperative vs Templates

**We use imperative code generation:**

```go
func generateGoMain(opts *StackOptions) string {
    return `package main

import (
    "context"
    "github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event interface{}) (interface{}, error) {
    return map[string]string{"message": "Hello from Forge!"}, nil
}

func main() {
    lambda.Start(handler)
}
`
}
```

**Why not templates?** See `TECHNICAL_DECISIONS.md`:
- ✅ Type safety - compiler catches errors
- ✅ Easy testing - no template rendering needed
- ✅ Composable - string manipulation is simpler than template logic
- ✅ IDE support - Go strings have better tooling than templates

## Helper Functions

```go
// Pure functions that generate code strings

func generateForgeHCL(opts *ProjectOptions) string
func generateGitignore() string
func generateReadme(opts *ProjectOptions) string
func generateStackHCL(opts *StackOptions) string

// Runtime-specific generators
func generateGoMain(opts *StackOptions) string
func generateGoMod(opts *StackOptions) string
func generateGoTerraform(opts *StackOptions) string

func generatePythonHandler(opts *StackOptions) string
func generateRequirementsTxt() string
func generatePythonTerraform(opts *StackOptions) string

func generateNodeIndex(opts *StackOptions) string
func generatePackageJson(opts *StackOptions) string
func generateNodeTerraform(opts *StackOptions) string

func generateJavaHandler(opts *StackOptions) string
func generatePomXml(opts *StackOptions) string
func generateJavaTerraform(opts *StackOptions) string
```

## Testing

```go
func TestGenerateProject(t *testing.T) {
    tmpDir := t.TempDir()

    opts := &scaffold.ProjectOptions{
        Name:   "test-app",
        Region: "us-east-1",
    }

    err := scaffold.GenerateProject(tmpDir, opts)

    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "forge.hcl"))
    assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
    assert.FileExists(t, filepath.Join(tmpDir, "README.md"))
}

func TestGenerateGoStack(t *testing.T) {
    tmpDir := t.TempDir()

    opts := &scaffold.StackOptions{
        Name:    "api",
        Runtime: "go1.x",
    }

    err := scaffold.GenerateStack(tmpDir, opts)

    assert.NoError(t, err)
    assert.FileExists(t, filepath.Join(tmpDir, "api", "main.go"))
    assert.FileExists(t, filepath.Join(tmpDir, "api", "go.mod"))
    assert.FileExists(t, filepath.Join(tmpDir, "api", "main.tf"))
}
```

## Files

- **`generator.go`** - `GenerateProject`, `GenerateStack`, all code generation functions
- **`generator_test.go`** - Tests for project and stack generation

## Design Principles

1. **Working code, not templates** - Generated code deploys immediately
2. **Approved patterns** - IAM, logging, best practices baked in
3. **Pure functions** - All generators are pure (same input = same output)
4. **Runtime parity** - All runtimes have equivalent features
5. **Customizable** - Generated code is meant to be edited

## Future Enhancements

- [ ] CI/CD workflow generation (GitHub Actions, GitLab CI)
- [ ] Multi-function templates (API + worker queue)
- [ ] Terraform module support (DynamoDB, SQS, SNS)
- [ ] Custom templates via user config
- [ ] Interactive prompts for advanced options
- [ ] VPC configuration scaffolding
- [ ] Observability stack (CloudWatch dashboards, alarms)
