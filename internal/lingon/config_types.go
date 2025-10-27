package lingon

import (
	"time"
)

// ForgeConfig is the root configuration structure for Forge infrastructure
type ForgeConfig struct {
	// Service name for the application
	Service string `json:"service"`

	// Provider configuration
	Provider ProviderConfig `json:"provider"`

	// Functions to deploy
	Functions map[string]FunctionConfig `json:"functions"`

	// API Gateway configuration
	APIGateway *APIGatewayConfig `json:"apiGateway,omitempty"`

	// DynamoDB tables
	Tables map[string]TableConfig `json:"tables,omitempty"`

	// EventBridge rules
	EventBridge map[string]EventBridgeConfig `json:"eventBridge,omitempty"`

	// Step Functions state machines
	StateMachines map[string]StateMachineConfig `json:"stateMachines,omitempty"`

	// SNS topics
	Topics map[string]TopicConfig `json:"topics,omitempty"`

	// SQS queues
	Queues map[string]QueueConfig `json:"queues,omitempty"`

	// S3 buckets
	Buckets map[string]BucketConfig `json:"buckets,omitempty"`

	// CloudWatch alarms
	Alarms map[string]AlarmConfig `json:"alarms,omitempty"`
}

// ProviderConfig contains AWS provider configuration
type ProviderConfig struct {
	Region  string            `json:"region"`
	Profile string            `json:"profile,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
}

// FunctionConfig contains complete Lambda function configuration
// Matches ALL terraform-aws-lambda module options (170+ parameters)
type FunctionConfig struct {
	// === Core Configuration ===
	Handler     string `json:"handler"`
	Runtime     string `json:"runtime"`
	Timeout     int    `json:"timeout,omitempty"`
	MemorySize  int    `json:"memorySize,omitempty"`
	Description string `json:"description,omitempty"`

	// === Source Configuration ===
	Source SourceConfig `json:"source"`

	// === Environment Variables ===
	Environment map[string]string `json:"environment,omitempty"`

	// === VPC Configuration ===
	VPC *VPCConfig `json:"vpc,omitempty"`

	// === IAM Configuration ===
	IAM IAMConfig `json:"iam,omitempty"`

	// === CloudWatch Logs ===
	Logs CloudWatchLogsConfig `json:"logs,omitempty"`

	// === Lambda Configuration ===
	ReservedConcurrentExecutions int     `json:"reservedConcurrentExecutions,omitempty"`
	ProvisionedConcurrency       int     `json:"provisionedConcurrency,omitempty"`
	Publish                      bool    `json:"publish,omitempty"`
	Architectures                []string `json:"architectures,omitempty"` // ["x86_64"] or ["arm64"]

	// === Layers ===
	Layers []string `json:"layers,omitempty"`

	// === Dead Letter Queue ===
	DeadLetterConfig *DeadLetterConfig `json:"deadLetterConfig,omitempty"`

	// === Tracing ===
	TracingMode string `json:"tracingMode,omitempty"` // "Active" or "PassThrough"

	// === File System ===
	FileSystemConfigs []FileSystemConfig `json:"fileSystemConfigs,omitempty"`

	// === Image Configuration (for container images) ===
	ImageConfig *ImageConfig `json:"imageConfig,omitempty"`

	// === Ephemeral Storage ===
	EphemeralStorage *EphemeralStorageConfig `json:"ephemeralStorage,omitempty"`

	// === Async Configuration ===
	AsyncConfig *AsyncConfig `json:"asyncConfig,omitempty"`

	// === Code Signing ===
	CodeSigningConfigArn string `json:"codeSigningConfigArn,omitempty"`

	// === Snap Start (for Java) ===
	SnapStart *SnapStartConfig `json:"snapStart,omitempty"`

	// === Event Source Mappings ===
	EventSourceMappings []EventSourceMappingConfig `json:"eventSourceMappings,omitempty"`

	// === HTTP Routing (API Gateway integration) ===
	HTTPRouting *HTTPRoutingConfig `json:"httpRouting,omitempty"`

	// === Tags ===
	Tags map[string]string `json:"tags,omitempty"`

	// === Package Configuration ===
	Package PackageConfig `json:"package,omitempty"`

	// === KMS Key ===
	KMSKeyArn string `json:"kmsKeyArn,omitempty"`

	// === CloudWatch Alarms ===
	Alarms []string `json:"alarms,omitempty"` // References to alarm names

	// === Replacement Strategy ===
	ReplaceSecurityGroupsOnDestroy bool `json:"replaceSecurityGroupsOnDestroy,omitempty"`
	ReplacementSecurityGroupIds    []string `json:"replacementSecurityGroupIds,omitempty"`

	// === Function URL ===
	FunctionURL *FunctionURLConfig `json:"functionUrl,omitempty"`

	// === Runtime Management ===
	RuntimeManagementConfig *RuntimeManagementConfig `json:"runtimeManagementConfig,omitempty"`

	// === Logging Configuration ===
	LoggingConfig *LoggingConfig `json:"loggingConfig,omitempty"`
}

// SourceConfig defines how to build and package the Lambda function
type SourceConfig struct {
	// Path to source code
	Path string `json:"path"`

	// Docker configuration for container-based Lambdas
	Docker *DockerConfig `json:"docker,omitempty"`

	// Build commands (npm install, pip install, etc.)
	BuildCommands []string `json:"buildCommands,omitempty"`

	// Install commands for dependencies
	InstallCommands []string `json:"installCommands,omitempty"`

	// Patterns to exclude from package
	Excludes []string `json:"excludes,omitempty"`

	// Patterns to include in package
	Includes []string `json:"includes,omitempty"`

	// Python-specific: Use poetry for dependency management
	Poetry *PoetryConfig `json:"poetry,omitempty"`

	// Python-specific: Use pip for dependency management
	Pip *PipConfig `json:"pip,omitempty"`

	// Node.js-specific: Use npm for dependency management
	Npm *NpmConfig `json:"npm,omitempty"`

	// Source artifact from S3
	S3Bucket string `json:"s3Bucket,omitempty"`
	S3Key    string `json:"s3Key,omitempty"`
	S3ObjectVersion string `json:"s3ObjectVersion,omitempty"`

	// Local file path (pre-built zip)
	Filename string `json:"filename,omitempty"`
}

// DockerConfig for container-based Lambda functions
type DockerConfig struct {
	File       string            `json:"file,omitempty"`             // Dockerfile path
	BuildArgs  map[string]string `json:"buildArgs,omitempty"`   // Docker build args
	Target     string            `json:"target,omitempty"`         // Multi-stage build target
	Platform   string            `json:"platform,omitempty"`     // linux/amd64 or linux/arm64
	Repository string            `json:"repository,omitempty"` // ECR repository
	Tag        string            `json:"tag,omitempty"`               // Image tag
}

// PoetryConfig for Python Poetry dependency management
type PoetryConfig struct {
	Version         string `json:"version,omitempty"`                 // Poetry version
	InstallArgs     string `json:"installArgs,omitempty"`         // Additional install args
	WithoutDev      bool   `json:"withoutDev,omitempty"`           // Exclude dev dependencies
	WithoutHashes   bool   `json:"withoutHashes,omitempty"`     // Skip hash verification
	ExportFormat    string `json:"exportFormat,omitempty"`       // requirements.txt format
	IncludeExtras   []string `json:"includeExtras,omitempty"`   // Include extra dependencies
}

// PipConfig for Python pip dependency management
type PipConfig struct {
	RequirementsFile string `json:"requirementsFile,omitempty"` // Path to requirements.txt
	InstallArgs      string `json:"installArgs,omitempty"`           // Additional pip install args
	UpgradePip       bool   `json:"upgradePip,omitempty"`             // Upgrade pip before install
	Target           string `json:"target,omitempty"`                     // Install target directory
}

// NpmConfig for Node.js npm dependency management
type NpmConfig struct {
	PackageManager string `json:"packageManager,omitempty"` // npm, yarn, or pnpm
	InstallArgs    string `json:"installArgs,omitempty"`       // Additional install args
	BuildScript    string `json:"buildScript,omitempty"`       // npm script to run for build
	ProductionOnly bool   `json:"productionOnly,omitempty"` // Only install production deps
}

// VPCConfig for Lambda VPC configuration
type VPCConfig struct {
	SubnetIds         []string `json:"subnetIds"`
	SecurityGroupIds  []string `json:"securityGroupIds"`
	IPv6AllowedForDualStack bool `json:"ipv6AllowedForDualStack,omitempty"`
}

// IAMConfig for Lambda IAM permissions
type IAMConfig struct {
	// Role ARN (if using existing role)
	RoleArn string `json:"roleArn,omitempty"`

	// Role name (if creating new role)
	RoleName string `json:"roleName,omitempty"`

	// Assume role policy (custom trust policy)
	AssumeRolePolicy string `json:"assumeRolePolicy,omitempty"`

	// Managed policy ARNs to attach
	ManagedPolicyArns []string `json:"managedPolicyArns,omitempty"`

	// Inline policies
	InlinePolicies []InlinePolicy `json:"inlinePolicies,omitempty"`

	// Additional policy statements
	PolicyStatements []PolicyStatement `json:"policyStatements,omitempty"`

	// Permissions boundary
	PermissionsBoundary string `json:"permissionsBoundary,omitempty"`

	// Maximum session duration
	MaxSessionDuration int `json:"maxSessionDuration,omitempty"`

	// Role path
	Path string `json:"path,omitempty"`

	// Role description
	Description string `json:"description,omitempty"`

	// Force detach policies on destroy
	ForceDetachPolicies bool `json:"forceDetachPolicies,omitempty"`

	// Tags for IAM role
	Tags map[string]string `json:"tags,omitempty"`
}

// InlinePolicy represents an inline IAM policy
type InlinePolicy struct {
	Name   string `json:"name"`
	Policy string `json:"policy"` // JSON policy document
}

// PolicyStatement represents an IAM policy statement
type PolicyStatement struct {
	Effect    string   `json:"effect"`       // Allow or Deny
	Actions   []string `json:"actions"`
	Resources []string `json:"resources"`
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// CloudWatchLogsConfig for Lambda logging
type CloudWatchLogsConfig struct {
	RetentionInDays       int    `json:"retentionInDays,omitempty"`
	LogGroupName          string `json:"logGroupName,omitempty"`
	KMSKeyId              string `json:"kmsKeyId,omitempty"`
	SkipDestroy           bool   `json:"skipDestroy,omitempty"`
	LogFormat             string `json:"logFormat,omitempty"`       // JSON or Text
	ApplicationLogLevel   string `json:"applicationLogLevel,omitempty"` // TRACE, DEBUG, INFO, WARN, ERROR, FATAL
	SystemLogLevel        string `json:"systemLogLevel,omitempty"`
	LogGroupClass         string `json:"logGroupClass,omitempty"` // STANDARD or INFREQUENT_ACCESS
	Tags                  map[string]string `json:"tags,omitempty"`
}

// DeadLetterConfig for Lambda DLQ
type DeadLetterConfig struct {
	TargetArn string `json:"targetArn"` // SNS topic or SQS queue ARN
}

// FileSystemConfig for Lambda EFS integration
type FileSystemConfig struct {
	Arn            string `json:"arn"`                       // EFS access point ARN
	LocalMountPath string `json:"localMountPath"` // Mount path in Lambda (e.g., /mnt/efs)
}

// ImageConfig for container image Lambda configuration
type ImageConfig struct {
	Command          []string `json:"command,omitempty"`
	EntryPoint       []string `json:"entryPoint,omitempty"`
	WorkingDirectory string   `json:"workingDirectory,omitempty"`
}

// EphemeralStorageConfig for Lambda /tmp storage
type EphemeralStorageConfig struct {
	Size int `json:"size"` // Size in MB (512 - 10240)
}

// AsyncConfig for Lambda asynchronous invocation
type AsyncConfig struct {
	MaximumRetryAttempts int `json:"maximumRetryAttempts,omitempty"` // 0-2
	MaximumEventAgeInSeconds int `json:"maximumEventAgeInSeconds,omitempty"` // 60-21600

	// Destination configuration
	OnSuccess *DestinationConfig `json:"onSuccess,omitempty"`
	OnFailure *DestinationConfig `json:"onFailure,omitempty"`
}

// DestinationConfig for async invocation destinations
type DestinationConfig struct {
	Destination string `json:"destination"` // SNS, SQS, Lambda, or EventBridge ARN
}

// SnapStartConfig for Lambda SnapStart (Java only)
type SnapStartConfig struct {
	ApplyOn string `json:"applyOn"` // "PublishedVersions" or "None"
}

// EventSourceMappingConfig for Lambda event sources
type EventSourceMappingConfig struct {
	// Event source ARN (DynamoDB Stream, Kinesis Stream, SQS Queue, Kafka, etc.)
	EventSourceArn string `json:"eventSourceArn"`

	// Starting position (LATEST, TRIM_HORIZON, AT_TIMESTAMP)
	StartingPosition string `json:"startingPosition,omitempty"`
	StartingPositionTimestamp *time.Time `json:"startingPositionTimestamp,omitempty"`

	// Batch configuration
	BatchSize int `json:"batchSize,omitempty"`
	MaximumBatchingWindowInSeconds int `json:"maximumBatchingWindowInSeconds,omitempty"`

	// Parallelization
	ParallelizationFactor int `json:"parallelizationFactor,omitempty"` // 1-10

	// Maximum record age
	MaximumRecordAgeInSeconds int `json:"maximumRecordAgeInSeconds,omitempty"` // -1 or 60-604800

	// Retry attempts
	MaximumRetryAttempts int `json:"maximumRetryAttempts,omitempty"` // -1 or 0-10000

	// Bisect batch on error
	BisectBatchOnFunctionError bool `json:"bisectBatchOnFunctionError,omitempty"`

	// Tumbling window
	TumblingWindowInSeconds int `json:"tumblingWindowInSeconds,omitempty"`

	// Destinations
	DestinationConfig *EventSourceDestinationConfig `json:"destinationConfig,omitempty"`

	// Filter criteria
	FilterCriteria *FilterCriteria `json:"filterCriteria,omitempty"`

	// Function response types (for streams)
	FunctionResponseTypes []string `json:"functionResponseTypes,omitempty"` // ReportBatchItemFailures

	// Enabled
	Enabled bool `json:"enabled,omitempty"`

	// SQS-specific
	ScalingConfig *ScalingConfig `json:"scalingConfig,omitempty"`

	// Kafka-specific
	SourceAccessConfigurations []SourceAccessConfiguration `json:"sourceAccessConfigurations,omitempty"`
	Topics []string `json:"topics,omitempty"` // Kafka topics

	// Self-managed Kafka
	SelfManagedEventSource *SelfManagedEventSourceConfig `json:"selfManagedEventSource,omitempty"`

	// Amazon MQ
	Queues []string `json:"queues,omitempty"` // Amazon MQ queues
}

// EventSourceDestinationConfig for event source mapping destinations
type EventSourceDestinationConfig struct {
	OnFailure *DestinationConfig `json:"onFailure,omitempty"`
}

// FilterCriteria for event filtering
type FilterCriteria struct {
	Filters []FilterPattern `json:"filters"`
}

// FilterPattern represents an event filter pattern
type FilterPattern struct {
	Pattern string `json:"pattern"` // JSON filter pattern
}

// ScalingConfig for SQS event source scaling
type ScalingConfig struct {
	MaximumConcurrency int `json:"maximumConcurrency"` // 2-1000
}

// SourceAccessConfiguration for Kafka authentication
type SourceAccessConfiguration struct {
	Type string `json:"type"` // BASIC_AUTH, VPC_SUBNET, VPC_SECURITY_GROUP, etc.
	URI  string `json:"uri"`
}

// SelfManagedEventSourceConfig for self-managed Kafka
type SelfManagedEventSourceConfig struct {
	Endpoints map[string][]string `json:"endpoints"` // KAFKA_BOOTSTRAP_SERVERS
}

// HTTPRoutingConfig for API Gateway integration
type HTTPRoutingConfig struct {
	// HTTP method (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD, ANY)
	Method string `json:"method"`

	// Route path (e.g., /users/{id})
	Path string `json:"path"`

	// Authorization type (NONE, AWS_IAM, CUSTOM, JWT)
	AuthorizationType string `json:"authorizationType,omitempty"`

	// Authorizer ID (for CUSTOM authorization)
	AuthorizerId string `json:"authorizerId,omitempty"`

	// Authorization scopes (for JWT)
	AuthorizationScopes []string `json:"authorizationScopes,omitempty"`

	// CORS configuration
	CORS *CORSConfig `json:"cors,omitempty"`

	// Request validation
	RequestValidator string `json:"requestValidator,omitempty"`

	// Request parameters
	RequestParameters map[string]bool `json:"requestParameters,omitempty"` // method.request.path.id: true

	// Throttling
	ThrottlingBurstLimit int `json:"throttlingBurstLimit,omitempty"`
	ThrottlingRateLimit  float64 `json:"throttlingRateLimit,omitempty"`
}

// CORSConfig for API Gateway CORS
type CORSConfig struct {
	AllowOrigins     []string `json:"allowOrigins"`
	AllowMethods     []string `json:"allowMethods,omitempty"`
	AllowHeaders     []string `json:"allowHeaders,omitempty"`
	ExposeHeaders    []string `json:"exposeHeaders,omitempty"`
	MaxAge           int      `json:"maxAge,omitempty"`
	AllowCredentials bool     `json:"allowCredentials,omitempty"`
}

// PackageConfig for Lambda deployment package
type PackageConfig struct {
	// Patterns to include
	Patterns []string `json:"patterns,omitempty"`

	// Individually package (for multi-function services)
	Individually bool `json:"individually,omitempty"`

	// Artifact directory
	Artifact string `json:"artifact,omitempty"`

	// Exclude dev dependencies
	ExcludeDevDependencies bool `json:"excludeDevDependencies,omitempty"`
}

// FunctionURLConfig for Lambda Function URLs
type FunctionURLConfig struct {
	// Authorization type (NONE or AWS_IAM)
	AuthorizationType string `json:"authorizationType"`

	// CORS configuration
	CORS *CORSConfig `json:"cors,omitempty"`

	// Invoke mode (BUFFERED or RESPONSE_STREAM)
	InvokeMode string `json:"invokeMode,omitempty"`

	// Qualifier (version or alias)
	Qualifier string `json:"qualifier,omitempty"`
}

// RuntimeManagementConfig for Lambda runtime updates
type RuntimeManagementConfig struct {
	UpdateRuntimeOn string `json:"updateRuntimeOn"` // Auto, Manual, or FunctionUpdate
	RuntimeVersionArn string `json:"runtimeVersionArn,omitempty"`
}

// LoggingConfig for advanced Lambda logging
type LoggingConfig struct {
	LogFormat string `json:"logFormat"` // JSON or Text
	ApplicationLogLevel string `json:"applicationLogLevel,omitempty"`
	SystemLogLevel string `json:"systemLogLevel,omitempty"`
	LogGroup string `json:"logGroup,omitempty"`
}

// APIGatewayConfig contains complete API Gateway v2 configuration
// Matches ALL terraform-aws-apigateway-v2 module options (80+ parameters)
type APIGatewayConfig struct {
	// === Core Configuration ===
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProtocolType string `json:"protocolType"` // HTTP or WEBSOCKET

	// === CORS Configuration ===
	CORS *CORSConfig `json:"cors,omitempty"`

	// === Domain Configuration ===
	Domain *DomainConfig `json:"domain,omitempty"`

	// === Stage Configuration ===
	Stages map[string]StageConfig `json:"stages,omitempty"`

	// === Authorizers ===
	Authorizers map[string]AuthorizerConfig `json:"authorizers,omitempty"`

	// === Access Logs ===
	AccessLogs *AccessLogsConfig `json:"accessLogs,omitempty"`

	// === Throttling ===
	DefaultRouteSettings *RouteSettings `json:"defaultRouteSettings,omitempty"`

	// === API Key ===
	APIKeySelectionExpression string `json:"apiKeySelectionExpression,omitempty"`
	DisableExecuteApiEndpoint bool `json:"disableExecuteApiEndpoint,omitempty"`

	// === Mutual TLS ===
	MutualTLSAuthentication *MutualTLSConfig `json:"mutualTlsAuthentication,omitempty"`

	// === Route Selection ===
	RouteSelectionExpression string `json:"routeSelectionExpression,omitempty"`

	// === Tags ===
	Tags map[string]string `json:"tags,omitempty"`

	// === VPC Link (for private integrations) ===
	VPCLinks map[string]VPCLinkConfig `json:"vpcLinks,omitempty"`

	// === API Mapping ===
	APIMappingKey string `json:"apiMappingKey,omitempty"`

	// === CloudWatch Metrics ===
	MetricsEnabled bool `json:"metricsEnabled,omitempty"`

	// === Request Validation ===
	RequestValidators map[string]RequestValidatorConfig `json:"requestValidators,omitempty"`

	// === Models ===
	Models map[string]ModelConfig `json:"models,omitempty"`
}

// DomainConfig for custom domain configuration
type DomainConfig struct {
	DomainName string `json:"domainName"`
	CertificateArn string `json:"certificateArn"`
	HostedZoneId string `json:"hostedZoneId,omitempty"`
	BasePath string `json:"basePath,omitempty"`
	EndpointType string `json:"endpointType,omitempty"` // REGIONAL or EDGE
	SecurityPolicy string `json:"securityPolicy,omitempty"` // TLS_1_0, TLS_1_2
	Tags map[string]string `json:"tags,omitempty"`
}

// StageConfig for API Gateway stage
type StageConfig struct {
	Name string `json:"name"`
	Description string `json:"description,omitempty"`
	AutoDeploy bool `json:"autoDeploy,omitempty"`

	// Access logs
	AccessLogs *AccessLogsConfig `json:"accessLogs,omitempty"`

	// Default route settings
	DefaultRouteSettings *RouteSettings `json:"defaultRouteSettings,omitempty"`

	// Route settings overrides
	RouteSettings map[string]RouteSettings `json:"routeSettings,omitempty"`

	// Stage variables
	Variables map[string]string `json:"variables,omitempty"`

	// Deployment ID
	DeploymentId string `json:"deploymentId,omitempty"`

	// Client certificate
	ClientCertificateId string `json:"clientCertificateId,omitempty"`

	// Tags
	Tags map[string]string `json:"tags,omitempty"`
}

// RouteSettings for API Gateway throttling and logging
type RouteSettings struct {
	DataTraceEnabled bool `json:"dataTraceEnabled,omitempty"`
	DetailedMetricsEnabled bool `json:"detailedMetricsEnabled,omitempty"`
	LoggingLevel string `json:"loggingLevel,omitempty"` // OFF, ERROR, INFO
	ThrottlingBurstLimit int `json:"throttlingBurstLimit,omitempty"`
	ThrottlingRateLimit float64 `json:"throttlingRateLimit,omitempty"`
}

// AuthorizerConfig for API Gateway authorizers
type AuthorizerConfig struct {
	Name string `json:"name"`
	Type string `json:"type"` // JWT or REQUEST

	// JWT configuration
	JWTConfiguration *JWTConfiguration `json:"jwtConfiguration,omitempty"`

	// Lambda authorizer configuration
	AuthorizerURI string `json:"authorizerUri,omitempty"`
	AuthorizerPayloadFormatVersion string `json:"authorizerPayloadFormatVersion,omitempty"`
	AuthorizerResultTtlInSeconds int `json:"authorizerResultTtlInSeconds,omitempty"`
	IdentitySource []string `json:"identitySource,omitempty"`
	AuthorizerCredentialsArn string `json:"authorizerCredentialsArn,omitempty"`
	EnableSimpleResponses bool `json:"enableSimpleResponses,omitempty"`
}

// JWTConfiguration for JWT authorizers
type JWTConfiguration struct {
	Issuer string `json:"issuer"`
	Audience []string `json:"audience"`
}

// AccessLogsConfig for API Gateway access logging
type AccessLogsConfig struct {
	DestinationArn string `json:"destinationArn"` // CloudWatch Logs ARN
	Format string `json:"format"` // JSON format string
}

// MutualTLSConfig for mTLS authentication
type MutualTLSConfig struct {
	TruststoreUri string `json:"truststoreUri"` // S3 URI
	TruststoreVersion string `json:"truststoreVersion,omitempty"`
}

// VPCLinkConfig for private integrations
type VPCLinkConfig struct {
	Name string `json:"name"`
	SecurityGroupIds []string `json:"securityGroupIds"`
	SubnetIds []string `json:"subnetIds"`
	Tags map[string]string `json:"tags,omitempty"`
}

// RequestValidatorConfig for request validation
type RequestValidatorConfig struct {
	Name string `json:"name"`
	ValidateRequestBody bool `json:"validateRequestBody,omitempty"`
	ValidateRequestParameters bool `json:"validateRequestParameters,omitempty"`
}

// ModelConfig for API models (schemas)
type ModelConfig struct {
	Name string `json:"name"`
	ContentType string `json:"contentType"`
	Description string `json:"description,omitempty"`
	Schema string `json:"schema"` // JSON Schema
}

// TableConfig contains complete DynamoDB table configuration
// Matches ALL terraform-aws-dynamodb-table options (50+ parameters)
type TableConfig struct {
	// === Core Configuration ===
	TableName string `json:"tableName"`
	BillingMode string `json:"billingMode,omitempty"` // PROVISIONED or PAY_PER_REQUEST

	// === Primary Key ===
	HashKey string `json:"hashKey"`
	RangeKey string `json:"rangeKey,omitempty"`

	// === Attributes ===
	Attributes []AttributeDefinition `json:"attributes"`

	// === Capacity (for PROVISIONED billing) ===
	ReadCapacity int `json:"readCapacity,omitempty"`
	WriteCapacity int `json:"writeCapacity,omitempty"`

	// === Global Secondary Indexes ===
	GlobalSecondaryIndexes []GlobalSecondaryIndex `json:"globalSecondaryIndexes,omitempty"`

	// === Local Secondary Indexes ===
	LocalSecondaryIndexes []LocalSecondaryIndex `json:"localSecondaryIndexes,omitempty"`

	// === Streams ===
	StreamEnabled bool `json:"streamEnabled,omitempty"`
	StreamViewType string `json:"streamViewType,omitempty"` // KEYS_ONLY, NEW_IMAGE, OLD_IMAGE, NEW_AND_OLD_IMAGES

	// === TTL ===
	TTL *TTLConfig `json:"ttl,omitempty"`

	// === Encryption ===
	ServerSideEncryption *ServerSideEncryptionConfig `json:"serverSideEncryption,omitempty"`

	// === Point-in-Time Recovery ===
	PointInTimeRecovery *PointInTimeRecoveryConfig `json:"pointInTimeRecovery,omitempty"`

	// === Table Class ===
	TableClass string `json:"tableClass,omitempty"` // STANDARD or STANDARD_INFREQUENT_ACCESS

	// === Tags ===
	Tags map[string]string `json:"tags,omitempty"`

	// === Deletion Protection ===
	DeletionProtectionEnabled bool `json:"deletionProtectionEnabled,omitempty"`

	// === Contributor Insights ===
	ContributorInsightsEnabled bool `json:"contributorInsightsEnabled,omitempty"`

	// === Replica Configuration (Global Tables) ===
	Replicas []ReplicaConfig `json:"replicas,omitempty"`

	// === Auto Scaling ===
	AutoScaling *AutoScalingConfig `json:"autoScaling,omitempty"`

	// === Import ===
	ImportTable *ImportTableConfig `json:"importTable,omitempty"`
}

// AttributeDefinition for DynamoDB attributes
type AttributeDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"` // S, N, or B (String, Number, Binary)
}

// GlobalSecondaryIndex for DynamoDB GSI
type GlobalSecondaryIndex struct {
	Name string `json:"name"`
	HashKey string `json:"hashKey"`
	RangeKey string `json:"rangeKey,omitempty"`
	ProjectionType string `json:"projectionType"` // ALL, KEYS_ONLY, or INCLUDE
	NonKeyAttributes []string `json:"nonKeyAttributes,omitempty"`
	ReadCapacity int `json:"readCapacity,omitempty"`
	WriteCapacity int `json:"writeCapacity,omitempty"`
}

// LocalSecondaryIndex for DynamoDB LSI
type LocalSecondaryIndex struct {
	Name string `json:"name"`
	RangeKey string `json:"rangeKey"`
	ProjectionType string `json:"projectionType"`
	NonKeyAttributes []string `json:"nonKeyAttributes,omitempty"`
}

// TTLConfig for DynamoDB TTL
type TTLConfig struct {
	Enabled bool `json:"enabled"`
	AttributeName string `json:"attributeName"`
}

// ServerSideEncryptionConfig for DynamoDB encryption
type ServerSideEncryptionConfig struct {
	Enabled bool `json:"enabled"`
	KMSKeyArn string `json:"kmsKeyArn,omitempty"`
}

// PointInTimeRecoveryConfig for DynamoDB PITR
type PointInTimeRecoveryConfig struct {
	Enabled bool `json:"enabled"`
}

// ReplicaConfig for DynamoDB global tables
type ReplicaConfig struct {
	RegionName string `json:"regionName"`
	KMSKeyArn string `json:"kmsKeyArn,omitempty"`
	PointInTimeRecovery bool `json:"pointInTimeRecovery,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

// AutoScalingConfig for DynamoDB auto scaling
type AutoScalingConfig struct {
	// Read capacity auto scaling
	ReadMinCapacity int `json:"readMinCapacity,omitempty"`
	ReadMaxCapacity int `json:"readMaxCapacity,omitempty"`
	ReadTargetUtilization float64 `json:"readTargetUtilization,omitempty"`

	// Write capacity auto scaling
	WriteMinCapacity int `json:"writeMinCapacity,omitempty"`
	WriteMaxCapacity int `json:"writeMaxCapacity,omitempty"`
	WriteTargetUtilization float64 `json:"writeTargetUtilization,omitempty"`
}

// ImportTableConfig for importing data into DynamoDB
type ImportTableConfig struct {
	S3BucketSource *S3BucketSource `json:"s3BucketSource"`
	InputFormat string `json:"inputFormat"` // CSV, DYNAMODB_JSON, or ION
	InputCompressionType string `json:"inputCompressionType,omitempty"` // GZIP, ZSTD, or NONE
	InputFormatOptions *InputFormatOptions `json:"inputFormatOptions,omitempty"`
}

// S3BucketSource for table import
type S3BucketSource struct {
	Bucket string `json:"bucket"`
	KeyPrefix string `json:"keyPrefix,omitempty"`
	BucketOwner string `json:"bucketOwner,omitempty"`
}

// InputFormatOptions for CSV import
type InputFormatOptions struct {
	Delimiter string `json:"delimiter,omitempty"`
	HeaderList []string `json:"headerList,omitempty"`
}

// EventBridgeConfig for EventBridge rules
type EventBridgeConfig struct {
	Name string `json:"name"`
	Description string `json:"description,omitempty"`
	EventPattern string `json:"eventPattern,omitempty"` // JSON event pattern
	ScheduleExpression string `json:"scheduleExpression,omitempty"` // rate() or cron()
	Targets []EventBridgeTarget `json:"targets"`
	EventBusName string `json:"eventBusName,omitempty"`
	State string `json:"state,omitempty"` // ENABLED or DISABLED
	Tags map[string]string `json:"tags,omitempty"`
}

// EventBridgeTarget for EventBridge rule targets
type EventBridgeTarget struct {
	Arn string `json:"arn"` // Target ARN (Lambda, SNS, SQS, etc.)
	RoleArn string `json:"roleArn,omitempty"`
	Input string `json:"input,omitempty"` // Static JSON input
	InputPath string `json:"inputPath,omitempty"` // JSONPath expression
	InputTransformer *InputTransformer `json:"inputTransformer,omitempty"`
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
	DeadLetterConfig *DeadLetterConfig `json:"deadLetterConfig,omitempty"`
}

// InputTransformer for EventBridge input transformation
type InputTransformer struct {
	InputPathsMap map[string]string `json:"inputPathsMap,omitempty"`
	InputTemplate string `json:"inputTemplate"`
}

// RetryPolicy for EventBridge targets
type RetryPolicy struct {
	MaximumRetryAttempts int `json:"maximumRetryAttempts"`
	MaximumEventAgeInSeconds int `json:"maximumEventAgeInSeconds"`
}

// StateMachineConfig for Step Functions state machines
type StateMachineConfig struct {
	Name string `json:"name"`
	Definition string `json:"definition"` // ASL JSON
	Type string `json:"type,omitempty"` // STANDARD or EXPRESS
	RoleArn string `json:"roleArn,omitempty"`
	LoggingConfiguration *StateMachineLoggingConfig `json:"loggingConfiguration,omitempty"`
	TracingConfiguration *TracingConfiguration `json:"tracingConfiguration,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

// StateMachineLoggingConfig for Step Functions logging
type StateMachineLoggingConfig struct {
	Level string `json:"level"` // ALL, ERROR, FATAL, OFF
	IncludeExecutionData bool `json:"includeExecutionData,omitempty"`
	Destinations []LogDestination `json:"destinations"`
}

// LogDestination for Step Functions logs
type LogDestination struct {
	CloudWatchLogsLogGroup *CloudWatchLogsLogGroup `json:"cloudWatchLogsLogGroup"`
}

// CloudWatchLogsLogGroup for log destination
type CloudWatchLogsLogGroup struct {
	LogGroupArn string `json:"logGroupArn"`
}

// TracingConfiguration for Step Functions X-Ray tracing
type TracingConfiguration struct {
	Enabled bool `json:"enabled"`
}

// TopicConfig for SNS topics
type TopicConfig struct {
	Name string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	DeliveryPolicy string `json:"deliveryPolicy,omitempty"` // JSON policy
	KMSMasterKeyId string `json:"kmsMasterKeyId,omitempty"`
	FifoTopic bool `json:"fifoTopic,omitempty"`
	ContentBasedDeduplication bool `json:"contentBasedDeduplication,omitempty"`
	Subscriptions []SubscriptionConfig `json:"subscriptions,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

// SubscriptionConfig for SNS subscriptions
type SubscriptionConfig struct {
	Protocol string `json:"protocol"` // sqs, lambda, email, etc.
	Endpoint string `json:"endpoint"` // ARN or email
	FilterPolicy string `json:"filterPolicy,omitempty"` // JSON filter
	RawMessageDelivery bool `json:"rawMessageDelivery,omitempty"`
}

// QueueConfig for SQS queues
type QueueConfig struct {
	Name string `json:"name"`
	FifoQueue bool `json:"fifoQueue,omitempty"`
	ContentBasedDeduplication bool `json:"contentBasedDeduplication,omitempty"`
	DelaySeconds int `json:"delaySeconds,omitempty"`
	MaxMessageSize int `json:"maxMessageSize,omitempty"`
	MessageRetentionSeconds int `json:"messageRetentionSeconds,omitempty"`
	ReceiveWaitTimeSeconds int `json:"receiveWaitTimeSeconds,omitempty"`
	VisibilityTimeoutSeconds int `json:"visibilityTimeoutSeconds,omitempty"`
	RedrivePolicy *RedrivePolicy `json:"redrivePolicy,omitempty"`
	KMSMasterKeyId string `json:"kmsMasterKeyId,omitempty"`
	KMSDataKeyReusePeriodSeconds int `json:"kmsDataKeyReusePeriodSeconds,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

// RedrivePolicy for SQS dead letter queue
type RedrivePolicy struct {
	DeadLetterTargetArn string `json:"deadLetterTargetArn"`
	MaxReceiveCount int `json:"maxReceiveCount"`
}

// BucketConfig for S3 buckets
type BucketConfig struct {
	Name string `json:"name"`
	Versioning *VersioningConfig `json:"versioning,omitempty"`
	LifecycleRules []LifecycleRule `json:"lifecycleRules,omitempty"`
	ServerSideEncryption *S3EncryptionConfig `json:"serverSideEncryption,omitempty"`
	PublicAccessBlock *PublicAccessBlockConfig `json:"publicAccessBlock,omitempty"`
	CORSRules []S3CORSRule `json:"corsRules,omitempty"`
	Notifications []S3NotificationConfig `json:"notifications,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}

// VersioningConfig for S3 versioning
type VersioningConfig struct {
	Enabled bool `json:"enabled"`
	MFADelete bool `json:"mfaDelete,omitempty"`
}

// LifecycleRule for S3 lifecycle policies
type LifecycleRule struct {
	ID string `json:"id"`
	Enabled bool `json:"enabled"`
	Prefix string `json:"prefix,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
	Transitions []Transition `json:"transitions,omitempty"`
	Expiration *Expiration `json:"expiration,omitempty"`
}

// Transition for S3 lifecycle transitions
type Transition struct {
	Days int `json:"days"`
	StorageClass string `json:"storageClass"`
}

// Expiration for S3 lifecycle expiration
type Expiration struct {
	Days int `json:"days"`
}

// S3EncryptionConfig for S3 server-side encryption
type S3EncryptionConfig struct {
	SSEAlgorithm string `json:"sseAlgorithm"` // AES256 or aws:kms
	KMSMasterKeyID string `json:"kmsMasterKeyId,omitempty"`
}

// PublicAccessBlockConfig for S3 public access settings
type PublicAccessBlockConfig struct {
	BlockPublicAcls bool `json:"blockPublicAcls"`
	BlockPublicPolicy bool `json:"blockPublicPolicy"`
	IgnorePublicAcls bool `json:"ignorePublicAcls"`
	RestrictPublicBuckets bool `json:"restrictPublicBuckets"`
}

// S3CORSRule for S3 CORS
type S3CORSRule struct {
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
	AllowedMethods []string `json:"allowedMethods"`
	AllowedOrigins []string `json:"allowedOrigins"`
	ExposeHeaders []string `json:"exposeHeaders,omitempty"`
	MaxAgeSeconds int `json:"maxAgeSeconds,omitempty"`
}

// S3NotificationConfig for S3 event notifications
type S3NotificationConfig struct {
	Events []string `json:"events"` // s3:ObjectCreated:*, etc.
	FilterPrefix string `json:"filterPrefix,omitempty"`
	FilterSuffix string `json:"filterSuffix,omitempty"`
	LambdaFunctionArn string `json:"lambdaFunctionArn,omitempty"`
	QueueArn string `json:"queueArn,omitempty"`
	TopicArn string `json:"topicArn,omitempty"`
}

// AlarmConfig for CloudWatch alarms
type AlarmConfig struct {
	Name string `json:"name"`
	Description string `json:"description,omitempty"`
	MetricName string `json:"metricName"`
	Namespace string `json:"namespace"`
	Statistic string `json:"statistic"` // Average, Sum, etc.
	Period int `json:"period"`
	EvaluationPeriods int `json:"evaluationPeriods"`
	Threshold float64 `json:"threshold"`
	ComparisonOperator string `json:"comparisonOperator"`
	Dimensions map[string]string `json:"dimensions,omitempty"`
	AlarmActions []string `json:"alarmActions,omitempty"`
	OKActions []string `json:"okActions,omitempty"`
	InsufficientDataActions []string `json:"insufficientDataActions,omitempty"`
	TreatMissingData string `json:"treatMissingData,omitempty"`
	Tags map[string]string `json:"tags,omitempty"`
}
