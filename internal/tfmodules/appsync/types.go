// Package appsync provides type-safe Terraform module definitions for terraform-aws-modules/appsync/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-appsync v2.0
package appsync

import "strconv"

// Module represents the terraform-aws-modules/appsync/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// CreateGraphQLAPI controls whether GraphQL API should be created
	CreateGraphQLAPI *bool `json:"create_graphql_api,omitempty" hcl:"create_graphql_api,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to add to all GraphQL resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// GraphQLAPITags are additional tags for the GraphQL API
	GraphQLAPITags map[string]string `json:"graphql_api_tags,omitempty" hcl:"graphql_api_tags,attr"`

	// ================================
	// GraphQL API Configuration
	// ================================

	// Name of the GraphQL API
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Schema is the GraphQL schema definition in GraphQL schema language format
	Schema *string `json:"schema,omitempty" hcl:"schema,attr"`

	// Visibility of the API
	// Valid values: "GLOBAL" | "PRIVATE"
	Visibility *string `json:"visibility,omitempty" hcl:"visibility,attr"`

	// AuthenticationType is the primary authentication type
	// Valid values: "API_KEY" | "AWS_IAM" | "AMAZON_COGNITO_USER_POOLS" | "OPENID_CONNECT" | "AWS_LAMBDA"
	AuthenticationType *string `json:"authentication_type,omitempty" hcl:"authentication_type,attr"`

	// AdditionalAuthenticationProvider are additional auth providers
	AdditionalAuthenticationProvider map[string]AuthenticationProvider `json:"additional_authentication_provider,omitempty" hcl:"additional_authentication_provider,attr"`

	// IntrospectionConfig controls API introspection
	// Valid values: "ENABLED" | "DISABLED"
	IntrospectionConfig *string `json:"introspection_config,omitempty" hcl:"introspection_config,attr"`

	// QueryDepthLimit is the maximum query depth in a single request
	QueryDepthLimit *int `json:"query_depth_limit,omitempty" hcl:"query_depth_limit,attr"`

	// ResolverCountLimit is the maximum number of resolvers per request
	ResolverCountLimit *int `json:"resolver_count_limit,omitempty" hcl:"resolver_count_limit,attr"`

	// ================================
	// Authentication Configuration
	// ================================

	// LambdaAuthorizerConfig for Lambda authorizer
	LambdaAuthorizerConfig map[string]string `json:"lambda_authorizer_config,omitempty" hcl:"lambda_authorizer_config,attr"`

	// OpenIDConnectConfig for OpenID Connect authentication
	OpenIDConnectConfig map[string]string `json:"openid_connect_config,omitempty" hcl:"openid_connect_config,attr"`

	// UserPoolConfig for Amazon Cognito User Pool authentication
	UserPoolConfig map[string]string `json:"user_pool_config,omitempty" hcl:"user_pool_config,attr"`

	// ================================
	// Logging Configuration
	// ================================

	// LoggingEnabled controls CloudWatch logging
	LoggingEnabled *bool `json:"logging_enabled,omitempty" hcl:"logging_enabled,attr"`

	// CreateLogsRole controls whether to create IAM role for logs
	CreateLogsRole *bool `json:"create_logs_role,omitempty" hcl:"create_logs_role,attr"`

	// LogsRoleName is the name of the CloudWatch logs IAM role
	LogsRoleName *string `json:"logs_role_name,omitempty" hcl:"logs_role_name,attr"`

	// LogsRoleDescription describes the logs IAM role
	LogsRoleDescription *string `json:"logs_role_description,omitempty" hcl:"logs_role_description,attr"`

	// LogsRoleTags are tags for the logs IAM role
	LogsRoleTags map[string]string `json:"logs_role_tags,omitempty" hcl:"logs_role_tags,attr"`

	// LogCloudwatchLogsRoleARN is the service role ARN for CloudWatch logs
	LogCloudwatchLogsRoleARN *string `json:"log_cloudwatch_logs_role_arn,omitempty" hcl:"log_cloudwatch_logs_role_arn,attr"`

	// LogFieldLogLevel is the field logging level
	// Valid values: "ALL" | "ERROR" | "NONE"
	LogFieldLogLevel *string `json:"log_field_log_level,omitempty" hcl:"log_field_log_level,attr"`

	// LogExcludeVerboseContent excludes verbose content from logs
	LogExcludeVerboseContent *bool `json:"log_exclude_verbose_content,omitempty" hcl:"log_exclude_verbose_content,attr"`

	// ================================
	// Monitoring & Tracing
	// ================================

	// XRayEnabled enables AWS X-Ray tracing
	XRayEnabled *bool `json:"xray_enabled,omitempty" hcl:"xray_enabled,attr"`

	// EnhancedMetricsConfig for Lambda enhanced metrics
	EnhancedMetricsConfig map[string]string `json:"enhanced_metrics_config,omitempty" hcl:"enhanced_metrics_config,attr"`

	// ================================
	// Caching Configuration
	// ================================

	// CachingEnabled enables Elasticache caching
	CachingEnabled *bool `json:"caching_enabled,omitempty" hcl:"caching_enabled,attr"`

	// CachingBehavior controls caching strategy
	// Valid values: "FULL_REQUEST_CACHING" | "PER_RESOLVER_CACHING"
	CachingBehavior *string `json:"caching_behavior,omitempty" hcl:"caching_behavior,attr"`

	// CacheType is the cache instance type
	// Valid values: "SMALL" | "MEDIUM" | "LARGE" | "XLARGE" | "LARGE_2X" | "LARGE_4X" | "LARGE_8X" | "LARGE_12X" |
	//               "T2_SMALL" | "T2_MEDIUM" | "R4_LARGE" | "R4_XLARGE" | "R4_2XLARGE" | "R4_4XLARGE" | "R4_8XLARGE"
	CacheType *string `json:"cache_type,omitempty" hcl:"cache_type,attr"`

	// CacheTTL is the TTL in seconds for cache entries
	CacheTTL *int `json:"cache_ttl,omitempty" hcl:"cache_ttl,attr"`

	// CacheAtRestEncryptionEnabled enables at-rest encryption for cache
	CacheAtRestEncryptionEnabled *bool `json:"cache_at_rest_encryption_enabled,omitempty" hcl:"cache_at_rest_encryption_enabled,attr"`

	// CacheTransitEncryptionEnabled enables transit encryption for cache
	CacheTransitEncryptionEnabled *bool `json:"cache_transit_encryption_enabled,omitempty" hcl:"cache_transit_encryption_enabled,attr"`

	// ResolverCachingTTL is the default caching TTL for resolvers
	ResolverCachingTTL *int `json:"resolver_caching_ttl,omitempty" hcl:"resolver_caching_ttl,attr"`

	// ================================
	// Domain Name Association
	// ================================

	// DomainNameAssociationEnabled enables domain name association
	DomainNameAssociationEnabled *bool `json:"domain_name_association_enabled,omitempty" hcl:"domain_name_association_enabled,attr"`

	// DomainName that AppSync gets associated with
	DomainName *string `json:"domain_name,omitempty" hcl:"domain_name,attr"`

	// DomainNameDescription describes the domain name
	DomainNameDescription *string `json:"domain_name_description,omitempty" hcl:"domain_name_description,attr"`

	// CertificateARN is the ACM certificate ARN for the domain
	CertificateARN *string `json:"certificate_arn,omitempty" hcl:"certificate_arn,attr"`

	// ================================
	// API Keys
	// ================================

	// APIKeys is a map of API keys to create
	APIKeys map[string]string `json:"api_keys,omitempty" hcl:"api_keys,attr"`

	// ================================
	// DataSources
	// ================================

	// DataSources is a map of datasources to create
	DataSources map[string]DataSource `json:"datasources,omitempty" hcl:"datasources,attr"`

	// ================================
	// Resolvers
	// ================================

	// Resolvers is a map of resolvers to create
	Resolvers map[string]Resolver `json:"resolvers,omitempty" hcl:"resolvers,attr"`

	// ================================
	// Functions (Pipeline Resolvers)
	// ================================

	// Functions is a map of pipeline functions to create
	Functions map[string]Function `json:"functions,omitempty" hcl:"functions,attr"`

	// ================================
	// VTL Templates
	// ================================

	// DirectLambdaRequestTemplate is the VTL template for direct Lambda integrations
	DirectLambdaRequestTemplate *string `json:"direct_lambda_request_template,omitempty" hcl:"direct_lambda_request_template,attr"`

	// DirectLambdaResponseTemplate is the VTL response template for direct Lambda
	DirectLambdaResponseTemplate *string `json:"direct_lambda_response_template,omitempty" hcl:"direct_lambda_response_template,attr"`

	// ================================
	// IAM Permissions
	// ================================

	// IAMPermissionsBoundary is the ARN for IAM permissions boundary
	IAMPermissionsBoundary *string `json:"iam_permissions_boundary,omitempty" hcl:"iam_permissions_boundary,attr"`

	// LambdaAllowedActions for datasources type AWS_LAMBDA
	LambdaAllowedActions []string `json:"lambda_allowed_actions,omitempty" hcl:"lambda_allowed_actions,attr"`

	// DynamoDBAllowedActions for datasources type AMAZON_DYNAMODB
	DynamoDBAllowedActions []string `json:"dynamodb_allowed_actions,omitempty" hcl:"dynamodb_allowed_actions,attr"`

	// ElasticsearchAllowedActions for datasources type AMAZON_ELASTICSEARCH
	ElasticsearchAllowedActions []string `json:"elasticsearch_allowed_actions,omitempty" hcl:"elasticsearch_allowed_actions,attr"`

	// OpenSearchServiceAllowedActions for datasources type AMAZON_OPENSEARCH_SERVICE
	OpenSearchServiceAllowedActions []string `json:"opensearchservice_allowed_actions,omitempty" hcl:"opensearchservice_allowed_actions,attr"`

	// EventBridgeAllowedActions for datasources type AMAZON_EVENTBRIDGE
	EventBridgeAllowedActions []string `json:"eventbridge_allowed_actions,omitempty" hcl:"eventbridge_allowed_actions,attr"`

	// RelationalDatabaseAllowedActions for datasources type RELATIONAL_DATABASE
	RelationalDatabaseAllowedActions []string `json:"relational_database_allowed_actions,omitempty" hcl:"relational_database_allowed_actions,attr"`

	// SecretsManagerAllowedActions for secrets manager datasources
	SecretsManagerAllowedActions []string `json:"secrets_manager_allowed_actions,omitempty" hcl:"secrets_manager_allowed_actions,attr"`
}

// AuthenticationProvider represents an additional authentication provider
type AuthenticationProvider struct {
	// AuthenticationType is the auth type
	// Valid values: "API_KEY" | "AWS_IAM" | "AMAZON_COGNITO_USER_POOLS" | "OPENID_CONNECT" | "AWS_LAMBDA"
	AuthenticationType string `json:"authentication_type" hcl:"authentication_type,attr"`

	// LambdaAuthorizerConfig for Lambda authorizer
	LambdaAuthorizerConfig map[string]string `json:"lambda_authorizer_config,omitempty" hcl:"lambda_authorizer_config,attr"`

	// OpenIDConnectConfig for OpenID Connect
	OpenIDConnectConfig map[string]string `json:"openid_connect_config,omitempty" hcl:"openid_connect_config,attr"`

	// UserPoolConfig for Cognito User Pool
	UserPoolConfig map[string]string `json:"user_pool_config,omitempty" hcl:"user_pool_config,attr"`
}

// DataSource represents an AppSync data source
type DataSource struct {
	// Type is the data source type
	// Valid values: "AWS_LAMBDA" | "AMAZON_DYNAMODB" | "AMAZON_ELASTICSEARCH" | "AMAZON_OPENSEARCH_SERVICE" |
	//               "AMAZON_EVENTBRIDGE" | "HTTP" | "RELATIONAL_DATABASE" | "NONE"
	Type string `json:"type" hcl:"type,attr"`

	// Description of the data source
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// ServiceRoleARN for accessing the resource
	ServiceRoleARN *string `json:"service_role_arn,omitempty" hcl:"service_role_arn,attr"`

	// LambdaConfig for AWS_LAMBDA type
	LambdaConfig *LambdaConfig `json:"lambda_config,omitempty" hcl:"lambda_config,attr"`

	// DynamoDBConfig for AMAZON_DYNAMODB type
	DynamoDBConfig *DynamoDBConfig `json:"dynamodb_config,omitempty" hcl:"dynamodb_config,attr"`

	// ElasticsearchConfig for AMAZON_ELASTICSEARCH type
	ElasticsearchConfig *ElasticsearchConfig `json:"elasticsearch_config,omitempty" hcl:"elasticsearch_config,attr"`

	// OpenSearchServiceConfig for AMAZON_OPENSEARCH_SERVICE type
	OpenSearchServiceConfig *OpenSearchServiceConfig `json:"opensearch_service_config,omitempty" hcl:"opensearch_service_config,attr"`

	// HTTPConfig for HTTP type
	HTTPConfig *HTTPConfig `json:"http_config,omitempty" hcl:"http_config,attr"`

	// RelationalDatabaseConfig for RELATIONAL_DATABASE type
	RelationalDatabaseConfig *RelationalDatabaseConfig `json:"relational_database_config,omitempty" hcl:"relational_database_config,attr"`

	// EventBridgeConfig for AMAZON_EVENTBRIDGE type
	EventBridgeConfig *EventBridgeConfig `json:"eventbridge_config,omitempty" hcl:"eventbridge_config,attr"`
}

// LambdaConfig represents Lambda data source configuration
type LambdaConfig struct {
	FunctionARN string `json:"function_arn" hcl:"function_arn,attr"`
}

// DynamoDBConfig represents DynamoDB data source configuration
type DynamoDBConfig struct {
	TableName             string  `json:"table_name" hcl:"table_name,attr"`
	Region                *string `json:"region,omitempty" hcl:"region,attr"`
	UseCallerCredentials  *bool   `json:"use_caller_credentials,omitempty" hcl:"use_caller_credentials,attr"`
	Versioned             *bool   `json:"versioned,omitempty" hcl:"versioned,attr"`
	DeltaSyncConfig       *DeltaSyncConfig `json:"delta_sync_config,omitempty" hcl:"delta_sync_config,attr"`
}

// DeltaSyncConfig for DynamoDB delta sync
type DeltaSyncConfig struct {
	DeltaSyncTableName       string  `json:"delta_sync_table_name" hcl:"delta_sync_table_name,attr"`
	DeltaSyncTableTTL        *int    `json:"delta_sync_table_ttl,omitempty" hcl:"delta_sync_table_ttl,attr"`
	BaseTableTTL             *int    `json:"base_table_ttl,omitempty" hcl:"base_table_ttl,attr"`
}

// ElasticsearchConfig represents Elasticsearch data source configuration
type ElasticsearchConfig struct {
	Endpoint string  `json:"endpoint" hcl:"endpoint,attr"`
	Region   *string `json:"region,omitempty" hcl:"region,attr"`
}

// OpenSearchServiceConfig represents OpenSearch data source configuration
type OpenSearchServiceConfig struct {
	Endpoint string  `json:"endpoint" hcl:"endpoint,attr"`
	Region   *string `json:"region,omitempty" hcl:"region,attr"`
}

// HTTPConfig represents HTTP data source configuration
type HTTPConfig struct {
	Endpoint           string          `json:"endpoint" hcl:"endpoint,attr"`
	AuthorizationConfig *AuthorizationConfig `json:"authorization_config,omitempty" hcl:"authorization_config,attr"`
}

// AuthorizationConfig for HTTP endpoints
type AuthorizationConfig struct {
	AuthorizationType string            `json:"authorization_type" hcl:"authorization_type,attr"`
	AWSIAMConfig      *AWSIAMConfig     `json:"aws_iam_config,omitempty" hcl:"aws_iam_config,attr"`
}

// AWSIAMConfig for HTTP authorization
type AWSIAMConfig struct {
	SigningRegion      string  `json:"signing_region" hcl:"signing_region,attr"`
	SigningServiceName string  `json:"signing_service_name" hcl:"signing_service_name,attr"`
}

// RelationalDatabaseConfig represents RDS data source configuration
type RelationalDatabaseConfig struct {
	HTTPEndpointConfig   *HTTPEndpointConfig   `json:"http_endpoint_config,omitempty" hcl:"http_endpoint_config,attr"`
	SourceType           *string               `json:"source_type,omitempty" hcl:"source_type,attr"`
}

// HTTPEndpointConfig for RDS Data API
type HTTPEndpointConfig struct {
	DBClusterIdentifier string  `json:"db_cluster_identifier" hcl:"db_cluster_identifier,attr"`
	AWSSecretStoreARN   string  `json:"aws_secret_store_arn" hcl:"aws_secret_store_arn,attr"`
	DatabaseName        *string `json:"database_name,omitempty" hcl:"database_name,attr"`
	Region              *string `json:"region,omitempty" hcl:"region,attr"`
	Schema              *string `json:"schema,omitempty" hcl:"schema,attr"`
}

// EventBridgeConfig represents EventBridge data source configuration
type EventBridgeConfig struct {
	EventBusARN string `json:"event_bus_arn" hcl:"event_bus_arn,attr"`
}

// Resolver represents a GraphQL resolver
type Resolver struct {
	// Type is the GraphQL type (Query, Mutation, Subscription, or custom type)
	Type string `json:"type" hcl:"type,attr"`

	// Field is the field name on the type
	Field string `json:"field" hcl:"field,attr"`

	// DataSource is the data source name
	DataSource *string `json:"data_source,omitempty" hcl:"data_source,attr"`

	// RequestTemplate is the VTL request mapping template
	RequestTemplate *string `json:"request_template,omitempty" hcl:"request_template,attr"`

	// ResponseTemplate is the VTL response mapping template
	ResponseTemplate *string `json:"response_template,omitempty" hcl:"response_template,attr"`

	// Kind is the resolver kind
	// Valid values: "UNIT" | "PIPELINE"
	Kind *string `json:"kind,omitempty" hcl:"kind,attr"`

	// PipelineConfig for pipeline resolvers
	PipelineConfig *PipelineConfig `json:"pipeline_config,omitempty" hcl:"pipeline_config,attr"`

	// CachingConfig for per-resolver caching
	CachingConfig *CachingConfig `json:"caching_config,omitempty" hcl:"caching_config,attr"`

	// Code is the resolver code (for JS resolvers)
	Code *string `json:"code,omitempty" hcl:"code,attr"`

	// Runtime for JS resolvers
	Runtime *Runtime `json:"runtime,omitempty" hcl:"runtime,attr"`

	// MaxBatchSize for batch resolvers
	MaxBatchSize *int `json:"max_batch_size,omitempty" hcl:"max_batch_size,attr"`
}

// PipelineConfig for pipeline resolvers
type PipelineConfig struct {
	Functions []string `json:"functions" hcl:"functions,attr"`
}

// CachingConfig for resolver caching
type CachingConfig struct {
	TTL            *int     `json:"ttl,omitempty" hcl:"ttl,attr"`
	CachingKeys    []string `json:"caching_keys,omitempty" hcl:"caching_keys,attr"`
}

// Runtime for JavaScript resolvers
type Runtime struct {
	Name            string `json:"name" hcl:"name,attr"`
	RuntimeVersion  string `json:"runtime_version" hcl:"runtime_version,attr"`
}

// Function represents a pipeline function
type Function struct {
	// DataSource is the data source name
	DataSource string `json:"data_source" hcl:"data_source,attr"`

	// Description of the function
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// RequestTemplate is the VTL request mapping template
	RequestTemplate *string `json:"request_template,omitempty" hcl:"request_template,attr"`

	// ResponseTemplate is the VTL response mapping template
	ResponseTemplate *string `json:"response_template,omitempty" hcl:"response_template,attr"`

	// Code is the function code (for JS functions)
	Code *string `json:"code,omitempty" hcl:"code,attr"`

	// Runtime for JS functions
	Runtime *Runtime `json:"runtime,omitempty" hcl:"runtime,attr"`

	// MaxBatchSize for batch functions
	MaxBatchSize *int `json:"max_batch_size,omitempty" hcl:"max_batch_size,attr"`
}

// NewModule creates a new AppSync module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/appsync/aws"
	version := "~> 2.0"
	create := true
	authType := "API_KEY"
	createLogsRole := true
	cachingBehavior := "FULL_REQUEST_CACHING"
	cacheType := "SMALL"

	return &Module{
		Source:             source,
		Version:            version,
		Name:               &name,
		CreateGraphQLAPI:   &create,
		AuthenticationType: &authType,
		CreateLogsRole:     &createLogsRole,
		CachingBehavior:    &cachingBehavior,
		CacheType:          &cacheType,
	}
}

// WithSchema sets the GraphQL schema
func (m *Module) WithSchema(schema string) *Module {
	m.Schema = &schema
	return m
}

// WithCognitoAuth configures Cognito User Pool authentication
func (m *Module) WithCognitoAuth(userPoolID, awsRegion string) *Module {
	authType := "AMAZON_COGNITO_USER_POOLS"
	m.AuthenticationType = &authType
	m.UserPoolConfig = map[string]string{
		"user_pool_id":   userPoolID,
		"aws_region":     awsRegion,
		"default_action": "ALLOW",
	}
	return m
}

// WithIAMAuth configures AWS IAM authentication
func (m *Module) WithIAMAuth() *Module {
	authType := "AWS_IAM"
	m.AuthenticationType = &authType
	return m
}

// WithLambdaAuth configures Lambda authorizer
func (m *Module) WithLambdaAuth(authorizerURI string, ttl int) *Module {
	authType := "AWS_LAMBDA"
	m.AuthenticationType = &authType
	m.LambdaAuthorizerConfig = map[string]string{
		"authorizer_uri":                   authorizerURI,
		"authorizer_result_ttl_in_seconds": strconv.Itoa(ttl),
	}
	return m
}

// WithLogging enables CloudWatch logging
func (m *Module) WithLogging(logLevel string, excludeVerbose bool) *Module {
	enabled := true
	m.LoggingEnabled = &enabled
	m.LogFieldLogLevel = &logLevel
	m.LogExcludeVerboseContent = &excludeVerbose
	return m
}

// WithXRayTracing enables X-Ray tracing
func (m *Module) WithXRayTracing() *Module {
	enabled := true
	m.XRayEnabled = &enabled
	return m
}

// WithCaching enables API caching
func (m *Module) WithCaching(cacheType string, ttl int, atRestEncryption, transitEncryption bool) *Module {
	enabled := true
	m.CachingEnabled = &enabled
	m.CacheType = &cacheType
	m.CacheTTL = &ttl
	m.CacheAtRestEncryptionEnabled = &atRestEncryption
	m.CacheTransitEncryptionEnabled = &transitEncryption
	return m
}

// WithDomainName associates a custom domain
func (m *Module) WithDomainName(domainName, certificateARN string) *Module {
	enabled := true
	m.DomainNameAssociationEnabled = &enabled
	m.DomainName = &domainName
	m.CertificateARN = &certificateARN
	return m
}

// WithAPIKey adds an API key
func (m *Module) WithAPIKey(name, description string) *Module {
	if m.APIKeys == nil {
		m.APIKeys = make(map[string]string)
	}
	m.APIKeys[name] = description
	return m
}

// WithDataSource adds a data source
func (m *Module) WithDataSource(name string, ds DataSource) *Module {
	if m.DataSources == nil {
		m.DataSources = make(map[string]DataSource)
	}
	m.DataSources[name] = ds
	return m
}

// WithLambdaDataSource adds a Lambda data source
func (m *Module) WithLambdaDataSource(name, functionARN string) *Module {
	ds := DataSource{
		Type: "AWS_LAMBDA",
		LambdaConfig: &LambdaConfig{
			FunctionARN: functionARN,
		},
	}
	return m.WithDataSource(name, ds)
}

// WithDynamoDBDataSource adds a DynamoDB data source
func (m *Module) WithDynamoDBDataSource(name, tableName string) *Module {
	ds := DataSource{
		Type: "AMAZON_DYNAMODB",
		DynamoDBConfig: &DynamoDBConfig{
			TableName: tableName,
		},
	}
	return m.WithDataSource(name, ds)
}

// WithResolver adds a resolver
func (m *Module) WithResolver(name string, resolver Resolver) *Module {
	if m.Resolvers == nil {
		m.Resolvers = make(map[string]Resolver)
	}
	m.Resolvers[name] = resolver
	return m
}

// WithFunction adds a pipeline function
func (m *Module) WithFunction(name string, fn Function) *Module {
	if m.Functions == nil {
		m.Functions = make(map[string]Function)
	}
	m.Functions[name] = fn
	return m
}

// WithTags adds tags to the GraphQL API
func (m *Module) WithTags(tags map[string]string) *Module {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	for k, v := range tags {
		m.Tags[k] = v
	}
	return m
}

// LocalName returns the local identifier for this module instance
func (m *Module) LocalName() string {
	if m.Name != nil {
		return *m.Name
	}
	return "graphql_api"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
