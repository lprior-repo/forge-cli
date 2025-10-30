package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lewis/forge/internal/generators/python"
)

// NewLambdaCmd creates the 'new lambda' command.
func NewLambdaCmd() *cobra.Command {
	var (
		runtime        string
		serviceName    string
		functionName   string
		description    string
		usePowertools  bool
		useIdempotency bool
		useDynamoDB    bool
		tableName      string
		apiPath        string
		httpMethod     string
	)

	cmd := &cobra.Command{
		Use:   "lambda [project-name]",
		Short: "Create production-ready Lambda project with infrastructure",
		Long: `
╭──────────────────────────────────────────────────────────────╮
│  ⚡ Forge Lambda Generator                                  │
╰──────────────────────────────────────────────────────────────╯

Generate production-ready Lambda projects with best practices built-in.
Complete with observability, validation, testing, and infrastructure.

🎯 What You Get (Python):
  • AWS Lambda Powertools integration
    - Structured logging with correlation IDs
    - Metrics and custom metrics
    - X-Ray tracing out of the box
  • Pydantic models with validation
  • Clean 3-layer architecture:
    - Handler layer (API contract)
    - Logic layer (business logic)
    - DAL layer (data access)
  • DynamoDB integration (optional)
  • Terraform infrastructure as code
  • uv-based builds (10-100x faster than pip)

📦 Included Infrastructure:
  • Lambda function (Python 3.13)
  • API Gateway v2 (HTTP API)
  • DynamoDB table with encryption
  • IAM roles with least privilege
  • CloudWatch logs and alarms
  • X-Ray tracing enabled

⚡ Build System:
  • Taskfile with 15+ commands
  • No Poetry installation required
  • Fast dependency resolution with uv
  • Automatic Lambda layer support

🚀 Examples:

  # Python Lambda with DynamoDB
  forge new lambda my-service

  # Python Lambda without DynamoDB
  forge new lambda my-service --no-dynamodb

  # Customize all options
  forge new lambda my-service \
    --runtime python \
    --function handler \
    --api-path /api/orders \
    --method POST

  # Coming soon: Go and Node.js
  forge new lambda my-service --runtime go
  forge new lambda my-service --runtime nodejs

💡 Pro Tips:
  • Start with defaults, customize later
  • All generated code is editable
  • Terraform infra is in infra/ directory
  • Use 'task' commands for common operations

📋 Available Runtimes:
  • python (default) - Python 3.13 with Powertools ✅
  • go - Coming soon
  • nodejs - Coming soon
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectName := args[0]

			// Validate runtime
			validRuntimes := map[string]bool{
				"python": true,
				"go":     true,
				"nodejs": true,
			}
			if !validRuntimes[runtime] {
				return fmt.Errorf("invalid runtime: %s (must be python, go, or nodejs)", runtime)
			}

			// Only Python is implemented currently
			if runtime != "python" {
				return fmt.Errorf("runtime %s not yet implemented (only python is available)", runtime)
			}

			return createLambdaProject(projectName, LambdaProjectOptions{
				Runtime:        runtime,
				ServiceName:    serviceName,
				FunctionName:   functionName,
				Description:    description,
				UsePowertools:  usePowertools,
				UseIdempotency: useIdempotency,
				UseDynamoDB:    useDynamoDB,
				TableName:      tableName,
				APIPath:        apiPath,
				HTTPMethod:     httpMethod,
			})
		},
	}

	// Flags
	cmd.Flags().StringVar(&runtime, "runtime", "python", "Runtime (python, go, nodejs)")
	cmd.Flags().StringVar(&serviceName, "service", "", "Service name (defaults to project name)")
	cmd.Flags().StringVar(&functionName, "function", "handler", "Function name")
	cmd.Flags().StringVar(&description, "description", "", "Project description")
	cmd.Flags().BoolVar(&usePowertools, "powertools", true, "Use AWS Lambda Powertools")
	cmd.Flags().BoolVar(&useIdempotency, "idempotency", true, "Enable idempotency support")
	cmd.Flags().BoolVar(&useDynamoDB, "dynamodb", true, "Include DynamoDB table")
	cmd.Flags().StringVar(&tableName, "table", "", "DynamoDB table name (defaults to service-table)")
	cmd.Flags().StringVar(&apiPath, "api-path", "/api/orders", "API Gateway path")
	cmd.Flags().StringVar(&httpMethod, "method", "POST", "HTTP method (GET, POST, PUT, DELETE)")

	return cmd
}

// LambdaProjectOptions holds configuration for creating a Lambda project.
type LambdaProjectOptions struct {
	Runtime        string
	ServiceName    string
	FunctionName   string
	Description    string
	UsePowertools  bool
	UseIdempotency bool
	UseDynamoDB    bool
	TableName      string
	APIPath        string
	HTTPMethod     string
}

// createLambdaProject creates a new Lambda project.
func createLambdaProject(projectName string, opts LambdaProjectOptions) error {
	projectDir := filepath.Join(".", projectName)

	// Check if directory already exists
	if _, err := os.Stat(projectDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory %s already exists", projectName)
	}

	// Set defaults
	if opts.ServiceName == "" {
		opts.ServiceName = strings.ReplaceAll(projectName, "-", "_")
	}
	if opts.Description == "" {
		opts.Description = opts.ServiceName + " Lambda service"
	}
	if opts.TableName == "" {
		opts.TableName = strings.ReplaceAll(opts.ServiceName, "_", "-") + "-table"
	}

	// Generate based on runtime
	switch opts.Runtime {
	case "python":
		return createPythonLambda(projectDir, projectName, opts)
	case "go":
		return errors.New("Go runtime not yet implemented")
	case "nodejs":
		return errors.New("Node.js runtime not yet implemented")
	default:
		return fmt.Errorf("unsupported runtime: %s", opts.Runtime)
	}
}

// createPythonLambda creates a Python Lambda project.
func createPythonLambda(projectDir, projectName string, opts LambdaProjectOptions) error {
	// Configure Python project
	config := python.ProjectConfig{
		ServiceName:    opts.ServiceName,
		FunctionName:   opts.FunctionName,
		Description:    opts.Description,
		PythonVersion:  "3.13",
		UsePowertools:  opts.UsePowertools,
		UseIdempotency: opts.UseIdempotency,
		UseDynamoDB:    opts.UseDynamoDB,
		TableName:      opts.TableName,
		APIPath:        opts.APIPath,
		HTTPMethod:     opts.HTTPMethod,
	}

	// Generate project
	if err := python.Generate(projectDir, config); err != nil {
		return fmt.Errorf("failed to generate Python project: %w", err)
	}

	// Success message
	fmt.Printf("✅ Created Python Lambda project: %s\n\n", projectName)

	fmt.Println("📦 Python Application:")
	if opts.UsePowertools {
		fmt.Println("  • AWS Lambda Powertools integrated")
	}
	fmt.Println("  • Pydantic models with validation")
	fmt.Println("  • 3-layer architecture")
	if opts.UseDynamoDB {
		fmt.Println("  • DynamoDB data access layer")
	}
	fmt.Println()

	fmt.Println("🏗️  Terraform Infrastructure:")
	fmt.Println("  • Lambda function with Python 3.13")
	fmt.Println("  • API Gateway v2 (HTTP API)")
	if opts.UseDynamoDB {
		fmt.Println("  • DynamoDB table with encryption")
	}
	fmt.Println("  • IAM roles and policies")
	fmt.Println("  • CloudWatch logs and X-Ray tracing")
	fmt.Println()

	fmt.Println("⚡ Build System:")
	fmt.Println("  • uv-based builds (10-100x faster than pip)")
	fmt.Println("  • No Poetry installation required")
	fmt.Println("  • Taskfile with 15+ commands")
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  task build                    # Build Lambda package with uv")
	fmt.Println("  task deploy                   # Deploy to AWS")
	fmt.Println("  task test-api                 # Test the deployed API")
	fmt.Println("  task logs                     # Tail CloudWatch logs")
	fmt.Println("  task destroy                  # Clean up AWS resources")
	fmt.Println()

	return nil
}
