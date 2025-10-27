package lingon

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/golingon/lingon/pkg/terra"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_api"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_authorizer"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_domain_name"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_integration"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_route"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_stage"
	"github.com/lewis/forge/internal/lingon/aws/aws_apigatewayv2_vpc_link"
	"github.com/lewis/forge/internal/lingon/aws/aws_cloudwatch_log_group"
	"github.com/lewis/forge/internal/lingon/aws/aws_dynamodb_table"
	"github.com/lewis/forge/internal/lingon/aws/aws_iam_role"
	"github.com/lewis/forge/internal/lingon/aws/aws_lambda_function"
	"github.com/lewis/forge/internal/lingon/aws/aws_lambda_permission"
)

// This file demonstrates the Lingon resource creation pattern.
// For actual AWS resource generation, run:
// terragen -out ./aws -provider aws=hashicorp/aws:5.0.0 -force

// LambdaFunctionResources contains all Terraform resources for a Lambda function
type LambdaFunctionResources struct {
	Function    *aws_lambda_function.Resource
	Role        *aws_iam_role.Resource
	LogGroup    *aws_cloudwatch_log_group.Resource
	Permissions []terra.Resource
}

// APIGatewayResources contains all Terraform resources for API Gateway
type APIGatewayResources struct {
	API          *aws_apigatewayv2_api.Resource
	Stage        *aws_apigatewayv2_stage.Resource
	Integrations map[string]terra.Resource
	Routes       map[string]terra.Resource
	Authorizers  map[string]*aws_apigatewayv2_authorizer.Resource
	DomainName   *aws_apigatewayv2_domain_name.Resource
	VPCLinks     map[string]*aws_apigatewayv2_vpc_link.Resource
	Permissions  []terra.Resource // Lambda permissions for API Gateway invocation
}

// DynamoDBTableResources contains all Terraform resources for a DynamoDB table
type DynamoDBTableResources struct {
	Table *aws_dynamodb_table.Resource
}

// createLambdaFunctionResources creates Terraform resources for a Lambda function
// This is a pure function that transforms config into Lingon resources
func createLambdaFunctionResources(service, name string, config FunctionConfig) (*LambdaFunctionResources, error) {
	// Validate configuration
	if err := validateFunction(name, config); err != nil {
		return nil, fmt.Errorf("invalid function config: %w", err)
	}

	resources := &LambdaFunctionResources{
		Permissions: make([]terra.Resource, 0),
	}

	functionName := fmt.Sprintf("%s-%s", service, name)
	roleName := fmt.Sprintf("%s-%s-role", service, name)

	// Create IAM role for Lambda execution
	assumeRolePolicy, _ := json.Marshal(map[string]interface{}{
		"Version": "2012-10-17",
		"Statement": []map[string]interface{}{
			{
				"Effect": "Allow",
				"Principal": map[string]interface{}{
					"Service": "lambda.amazonaws.com",
				},
				"Action": "sts:AssumeRole",
			},
		},
	})

	resources.Role = &aws_iam_role.Resource{
		Name: roleName,
		Args: aws_iam_role.Args{
			Name:             terra.String(roleName),
			AssumeRolePolicy: terra.String(string(assumeRolePolicy)),
		},
	}

	// Create Lambda function
	funcArgs := aws_lambda_function.Args{
		FunctionName: terra.String(functionName),
		Runtime:      terra.String(config.Runtime),
		Handler:      terra.String(config.Handler),
		Role:         resources.Role.Attributes().Arn(),
	}

	// Add source configuration
	if config.Source.Path != "" {
		funcArgs.Filename = terra.String("lambda.zip") // Placeholder, would be actual zip path
	}
	if config.Source.S3Bucket != "" {
		funcArgs.S3Bucket = terra.String(config.Source.S3Bucket)
		funcArgs.S3Key = terra.String(config.Source.S3Key)
	}

	// Add environment variables
	if len(config.Environment) > 0 {
		funcArgs.Environment = &aws_lambda_function.Environment{
			Variables: terra.MapString(config.Environment),
		}
	}

	// Add memory, timeout, etc.
	if config.MemorySize > 0 {
		funcArgs.MemorySize = terra.Number(config.MemorySize)
	}
	if config.Timeout > 0 {
		funcArgs.Timeout = terra.Number(config.Timeout)
	}

	// Add description
	if config.Description != "" {
		funcArgs.Description = terra.String(config.Description)
	}

	// Add VPC configuration
	if config.VPC != nil && len(config.VPC.SubnetIds) > 0 {
		funcArgs.VpcConfig = &aws_lambda_function.VpcConfig{
			SubnetIds:        terra.SetString(config.VPC.SubnetIds...),
			SecurityGroupIds: terra.SetString(config.VPC.SecurityGroupIds...),
		}
	}

	// Add layers
	if len(config.Layers) > 0 {
		funcArgs.Layers = terra.ListString(config.Layers...)
	}

	// Add dead letter config
	if config.DeadLetterConfig != nil && config.DeadLetterConfig.TargetArn != "" {
		funcArgs.DeadLetterConfig = &aws_lambda_function.DeadLetterConfig{
			TargetArn: terra.String(config.DeadLetterConfig.TargetArn),
		}
	}

	// Add tracing
	if config.TracingMode != "" {
		funcArgs.TracingConfig = &aws_lambda_function.TracingConfig{
			Mode: terra.String(config.TracingMode),
		}
	}

	// Add reserved concurrent executions
	if config.ReservedConcurrentExecutions > 0 {
		funcArgs.ReservedConcurrentExecutions = terra.Number(config.ReservedConcurrentExecutions)
	}

	// Add architectures
	if len(config.Architectures) > 0 {
		funcArgs.Architectures = terra.ListString(config.Architectures...)
	}

	// Add ephemeral storage
	if config.EphemeralStorage != nil && config.EphemeralStorage.Size > 0 {
		funcArgs.EphemeralStorage = &aws_lambda_function.EphemeralStorage{
			Size: terra.Number(config.EphemeralStorage.Size),
		}
	}

	// Add KMS key
	if config.KMSKeyArn != "" {
		funcArgs.KmsKeyArn = terra.String(config.KMSKeyArn)
	}

	// Add code signing config
	if config.CodeSigningConfigArn != "" {
		funcArgs.CodeSigningConfigArn = terra.String(config.CodeSigningConfigArn)
	}

	// Add publish flag
	if config.Publish {
		funcArgs.Publish = terra.Bool(true)
	}

	// Add tags
	if len(config.Tags) > 0 {
		funcArgs.Tags = terra.MapString(config.Tags)
	}

	resources.Function = &aws_lambda_function.Resource{
		Name: functionName,
		Args: funcArgs,
	}

	// Create log group if logging is configured
	if config.Logs.RetentionInDays > 0 || config.Logs.LogGroupName != "" {
		logGroupName := config.Logs.LogGroupName
		if logGroupName == "" {
			logGroupName = fmt.Sprintf("/aws/lambda/%s-%s", service, name)
		}
		resources.LogGroup = &aws_cloudwatch_log_group.Resource{
			Name: fmt.Sprintf("%s_logs", name),
			Args: aws_cloudwatch_log_group.Args{
				Name:            terra.String(logGroupName),
				RetentionInDays: terra.Number(config.Logs.RetentionInDays),
			},
		}
	}

	return resources, nil
}

// createAPIGatewayResources creates Terraform resources for API Gateway
func createAPIGatewayResources(service string, config APIGatewayConfig, functions map[string]FunctionConfig) (*APIGatewayResources, error) {
	apiName := config.Name
	if apiName == "" {
		apiName = fmt.Sprintf("%s-api", service)
	}

	resources := &APIGatewayResources{
		Integrations: make(map[string]terra.Resource),
		Routes:       make(map[string]terra.Resource),
		Authorizers:  make(map[string]*aws_apigatewayv2_authorizer.Resource),
		VPCLinks:     make(map[string]*aws_apigatewayv2_vpc_link.Resource),
		Permissions:  make([]terra.Resource, 0),
	}

	// Create API Gateway
	apiArgs := aws_apigatewayv2_api.Args{
		Name:         terra.String(apiName),
		ProtocolType: terra.String(config.ProtocolType),
	}

	// Add description
	if config.Description != "" {
		apiArgs.Description = terra.String(config.Description)
	}

	// Add CORS configuration for HTTP APIs
	if config.CORS != nil && config.ProtocolType == "HTTP" {
		corsConfig := &aws_apigatewayv2_api.CorsConfiguration{}

		if len(config.CORS.AllowOrigins) > 0 {
			corsConfig.AllowOrigins = terra.SetString(config.CORS.AllowOrigins...)
		}
		if len(config.CORS.AllowMethods) > 0 {
			corsConfig.AllowMethods = terra.SetString(config.CORS.AllowMethods...)
		}
		if len(config.CORS.AllowHeaders) > 0 {
			corsConfig.AllowHeaders = terra.SetString(config.CORS.AllowHeaders...)
		}
		if len(config.CORS.ExposeHeaders) > 0 {
			corsConfig.ExposeHeaders = terra.SetString(config.CORS.ExposeHeaders...)
		}
		if config.CORS.AllowCredentials {
			corsConfig.AllowCredentials = terra.Bool(true)
		}
		if config.CORS.MaxAge > 0 {
			corsConfig.MaxAge = terra.Number(config.CORS.MaxAge)
		}

		apiArgs.CorsConfiguration = corsConfig
	}

	// Add disable execute API endpoint
	if config.DisableExecuteApiEndpoint {
		apiArgs.DisableExecuteApiEndpoint = terra.Bool(true)
	}

	// Add tags
	if len(config.Tags) > 0 {
		apiArgs.Tags = terra.MapString(config.Tags)
	}

	resources.API = &aws_apigatewayv2_api.Resource{
		Name: "api",
		Args: apiArgs,
	}

	// Create default stage
	stageArgs := aws_apigatewayv2_stage.Args{
		ApiId:      resources.API.Attributes().Id(),
		Name:       terra.String("$default"),
		AutoDeploy: terra.Bool(true),
	}

	// Add access logs
	if config.AccessLogs != nil && config.AccessLogs.DestinationArn != "" {
		stageArgs.AccessLogSettings = &aws_apigatewayv2_stage.AccessLogSettings{
			DestinationArn: terra.String(config.AccessLogs.DestinationArn),
			Format:         terra.String(config.AccessLogs.Format),
		}
	}

	// Add default route settings (throttling)
	if config.DefaultRouteSettings != nil {
		routeSettings := &aws_apigatewayv2_stage.DefaultRouteSettings{}

		if config.DefaultRouteSettings.ThrottlingBurstLimit > 0 {
			routeSettings.ThrottlingBurstLimit = terra.Number(config.DefaultRouteSettings.ThrottlingBurstLimit)
		}
		if config.DefaultRouteSettings.ThrottlingRateLimit > 0 {
			routeSettings.ThrottlingRateLimit = terra.Number(int(config.DefaultRouteSettings.ThrottlingRateLimit))
		}
		if config.DefaultRouteSettings.DetailedMetricsEnabled {
			routeSettings.DetailedMetricsEnabled = terra.Bool(true)
		}
		if config.DefaultRouteSettings.LoggingLevel != "" {
			routeSettings.LoggingLevel = terra.String(config.DefaultRouteSettings.LoggingLevel)
		}
		if config.DefaultRouteSettings.DataTraceEnabled {
			routeSettings.DataTraceEnabled = terra.Bool(true)
		}

		stageArgs.DefaultRouteSettings = routeSettings
	}

	// Add stage tags
	if len(config.Tags) > 0 {
		stageArgs.Tags = terra.MapString(config.Tags)
	}

	resources.Stage = &aws_apigatewayv2_stage.Resource{
		Name: "default",
		Args: stageArgs,
	}

	// Create authorizers
	for authName, authConfig := range config.Authorizers {
		authArgs := aws_apigatewayv2_authorizer.Args{
			ApiId:          resources.API.Attributes().Id(),
			AuthorizerType: terra.String(authConfig.Type),
			Name:           terra.String(authName),
		}

		// Configure based on authorizer type
		if authConfig.Type == "JWT" {
			authArgs.IdentitySources = terra.SetString(authConfig.IdentitySource...)

			if authConfig.JWTConfiguration != nil {
				jwtConfig := &aws_apigatewayv2_authorizer.JwtConfiguration{}
				if len(authConfig.JWTConfiguration.Audience) > 0 {
					jwtConfig.Audience = terra.SetString(authConfig.JWTConfiguration.Audience...)
				}
				if authConfig.JWTConfiguration.Issuer != "" {
					jwtConfig.Issuer = terra.String(authConfig.JWTConfiguration.Issuer)
				}
				authArgs.JwtConfiguration = jwtConfig
			}
		} else if authConfig.Type == "REQUEST" {
			authArgs.AuthorizerUri = terra.String(authConfig.AuthorizerURI)
			if len(authConfig.IdentitySource) > 0 {
				authArgs.IdentitySources = terra.SetString(authConfig.IdentitySource...)
			}
			if authConfig.AuthorizerResultTtlInSeconds > 0 {
				authArgs.AuthorizerResultTtlInSeconds = terra.Number(authConfig.AuthorizerResultTtlInSeconds)
			}
			if authConfig.AuthorizerPayloadFormatVersion != "" {
				authArgs.AuthorizerPayloadFormatVersion = terra.String(authConfig.AuthorizerPayloadFormatVersion)
			}
			if authConfig.EnableSimpleResponses {
				authArgs.EnableSimpleResponses = terra.Bool(true)
			}
		}

		resources.Authorizers[authName] = &aws_apigatewayv2_authorizer.Resource{
			Name: fmt.Sprintf("%s-authorizer", authName),
			Args: authArgs,
		}
	}

	// Create integrations and routes for functions with HTTP routing
	for fnName, fnConfig := range functions {
		if fnConfig.HTTPRouting != nil {
			// Create integration
			integrationName := fmt.Sprintf("%s-integration", fnName)
			resources.Integrations[integrationName] = &aws_apigatewayv2_integration.Resource{
				Name: integrationName,
				Args: aws_apigatewayv2_integration.Args{
					ApiId:           resources.API.Attributes().Id(),
					IntegrationType: terra.String("AWS_PROXY"),
					IntegrationUri:  terra.String(fmt.Sprintf("${aws_lambda_function.%s.invoke_arn}", fnName)),
				},
			}

			// Create route
			routeName := fmt.Sprintf("%s-route", fnName)
			routeKey := fmt.Sprintf("%s %s", fnConfig.HTTPRouting.Method, fnConfig.HTTPRouting.Path)
			routeArgs := aws_apigatewayv2_route.Args{
				ApiId:    resources.API.Attributes().Id(),
				RouteKey: terra.String(routeKey),
				Target:   terra.String(fmt.Sprintf("integrations/${aws_apigatewayv2_integration.%s.id}", integrationName)),
			}

			// Add authorizer if specified
			if fnConfig.HTTPRouting.AuthorizerId != "" {
				// Check if it's a reference to a named authorizer
				if auth, exists := resources.Authorizers[fnConfig.HTTPRouting.AuthorizerId]; exists {
					routeArgs.AuthorizerId = auth.Attributes().Id()
				} else {
					// Use the ID directly
					routeArgs.AuthorizerId = terra.String(fnConfig.HTTPRouting.AuthorizerId)
				}
				if fnConfig.HTTPRouting.AuthorizationType != "" {
					routeArgs.AuthorizationType = terra.String(fnConfig.HTTPRouting.AuthorizationType)
				}
			}

			resources.Routes[routeName] = &aws_apigatewayv2_route.Resource{
				Name: routeName,
				Args: routeArgs,
			}

			// Create Lambda permission for API Gateway to invoke the function
			permissionName := fmt.Sprintf("%s-api-permission", fnName)
			permission := &aws_lambda_permission.Resource{
				Name: permissionName,
				Args: aws_lambda_permission.Args{
					StatementId:  terra.String(permissionName),
					Action:       terra.String("lambda:InvokeFunction"),
					FunctionName: terra.String(fmt.Sprintf("${aws_lambda_function.%s.function_name}", fnName)),
					Principal:    terra.String("apigateway.amazonaws.com"),
					SourceArn:    terra.String(fmt.Sprintf("${aws_apigatewayv2_api.%s.execution_arn}/*/*", "api")),
				},
			}
			resources.Permissions = append(resources.Permissions, permission)
		}
	}

	return resources, nil
}

// createDynamoDBTableResources creates Terraform resources for a DynamoDB table
func createDynamoDBTableResources(service, name string, config TableConfig) (*DynamoDBTableResources, error) {
	tableName := config.TableName
	if tableName == "" {
		tableName = fmt.Sprintf("%s-%s", service, name)
	}

	// Build attributes
	attributes := make([]aws_dynamodb_table.Attribute, len(config.Attributes))
	for i, attr := range config.Attributes {
		attributes[i] = aws_dynamodb_table.Attribute{
			Name: terra.String(attr.Name),
			Type: terra.String(attr.Type),
		}
	}

	resources := &DynamoDBTableResources{
		Table: &aws_dynamodb_table.Resource{
			Name: tableName,
			Args: aws_dynamodb_table.Args{
				Name:        terra.String(tableName),
				BillingMode: terra.String(config.BillingMode),
				HashKey:     terra.String(config.HashKey),
				Attribute:   attributes,
			},
		},
	}

	return resources, nil
}

// PlaceholderResource implements terra.Resource interface for demonstration
type PlaceholderResource struct {
	resourceType string
	resourceName string
	attributes   map[string]interface{}
}

func createPlaceholderResource(resourceType, resourceName string) *PlaceholderResource {
	return &PlaceholderResource{
		resourceType: resourceType,
		resourceName: resourceName,
		attributes:   make(map[string]interface{}),
	}
}

// Type implements terra.Resource - returns the resource type (e.g. aws_iam_role)
func (r *PlaceholderResource) Type() string {
	return r.resourceType
}

// LocalName implements terra.Resource - returns the unique name in state
func (r *PlaceholderResource) LocalName() string {
	return r.resourceName
}

// Configuration implements terra.Resource - returns the resource arguments
func (r *PlaceholderResource) Configuration() interface{} {
	return r.attributes
}

// Dependencies implements terra.Resource - returns resource dependencies
func (r *PlaceholderResource) Dependencies() terra.Dependencies {
	return terra.Dependencies{}
}

// LifecycleManagement implements terra.Resource - returns lifecycle config
func (r *PlaceholderResource) LifecycleManagement() *terra.Lifecycle {
	return nil
}

// ImportState implements terra.Resource - imports state from Terraform
func (r *PlaceholderResource) ImportState(attributes io.Reader) error {
	return nil
}

// Arn returns a placeholder ARN reference
func (r *PlaceholderResource) Arn() string {
	return fmt.Sprintf("${aws_%s.%s.arn}", r.resourceType, r.resourceName)
}

// Name returns a placeholder name reference
func (r *PlaceholderResource) Name() string {
	return fmt.Sprintf("${aws_%s.%s.name}", r.resourceType, r.resourceName)
}

// ID returns a placeholder ID reference
func (r *PlaceholderResource) ID() string {
	return fmt.Sprintf("${aws_%s.%s.id}", r.resourceType, r.resourceName)
}
