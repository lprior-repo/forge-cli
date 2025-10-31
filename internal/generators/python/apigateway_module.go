// Package python provides Python Lambda project generation with type-safe Terraform modules.
package python

import (
	"fmt"
	"strings"

	"github.com/lewis/forge/internal/tfmodules/apigatewayv2"
)

// GenerateAPIGatewayModule creates a type-safe API Gateway V2 module configuration.
// PURE: Calculation - same inputs always produce same outputs.
func GenerateAPIGatewayModule(config ProjectConfig) *apigatewayv2.Module {
	apiName := fmt.Sprintf("%s-%s", config.ServiceName, "${var.environment}")

	// Create base API Gateway module
	api := apigatewayv2.NewModule(apiName)

	// Configure as HTTP API
	protocolType := "HTTP"
	api.ProtocolType = &protocolType

	// Set description
	api.Description = &config.Description

	// Disable custom domain features (Route53/ACM)
	createDomainName := false
	api.CreateDomainName = &createDomainName

	// Enable CORS
	corsConfig := &apigatewayv2.CORSConfiguration{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"content-type", "x-amz-date", "authorization", "x-api-key"},
	}
	maxAge := 300
	corsConfig.MaxAge = &maxAge
	api.CORSConfiguration = corsConfig

	// Create default stage with auto-deploy
	createStage := true
	api.CreateStage = &createStage

	stageName := "$default"
	api.StageName = &stageName

	autoDeploy := true
	api.AutoDeploy = &autoDeploy

	// Configure Lambda integration
	lambdaFunctionName := fmt.Sprintf("%s-%s", config.ServiceName, "${var.environment}")
	integrationURI := fmt.Sprintf("${module.%s.lambda_function_invoke_arn}", sanitizeModuleName(lambdaFunctionName))

	// Define integration
	api.Integrations = map[string]apigatewayv2.Integration{
		"lambda": {
			IntegrationType:      "AWS_PROXY",
			IntegrationURI:       &integrationURI,
			IntegrationMethod:    stringPtr("POST"),
			PayloadFormatVersion: stringPtr("2.0"),
		},
	}

	// Define route
	routeKey := fmt.Sprintf("%s %s", config.HTTPMethod, config.APIPath)
	api.Routes = map[string]apigatewayv2.Route{
		"main": {
			RouteKey:       routeKey,
			IntegrationKey: stringPtr("lambda"),
		},
	}

	// Configure access logging for default stage with throttling
	logDestination := "${aws_cloudwatch_log_group.api_gateway.arn}"
	logFormat := `{
  "requestId": "$context.requestId",
  "ip": "$context.identity.sourceIp",
  "requestTime": "$context.requestTime",
  "httpMethod": "$context.httpMethod",
  "routeKey": "$context.routeKey",
  "status": "$context.status",
  "protocol": "$context.protocol",
  "responseLength": "$context.responseLength"
}`
	// Add throttling settings to prevent API abuse
	throttleSettings := &apigatewayv2.ThrottleSettings{
		BurstLimit: intPtr(100),
		RateLimit:  float64Ptr(50.0),
	}

	api.Stages = map[string]apigatewayv2.Stage{
		"$default": {
			Name:             stringPtr("$default"),
			AutoDeploy:       &autoDeploy,
			ThrottleSettings: throttleSettings,
			AccessLogSettings: &apigatewayv2.AccessLogSettings{
				DestinationARN: logDestination,
				Format:         logFormat,
			},
		},
	}

	return api
}

// GenerateAPIGatewayModuleHCL converts API Gateway module to HCL string.
// PURE: Calculation - deterministic output from module configuration.
func GenerateAPIGatewayModuleHCL(module *apigatewayv2.Module, lambdaModuleName string) string {
	// Use a static module name (variable interpolation not allowed in module names)
	moduleName := "api_gateway"

	// Get the route key from the module's Routes map
	routeKey := "POST /"
	if len(module.Routes) > 0 {
		for _, route := range module.Routes {
			routeKey = route.RouteKey
			break // Use the first route key
		}
	}

	hcl := fmt.Sprintf(`# API Gateway HTTP API (v2) module
module "%s" {
  source  = "%s"
  version = "%s"

  name          = "%s"
  description   = "%s"
  protocol_type = "%s"

  # Disable custom domain features (Route53/ACM)
  create_domain_name = false

  cors_configuration = {
    allow_origins = %s
    allow_methods = %s
    allow_headers = %s
    max_age       = 300
  }

  routes = {
    "%s" = {
      integration = {
        uri                    = module.%s.lambda_function_invoke_arn
        payload_format_version = "2.0"
        timeout_milliseconds   = 12000
      }
    }
  }

  tags = {
    ManagedBy   = "Terraform"
    Generator   = "Forge"
    Service     = var.service_name
    Environment = var.environment
  }
}

`, moduleName, module.Source, module.Version,
		*module.Name, *module.Description, *module.ProtocolType,
		formatStringList(module.CORSConfiguration.AllowOrigins),
		formatStringList(module.CORSConfiguration.AllowMethods),
		formatStringList(module.CORSConfiguration.AllowHeaders),
		routeKey, lambdaModuleName)

	// CloudWatch Log Group for API Gateway
	hcl += fmt.Sprintf(`# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.service_name}-${var.environment}"
  retention_in_days = 7

  tags = {
    ManagedBy   = "Terraform"
    Generator   = "Forge"
    Service     = var.service_name
    Environment = var.environment
  }
}

# Lambda permission for API Gateway invocation
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = module.%s.lambda_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${module.%s.api_execution_arn}/*/*"
}

`, lambdaModuleName, moduleName)

	return hcl
}

// formatStringList formats a string slice for HCL.
// PURE: Calculation - deterministic string formatting.
func formatStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("%q", item)
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}

// stringPtr returns a pointer to a string.
// PURE: Helper function for pointer creation.
func stringPtr(s string) *string {
	return &s
}

// intPtr returns a pointer to an int.
// PURE: Helper function for pointer creation.
func intPtr(i int) *int {
	return &i
}

// float64Ptr returns a pointer to a float64.
// PURE: Helper function for pointer creation.
func float64Ptr(f float64) *float64 {
	return &f
}
