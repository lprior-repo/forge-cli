// Package apigatewayv2 provides type-safe Terraform module definitions for terraform-aws-modules/apigateway-v2/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-apigateway-v2 v5.0
package apigatewayv2

// Module represents the terraform-aws-modules/apigateway-v2/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create controls if resources should be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// Tags to assign to API gateway resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// API Gateway Configuration
	// ================================

	// Name is the name of the API (max 128 characters)
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Description of the API (max 1024 characters)
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// ProtocolType is the API protocol
	// Valid values: "HTTP" | "WEBSOCKET"
	ProtocolType *string `json:"protocol_type,omitempty" hcl:"protocol_type,attr"`

	// APIVersion is a version identifier (1-64 characters)
	APIVersion *string `json:"api_version,omitempty" hcl:"api_version,attr"`

	// Body is an OpenAPI specification for HTTP APIs
	Body *string `json:"body,omitempty" hcl:"body,attr"`

	// DisableExecuteAPIEndpoint disables the default execute-api endpoint
	DisableExecuteAPIEndpoint *bool `json:"disable_execute_api_endpoint,omitempty" hcl:"disable_execute_api_endpoint,attr"`

	// FailOnWarnings returns error on warnings (HTTP APIs)
	FailOnWarnings *bool `json:"fail_on_warnings,omitempty" hcl:"fail_on_warnings,attr"`

	// IPAddressType for API invocation
	// Valid values: "ipv4" | "dualstack"
	IPAddressType *string `json:"ip_address_type,omitempty" hcl:"ip_address_type,attr"`

	// RouteSelectionExpression for route selection (default: $request.method $request.path)
	RouteSelectionExpression *string `json:"route_selection_expression,omitempty" hcl:"route_selection_expression,attr"`

	// APIKeySelectionExpression for WebSocket APIs
	// Valid values: "$context.authorizer.usageIdentifierKey" | "$request.header.x-api-key"
	APIKeySelectionExpression *string `json:"api_key_selection_expression,omitempty" hcl:"api_key_selection_expression,attr"`

	// APIMappingKey for API mapping
	APIMappingKey *string `json:"api_mapping_key,omitempty" hcl:"api_mapping_key,attr"`

	// ================================
	// Quick Create (HTTP APIs)
	// ================================

	// Target is the integration target (Lambda ARN or HTTP URL)
	Target *string `json:"target,omitempty" hcl:"target,attr"`

	// RouteKey specifies the route key for quick create
	RouteKey *string `json:"route_key,omitempty" hcl:"route_key,attr"`

	// CredentialsARN for the integration
	CredentialsARN *string `json:"credentials_arn,omitempty" hcl:"credentials_arn,attr"`

	// ================================
	// CORS Configuration
	// ================================

	// CORSConfiguration for HTTP APIs
	CORSConfiguration *CORSConfiguration `json:"cors_configuration,omitempty" hcl:"cors_configuration,attr"`

	// ================================
	// Authorizers
	// ================================

	// Authorizers is a map of authorizer configurations
	Authorizers map[string]Authorizer `json:"authorizers,omitempty" hcl:"authorizers,attr"`

	// ================================
	// Domain Name
	// ================================

	// CreateDomainName controls whether to create domain name resource
	CreateDomainName *bool `json:"create_domain_name,omitempty" hcl:"create_domain_name,attr"`

	// DomainName for the API
	DomainName *string `json:"domain_name,omitempty" hcl:"domain_name,attr"`

	// DomainNameCertificateARN is the ACM certificate ARN
	DomainNameCertificateARN *string `json:"domain_name_certificate_arn,omitempty" hcl:"domain_name_certificate_arn,attr"`

	// DomainNameOwnershipVerificationCertificateARN for verification
	DomainNameOwnershipVerificationCertificateARN *string `json:"domain_name_ownership_verification_certificate_arn,omitempty" hcl:"domain_name_ownership_verification_certificate_arn,attr"`

	// MutualTLSAuthentication configures mTLS
	MutualTLSAuthentication *MutualTLSAuthentication `json:"mutual_tls_authentication,omitempty" hcl:"mutual_tls_authentication,attr"`

	// ================================
	// Stages
	// ================================

	// Stages is a map of stage configurations
	Stages map[string]Stage `json:"stages,omitempty" hcl:"stages,attr"`

	// CreateStage controls whether to create default stage
	CreateStage *bool `json:"create_stage,omitempty" hcl:"create_stage,attr"`

	// StageName is the default stage name
	StageName *string `json:"stage_name,omitempty" hcl:"stage_name,attr"`

	// StageDescription is the default stage description
	StageDescription *string `json:"stage_description,omitempty" hcl:"stage_description,attr"`

	// AutoDeploy automatically deploys changes to default stage
	AutoDeploy *bool `json:"auto_deploy,omitempty" hcl:"auto_deploy,attr"`

	// DefaultRouteSettings for default stage
	DefaultRouteSettings *RouteSettings `json:"default_route_settings,omitempty" hcl:"default_route_settings,attr"`

	// ================================
	// Routes & Integrations
	// ================================

	// Routes is a map of route configurations
	Routes map[string]Route `json:"routes,omitempty" hcl:"routes,attr"`

	// Integrations is a map of integration configurations
	Integrations map[string]Integration `json:"integrations,omitempty" hcl:"integrations,attr"`

	// ================================
	// VPC Links
	// ================================

	// VPCLinks is a map of VPC link configurations
	VPCLinks map[string]VPCLink `json:"vpc_links,omitempty" hcl:"vpc_links,attr"`
}

// CORSConfiguration represents CORS settings for HTTP APIs.
type CORSConfiguration struct {
	// AllowCredentials allows credentials
	AllowCredentials *bool `json:"allow_credentials,omitempty" hcl:"allow_credentials,attr"`

	// AllowHeaders are allowed headers
	AllowHeaders []string `json:"allow_headers,omitempty" hcl:"allow_headers,attr"`

	// AllowMethods are allowed HTTP methods
	AllowMethods []string `json:"allow_methods,omitempty" hcl:"allow_methods,attr"`

	// AllowOrigins are allowed origins
	AllowOrigins []string `json:"allow_origins,omitempty" hcl:"allow_origins,attr"`

	// ExposeHeaders are exposed headers
	ExposeHeaders []string `json:"expose_headers,omitempty" hcl:"expose_headers,attr"`

	// MaxAge is the cache duration in seconds
	MaxAge *int `json:"max_age,omitempty" hcl:"max_age,attr"`
}

// Authorizer represents an API Gateway authorizer.
type Authorizer struct {
	// Name is the authorizer name
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// AuthorizerType is the type of authorizer
	// Valid values: "REQUEST" | "JWT"
	AuthorizerType *string `json:"authorizer_type,omitempty" hcl:"authorizer_type,attr"`

	// AuthorizerURI is the Lambda function URI
	AuthorizerURI *string `json:"authorizer_uri,omitempty" hcl:"authorizer_uri,attr"`

	// AuthorizerCredentialsARN is the IAM role ARN
	AuthorizerCredentialsARN *string `json:"authorizer_credentials_arn,omitempty" hcl:"authorizer_credentials_arn,attr"`

	// AuthorizerPayloadFormatVersion is the payload format version
	AuthorizerPayloadFormatVersion *string `json:"authorizer_payload_format_version,omitempty" hcl:"authorizer_payload_format_version,attr"`

	// AuthorizerResultTTLInSeconds is the TTL for cached results (0-3600)
	AuthorizerResultTTLInSeconds *int `json:"authorizer_result_ttl_in_seconds,omitempty" hcl:"authorizer_result_ttl_in_seconds,attr"`

	// EnableSimpleResponses enables simple boolean responses
	EnableSimpleResponses *bool `json:"enable_simple_responses,omitempty" hcl:"enable_simple_responses,attr"`

	// IdentitySources are identity source expressions
	IdentitySources []string `json:"identity_sources,omitempty" hcl:"identity_sources,attr"`

	// JWTConfiguration for JWT authorizers
	JWTConfiguration *JWTConfiguration `json:"jwt_configuration,omitempty" hcl:"jwt_configuration,attr"`
}

// JWTConfiguration represents JWT authorizer configuration.
type JWTConfiguration struct {
	// Audience is the list of allowed audiences
	Audience []string `json:"audience,omitempty" hcl:"audience,attr"`

	// Issuer is the JWT issuer
	Issuer *string `json:"issuer,omitempty" hcl:"issuer,attr"`
}

// MutualTLSAuthentication represents mTLS configuration.
type MutualTLSAuthentication struct {
	// TruststoreURI is the S3 URI of the truststore
	TruststoreURI string `json:"truststore_uri" hcl:"truststore_uri,attr"`

	// TruststoreVersion is the version of the truststore
	TruststoreVersion *string `json:"truststore_version,omitempty" hcl:"truststore_version,attr"`
}

// Stage represents an API Gateway stage.
type Stage struct {
	// Name is the stage name
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Description of the stage
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// AutoDeploy automatically deploys changes
	AutoDeploy *bool `json:"auto_deploy,omitempty" hcl:"auto_deploy,attr"`

	// DefaultRouteSettings for the stage
	DefaultRouteSettings *RouteSettings `json:"default_route_settings,omitempty" hcl:"default_route_settings,attr"`

	// AccessLogSettings configures access logging
	AccessLogSettings *AccessLogSettings `json:"access_log_settings,omitempty" hcl:"access_log_settings,attr"`

	// ThrottleSettings configures throttling
	ThrottleSettings *ThrottleSettings `json:"throttle_settings,omitempty" hcl:"throttle_settings,attr"`

	// StageVariables are stage-specific variables
	StageVariables map[string]string `json:"stage_variables,omitempty" hcl:"stage_variables,attr"`

	// Tags for the stage
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`
}

// RouteSettings represents route-level settings.
type RouteSettings struct {
	// DataTraceEnabled enables data trace logging
	DataTraceEnabled *bool `json:"data_trace_enabled,omitempty" hcl:"data_trace_enabled,attr"`

	// DetailedMetricsEnabled enables detailed CloudWatch metrics
	DetailedMetricsEnabled *bool `json:"detailed_metrics_enabled,omitempty" hcl:"detailed_metrics_enabled,attr"`

	// LoggingLevel is the logging level
	// Valid values: "OFF" | "ERROR" | "INFO"
	LoggingLevel *string `json:"logging_level,omitempty" hcl:"logging_level,attr"`

	// ThrottlingBurstLimit is the throttle burst limit
	ThrottlingBurstLimit *int `json:"throttling_burst_limit,omitempty" hcl:"throttling_burst_limit,attr"`

	// ThrottlingRateLimit is the throttle rate limit
	ThrottlingRateLimit *float64 `json:"throttling_rate_limit,omitempty" hcl:"throttling_rate_limit,attr"`
}

// AccessLogSettings represents access logging configuration.
type AccessLogSettings struct {
	// DestinationARN is the CloudWatch Logs group or Kinesis Data Firehose ARN
	DestinationARN string `json:"destination_arn" hcl:"destination_arn,attr"`

	// Format is the log format
	Format string `json:"format" hcl:"format,attr"`
}

// ThrottleSettings represents throttling configuration.
type ThrottleSettings struct {
	// BurstLimit is the burst limit
	BurstLimit *int `json:"burst_limit,omitempty" hcl:"burst_limit,attr"`

	// RateLimit is the rate limit
	RateLimit *float64 `json:"rate_limit,omitempty" hcl:"rate_limit,attr"`
}

// Route represents an API Gateway route.
type Route struct {
	// RouteKey is the route key
	RouteKey string `json:"route_key" hcl:"route_key,attr"`

	// IntegrationKey references an integration
	IntegrationKey *string `json:"integration_key,omitempty" hcl:"integration_key,attr"`

	// AuthorizationKey references an authorizer
	AuthorizationKey *string `json:"authorization_key,omitempty" hcl:"authorization_key,attr"`

	// AuthorizationType is the authorization type
	// Valid values: "NONE" | "AWS_IAM" | "CUSTOM" | "JWT"
	AuthorizationType *string `json:"authorization_type,omitempty" hcl:"authorization_type,attr"`

	// APIKeyRequired indicates if API key is required
	APIKeyRequired *bool `json:"api_key_required,omitempty" hcl:"api_key_required,attr"`

	// OperationName is a friendly operation name
	OperationName *string `json:"operation_name,omitempty" hcl:"operation_name,attr"`
}

// Integration represents an API Gateway integration.
type Integration struct {
	// IntegrationType is the type of integration
	// Valid values: "AWS" | "AWS_PROXY" | "HTTP" | "HTTP_PROXY" | "MOCK"
	IntegrationType string `json:"integration_type" hcl:"integration_type,attr"`

	// IntegrationURI is the URI of the integration
	IntegrationURI *string `json:"integration_uri,omitempty" hcl:"integration_uri,attr"`

	// IntegrationMethod is the HTTP method for the integration
	IntegrationMethod *string `json:"integration_method,omitempty" hcl:"integration_method,attr"`

	// ConnectionType is the connection type
	// Valid values: "INTERNET" | "VPC_LINK"
	ConnectionType *string `json:"connection_type,omitempty" hcl:"connection_type,attr"`

	// ConnectionID is the VPC link ID
	ConnectionID *string `json:"connection_id,omitempty" hcl:"connection_id,attr"`

	// PayloadFormatVersion is the payload format version
	PayloadFormatVersion *string `json:"payload_format_version,omitempty" hcl:"payload_format_version,attr"`

	// TimeoutMilliseconds is the integration timeout (50-30000)
	TimeoutMilliseconds *int `json:"timeout_milliseconds,omitempty" hcl:"timeout_milliseconds,attr"`
}

// VPCLink represents a VPC link.
type VPCLink struct {
	// Name is the VPC link name
	Name string `json:"name" hcl:"name,attr"`

	// SecurityGroupIDs are the security group IDs
	SecurityGroupIDs []string `json:"security_group_ids" hcl:"security_group_ids,attr"`

	// SubnetIDs are the subnet IDs
	SubnetIDs []string `json:"subnet_ids" hcl:"subnet_ids,attr"`

	// Tags for the VPC link
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`
}

// NewModule creates a new API Gateway V2 module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/apigateway-v2/aws"
	version := "~> 5.0"
	create := true
	protocolType := "HTTP"
	createStage := true
	stageName := "default"
	autoDeploy := true

	return &Module{
		Source:       source,
		Version:      version,
		Name:         &name,
		Create:       &create,
		ProtocolType: &protocolType,
		CreateStage:  &createStage,
		StageName:    &stageName,
		AutoDeploy:   &autoDeploy,
	}
}

// WithCORS configures CORS for HTTP APIs.
func (m *Module) WithCORS(allowOrigins, allowMethods, allowHeaders []string) *Module {
	m.CORSConfiguration = &CORSConfiguration{
		AllowOrigins: allowOrigins,
		AllowMethods: allowMethods,
		AllowHeaders: allowHeaders,
	}
	return m
}

// WithDomainName configures custom domain.
func (m *Module) WithDomainName(domain, certificateARN string) *Module {
	createDomain := true
	m.CreateDomainName = &createDomain
	m.DomainName = &domain
	m.DomainNameCertificateARN = &certificateARN
	return m
}

// WithJWTAuthorizer adds a JWT authorizer.
func (m *Module) WithJWTAuthorizer(name, issuer string, audience []string) *Module {
	if m.Authorizers == nil {
		m.Authorizers = make(map[string]Authorizer)
	}
	authType := "JWT"
	m.Authorizers[name] = Authorizer{
		Name:           &name,
		AuthorizerType: &authType,
		JWTConfiguration: &JWTConfiguration{
			Issuer:   &issuer,
			Audience: audience,
		},
	}
	return m
}

// WithLambdaAuthorizer adds a Lambda authorizer.
func (m *Module) WithLambdaAuthorizer(name, lambdaURI string, identitySources []string) *Module {
	if m.Authorizers == nil {
		m.Authorizers = make(map[string]Authorizer)
	}
	authType := "REQUEST"
	m.Authorizers[name] = Authorizer{
		Name:            &name,
		AuthorizerType:  &authType,
		AuthorizerURI:   &lambdaURI,
		IdentitySources: identitySources,
	}
	return m
}

// WithRoute adds a route.
func (m *Module) WithRoute(key string, route Route) *Module {
	if m.Routes == nil {
		m.Routes = make(map[string]Route)
	}
	m.Routes[key] = route
	return m
}

// WithIntegration adds an integration.
func (m *Module) WithIntegration(key string, integration Integration) *Module {
	if m.Integrations == nil {
		m.Integrations = make(map[string]Integration)
	}
	m.Integrations[key] = integration
	return m
}

// WithTags adds tags to the API.
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
	if m.Name != nil {
		return *m.Name
	}
	return "api_gateway"
}

// Configuration generates the HCL configuration for this module.
func (_ *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
