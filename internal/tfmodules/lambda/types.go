// Package lambda provides type-safe Terraform module definitions for terraform-aws-modules/lambda/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-lambda v7.0
package lambda

// Module represents the terraform-aws-modules/lambda/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create controls whether resources should be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// CreateFunction controls whether Lambda Function resource should be created
	CreateFunction *bool `json:"create_function,omitempty" hcl:"create_function,attr"`

	// CreatePackage controls whether Lambda package should be created
	CreatePackage *bool `json:"create_package,omitempty" hcl:"create_package,attr"`

	// CreateRole controls whether IAM role should be created
	CreateRole *bool `json:"create_role,omitempty" hcl:"create_role,attr"`

	// CreateLayer controls whether Lambda Layer should be created
	CreateLayer *bool `json:"create_layer,omitempty" hcl:"create_layer,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// FunctionTags are tags to assign only to the lambda function
	FunctionTags map[string]string `json:"function_tags,omitempty" hcl:"function_tags,attr"`

	// ================================
	// Function Configuration
	// ================================

	// FunctionName is a unique name for your Lambda Function
	FunctionName *string `json:"function_name,omitempty" hcl:"function_name,attr"`

	// Description of your Lambda Function
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// Handler is the Lambda Function entrypoint
	Handler *string `json:"handler,omitempty" hcl:"handler,attr"`

	// Runtime is the Lambda Function runtime
	// Valid values: nodejs20.x, python3.13, java21, go1.x, dotnet8, ruby3.3, provided.al2023, etc.
	Runtime *string `json:"runtime,omitempty" hcl:"runtime,attr"`

	// LambdaRole is the IAM role ARN attached to the Lambda Function
	LambdaRole *string `json:"lambda_role,omitempty" hcl:"lambda_role,attr"`

	// MemorySize is the amount of memory in MB (128-10240)
	MemorySize *int `json:"memory_size,omitempty" validate:"min=128,max=10240" hcl:"memory_size,attr"`

	// Timeout is the execution timeout in seconds (1-900)
	Timeout *int `json:"timeout,omitempty" validate:"min=1,max=900" hcl:"timeout,attr"`

	// EphemeralStorageSize is the /tmp storage in MB (512-10240)
	EphemeralStorageSize *int `json:"ephemeral_storage_size,omitempty" validate:"min=512,max=10240" hcl:"ephemeral_storage_size,attr"`

	// ReservedConcurrentExecutions limits concurrent executions (-1 for unlimited, 0 to disable)
	ReservedConcurrentExecutions *int `json:"reserved_concurrent_executions,omitempty" hcl:"reserved_concurrent_executions,attr"`

	// Publish whether to publish creation/change as new version
	Publish *bool `json:"publish,omitempty" hcl:"publish,attr"`

	// ================================
	// Package Configuration
	// ================================

	// PackageType is the deployment package type
	// Valid values: "Zip" | "Image"
	PackageType *string `json:"package_type,omitempty" hcl:"package_type,attr"`

	// ImageURI is the ECR image URI (for Image package type)
	ImageURI *string `json:"image_uri,omitempty" hcl:"image_uri,attr"`

	// ImageConfigEntryPoint is the ENTRYPOINT for docker image
	ImageConfigEntryPoint []string `json:"image_config_entry_point,omitempty" hcl:"image_config_entry_point,attr"`

	// ImageConfigCommand is the CMD for docker image
	ImageConfigCommand []string `json:"image_config_command,omitempty" hcl:"image_config_command,attr"`

	// ImageConfigWorkingDirectory is the working directory for docker image
	ImageConfigWorkingDirectory *string `json:"image_config_working_directory,omitempty" hcl:"image_config_working_directory,attr"`

	// LocalExistingPackage is the path to an existing package file
	LocalExistingPackage *string `json:"local_existing_package,omitempty" hcl:"local_existing_package,attr"`

	// S3Bucket is the S3 bucket for Lambda deployment package
	S3Bucket *string `json:"s3_bucket,omitempty" hcl:"s3_bucket,attr"`

	// S3Key is the S3 key for Lambda deployment package
	S3Key *string `json:"s3_key,omitempty" hcl:"s3_key,attr"`

	// S3ObjectVersion is the S3 object version
	S3ObjectVersion *string `json:"s3_object_version,omitempty" hcl:"s3_object_version,attr"`

	// ================================
	// Architecture & Layers
	// ================================

	// Architectures is the instruction set architecture
	// Valid values: ["x86_64"] | ["arm64"]
	Architectures []string `json:"architectures,omitempty" hcl:"architectures,attr"`

	// Layers are Lambda Layer Version ARNs (maximum 5)
	Layers []string `json:"layers,omitempty" hcl:"layers,attr"`

	// ================================
	// Environment & Configuration
	// ================================

	// EnvironmentVariables defines environment variables
	EnvironmentVariables map[string]string `json:"environment_variables,omitempty" hcl:"environment_variables,attr"`

	// DeadLetterTargetARN is the ARN of SNS/SQS for failed invocations
	DeadLetterTargetARN *string `json:"dead_letter_target_arn,omitempty" hcl:"dead_letter_target_arn,attr"`

	// KMSKeyARN is the ARN of KMS key for encryption
	KMSKeyARN *string `json:"kms_key_arn,omitempty" hcl:"kms_key_arn,attr"`

	// CodeSigningConfigARN is the ARN for Code Signing Configuration
	CodeSigningConfigARN *string `json:"code_signing_config_arn,omitempty" hcl:"code_signing_config_arn,attr"`

	// ================================
	// VPC Configuration
	// ================================

	// VPCSubnetIDs are subnet IDs for VPC
	VPCSubnetIDs []string `json:"vpc_subnet_ids,omitempty" hcl:"vpc_subnet_ids,attr"`

	// VPCSecurityGroupIDs are security group IDs for VPC
	VPCSecurityGroupIDs []string `json:"vpc_security_group_ids,omitempty" hcl:"vpc_security_group_ids,attr"`

	// IPv6AllowedForDualStack allows outbound IPv6 traffic
	IPv6AllowedForDualStack *bool `json:"ipv6_allowed_for_dual_stack,omitempty" hcl:"ipv6_allowed_for_dual_stack,attr"`

	// ReplaceSecurityGroupsOnDestroy replaces security groups after destruction
	ReplaceSecurityGroupsOnDestroy *bool `json:"replace_security_groups_on_destroy,omitempty" hcl:"replace_security_groups_on_destroy,attr"`

	// ReplacementSecurityGroupIDs are security groups for replacement
	ReplacementSecurityGroupIDs []string `json:"replacement_security_group_ids,omitempty" hcl:"replacement_security_group_ids,attr"`

	// ================================
	// Observability
	// ================================

	// TracingMode is the X-Ray tracing mode
	// Valid values: "PassThrough" | "Active"
	TracingMode *string `json:"tracing_mode,omitempty" hcl:"tracing_mode,attr"`

	// CloudwatchLogsRetentionInDays is the log retention period
	CloudwatchLogsRetentionInDays *int `json:"cloudwatch_logs_retention_in_days,omitempty" hcl:"cloudwatch_logs_retention_in_days,attr"`

	// CloudwatchLogsKMSKeyID is the KMS key for log encryption
	CloudwatchLogsKMSKeyID *string `json:"cloudwatch_logs_kms_key_id,omitempty" hcl:"cloudwatch_logs_kms_key_id,attr"`

	// CloudwatchLogsLogGroupClass is the log class
	// Valid values: "STANDARD" | "INFREQUENT_ACCESS"
	CloudwatchLogsLogGroupClass *string `json:"cloudwatch_logs_log_group_class,omitempty" hcl:"cloudwatch_logs_log_group_class,attr"`

	// ================================
	// Lambda@Edge
	// ================================

	// LambdaAtEdge enables Lambda@Edge configuration
	LambdaAtEdge *bool `json:"lambda_at_edge,omitempty" hcl:"lambda_at_edge,attr"`

	// LambdaAtEdgeLogsAllRegions allows logging in all regions
	LambdaAtEdgeLogsAllRegions *bool `json:"lambda_at_edge_logs_all_regions,omitempty" hcl:"lambda_at_edge_logs_all_regions,attr"`

	// ================================
	// Function URL
	// ================================

	// CreateLambdaFunctionURL controls whether Function URL should be created
	CreateLambdaFunctionURL *bool `json:"create_lambda_function_url,omitempty" hcl:"create_lambda_function_url,attr"`

	// AuthorizationType is the authentication type for Function URL
	// Valid values: "AWS_IAM" | "NONE"
	AuthorizationType *string `json:"authorization_type,omitempty" hcl:"authorization_type,attr"`

	// CORS settings for Function URL
	CORS *CORSConfig `json:"cors,omitempty" hcl:"cors,attr"`

	// InvokeMode is the invoke mode for Function URL
	// Valid values: "BUFFERED" | "RESPONSE_STREAM"
	InvokeMode *string `json:"invoke_mode,omitempty" hcl:"invoke_mode,attr"`

	// ================================
	// Async Event Config
	// ================================

	// CreateAsyncEventConfig controls async event configuration
	CreateAsyncEventConfig *bool `json:"create_async_event_config,omitempty" hcl:"create_async_event_config,attr"`

	// MaximumEventAgeInSeconds is the max age for async events (60-21600)
	MaximumEventAgeInSeconds *int `json:"maximum_event_age_in_seconds,omitempty" validate:"min=60,max=21600" hcl:"maximum_event_age_in_seconds,attr"`

	// MaximumRetryAttempts is the max retry attempts (0-2)
	MaximumRetryAttempts *int `json:"maximum_retry_attempts,omitempty" validate:"min=0,max=2" hcl:"maximum_retry_attempts,attr"`

	// DestinationOnFailure is the ARN for failed invocations
	DestinationOnFailure *string `json:"destination_on_failure,omitempty" hcl:"destination_on_failure,attr"`

	// DestinationOnSuccess is the ARN for successful invocations
	DestinationOnSuccess *string `json:"destination_on_success,omitempty" hcl:"destination_on_success,attr"`

	// ================================
	// Provisioned Concurrency
	// ================================

	// ProvisionedConcurrentExecutions sets provisioned concurrency
	ProvisionedConcurrentExecutions *int `json:"provisioned_concurrent_executions,omitempty" hcl:"provisioned_concurrent_executions,attr"`

	// ================================
	// SnapStart (Java)
	// ================================

	// SnapStart enables SnapStart for low-latency startups
	SnapStart *bool `json:"snap_start,omitempty" hcl:"snap_start,attr"`

	// ================================
	// IAM Configuration
	// ================================

	// AttachPolicyJSON attaches custom JSON IAM policy
	AttachPolicyJSON *bool `json:"attach_policy_json,omitempty" hcl:"attach_policy_json,attr"`

	// PolicyJSON is the custom IAM policy JSON
	PolicyJSON *string `json:"policy_json,omitempty" hcl:"policy_json,attr"`

	// AttachPolicyStatements attaches IAM policy statements
	AttachPolicyStatements *bool `json:"attach_policy_statements,omitempty" hcl:"attach_policy_statements,attr"`

	// PolicyStatements is a map of IAM policy statements
	PolicyStatements map[string]PolicyStatement `json:"policy_statements,omitempty" hcl:"policy_statements,attr"`

	// AttachCloudwatchLogsPolicy attaches CloudWatch Logs policy
	AttachCloudwatchLogsPolicy *bool `json:"attach_cloudwatch_logs_policy,omitempty" hcl:"attach_cloudwatch_logs_policy,attr"`

	// AttachDeadLetterPolicy attaches dead letter queue policy
	AttachDeadLetterPolicy *bool `json:"attach_dead_letter_policy,omitempty" hcl:"attach_dead_letter_policy,attr"`

	// AttachNetworkPolicy attaches VPC network policy
	AttachNetworkPolicy *bool `json:"attach_network_policy,omitempty" hcl:"attach_network_policy,attr"`

	// AttachTracingPolicy attaches X-Ray tracing policy
	AttachTracingPolicy *bool `json:"attach_tracing_policy,omitempty" hcl:"attach_tracing_policy,attr"`

	// ================================
	// Event Source Mappings
	// ================================

	// EventSourceMapping configures event source mappings
	EventSourceMapping map[string]EventSourceMapping `json:"event_source_mapping,omitempty" hcl:"event_source_mapping,attr"`

	// AllowedTriggers configures allowed triggers
	AllowedTriggers map[string]AllowedTrigger `json:"allowed_triggers,omitempty" hcl:"allowed_triggers,attr"`

	// ================================
	// Lambda Layer
	// ================================

	// LayerName is the name of Lambda Layer
	LayerName *string `json:"layer_name,omitempty" hcl:"layer_name,attr"`

	// LicenseInfo is the license for Lambda Layer
	LicenseInfo *string `json:"license_info,omitempty" hcl:"license_info,attr"`

	// CompatibleRuntimes are runtimes compatible with layer (max 5)
	CompatibleRuntimes []string `json:"compatible_runtimes,omitempty" hcl:"compatible_runtimes,attr"`

	// CompatibleArchitectures are architectures compatible with layer
	CompatibleArchitectures []string `json:"compatible_architectures,omitempty" hcl:"compatible_architectures,attr"`

	// ================================
	// Lifecycle
	// ================================

	// Timeouts for Terraform resource management
	Timeouts map[string]string `json:"timeouts,omitempty" hcl:"timeouts,attr"`

	// SkipDestroy prevents deletion at destroy time
	SkipDestroy *bool `json:"skip_destroy,omitempty" hcl:"skip_destroy,attr"`
}

// CORSConfig represents CORS configuration for Function URL.
type CORSConfig struct {
	// AllowCredentials allows credentials
	AllowCredentials *bool `json:"allow_credentials,omitempty" hcl:"allow_credentials,attr"`

	// AllowHeaders are allowed headers
	AllowHeaders []string `json:"allow_headers,omitempty" hcl:"allow_headers,attr"`

	// AllowMethods are allowed methods
	AllowMethods []string `json:"allow_methods,omitempty" hcl:"allow_methods,attr"`

	// AllowOrigins are allowed origins
	AllowOrigins []string `json:"allow_origins,omitempty" hcl:"allow_origins,attr"`

	// ExposeHeaders are exposed headers
	ExposeHeaders []string `json:"expose_headers,omitempty" hcl:"expose_headers,attr"`

	// MaxAge is the cache duration in seconds
	MaxAge *int `json:"max_age,omitempty" hcl:"max_age,attr"`
}

// PolicyStatement represents an IAM policy statement.
type PolicyStatement struct {
	// Effect is Allow or Deny
	Effect *string `json:"effect,omitempty" hcl:"effect,attr"`

	// Actions are the allowed actions
	Actions []string `json:"actions,omitempty" hcl:"actions,attr"`

	// Resources are the resources this statement applies to
	Resources []string `json:"resources,omitempty" hcl:"resources,attr"`
}

// EventSourceMapping represents an event source mapping configuration.
type EventSourceMapping struct {
	// EventSourceARN is the ARN of the event source
	EventSourceARN string `json:"event_source_arn" hcl:"event_source_arn,attr"`

	// BatchSize is the maximum batch size (1-10000)
	BatchSize *int `json:"batch_size,omitempty" hcl:"batch_size,attr"`

	// MaximumBatchingWindowInSeconds is the max batching window (0-300)
	MaximumBatchingWindowInSeconds *int `json:"maximum_batching_window_in_seconds,omitempty" hcl:"maximum_batching_window_in_seconds,attr"`

	// StartingPosition is the position in stream
	// Valid values: "TRIM_HORIZON" | "LATEST"
	StartingPosition *string `json:"starting_position,omitempty" hcl:"starting_position,attr"`

	// Enabled indicates if mapping is enabled
	Enabled *bool `json:"enabled,omitempty" hcl:"enabled,attr"`

	// FilterCriteria defines event filtering
	FilterCriteria map[string]interface{} `json:"filter_criteria,omitempty" hcl:"filter_criteria,attr"`
}

// AllowedTrigger represents an allowed trigger configuration.
type AllowedTrigger struct {
	// Service is the AWS service
	Service string `json:"service" hcl:"service,attr"`

	// SourceARN is the source ARN
	SourceARN *string `json:"source_arn,omitempty" hcl:"source_arn,attr"`

	// Principal is the service principal
	Principal *string `json:"principal,omitempty" hcl:"principal,attr"`
}

// NewModule creates a new Lambda module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/lambda/aws"
	version := "~> 7.0"
	create := true
	createFunction := true
	createPackage := true
	createRole := true
	memory := 128
	timeout := 3
	ephemeral := 512
	packageType := "Zip"
	attachLogs := true

	return &Module{
		Source:                     source,
		Version:                    version,
		FunctionName:               &name,
		Create:                     &create,
		CreateFunction:             &createFunction,
		CreatePackage:              &createPackage,
		CreateRole:                 &createRole,
		MemorySize:                 &memory,
		Timeout:                    &timeout,
		EphemeralStorageSize:       &ephemeral,
		PackageType:                &packageType,
		AttachCloudwatchLogsPolicy: &attachLogs,
		Timeouts: map[string]string{
			"create": "10m",
			"update": "10m",
			"delete": "10m",
		},
	}
}

// WithRuntime sets the runtime and handler.
func (m *Module) WithRuntime(runtime, handler string) *Module {
	m.Runtime = &runtime
	m.Handler = &handler
	return m
}

// WithMemoryAndTimeout configures memory and timeout.
func (m *Module) WithMemoryAndTimeout(memoryMB, timeoutSeconds int) *Module {
	m.MemorySize = &memoryMB
	m.Timeout = &timeoutSeconds
	return m
}

// WithVPC configures VPC settings.
func (m *Module) WithVPC(subnetIDs, securityGroupIDs []string) *Module {
	m.VPCSubnetIDs = subnetIDs
	m.VPCSecurityGroupIDs = securityGroupIDs
	attachNetwork := true
	m.AttachNetworkPolicy = &attachNetwork
	return m
}

// WithEnvironment sets environment variables.
func (m *Module) WithEnvironment(envVars map[string]string) *Module {
	if m.EnvironmentVariables == nil {
		m.EnvironmentVariables = make(map[string]string)
	}
	for k, v := range envVars {
		m.EnvironmentVariables[k] = v
	}
	return m
}

// WithTracing enables X-Ray tracing.
func (m *Module) WithTracing(mode string) *Module {
	m.TracingMode = &mode
	attachTracing := true
	m.AttachTracingPolicy = &attachTracing
	return m
}

// WithLayers adds Lambda layers.
func (m *Module) WithLayers(layerARNs ...string) *Module {
	m.Layers = append(m.Layers, layerARNs...)
	return m
}

// WithFunctionURL enables Function URL.
func (m *Module) WithFunctionURL(authType string, cors *CORSConfig) *Module {
	createURL := true
	m.CreateLambdaFunctionURL = &createURL
	m.AuthorizationType = &authType
	m.CORS = cors
	return m
}

// WithDeadLetterQueue configures DLQ.
func (m *Module) WithDeadLetterQueue(targetARN string) *Module {
	m.DeadLetterTargetARN = &targetARN
	attachDLQ := true
	m.AttachDeadLetterPolicy = &attachDLQ
	return m
}

// WithEventSourceMapping adds an event source mapping.
func (m *Module) WithEventSourceMapping(name string, mapping EventSourceMapping) *Module {
	if m.EventSourceMapping == nil {
		m.EventSourceMapping = make(map[string]EventSourceMapping)
	}
	m.EventSourceMapping[name] = mapping
	return m
}

// WithTags adds tags to the function.
func (m *Module) WithTags(tags map[string]string) *Module {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	for k, v := range tags {
		m.Tags[k] = v
	}
	return m
}

// LocalName returns the local identifier for this module instance.
func (m *Module) LocalName() string {
	if m.FunctionName != nil {
		return *m.FunctionName
	}
	return "lambda_function"
}

// Configuration generates the HCL configuration for this module.
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
