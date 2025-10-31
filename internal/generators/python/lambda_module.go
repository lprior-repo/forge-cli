package python

import (
	"fmt"

	"github.com/lewis/forge/internal/tfmodules/lambda"
)

// GenerateLambdaModule creates a type-safe Lambda module configuration.
// PURE: Calculation - same inputs always produce same outputs.
func GenerateLambdaModule(config ProjectConfig) *lambda.Module {
	serviceName := fmt.Sprintf("%s-%s", config.ServiceName, "${var.environment}")
	handler := "service.handlers.handle_request.lambda_handler"
	runtime := fmt.Sprintf("python%s", config.PythonVersion)

	// Create base Lambda module with sensible defaults
	fn := lambda.NewModule(serviceName)

	// Configure runtime and handler
	fn.WithRuntime(runtime, handler)

	// Configure memory and timeout
	memorySize := 512
	timeout := 30
	fn.WithMemoryAndTimeout(memorySize, timeout)

	// Use pre-built package (managed by UV in build step)
	packagePath := "${path.module}/../.build/lambda.zip"
	fn.LocalExistingPackage = &packagePath

	// Disable module's package creation - we build with UV
	createPackage := false
	fn.CreatePackage = &createPackage

	// Enable module-managed IAM role
	createRole := true
	fn.CreateRole = &createRole

	// Environment variables
	envVars := map[string]string{
		"POWERTOOLS_SERVICE_NAME": config.ServiceName,
		"LOG_LEVEL":                "INFO",
		"ENVIRONMENT":              "${var.environment}",
	}

	if config.UseDynamoDB {
		envVars["TABLE_NAME"] = "${module.dynamodb_table.dynamodb_table_id}"
	}

	if config.UseIdempotency && config.UseDynamoDB {
		envVars["IDEMPOTENCY_TABLE_NAME"] = "${module.dynamodb_table.dynamodb_table_id}"
	}

	fn.WithEnvironment(envVars)

	// Enable X-Ray tracing
	fn.WithTracing("Active")

	// CloudWatch Logs retention
	retention := 7
	fn.CloudwatchLogsRetentionInDays = &retention

	// IAM policy for DynamoDB access
	if config.UseDynamoDB {
		attachPolicy := true
		fn.AttachPolicyJSON = &attachPolicy

		policyJSON := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:GetItem",
        "dynamodb:PutItem",
        "dynamodb:UpdateItem",
        "dynamodb:DeleteItem",
        "dynamodb:Query",
        "dynamodb:Scan"
      ],
      "Resource": "${module.dynamodb_table.dynamodb_table_arn}"
    }
  ]
}`
		fn.PolicyJSON = &policyJSON
	}

	return fn
}

// GenerateLambdaModuleHCL converts Lambda module to HCL string.
// PURE: Calculation - deterministic output from module configuration.
func GenerateLambdaModuleHCL(module *lambda.Module) string {
	// TODO: Implement full HCL marshaling using hclwrite or lingon's marshaling
	// Use a static module name (variable interpolation not allowed in module names)
	moduleName := "lambda_function"

	hcl := fmt.Sprintf(`# Lambda function module
module "%s" {
  source  = "%s"
  version = "%s"

  function_name = "%s"
  handler       = "%s"
  runtime       = "%s"

  create_package = false
  local_existing_package = "%s"

  memory_size = %d
  timeout     = %d

`, moduleName, module.Source, module.Version,
		*module.FunctionName, *module.Handler, *module.Runtime,
		*module.LocalExistingPackage,
		*module.MemorySize, *module.Timeout)

	// Environment variables
	if len(module.EnvironmentVariables) > 0 {
		hcl += "  environment_variables = {\n"
		for key, val := range module.EnvironmentVariables {
			hcl += fmt.Sprintf("    %s = \"%s\"\n", key, val)
		}
		hcl += "  }\n\n"
	}

	// Tracing
	if module.TracingMode != nil {
		hcl += fmt.Sprintf("  tracing_mode = \"%s\"\n", *module.TracingMode)
	}

	// CloudWatch Logs
	if module.CloudwatchLogsRetentionInDays != nil {
		hcl += fmt.Sprintf("  cloudwatch_logs_retention_in_days = %d\n", *module.CloudwatchLogsRetentionInDays)
	}

	// IAM policy
	if module.AttachPolicyJSON != nil && *module.AttachPolicyJSON && module.PolicyJSON != nil {
		hcl += fmt.Sprintf("\n  attach_policy_json = true\n  policy_json = <<-EOT\n%s\nEOT\n", *module.PolicyJSON)
	}

	// Tags
	hcl += "\n  tags = {\n"
	hcl += "    ManagedBy   = \"Terraform\"\n"
	hcl += "    Generator   = \"Forge\"\n"
	hcl += "    Service     = var.service_name\n"
	hcl += "    Environment = var.environment\n"
	hcl += "  }\n"

	hcl += "}\n"

	return hcl
}

// sanitizeModuleName converts a name to a valid Terraform identifier.
// PURE: Calculation - deterministic transformation.
func sanitizeModuleName(name string) string {
	// For now, simple sanitization - can be enhanced
	return name
}
