package lingon

import (
	"time"
)

// ForgeConfig is the root configuration structure for Forge infrastructure
type ForgeConfig struct {
	// Service name for the application
	Service string `yaml:"service" json:"service"`

	// Provider configuration
	Provider ProviderConfig `yaml:"provider" json:"provider"`

	// Functions to deploy
	Functions map[string]FunctionConfig `yaml:"functions" json:"functions"`

	// API Gateway configuration
	APIGateway *APIGatewayConfig `yaml:"apiGateway,omitempty" json:"apiGateway,omitempty"`

	// DynamoDB tables
	Tables map[string]TableConfig `yaml:"tables,omitempty" json:"tables,omitempty"`

	// EventBridge rules
	EventBridge map[string]EventBridgeConfig `yaml:"eventBridge,omitempty" json:"eventBridge,omitempty"`

	// Step Functions state machines
	StateMachines map[string]StateMachineConfig `yaml:"stateMachines,omitempty" json:"stateMachines,omitempty"`

	// SNS topics
	Topics map[string]TopicConfig `yaml:"topics,omitempty" json:"topics,omitempty"`

	// SQS queues
	Queues map[string]QueueConfig `yaml:"queues,omitempty" json:"queues,omitempty"`

	// S3 buckets
	Buckets map[string]BucketConfig `yaml:"buckets,omitempty" json:"buckets,omitempty"`

	// CloudWatch alarms
	Alarms map[string]AlarmConfig `yaml:"alarms,omitempty" json:"alarms,omitempty"`
}

// ProviderConfig contains AWS provider configuration
type ProviderConfig struct {
	Region  string            `yaml:"region" json:"region"`
	Profile string            `yaml:"profile,omitempty" json:"profile,omitempty"`
	Tags    map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// FunctionConfig contains complete Lambda function configuration
// Matches ALL terraform-aws-lambda module options (170+ parameters)
type FunctionConfig struct {
	// === Core Configuration ===
	Handler     string `yaml:"handler" json:"handler"`
	Runtime     string `yaml:"runtime" json:"runtime"`
	Timeout     int    `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	MemorySize  int    `yaml:"memorySize,omitempty" json:"memorySize,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// === Source Configuration ===
	Source SourceConfig `yaml:"source" json:"source"`

	// === Environment Variables ===
	Environment map[string]string `yaml:"environment,omitempty" json:"environment,omitempty"`

	// === VPC Configuration ===
	VPC *VPCConfig `yaml:"vpc,omitempty" json:"vpc,omitempty"`

	// === IAM Configuration ===
	IAM IAMConfig `yaml:"iam,omitempty" json:"iam,omitempty"`

	// === CloudWatch Logs ===
	Logs CloudWatchLogsConfig `yaml:"logs,omitempty" json:"logs,omitempty"`

	// === Lambda Configuration ===
	ReservedConcurrentExecutions int     `yaml:"reservedConcurrentExecutions,omitempty" json:"reservedConcurrentExecutions,omitempty"`
	ProvisionedConcurrency       int     `yaml:"provisionedConcurrency,omitempty" json:"provisionedConcurrency,omitempty"`
	Publish                      bool    `yaml:"publish,omitempty" json:"publish,omitempty"`
	Architectures                []string `yaml:"architectures,omitempty" json:"architectures,omitempty"` // ["x86_64"] or ["arm64"]

	// === Layers ===
	Layers []string `yaml:"layers,omitempty" json:"layers,omitempty"`

	// === Dead Letter Queue ===
	DeadLetterConfig *DeadLetterConfig `yaml:"deadLetterConfig,omitempty" json:"deadLetterConfig,omitempty"`

	// === Tracing ===
	TracingMode string `yaml:"tracingMode,omitempty" json:"tracingMode,omitempty"` // "Active" or "PassThrough"

	// === File System ===
	FileSystemConfigs []FileSystemConfig `yaml:"fileSystemConfigs,omitempty" json:"fileSystemConfigs,omitempty"`

	// === Image Configuration (for container images) ===
	ImageConfig *ImageConfig `yaml:"imageConfig,omitempty" json:"imageConfig,omitempty"`

	// === Ephemeral Storage ===
	EphemeralStorage *EphemeralStorageConfig `yaml:"ephemeralStorage,omitempty" json:"ephemeralStorage,omitempty"`

	// === Async Configuration ===
	AsyncConfig *AsyncConfig `yaml:"asyncConfig,omitempty" json:"asyncConfig,omitempty"`

	// === Code Signing ===
	CodeSigningConfigArn string `yaml:"codeSigningConfigArn,omitempty" json:"codeSigningConfigArn,omitempty"`

	// === Snap Start (for Java) ===
	SnapStart *SnapStartConfig `yaml:"snapStart,omitempty" json:"snapStart,omitempty"`

	// === Event Source Mappings ===
	EventSourceMappings []EventSourceMappingConfig `yaml:"eventSourceMappings,omitempty" json:"eventSourceMappings,omitempty"`

	// === HTTP Routing (API Gateway integration) ===
	HTTPRouting *HTTPRoutingConfig `yaml:"httpRouting,omitempty" json:"httpRouting,omitempty"`

	// === Tags ===
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// === Package Configuration ===
	Package PackageConfig `yaml:"package,omitempty" json:"package,omitempty"`

	// === KMS Key ===
	KMSKeyArn string `yaml:"kmsKeyArn,omitempty" json:"kmsKeyArn,omitempty"`

	// === CloudWatch Alarms ===
	Alarms []string `yaml:"alarms,omitempty" json:"alarms,omitempty"` // References to alarm names

	// === Replacement Strategy ===
	ReplaceSecurityGroupsOnDestroy bool `yaml:"replaceSecurityGroupsOnDestroy,omitempty" json:"replaceSecurityGroupsOnDestroy,omitempty"`
	ReplacementSecurityGroupIds    []string `yaml:"replacementSecurityGroupIds,omitempty" json:"replacementSecurityGroupIds,omitempty"`

	// === Function URL ===
	FunctionURL *FunctionURLConfig `yaml:"functionUrl,omitempty" json:"functionUrl,omitempty"`

	// === Runtime Management ===
	RuntimeManagementConfig *RuntimeManagementConfig `yaml:"runtimeManagementConfig,omitempty" json:"runtimeManagementConfig,omitempty"`

	// === Logging Configuration ===
	LoggingConfig *LoggingConfig `yaml:"loggingConfig,omitempty" json:"loggingConfig,omitempty"`
}

// SourceConfig defines how to build and package the Lambda function
type SourceConfig struct {
	// Path to source code
	Path string `yaml:"path" json:"path"`

	// Docker configuration for container-based Lambdas
	Docker *DockerConfig `yaml:"docker,omitempty" json:"docker,omitempty"`

	// Build commands (npm install, pip install, etc.)
	BuildCommands []string `yaml:"buildCommands,omitempty" json:"buildCommands,omitempty"`

	// Install commands for dependencies
	InstallCommands []string `yaml:"installCommands,omitempty" json:"installCommands,omitempty"`

	// Patterns to exclude from package
	Excludes []string `yaml:"excludes,omitempty" json:"excludes,omitempty"`

	// Patterns to include in package
	Includes []string `yaml:"includes,omitempty" json:"includes,omitempty"`

	// Python-specific: Use poetry for dependency management
	Poetry *PoetryConfig `yaml:"poetry,omitempty" json:"poetry,omitempty"`

	// Python-specific: Use pip for dependency management
	Pip *PipConfig `yaml:"pip,omitempty" json:"pip,omitempty"`

	// Node.js-specific: Use npm for dependency management
	Npm *NpmConfig `yaml:"npm,omitempty" json:"npm,omitempty"`

	// Source artifact from S3
	S3Bucket string `yaml:"s3Bucket,omitempty" json:"s3Bucket,omitempty"`
	S3Key    string `yaml:"s3Key,omitempty" json:"s3Key,omitempty"`
	S3ObjectVersion string `yaml:"s3ObjectVersion,omitempty" json:"s3ObjectVersion,omitempty"`

	// Local file path (pre-built zip)
	Filename string `yaml:"filename,omitempty" json:"filename,omitempty"`
}

// DockerConfig for container-based Lambda functions
type DockerConfig struct {
	File       string            `yaml:"file,omitempty" json:"file,omitempty"`             // Dockerfile path
	BuildArgs  map[string]string `yaml:"buildArgs,omitempty" json:"buildArgs,omitempty"`   // Docker build args
	Target     string            `yaml:"target,omitempty" json:"target,omitempty"`         // Multi-stage build target
	Platform   string            `yaml:"platform,omitempty" json:"platform,omitempty"`     // linux/amd64 or linux/arm64
	Repository string            `yaml:"repository,omitempty" json:"repository,omitempty"` // ECR repository
	Tag        string            `yaml:"tag,omitempty" json:"tag,omitempty"`               // Image tag
}

// PoetryConfig for Python Poetry dependency management
type PoetryConfig struct {
	Version         string `yaml:"version,omitempty" json:"version,omitempty"`                 // Poetry version
	InstallArgs     string `yaml:"installArgs,omitempty" json:"installArgs,omitempty"`         // Additional install args
	WithoutDev      bool   `yaml:"withoutDev,omitempty" json:"withoutDev,omitempty"`           // Exclude dev dependencies
	WithoutHashes   bool   `yaml:"withoutHashes,omitempty" json:"withoutHashes,omitempty"`     // Skip hash verification
	ExportFormat    string `yaml:"exportFormat,omitempty" json:"exportFormat,omitempty"`       // requirements.txt format
	IncludeExtras   []string `yaml:"includeExtras,omitempty" json:"includeExtras,omitempty"`   // Include extra dependencies
}

// PipConfig for Python pip dependency management
type PipConfig struct {
	RequirementsFile string `yaml:"requirementsFile,omitempty" json:"requirementsFile,omitempty"` // Path to requirements.txt
	InstallArgs      string `yaml:"installArgs,omitempty" json:"installArgs,omitempty"`           // Additional pip install args
	UpgradePip       bool   `yaml:"upgradePip,omitempty" json:"upgradePip,omitempty"`             // Upgrade pip before install
	Target           string `yaml:"target,omitempty" json:"target,omitempty"`                     // Install target directory
}

// NpmConfig for Node.js npm dependency management
type NpmConfig struct {
	PackageManager string `yaml:"packageManager,omitempty" json:"packageManager,omitempty"` // npm, yarn, or pnpm
	InstallArgs    string `yaml:"installArgs,omitempty" json:"installArgs,omitempty"`       // Additional install args
	BuildScript    string `yaml:"buildScript,omitempty" json:"buildScript,omitempty"`       // npm script to run for build
	ProductionOnly bool   `yaml:"productionOnly,omitempty" json:"productionOnly,omitempty"` // Only install production deps
}

// VPCConfig for Lambda VPC configuration
type VPCConfig struct {
	SubnetIds         []string `yaml:"subnetIds" json:"subnetIds"`
	SecurityGroupIds  []string `yaml:"securityGroupIds" json:"securityGroupIds"`
	IPv6AllowedForDualStack bool `yaml:"ipv6AllowedForDualStack,omitempty" json:"ipv6AllowedForDualStack,omitempty"`
}

// IAMConfig for Lambda IAM permissions
type IAMConfig struct {
	// Role ARN (if using existing role)
	RoleArn string `yaml:"roleArn,omitempty" json:"roleArn,omitempty"`

	// Role name (if creating new role)
	RoleName string `yaml:"roleName,omitempty" json:"roleName,omitempty"`

	// Assume role policy (custom trust policy)
	AssumeRolePolicy string `yaml:"assumeRolePolicy,omitempty" json:"assumeRolePolicy,omitempty"`

	// Managed policy ARNs to attach
	ManagedPolicyArns []string `yaml:"managedPolicyArns,omitempty" json:"managedPolicyArns,omitempty"`

	// Inline policies
	InlinePolicies []InlinePolicy `yaml:"inlinePolicies,omitempty" json:"inlinePolicies,omitempty"`

	// Additional policy statements
	PolicyStatements []PolicyStatement `yaml:"policyStatements,omitempty" json:"policyStatements,omitempty"`

	// Permissions boundary
	PermissionsBoundary string `yaml:"permissionsBoundary,omitempty" json:"permissionsBoundary,omitempty"`

	// Maximum session duration
	MaxSessionDuration int `yaml:"maxSessionDuration,omitempty" json:"maxSessionDuration,omitempty"`

	// Role path
	Path string `yaml:"path,omitempty" json:"path,omitempty"`

	// Role description
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Force detach policies on destroy
	ForceDetachPolicies bool `yaml:"forceDetachPolicies,omitempty" json:"forceDetachPolicies,omitempty"`

	// Tags for IAM role
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// InlinePolicy represents an inline IAM policy
type InlinePolicy struct {
	Name   string `yaml:"name" json:"name"`
	Policy string `yaml:"policy" json:"policy"` // JSON policy document
}

// PolicyStatement represents an IAM policy statement
type PolicyStatement struct {
	Effect    string   `yaml:"effect" json:"effect"`       // Allow or Deny
	Actions   []string `yaml:"actions" json:"actions"`
	Resources []string `yaml:"resources" json:"resources"`
	Condition map[string]interface{} `yaml:"condition,omitempty" json:"condition,omitempty"`
}

// CloudWatchLogsConfig for Lambda logging
type CloudWatchLogsConfig struct {
	RetentionInDays       int    `yaml:"retentionInDays,omitempty" json:"retentionInDays,omitempty"`
	LogGroupName          string `yaml:"logGroupName,omitempty" json:"logGroupName,omitempty"`
	KMSKeyId              string `yaml:"kmsKeyId,omitempty" json:"kmsKeyId,omitempty"`
	SkipDestroy           bool   `yaml:"skipDestroy,omitempty" json:"skipDestroy,omitempty"`
	LogFormat             string `yaml:"logFormat,omitempty" json:"logFormat,omitempty"`       // JSON or Text
	ApplicationLogLevel   string `yaml:"applicationLogLevel,omitempty" json:"applicationLogLevel,omitempty"` // TRACE, DEBUG, INFO, WARN, ERROR, FATAL
	SystemLogLevel        string `yaml:"systemLogLevel,omitempty" json:"systemLogLevel,omitempty"`
	LogGroupClass         string `yaml:"logGroupClass,omitempty" json:"logGroupClass,omitempty"` // STANDARD or INFREQUENT_ACCESS
	Tags                  map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// DeadLetterConfig for Lambda DLQ
type DeadLetterConfig struct {
	TargetArn string `yaml:"targetArn" json:"targetArn"` // SNS topic or SQS queue ARN
}

// FileSystemConfig for Lambda EFS integration
type FileSystemConfig struct {
	Arn            string `yaml:"arn" json:"arn"`                       // EFS access point ARN
	LocalMountPath string `yaml:"localMountPath" json:"localMountPath"` // Mount path in Lambda (e.g., /mnt/efs)
}

// ImageConfig for container image Lambda configuration
type ImageConfig struct {
	Command          []string `yaml:"command,omitempty" json:"command,omitempty"`
	EntryPoint       []string `yaml:"entryPoint,omitempty" json:"entryPoint,omitempty"`
	WorkingDirectory string   `yaml:"workingDirectory,omitempty" json:"workingDirectory,omitempty"`
}

// EphemeralStorageConfig for Lambda /tmp storage
type EphemeralStorageConfig struct {
	Size int `yaml:"size" json:"size"` // Size in MB (512 - 10240)
}

// AsyncConfig for Lambda asynchronous invocation
type AsyncConfig struct {
	MaximumRetryAttempts int `yaml:"maximumRetryAttempts,omitempty" json:"maximumRetryAttempts,omitempty"` // 0-2
	MaximumEventAgeInSeconds int `yaml:"maximumEventAgeInSeconds,omitempty" json:"maximumEventAgeInSeconds,omitempty"` // 60-21600

	// Destination configuration
	OnSuccess *DestinationConfig `yaml:"onSuccess,omitempty" json:"onSuccess,omitempty"`
	OnFailure *DestinationConfig `yaml:"onFailure,omitempty" json:"onFailure,omitempty"`
}

// DestinationConfig for async invocation destinations
type DestinationConfig struct {
	Destination string `yaml:"destination" json:"destination"` // SNS, SQS, Lambda, or EventBridge ARN
}

// SnapStartConfig for Lambda SnapStart (Java only)
type SnapStartConfig struct {
	ApplyOn string `yaml:"applyOn" json:"applyOn"` // "PublishedVersions" or "None"
}

// EventSourceMappingConfig for Lambda event sources
type EventSourceMappingConfig struct {
	// Event source ARN (DynamoDB Stream, Kinesis Stream, SQS Queue, Kafka, etc.)
	EventSourceArn string `yaml:"eventSourceArn" json:"eventSourceArn"`

	// Starting position (LATEST, TRIM_HORIZON, AT_TIMESTAMP)
	StartingPosition string `yaml:"startingPosition,omitempty" json:"startingPosition,omitempty"`
	StartingPositionTimestamp *time.Time `yaml:"startingPositionTimestamp,omitempty" json:"startingPositionTimestamp,omitempty"`

	// Batch configuration
	BatchSize int `yaml:"batchSize,omitempty" json:"batchSize,omitempty"`
	MaximumBatchingWindowInSeconds int `yaml:"maximumBatchingWindowInSeconds,omitempty" json:"maximumBatchingWindowInSeconds,omitempty"`

	// Parallelization
	ParallelizationFactor int `yaml:"parallelizationFactor,omitempty" json:"parallelizationFactor,omitempty"` // 1-10

	// Maximum record age
	MaximumRecordAgeInSeconds int `yaml:"maximumRecordAgeInSeconds,omitempty" json:"maximumRecordAgeInSeconds,omitempty"` // -1 or 60-604800

	// Retry attempts
	MaximumRetryAttempts int `yaml:"maximumRetryAttempts,omitempty" json:"maximumRetryAttempts,omitempty"` // -1 or 0-10000

	// Bisect batch on error
	BisectBatchOnFunctionError bool `yaml:"bisectBatchOnFunctionError,omitempty" json:"bisectBatchOnFunctionError,omitempty"`

	// Tumbling window
	TumblingWindowInSeconds int `yaml:"tumblingWindowInSeconds,omitempty" json:"tumblingWindowInSeconds,omitempty"`

	// Destinations
	DestinationConfig *EventSourceDestinationConfig `yaml:"destinationConfig,omitempty" json:"destinationConfig,omitempty"`

	// Filter criteria
	FilterCriteria *FilterCriteria `yaml:"filterCriteria,omitempty" json:"filterCriteria,omitempty"`

	// Function response types (for streams)
	FunctionResponseTypes []string `yaml:"functionResponseTypes,omitempty" json:"functionResponseTypes,omitempty"` // ReportBatchItemFailures

	// Enabled
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`

	// SQS-specific
	ScalingConfig *ScalingConfig `yaml:"scalingConfig,omitempty" json:"scalingConfig,omitempty"`

	// Kafka-specific
	SourceAccessConfigurations []SourceAccessConfiguration `yaml:"sourceAccessConfigurations,omitempty" json:"sourceAccessConfigurations,omitempty"`
	Topics []string `yaml:"topics,omitempty" json:"topics,omitempty"` // Kafka topics

	// Self-managed Kafka
	SelfManagedEventSource *SelfManagedEventSourceConfig `yaml:"selfManagedEventSource,omitempty" json:"selfManagedEventSource,omitempty"`

	// Amazon MQ
	Queues []string `yaml:"queues,omitempty" json:"queues,omitempty"` // Amazon MQ queues
}

// EventSourceDestinationConfig for event source mapping destinations
type EventSourceDestinationConfig struct {
	OnFailure *DestinationConfig `yaml:"onFailure,omitempty" json:"onFailure,omitempty"`
}

// FilterCriteria for event filtering
type FilterCriteria struct {
	Filters []FilterPattern `yaml:"filters" json:"filters"`
}

// FilterPattern represents an event filter pattern
type FilterPattern struct {
	Pattern string `yaml:"pattern" json:"pattern"` // JSON filter pattern
}

// ScalingConfig for SQS event source scaling
type ScalingConfig struct {
	MaximumConcurrency int `yaml:"maximumConcurrency" json:"maximumConcurrency"` // 2-1000
}

// SourceAccessConfiguration for Kafka authentication
type SourceAccessConfiguration struct {
	Type string `yaml:"type" json:"type"` // BASIC_AUTH, VPC_SUBNET, VPC_SECURITY_GROUP, etc.
	URI  string `yaml:"uri" json:"uri"`
}

// SelfManagedEventSourceConfig for self-managed Kafka
type SelfManagedEventSourceConfig struct {
	Endpoints map[string][]string `yaml:"endpoints" json:"endpoints"` // KAFKA_BOOTSTRAP_SERVERS
}

// HTTPRoutingConfig for API Gateway integration
type HTTPRoutingConfig struct {
	// HTTP method (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, ANY)
	Method string `yaml:"method" json:"method"`

	// Route path (e.g., /users/{id})
	Path string `yaml:"path" json:"path"`

	// Authorization type (NONE, AWS_IAM, CUSTOM, JWT)
	AuthorizationType string `yaml:"authorizationType,omitempty" json:"authorizationType,omitempty"`

	// Authorizer ID (for CUSTOM authorization)
	AuthorizerId string `yaml:"authorizerId,omitempty" json:"authorizerId,omitempty"`

	// Authorization scopes (for JWT)
	AuthorizationScopes []string `yaml:"authorizationScopes,omitempty" json:"authorizationScopes,omitempty"`

	// CORS configuration
	CORS *CORSConfig `yaml:"cors,omitempty" json:"cors,omitempty"`

	// Request validation
	RequestValidator string `yaml:"requestValidator,omitempty" json:"requestValidator,omitempty"`

	// Request parameters
	RequestParameters map[string]bool `yaml:"requestParameters,omitempty" json:"requestParameters,omitempty"` // method.request.path.id: true

	// Throttling
	ThrottlingBurstLimit int `yaml:"throttlingBurstLimit,omitempty" json:"throttlingBurstLimit,omitempty"`
	ThrottlingRateLimit  float64 `yaml:"throttlingRateLimit,omitempty" json:"throttlingRateLimit,omitempty"`
}

// CORSConfig for API Gateway CORS
type CORSConfig struct {
	AllowOrigins     []string `yaml:"allowOrigins" json:"allowOrigins"`
	AllowMethods     []string `yaml:"allowMethods,omitempty" json:"allowMethods,omitempty"`
	AllowHeaders     []string `yaml:"allowHeaders,omitempty" json:"allowHeaders,omitempty"`
	ExposeHeaders    []string `yaml:"exposeHeaders,omitempty" json:"exposeHeaders,omitempty"`
	MaxAge           int      `yaml:"maxAge,omitempty" json:"maxAge,omitempty"`
	AllowCredentials bool     `yaml:"allowCredentials,omitempty" json:"allowCredentials,omitempty"`
}

// PackageConfig for Lambda deployment package
type PackageConfig struct {
	// Patterns to include
	Patterns []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`

	// Individually package (for multi-function services)
	Individually bool `yaml:"individually,omitempty" json:"individually,omitempty"`

	// Artifact directory
	Artifact string `yaml:"artifact,omitempty" json:"artifact,omitempty"`

	// Exclude dev dependencies
	ExcludeDevDependencies bool `yaml:"excludeDevDependencies,omitempty" json:"excludeDevDependencies,omitempty"`
}

// FunctionURLConfig for Lambda Function URLs
type FunctionURLConfig struct {
	// Authorization type (NONE or AWS_IAM)
	AuthorizationType string `yaml:"authorizationType" json:"authorizationType"`

	// CORS configuration
	CORS *CORSConfig `yaml:"cors,omitempty" json:"cors,omitempty"`

	// Invoke mode (BUFFERED or RESPONSE_STREAM)
	InvokeMode string `yaml:"invokeMode,omitempty" json:"invokeMode,omitempty"`

	// Qualifier (version or alias)
	Qualifier string `yaml:"qualifier,omitempty" json:"qualifier,omitempty"`
}

// RuntimeManagementConfig for Lambda runtime updates
type RuntimeManagementConfig struct {
	UpdateRuntimeOn string `yaml:"updateRuntimeOn" json:"updateRuntimeOn"` // Auto, Manual, or FunctionUpdate
	RuntimeVersionArn string `yaml:"runtimeVersionArn,omitempty" json:"runtimeVersionArn,omitempty"`
}

// LoggingConfig for advanced Lambda logging
type LoggingConfig struct {
	LogFormat string `yaml:"logFormat" json:"logFormat"` // JSON or Text
	ApplicationLogLevel string `yaml:"applicationLogLevel,omitempty" json:"applicationLogLevel,omitempty"`
	SystemLogLevel string `yaml:"systemLogLevel,omitempty" json:"systemLogLevel,omitempty"`
	LogGroup string `yaml:"logGroup,omitempty" json:"logGroup,omitempty"`
}

// APIGatewayConfig contains complete API Gateway v2 configuration
// Matches ALL terraform-aws-apigateway-v2 module options (80+ parameters)
type APIGatewayConfig struct {
	// === Core Configuration ===
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	ProtocolType string `yaml:"protocolType" json:"protocolType"` // HTTP or WEBSOCKET

	// === CORS Configuration ===
	CORS *CORSConfig `yaml:"cors,omitempty" json:"cors,omitempty"`

	// === Domain Configuration ===
	Domain *DomainConfig `yaml:"domain,omitempty" json:"domain,omitempty"`

	// === Stage Configuration ===
	Stages map[string]StageConfig `yaml:"stages,omitempty" json:"stages,omitempty"`

	// === Authorizers ===
	Authorizers map[string]AuthorizerConfig `yaml:"authorizers,omitempty" json:"authorizers,omitempty"`

	// === Access Logs ===
	AccessLogs *AccessLogsConfig `yaml:"accessLogs,omitempty" json:"accessLogs,omitempty"`

	// === Throttling ===
	DefaultRouteSettings *RouteSettings `yaml:"defaultRouteSettings,omitempty" json:"defaultRouteSettings,omitempty"`

	// === API Key ===
	APIKeySelectionExpression string `yaml:"apiKeySelectionExpression,omitempty" json:"apiKeySelectionExpression,omitempty"`
	DisableExecuteApiEndpoint bool `yaml:"disableExecuteApiEndpoint,omitempty" json:"disableExecuteApiEndpoint,omitempty"`

	// === Mutual TLS ===
	MutualTLSAuthentication *MutualTLSConfig `yaml:"mutualTlsAuthentication,omitempty" json:"mutualTlsAuthentication,omitempty"`

	// === Route Selection ===
	RouteSelectionExpression string `yaml:"routeSelectionExpression,omitempty" json:"routeSelectionExpression,omitempty"`

	// === Tags ===
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// === VPC Link (for private integrations) ===
	VPCLinks map[string]VPCLinkConfig `yaml:"vpcLinks,omitempty" json:"vpcLinks,omitempty"`

	// === API Mapping ===
	APIMappingKey string `yaml:"apiMappingKey,omitempty" json:"apiMappingKey,omitempty"`

	// === CloudWatch Metrics ===
	MetricsEnabled bool `yaml:"metricsEnabled,omitempty" json:"metricsEnabled,omitempty"`

	// === Request Validation ===
	RequestValidators map[string]RequestValidatorConfig `yaml:"requestValidators,omitempty" json:"requestValidators,omitempty"`

	// === Models ===
	Models map[string]ModelConfig `yaml:"models,omitempty" json:"models,omitempty"`
}

// DomainConfig for custom domain configuration
type DomainConfig struct {
	DomainName string `yaml:"domainName" json:"domainName"`
	CertificateArn string `yaml:"certificateArn" json:"certificateArn"`
	HostedZoneId string `yaml:"hostedZoneId,omitempty" json:"hostedZoneId,omitempty"`
	BasePath string `yaml:"basePath,omitempty" json:"basePath,omitempty"`
	EndpointType string `yaml:"endpointType,omitempty" json:"endpointType,omitempty"` // REGIONAL or EDGE
	SecurityPolicy string `yaml:"securityPolicy,omitempty" json:"securityPolicy,omitempty"` // TLS_1_0, TLS_1_2
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// StageConfig for API Gateway stage
type StageConfig struct {
	Name string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	AutoDeploy bool `yaml:"autoDeploy,omitempty" json:"autoDeploy,omitempty"`

	// Access logs
	AccessLogs *AccessLogsConfig `yaml:"accessLogs,omitempty" json:"accessLogs,omitempty"`

	// Default route settings
	DefaultRouteSettings *RouteSettings `yaml:"defaultRouteSettings,omitempty" json:"defaultRouteSettings,omitempty"`

	// Route settings overrides
	RouteSettings map[string]RouteSettings `yaml:"routeSettings,omitempty" json:"routeSettings,omitempty"`

	// Stage variables
	Variables map[string]string `yaml:"variables,omitempty" json:"variables,omitempty"`

	// Deployment ID
	DeploymentId string `yaml:"deploymentId,omitempty" json:"deploymentId,omitempty"`

	// Client certificate
	ClientCertificateId string `yaml:"clientCertificateId,omitempty" json:"clientCertificateId,omitempty"`

	// Tags
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// RouteSettings for API Gateway throttling and logging
type RouteSettings struct {
	DataTraceEnabled bool `yaml:"dataTraceEnabled,omitempty" json:"dataTraceEnabled,omitempty"`
	DetailedMetricsEnabled bool `yaml:"detailedMetricsEnabled,omitempty" json:"detailedMetricsEnabled,omitempty"`
	LoggingLevel string `yaml:"loggingLevel,omitempty" json:"loggingLevel,omitempty"` // OFF, ERROR, INFO
	ThrottlingBurstLimit int `yaml:"throttlingBurstLimit,omitempty" json:"throttlingBurstLimit,omitempty"`
	ThrottlingRateLimit float64 `yaml:"throttlingRateLimit,omitempty" json:"throttlingRateLimit,omitempty"`
}

// AuthorizerConfig for API Gateway authorizers
type AuthorizerConfig struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"` // JWT or REQUEST

	// JWT configuration
	JWTConfiguration *JWTConfiguration `yaml:"jwtConfiguration,omitempty" json:"jwtConfiguration,omitempty"`

	// Lambda authorizer configuration
	AuthorizerURI string `yaml:"authorizerUri,omitempty" json:"authorizerUri,omitempty"`
	AuthorizerPayloadFormatVersion string `yaml:"authorizerPayloadFormatVersion,omitempty" json:"authorizerPayloadFormatVersion,omitempty"`
	AuthorizerResultTtlInSeconds int `yaml:"authorizerResultTtlInSeconds,omitempty" json:"authorizerResultTtlInSeconds,omitempty"`
	IdentitySource []string `yaml:"identitySource,omitempty" json:"identitySource,omitempty"`
	AuthorizerCredentialsArn string `yaml:"authorizerCredentialsArn,omitempty" json:"authorizerCredentialsArn,omitempty"`
	EnableSimpleResponses bool `yaml:"enableSimpleResponses,omitempty" json:"enableSimpleResponses,omitempty"`
}

// JWTConfiguration for JWT authorizers
type JWTConfiguration struct {
	Issuer string `yaml:"issuer" json:"issuer"`
	Audience []string `yaml:"audience" json:"audience"`
}

// AccessLogsConfig for API Gateway access logging
type AccessLogsConfig struct {
	DestinationArn string `yaml:"destinationArn" json:"destinationArn"` // CloudWatch Logs ARN
	Format string `yaml:"format" json:"format"` // JSON format string
}

// MutualTLSConfig for mTLS authentication
type MutualTLSConfig struct {
	TruststoreUri string `yaml:"truststoreUri" json:"truststoreUri"` // S3 URI
	TruststoreVersion string `yaml:"truststoreVersion,omitempty" json:"truststoreVersion,omitempty"`
}

// VPCLinkConfig for private integrations
type VPCLinkConfig struct {
	Name string `yaml:"name" json:"name"`
	SecurityGroupIds []string `yaml:"securityGroupIds" json:"securityGroupIds"`
	SubnetIds []string `yaml:"subnetIds" json:"subnetIds"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// RequestValidatorConfig for request validation
type RequestValidatorConfig struct {
	Name string `yaml:"name" json:"name"`
	ValidateRequestBody bool `yaml:"validateRequestBody,omitempty" json:"validateRequestBody,omitempty"`
	ValidateRequestParameters bool `yaml:"validateRequestParameters,omitempty" json:"validateRequestParameters,omitempty"`
}

// ModelConfig for API models (schemas)
type ModelConfig struct {
	Name string `yaml:"name" json:"name"`
	ContentType string `yaml:"contentType" json:"contentType"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Schema string `yaml:"schema" json:"schema"` // JSON Schema
}

// TableConfig contains complete DynamoDB table configuration
// Matches ALL terraform-aws-dynamodb-table options (50+ parameters)
type TableConfig struct {
	// === Core Configuration ===
	TableName string `yaml:"tableName" json:"tableName"`
	BillingMode string `yaml:"billingMode,omitempty" json:"billingMode,omitempty"` // PROVISIONED or PAY_PER_REQUEST

	// === Primary Key ===
	HashKey string `yaml:"hashKey" json:"hashKey"`
	RangeKey string `yaml:"rangeKey,omitempty" json:"rangeKey,omitempty"`

	// === Attributes ===
	Attributes []AttributeDefinition `yaml:"attributes" json:"attributes"`

	// === Capacity (for PROVISIONED billing) ===
	ReadCapacity int `yaml:"readCapacity,omitempty" json:"readCapacity,omitempty"`
	WriteCapacity int `yaml:"writeCapacity,omitempty" json:"writeCapacity,omitempty"`

	// === Global Secondary Indexes ===
	GlobalSecondaryIndexes []GlobalSecondaryIndex `yaml:"globalSecondaryIndexes,omitempty" json:"globalSecondaryIndexes,omitempty"`

	// === Local Secondary Indexes ===
	LocalSecondaryIndexes []LocalSecondaryIndex `yaml:"localSecondaryIndexes,omitempty" json:"localSecondaryIndexes,omitempty"`

	// === Streams ===
	StreamEnabled bool `yaml:"streamEnabled,omitempty" json:"streamEnabled,omitempty"`
	StreamViewType string `yaml:"streamViewType,omitempty" json:"streamViewType,omitempty"` // KEYS_ONLY, NEW_IMAGE, OLD_IMAGE, NEW_AND_OLD_IMAGES

	// === TTL ===
	TTL *TTLConfig `yaml:"ttl,omitempty" json:"ttl,omitempty"`

	// === Encryption ===
	ServerSideEncryption *ServerSideEncryptionConfig `yaml:"serverSideEncryption,omitempty" json:"serverSideEncryption,omitempty"`

	// === Point-in-Time Recovery ===
	PointInTimeRecovery *PointInTimeRecoveryConfig `yaml:"pointInTimeRecovery,omitempty" json:"pointInTimeRecovery,omitempty"`

	// === Table Class ===
	TableClass string `yaml:"tableClass,omitempty" json:"tableClass,omitempty"` // STANDARD or STANDARD_INFREQUENT_ACCESS

	// === Tags ===
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// === Deletion Protection ===
	DeletionProtectionEnabled bool `yaml:"deletionProtectionEnabled,omitempty" json:"deletionProtectionEnabled,omitempty"`

	// === Contributor Insights ===
	ContributorInsightsEnabled bool `yaml:"contributorInsightsEnabled,omitempty" json:"contributorInsightsEnabled,omitempty"`

	// === Replica Configuration (Global Tables) ===
	Replicas []ReplicaConfig `yaml:"replicas,omitempty" json:"replicas,omitempty"`

	// === Auto Scaling ===
	AutoScaling *AutoScalingConfig `yaml:"autoScaling,omitempty" json:"autoScaling,omitempty"`

	// === Import ===
	ImportTable *ImportTableConfig `yaml:"importTable,omitempty" json:"importTable,omitempty"`
}

// AttributeDefinition for DynamoDB attributes
type AttributeDefinition struct {
	Name string `yaml:"name" json:"name"`
	Type string `yaml:"type" json:"type"` // S, N, or B (String, Number, Binary)
}

// GlobalSecondaryIndex for DynamoDB GSI
type GlobalSecondaryIndex struct {
	Name string `yaml:"name" json:"name"`
	HashKey string `yaml:"hashKey" json:"hashKey"`
	RangeKey string `yaml:"rangeKey,omitempty" json:"rangeKey,omitempty"`
	ProjectionType string `yaml:"projectionType" json:"projectionType"` // ALL, KEYS_ONLY, or INCLUDE
	NonKeyAttributes []string `yaml:"nonKeyAttributes,omitempty" json:"nonKeyAttributes,omitempty"`
	ReadCapacity int `yaml:"readCapacity,omitempty" json:"readCapacity,omitempty"`
	WriteCapacity int `yaml:"writeCapacity,omitempty" json:"writeCapacity,omitempty"`
}

// LocalSecondaryIndex for DynamoDB LSI
type LocalSecondaryIndex struct {
	Name string `yaml:"name" json:"name"`
	RangeKey string `yaml:"rangeKey" json:"rangeKey"`
	ProjectionType string `yaml:"projectionType" json:"projectionType"`
	NonKeyAttributes []string `yaml:"nonKeyAttributes,omitempty" json:"nonKeyAttributes,omitempty"`
}

// TTLConfig for DynamoDB TTL
type TTLConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	AttributeName string `yaml:"attributeName" json:"attributeName"`
}

// ServerSideEncryptionConfig for DynamoDB encryption
type ServerSideEncryptionConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	KMSKeyArn string `yaml:"kmsKeyArn,omitempty" json:"kmsKeyArn,omitempty"`
}

// PointInTimeRecoveryConfig for DynamoDB PITR
type PointInTimeRecoveryConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// ReplicaConfig for DynamoDB global tables
type ReplicaConfig struct {
	RegionName string `yaml:"regionName" json:"regionName"`
	KMSKeyArn string `yaml:"kmsKeyArn,omitempty" json:"kmsKeyArn,omitempty"`
	PointInTimeRecovery bool `yaml:"pointInTimeRecovery,omitempty" json:"pointInTimeRecovery,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// AutoScalingConfig for DynamoDB auto scaling
type AutoScalingConfig struct {
	// Read capacity auto scaling
	ReadMinCapacity int `yaml:"readMinCapacity,omitempty" json:"readMinCapacity,omitempty"`
	ReadMaxCapacity int `yaml:"readMaxCapacity,omitempty" json:"readMaxCapacity,omitempty"`
	ReadTargetUtilization float64 `yaml:"readTargetUtilization,omitempty" json:"readTargetUtilization,omitempty"`

	// Write capacity auto scaling
	WriteMinCapacity int `yaml:"writeMinCapacity,omitempty" json:"writeMinCapacity,omitempty"`
	WriteMaxCapacity int `yaml:"writeMaxCapacity,omitempty" json:"writeMaxCapacity,omitempty"`
	WriteTargetUtilization float64 `yaml:"writeTargetUtilization,omitempty" json:"writeTargetUtilization,omitempty"`
}

// ImportTableConfig for importing data into DynamoDB
type ImportTableConfig struct {
	S3BucketSource *S3BucketSource `yaml:"s3BucketSource" json:"s3BucketSource"`
	InputFormat string `yaml:"inputFormat" json:"inputFormat"` // CSV, DYNAMODB_JSON, or ION
	InputCompressionType string `yaml:"inputCompressionType,omitempty" json:"inputCompressionType,omitempty"` // GZIP, ZSTD, or NONE
	InputFormatOptions *InputFormatOptions `yaml:"inputFormatOptions,omitempty" json:"inputFormatOptions,omitempty"`
}

// S3BucketSource for table import
type S3BucketSource struct {
	Bucket string `yaml:"bucket" json:"bucket"`
	KeyPrefix string `yaml:"keyPrefix,omitempty" json:"keyPrefix,omitempty"`
	BucketOwner string `yaml:"bucketOwner,omitempty" json:"bucketOwner,omitempty"`
}

// InputFormatOptions for CSV import
type InputFormatOptions struct {
	Delimiter string `yaml:"delimiter,omitempty" json:"delimiter,omitempty"`
	HeaderList []string `yaml:"headerList,omitempty" json:"headerList,omitempty"`
}

// EventBridgeConfig for EventBridge rules
type EventBridgeConfig struct {
	Name string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	EventPattern string `yaml:"eventPattern,omitempty" json:"eventPattern,omitempty"` // JSON event pattern
	ScheduleExpression string `yaml:"scheduleExpression,omitempty" json:"scheduleExpression,omitempty"` // rate() or cron()
	Targets []EventBridgeTarget `yaml:"targets" json:"targets"`
	EventBusName string `yaml:"eventBusName,omitempty" json:"eventBusName,omitempty"`
	State string `yaml:"state,omitempty" json:"state,omitempty"` // ENABLED or DISABLED
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// EventBridgeTarget for EventBridge rule targets
type EventBridgeTarget struct {
	Arn string `yaml:"arn" json:"arn"` // Target ARN (Lambda, SNS, SQS, etc.)
	RoleArn string `yaml:"roleArn,omitempty" json:"roleArn,omitempty"`
	Input string `yaml:"input,omitempty" json:"input,omitempty"` // Static JSON input
	InputPath string `yaml:"inputPath,omitempty" json:"inputPath,omitempty"` // JSONPath expression
	InputTransformer *InputTransformer `yaml:"inputTransformer,omitempty" json:"inputTransformer,omitempty"`
	RetryPolicy *RetryPolicy `yaml:"retryPolicy,omitempty" json:"retryPolicy,omitempty"`
	DeadLetterConfig *DeadLetterConfig `yaml:"deadLetterConfig,omitempty" json:"deadLetterConfig,omitempty"`
}

// InputTransformer for EventBridge input transformation
type InputTransformer struct {
	InputPathsMap map[string]string `yaml:"inputPathsMap,omitempty" json:"inputPathsMap,omitempty"`
	InputTemplate string `yaml:"inputTemplate" json:"inputTemplate"`
}

// RetryPolicy for EventBridge targets
type RetryPolicy struct {
	MaximumRetryAttempts int `yaml:"maximumRetryAttempts" json:"maximumRetryAttempts"`
	MaximumEventAgeInSeconds int `yaml:"maximumEventAgeInSeconds" json:"maximumEventAgeInSeconds"`
}

// StateMachineConfig for Step Functions state machines
type StateMachineConfig struct {
	Name string `yaml:"name" json:"name"`
	Definition string `yaml:"definition" json:"definition"` // ASL JSON
	Type string `yaml:"type,omitempty" json:"type,omitempty"` // STANDARD or EXPRESS
	RoleArn string `yaml:"roleArn,omitempty" json:"roleArn,omitempty"`
	LoggingConfiguration *StateMachineLoggingConfig `yaml:"loggingConfiguration,omitempty" json:"loggingConfiguration,omitempty"`
	TracingConfiguration *TracingConfiguration `yaml:"tracingConfiguration,omitempty" json:"tracingConfiguration,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// StateMachineLoggingConfig for Step Functions logging
type StateMachineLoggingConfig struct {
	Level string `yaml:"level" json:"level"` // ALL, ERROR, FATAL, OFF
	IncludeExecutionData bool `yaml:"includeExecutionData,omitempty" json:"includeExecutionData,omitempty"`
	Destinations []LogDestination `yaml:"destinations" json:"destinations"`
}

// LogDestination for Step Functions logs
type LogDestination struct {
	CloudWatchLogsLogGroup *CloudWatchLogsLogGroup `yaml:"cloudWatchLogsLogGroup" json:"cloudWatchLogsLogGroup"`
}

// CloudWatchLogsLogGroup for log destination
type CloudWatchLogsLogGroup struct {
	LogGroupArn string `yaml:"logGroupArn" json:"logGroupArn"`
}

// TracingConfiguration for Step Functions X-Ray tracing
type TracingConfiguration struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// TopicConfig for SNS topics
type TopicConfig struct {
	Name string `yaml:"name" json:"name"`
	DisplayName string `yaml:"displayName,omitempty" json:"displayName,omitempty"`
	DeliveryPolicy string `yaml:"deliveryPolicy,omitempty" json:"deliveryPolicy,omitempty"` // JSON policy
	KMSMasterKeyId string `yaml:"kmsMasterKeyId,omitempty" json:"kmsMasterKeyId,omitempty"`
	FifoTopic bool `yaml:"fifoTopic,omitempty" json:"fifoTopic,omitempty"`
	ContentBasedDeduplication bool `yaml:"contentBasedDeduplication,omitempty" json:"contentBasedDeduplication,omitempty"`
	Subscriptions []SubscriptionConfig `yaml:"subscriptions,omitempty" json:"subscriptions,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// SubscriptionConfig for SNS subscriptions
type SubscriptionConfig struct {
	Protocol string `yaml:"protocol" json:"protocol"` // sqs, lambda, email, etc.
	Endpoint string `yaml:"endpoint" json:"endpoint"` // ARN or email
	FilterPolicy string `yaml:"filterPolicy,omitempty" json:"filterPolicy,omitempty"` // JSON filter
	RawMessageDelivery bool `yaml:"rawMessageDelivery,omitempty" json:"rawMessageDelivery,omitempty"`
}

// QueueConfig for SQS queues
type QueueConfig struct {
	Name string `yaml:"name" json:"name"`
	FifoQueue bool `yaml:"fifoQueue,omitempty" json:"fifoQueue,omitempty"`
	ContentBasedDeduplication bool `yaml:"contentBasedDeduplication,omitempty" json:"contentBasedDeduplication,omitempty"`
	DelaySeconds int `yaml:"delaySeconds,omitempty" json:"delaySeconds,omitempty"`
	MaxMessageSize int `yaml:"maxMessageSize,omitempty" json:"maxMessageSize,omitempty"`
	MessageRetentionSeconds int `yaml:"messageRetentionSeconds,omitempty" json:"messageRetentionSeconds,omitempty"`
	ReceiveWaitTimeSeconds int `yaml:"receiveWaitTimeSeconds,omitempty" json:"receiveWaitTimeSeconds,omitempty"`
	VisibilityTimeoutSeconds int `yaml:"visibilityTimeoutSeconds,omitempty" json:"visibilityTimeoutSeconds,omitempty"`
	RedrivePolicy *RedrivePolicy `yaml:"redrivePolicy,omitempty" json:"redrivePolicy,omitempty"`
	KMSMasterKeyId string `yaml:"kmsMasterKeyId,omitempty" json:"kmsMasterKeyId,omitempty"`
	KMSDataKeyReusePeriodSeconds int `yaml:"kmsDataKeyReusePeriodSeconds,omitempty" json:"kmsDataKeyReusePeriodSeconds,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// RedrivePolicy for SQS dead letter queue
type RedrivePolicy struct {
	DeadLetterTargetArn string `yaml:"deadLetterTargetArn" json:"deadLetterTargetArn"`
	MaxReceiveCount int `yaml:"maxReceiveCount" json:"maxReceiveCount"`
}

// BucketConfig for S3 buckets
type BucketConfig struct {
	Name string `yaml:"name" json:"name"`
	Versioning *VersioningConfig `yaml:"versioning,omitempty" json:"versioning,omitempty"`
	LifecycleRules []LifecycleRule `yaml:"lifecycleRules,omitempty" json:"lifecycleRules,omitempty"`
	ServerSideEncryption *S3EncryptionConfig `yaml:"serverSideEncryption,omitempty" json:"serverSideEncryption,omitempty"`
	PublicAccessBlock *PublicAccessBlockConfig `yaml:"publicAccessBlock,omitempty" json:"publicAccessBlock,omitempty"`
	CORSRules []S3CORSRule `yaml:"corsRules,omitempty" json:"corsRules,omitempty"`
	Notifications []S3NotificationConfig `yaml:"notifications,omitempty" json:"notifications,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}

// VersioningConfig for S3 versioning
type VersioningConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
	MFADelete bool `yaml:"mfaDelete,omitempty" json:"mfaDelete,omitempty"`
}

// LifecycleRule for S3 lifecycle policies
type LifecycleRule struct {
	ID string `yaml:"id" json:"id"`
	Enabled bool `yaml:"enabled" json:"enabled"`
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
	Transitions []Transition `yaml:"transitions,omitempty" json:"transitions,omitempty"`
	Expiration *Expiration `yaml:"expiration,omitempty" json:"expiration,omitempty"`
}

// Transition for S3 lifecycle transitions
type Transition struct {
	Days int `yaml:"days" json:"days"`
	StorageClass string `yaml:"storageClass" json:"storageClass"`
}

// Expiration for S3 lifecycle expiration
type Expiration struct {
	Days int `yaml:"days" json:"days"`
}

// S3EncryptionConfig for S3 server-side encryption
type S3EncryptionConfig struct {
	SSEAlgorithm string `yaml:"sseAlgorithm" json:"sseAlgorithm"` // AES256 or aws:kms
	KMSMasterKeyID string `yaml:"kmsMasterKeyId,omitempty" json:"kmsMasterKeyId,omitempty"`
}

// PublicAccessBlockConfig for S3 public access settings
type PublicAccessBlockConfig struct {
	BlockPublicAcls bool `yaml:"blockPublicAcls" json:"blockPublicAcls"`
	BlockPublicPolicy bool `yaml:"blockPublicPolicy" json:"blockPublicPolicy"`
	IgnorePublicAcls bool `yaml:"ignorePublicAcls" json:"ignorePublicAcls"`
	RestrictPublicBuckets bool `yaml:"restrictPublicBuckets" json:"restrictPublicBuckets"`
}

// S3CORSRule for S3 CORS
type S3CORSRule struct {
	AllowedHeaders []string `yaml:"allowedHeaders,omitempty" json:"allowedHeaders,omitempty"`
	AllowedMethods []string `yaml:"allowedMethods" json:"allowedMethods"`
	AllowedOrigins []string `yaml:"allowedOrigins" json:"allowedOrigins"`
	ExposeHeaders []string `yaml:"exposeHeaders,omitempty" json:"exposeHeaders,omitempty"`
	MaxAgeSeconds int `yaml:"maxAgeSeconds,omitempty" json:"maxAgeSeconds,omitempty"`
}

// S3NotificationConfig for S3 event notifications
type S3NotificationConfig struct {
	Events []string `yaml:"events" json:"events"` // s3:ObjectCreated:*, etc.
	FilterPrefix string `yaml:"filterPrefix,omitempty" json:"filterPrefix,omitempty"`
	FilterSuffix string `yaml:"filterSuffix,omitempty" json:"filterSuffix,omitempty"`
	LambdaFunctionArn string `yaml:"lambdaFunctionArn,omitempty" json:"lambdaFunctionArn,omitempty"`
	QueueArn string `yaml:"queueArn,omitempty" json:"queueArn,omitempty"`
	TopicArn string `yaml:"topicArn,omitempty" json:"topicArn,omitempty"`
}

// AlarmConfig for CloudWatch alarms
type AlarmConfig struct {
	Name string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	MetricName string `yaml:"metricName" json:"metricName"`
	Namespace string `yaml:"namespace" json:"namespace"`
	Statistic string `yaml:"statistic" json:"statistic"` // Average, Sum, etc.
	Period int `yaml:"period" json:"period"`
	EvaluationPeriods int `yaml:"evaluationPeriods" json:"evaluationPeriods"`
	Threshold float64 `yaml:"threshold" json:"threshold"`
	ComparisonOperator string `yaml:"comparisonOperator" json:"comparisonOperator"`
	Dimensions map[string]string `yaml:"dimensions,omitempty" json:"dimensions,omitempty"`
	AlarmActions []string `yaml:"alarmActions,omitempty" json:"alarmActions,omitempty"`
	OKActions []string `yaml:"okActions,omitempty" json:"okActions,omitempty"`
	InsufficientDataActions []string `yaml:"insufficientDataActions,omitempty" json:"insufficientDataActions,omitempty"`
	TreatMissingData string `yaml:"treatMissingData,omitempty" json:"treatMissingData,omitempty"`
	Tags map[string]string `yaml:"tags,omitempty" json:"tags,omitempty"`
}
