package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Generator generates project scaffolding using pure Go code
type Generator struct {
	projectRoot string
}

// NewGenerator creates a new scaffold generator
func NewGenerator(projectRoot string) (*Generator, error) {
	return &Generator{
		projectRoot: projectRoot,
	}, nil
}

// ProjectOptions configures project generation
type ProjectOptions struct {
	Name   string
	Region string
}

// StackOptions configures stack generation
type StackOptions struct {
	Name        string
	Runtime     string
	Description string
}

// GenerateProject creates a new forge project structure
func (g *Generator) GenerateProject(opts *ProjectOptions) error {
	// Create project directory
	if err := os.MkdirAll(g.projectRoot, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate forge.hcl
	forgeHCL := generateForgeHCL(opts)
	if err := os.WriteFile(filepath.Join(g.projectRoot, "forge.hcl"), []byte(forgeHCL), 0644); err != nil {
		return fmt.Errorf("failed to write forge.hcl: %w", err)
	}

	// Generate .gitignore
	gitignore := generateGitignore()
	if err := os.WriteFile(filepath.Join(g.projectRoot, ".gitignore"), []byte(gitignore), 0644); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	// Generate README.md
	readme := generateReadme(opts)
	if err := os.WriteFile(filepath.Join(g.projectRoot, "README.md"), []byte(readme), 0644); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

// GenerateStack creates a new stack with the specified runtime
func (g *Generator) GenerateStack(opts *StackOptions) error {
	stackDir := filepath.Join(g.projectRoot, opts.Name)

	// Create stack directory
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		return fmt.Errorf("failed to create stack directory: %w", err)
	}

	// Generate stack.forge.hcl
	stackHCL := generateStackHCL(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0644); err != nil {
		return fmt.Errorf("failed to write stack.forge.hcl: %w", err)
	}

	// Generate runtime-specific files
	switch {
	case strings.HasPrefix(opts.Runtime, "go"), strings.HasPrefix(opts.Runtime, "provided"):
		return g.generateGoStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "python"):
		return g.generatePythonStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "nodejs"):
		return g.generateNodeStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "java"):
		return g.generateJavaStack(stackDir, opts)
	default:
		return fmt.Errorf("unsupported runtime: %s", opts.Runtime)
	}
}

// generateGoStack creates Go-specific files
func (g *Generator) generateGoStack(stackDir string, opts *StackOptions) error {
	// Generate main.go
	mainGo := generateGoMain(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "main.go"), []byte(mainGo), 0644); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	// Generate go.mod
	goMod := generateGoMod(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "go.mod"), []byte(goMod), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Generate main.tf
	mainTF := generateGoTerraform(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "main.tf"), []byte(mainTF), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}

// generatePythonStack creates Python-specific files
func (g *Generator) generatePythonStack(stackDir string, opts *StackOptions) error {
	// Generate handler.py
	handler := generatePythonHandler(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "handler.py"), []byte(handler), 0644); err != nil {
		return fmt.Errorf("failed to write handler.py: %w", err)
	}

	// Generate requirements.txt
	requirements := generateRequirementsTxt()
	if err := os.WriteFile(filepath.Join(stackDir, "requirements.txt"), []byte(requirements), 0644); err != nil {
		return fmt.Errorf("failed to write requirements.txt: %w", err)
	}

	// Generate main.tf
	mainTF := generatePythonTerraform(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "main.tf"), []byte(mainTF), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}

// generateNodeStack creates Node.js-specific files
func (g *Generator) generateNodeStack(stackDir string, opts *StackOptions) error {
	// Generate index.js
	index := generateNodeIndex(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "index.js"), []byte(index), 0644); err != nil {
		return fmt.Errorf("failed to write index.js: %w", err)
	}

	// Generate package.json
	pkg := generatePackageJson(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "package.json"), []byte(pkg), 0644); err != nil {
		return fmt.Errorf("failed to write package.json: %w", err)
	}

	// Generate main.tf
	mainTF := generateNodeTerraform(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "main.tf"), []byte(mainTF), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}

// generateJavaStack creates Java-specific files
func (g *Generator) generateJavaStack(stackDir string, opts *StackOptions) error {
	// Create Java directory structure
	javaDir := filepath.Join(stackDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return fmt.Errorf("failed to create Java source directory: %w", err)
	}

	// Generate Handler.java
	handler := generateJavaHandler(opts)
	if err := os.WriteFile(filepath.Join(javaDir, "Handler.java"), []byte(handler), 0644); err != nil {
		return fmt.Errorf("failed to write Handler.java: %w", err)
	}

	// Generate pom.xml
	pom := generatePomXml(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "pom.xml"), []byte(pom), 0644); err != nil {
		return fmt.Errorf("failed to write pom.xml: %w", err)
	}

	// Generate main.tf
	mainTF := generateJavaTerraform(opts)
	if err := os.WriteFile(filepath.Join(stackDir, "main.tf"), []byte(mainTF), 0644); err != nil {
		return fmt.Errorf("failed to write main.tf: %w", err)
	}

	return nil
}

// Code generation functions

func generateForgeHCL(opts *ProjectOptions) string {
	return fmt.Sprintf(`project {
  name   = "%s"
  region = "%s"
}
`, opts.Name, opts.Region)
}

func generateGitignore() string {
	return `.terraform/
*.tfstate
*.tfstate.backup
.terraform.lock.hcl
terraform.tfvars
bin/
dist/
*.zip
.DS_Store
`
}

func generateReadme(opts *ProjectOptions) string {
	return fmt.Sprintf(`# %s

A Forge serverless project.

## Getting Started

Deploy your stacks:

'''bash
forge deploy
'''

## Project Structure

- Each directory is a stack
- Each stack has a 'stack.forge.hcl' configuration
- Stacks are deployed independently

## Commands

- 'forge deploy [stack]' - Deploy stack(s)
- 'forge destroy [stack]' - Destroy stack(s)
- 'forge version' - Show version
`, opts.Name)
}

func generateStackHCL(opts *StackOptions) string {
	desc := opts.Description
	if desc == "" {
		desc = fmt.Sprintf("%s stack", opts.Name)
	}
	return fmt.Sprintf(`stack {
  name        = "%s"
  runtime     = "%s"
  description = "%s"
}
`, opts.Name, opts.Runtime, desc)
}

func generateGoMain(opts *StackOptions) string {
	return `package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	response := map[string]interface{}{
		"message": "Hello from Forge!",
		"path":    event.Path,
		"method":  event.HTTPMethod,
	}

	body, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       fmt.Sprintf("error marshaling response: %v", err),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
`
}

func generateGoMod(opts *StackOptions) string {
	return fmt.Sprintf(`module %s

go 1.21

require github.com/aws/aws-lambda-go v1.41.0
`, opts.Name)
}

func generateGoTerraform(opts *StackOptions) string {
	return fmt.Sprintf(`# Lambda function
resource "aws_lambda_function" "%s" {
  function_name = "%s"
  role          = aws_iam_role.lambda.arn
  handler       = "bootstrap"
  runtime       = "%s"
  filename      = "bootstrap.zip"

  environment {
    variables = {
      LOG_LEVEL = "info"
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "%s-lambda-role"

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

# Attach basic execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Outputs
output "function_name" {
  value = aws_lambda_function.%s.function_name
}

output "function_arn" {
  value = aws_lambda_function.%s.arn
}
`, opts.Name, opts.Name, opts.Runtime, opts.Name, opts.Name, opts.Name)
}

func generatePythonHandler(opts *StackOptions) string {
	return `import json
import logging

logger = logging.getLogger()
logger.setLevel(logging.INFO)

def handler(event, context):
    """Lambda handler function"""
    logger.info(f"Received event: {json.dumps(event)}")

    response = {
        "message": "Hello from Forge!",
        "path": event.get("path", "/"),
        "method": event.get("httpMethod", "GET"),
    }

    return {
        "statusCode": 200,
        "headers": {
            "Content-Type": "application/json"
        },
        "body": json.dumps(response)
    }
`
}

func generateRequirementsTxt() string {
	return `# Add your Python dependencies here
# boto3==1.34.0
`
}

func generatePythonTerraform(opts *StackOptions) string {
	return fmt.Sprintf(`# Lambda function
resource "aws_lambda_function" "%s" {
  function_name = "%s"
  role          = aws_iam_role.lambda.arn
  handler       = "handler.handler"
  runtime       = "%s"
  filename      = "function.zip"

  environment {
    variables = {
      LOG_LEVEL = "INFO"
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "%s-lambda-role"

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

# Attach basic execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Outputs
output "function_name" {
  value = aws_lambda_function.%s.function_name
}

output "function_arn" {
  value = aws_lambda_function.%s.arn
}
`, opts.Name, opts.Name, opts.Runtime, opts.Name, opts.Name, opts.Name)
}

func generateNodeIndex(opts *StackOptions) string {
	return `exports.handler = async (event) => {
  console.log('Received event:', JSON.stringify(event, null, 2));

  const response = {
    message: 'Hello from Forge!',
    path: event.path || '/',
    method: event.httpMethod || 'GET',
  };

  return {
    statusCode: 200,
    headers: {
      'Content-Type': 'application/json'
    },
    body: JSON.stringify(response)
  };
};
`
}

func generatePackageJson(opts *StackOptions) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "description": "%s",
  "main": "index.js",
  "dependencies": {}
}
`, opts.Name, opts.Description)
}

func generateNodeTerraform(opts *StackOptions) string {
	return fmt.Sprintf(`# Lambda function
resource "aws_lambda_function" "%s" {
  function_name = "%s"
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "%s"
  filename      = "function.zip"

  environment {
    variables = {
      LOG_LEVEL = "info"
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "%s-lambda-role"

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

# Attach basic execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Outputs
output "function_name" {
  value = aws_lambda_function.%s.function_name
}

output "function_arn" {
  value = aws_lambda_function.%s.arn
}
`, opts.Name, opts.Name, opts.Runtime, opts.Name, opts.Name, opts.Name)
}

func generateJavaHandler(opts *StackOptions) string {
	return `package com.example;

import com.amazonaws.services.lambda.runtime.Context;
import com.amazonaws.services.lambda.runtime.RequestHandler;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyRequestEvent;
import com.amazonaws.services.lambda.runtime.events.APIGatewayProxyResponseEvent;

import java.util.HashMap;
import java.util.Map;

public class Handler implements RequestHandler<APIGatewayProxyRequestEvent, APIGatewayProxyResponseEvent> {

    @Override
    public APIGatewayProxyResponseEvent handleRequest(APIGatewayProxyRequestEvent event, Context context) {
        context.getLogger().log("Received event: " + event.toString());

        Map<String, String> headers = new HashMap<>();
        headers.put("Content-Type", "application/json");

        String body = String.format(
            "{\"message\":\"Hello from Forge!\",\"path\":\"%s\",\"method\":\"%s\"}",
            event.getPath(),
            event.getHttpMethod()
        );

        APIGatewayProxyResponseEvent response = new APIGatewayProxyResponseEvent();
        response.setStatusCode(200);
        response.setHeaders(headers);
        response.setBody(body);

        return response;
    }
}
`
}

func generatePomXml(opts *StackOptions) string {
	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>com.example</groupId>
    <artifactId>%s</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>

    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
    </properties>

    <dependencies>
        <dependency>
            <groupId>com.amazonaws</groupId>
            <artifactId>aws-lambda-java-core</artifactId>
            <version>1.2.3</version>
        </dependency>
        <dependency>
            <groupId>com.amazonaws</groupId>
            <artifactId>aws-lambda-java-events</artifactId>
            <version>3.11.3</version>
        </dependency>
    </dependencies>

    <build>
        <plugins>
            <plugin>
                <groupId>org.apache.maven.plugins</groupId>
                <artifactId>maven-shade-plugin</artifactId>
                <version>3.5.0</version>
                <executions>
                    <execution>
                        <phase>package</phase>
                        <goals>
                            <goal>shade</goal>
                        </goals>
                    </execution>
                </executions>
            </plugin>
        </plugins>
    </build>
</project>
`, opts.Name)
}

func generateJavaTerraform(opts *StackOptions) string {
	return fmt.Sprintf(`# Lambda function
resource "aws_lambda_function" "%s" {
  function_name = "%s"
  role          = aws_iam_role.lambda.arn
  handler       = "com.example.Handler::handleRequest"
  runtime       = "%s"
  filename      = "function.jar"

  environment {
    variables = {
      LOG_LEVEL = "INFO"
    }
  }
}

# IAM role for Lambda
resource "aws_iam_role" "lambda" {
  name = "%s-lambda-role"

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

# Attach basic execution policy
resource "aws_iam_role_policy_attachment" "lambda_basic" {
  role       = aws_iam_role.lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# Outputs
output "function_name" {
  value = aws_lambda_function.%s.function_name
}

output "function_arn" {
  value = aws_lambda_function.%s.arn
}
`, opts.Name, opts.Name, opts.Runtime, opts.Name, opts.Name, opts.Name)
}
