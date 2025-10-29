// Example: Generate a complete Python Lambda project with Terraform
//
// Usage: go run examples/generate-python-lambda.go
//
// This generates a production-ready Python Lambda service with:
// - AWS Lambda Powertools (logging, metrics, tracing)
// - Pydantic models with validation
// - 3-layer architecture (handler → logic → dal)
// - DynamoDB integration
// - Terraform infrastructure (NO CloudFormation!)
// - Complete project structure

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lewis/forge/internal/generators/python"
)

func main() {
	// Configure the Python Lambda project
	config := python.ProjectConfig{
		ServiceName:    "orders-service",
		FunctionName:   "create-order",
		Description:    "Orders service API - Create order endpoint",
		PythonVersion:  "3.13",
		UsePowertools:  true,
		UseIdempotency: true,
		UseDynamoDB:    true,
		TableName:      "orders-table",
		APIPath:        "/api/orders",
		HTTPMethod:     "POST",
	}

	// Generate in examples directory
	projectRoot := filepath.Join("examples", "generated-python-lambda")

	// Clean up if exists
	os.RemoveAll(projectRoot)

	fmt.Println("🔨 Generating Python Lambda project with Terraform...")
	fmt.Printf("   Service: %s\n", config.ServiceName)
	fmt.Printf("   Python: %s\n", config.PythonVersion)
	fmt.Printf("   Powertools: %v\n", config.UsePowertools)
	fmt.Printf("   DynamoDB: %v\n", config.UseDynamoDB)
	fmt.Println()

	generator := python.NewGenerator(projectRoot, config)
	if err := generator.Generate(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error generating project: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Project generated successfully!")
	fmt.Printf("\n📁 Location: %s\n\n", projectRoot)

	fmt.Println("📦 Python Application:")
	fmt.Println("  • AWS Lambda Powertools integrated")
	fmt.Println("  • Pydantic models with validation")
	fmt.Println("  • 3-layer architecture")
	fmt.Println("  • DynamoDB data access layer")
	fmt.Println()

	fmt.Println("🏗️  Terraform Infrastructure (NO CloudFormation!):")
	fmt.Println("  • Lambda function with Python 3.13")
	fmt.Println("  • API Gateway v2 (HTTP API)")
	fmt.Println("  • DynamoDB table with encryption")
	fmt.Println("  • IAM roles and policies")
	fmt.Println("  • CloudWatch logs and X-Ray tracing")
	fmt.Println()

	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", projectRoot)
	fmt.Println("  poetry install                    # Install Python dependencies")
	fmt.Println("  poetry run pytest                 # Run tests")
	fmt.Println("  cd terraform && terraform init    # Initialize Terraform")
	fmt.Println("  terraform plan                    # Preview infrastructure")
	fmt.Println("  terraform apply                   # Deploy to AWS")
	fmt.Println()
}
