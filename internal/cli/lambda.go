package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lewis/forge/internal/generators/python"
	"github.com/spf13/cobra"
)

// NewLambdaCmd creates the 'new lambda' command
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
		Short: "Create a new Lambda function project",
		Long: `Create a new Lambda function project with complete infrastructure.

Generates a production-ready Lambda project with:
  • AWS Lambda Powertools (logging, metrics, tracing)
  • Pydantic models with validation
  • 3-layer architecture (handler → logic → dal)
  • DynamoDB integration (optional)
  • Terraform infrastructure
  • uv-based build system (10-100x faster than pip)

Examples:
  forge new lambda my-service
  forge new lambda my-service --runtime python
  forge new lambda my-service --runtime python --dynamodb
  forge new lambda my-service --runtime go
  forge new lambda my-service --runtime nodejs`,
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

// LambdaProjectOptions holds configuration for creating a Lambda project
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

// createLambdaProject creates a new Lambda project
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
		opts.Description = fmt.Sprintf("%s Lambda service", opts.ServiceName)
	}
	if opts.TableName == "" {
		opts.TableName = fmt.Sprintf("%s-table", strings.ReplaceAll(opts.ServiceName, "_", "-"))
	}

	// Generate based on runtime
	switch opts.Runtime {
	case "python":
		return createPythonLambda(projectDir, projectName, opts)
	case "go":
		return fmt.Errorf("Go runtime not yet implemented")
	case "nodejs":
		return fmt.Errorf("Node.js runtime not yet implemented")
	default:
		return fmt.Errorf("unsupported runtime: %s", opts.Runtime)
	}
}

// createPythonLambda creates a Python Lambda project
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
