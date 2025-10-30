// Package cloudfront provides type-safe Terraform module definitions for terraform-aws-modules/cloudfront/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-cloudfront v3.0
package cloudfront

// Module represents the terraform-aws-modules/cloudfront/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// CreateDistribution controls if CloudFront distribution should be created
	CreateDistribution *bool `json:"create_distribution,omitempty" hcl:"create_distribution,attr"`

	// Tags to assign to the resource
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Distribution Configuration
	// ================================

	// Aliases are extra CNAMEs (alternate domain names) for this distribution
	Aliases []string `json:"aliases,omitempty" hcl:"aliases,attr"`

	// Comment about the distribution
	Comment *string `json:"comment,omitempty" hcl:"comment,attr"`

	// ContinuousDeploymentPolicyID for continuous deployment
	ContinuousDeploymentPolicyID *string `json:"continuous_deployment_policy_id,omitempty" hcl:"continuous_deployment_policy_id,attr"`

	// DefaultRootObject is the object CloudFront returns when requesting root URL (e.g., index.html)
	DefaultRootObject *string `json:"default_root_object,omitempty" hcl:"default_root_object,attr"`

	// Enabled controls whether the distribution accepts end user requests
	Enabled *bool `json:"enabled,omitempty" hcl:"enabled,attr"`

	// HTTPVersion is the maximum HTTP version to support
	// Valid values: "http1.1" | "http2" | "http2and3" | "http3"
	HTTPVersion *string `json:"http_version,omitempty" hcl:"http_version,attr"`

	// IsIPv6Enabled controls IPv6 support
	IsIPv6Enabled *bool `json:"is_ipv6_enabled,omitempty" hcl:"is_ipv6_enabled,attr"`

	// PriceClass is the price class for this distribution
	// Valid values: "PriceClass_All" | "PriceClass_200" | "PriceClass_100"
	PriceClass *string `json:"price_class,omitempty" hcl:"price_class,attr"`

	// RetainOnDelete disables distribution instead of deleting when destroying
	RetainOnDelete *bool `json:"retain_on_delete,omitempty" hcl:"retain_on_delete,attr"`

	// WaitForDeployment waits for distribution status to change to Deployed
	WaitForDeployment *bool `json:"wait_for_deployment,omitempty" hcl:"wait_for_deployment,attr"`

	// WebACLID is the AWS WAF web ACL associated with the distribution
	WebACLID *string `json:"web_acl_id,omitempty" hcl:"web_acl_id,attr"`

	// Staging indicates if this is a staging distribution
	Staging *bool `json:"staging,omitempty" hcl:"staging,attr"`

	// ================================
	// Origin Configuration
	// ================================

	// Origin configurations (one or more)
	Origin map[string]Origin `json:"origin,omitempty" hcl:"origin,attr"`

	// OriginGroup configurations for failover
	OriginGroup map[string]OriginGroup `json:"origin_group,omitempty" hcl:"origin_group,attr"`

	// ================================
	// Origin Access Control
	// ================================

	// CreateOriginAccessIdentity controls if origin access identity should be created (legacy)
	CreateOriginAccessIdentity *bool `json:"create_origin_access_identity,omitempty" hcl:"create_origin_access_identity,attr"`

	// OriginAccessIdentities is a map of origin access identities
	OriginAccessIdentities map[string]string `json:"origin_access_identities,omitempty" hcl:"origin_access_identities,attr"`

	// CreateOriginAccessControl controls if origin access control should be created (recommended)
	CreateOriginAccessControl *bool `json:"create_origin_access_control,omitempty" hcl:"create_origin_access_control,attr"`

	// OriginAccessControl is a map of origin access controls
	OriginAccessControl map[string]OriginAccessControl `json:"origin_access_control,omitempty" hcl:"origin_access_control,attr"`

	// ================================
	// Cache Behavior
	// ================================

	// DefaultCacheBehavior is the default cache behavior
	DefaultCacheBehavior *CacheBehavior `json:"default_cache_behavior,omitempty" hcl:"default_cache_behavior,attr"`

	// OrderedCacheBehavior is an ordered list of cache behaviors (precedence 0 to N)
	OrderedCacheBehavior []CacheBehavior `json:"ordered_cache_behavior,omitempty" hcl:"ordered_cache_behavior,attr"`

	// ================================
	// Security & Access
	// ================================

	// ViewerCertificate is the SSL configuration
	ViewerCertificate *ViewerCertificate `json:"viewer_certificate,omitempty" hcl:"viewer_certificate,attr"`

	// GeoRestriction configures geographic restrictions
	GeoRestriction *GeoRestriction `json:"geo_restriction,omitempty" hcl:"geo_restriction,attr"`

	// ================================
	// Logging & Monitoring
	// ================================

	// LoggingConfig controls how logs are written
	LoggingConfig *LoggingConfig `json:"logging_config,omitempty" hcl:"logging_config,attr"`

	// CreateMonitoringSubscription enables additional CloudWatch metrics
	CreateMonitoringSubscription *bool `json:"create_monitoring_subscription,omitempty" hcl:"create_monitoring_subscription,attr"`

	// RealtimeMetricsSubscriptionStatus enables real-time metrics
	// Valid values: "Enabled" | "Disabled"
	RealtimeMetricsSubscriptionStatus *string `json:"realtime_metrics_subscription_status,omitempty" hcl:"realtime_metrics_subscription_status,attr"`

	// ================================
	// Error Pages
	// ================================

	// CustomErrorResponse defines custom error pages
	CustomErrorResponse map[string]CustomErrorResponse `json:"custom_error_response,omitempty" hcl:"custom_error_response,attr"`

	// ================================
	// VPC Origin (Advanced)
	// ================================

	// CreateVPCOrigin enables VPC origin resource
	CreateVPCOrigin *bool `json:"create_vpc_origin,omitempty" hcl:"create_vpc_origin,attr"`

	// VPCOrigin configurations
	VPCOrigin map[string]VPCOrigin `json:"vpc_origin,omitempty" hcl:"vpc_origin,attr"`

	// VPCOriginTimeouts for create, update, delete operations
	VPCOriginTimeouts map[string]string `json:"vpc_origin_timeouts,omitempty" hcl:"vpc_origin_timeouts,attr"`
}

// Origin represents a CloudFront origin.
type Origin struct {
	// DomainName is the DNS domain name of the origin (S3 bucket, ALB, etc.)
	DomainName string `json:"domain_name" hcl:"domain_name,attr"`

	// OriginID is a unique identifier for this origin
	OriginID string `json:"origin_id" hcl:"origin_id,attr"`

	// OriginPath is a directory path to append to origin requests
	OriginPath *string `json:"origin_path,omitempty" hcl:"origin_path,attr"`

	// ConnectionAttempts is the number of connection attempts (1-3)
	ConnectionAttempts *int `json:"connection_attempts,omitempty" validate:"min=1,max=3" hcl:"connection_attempts,attr"`

	// ConnectionTimeout is the connection timeout in seconds (1-10)
	ConnectionTimeout *int `json:"connection_timeout,omitempty" validate:"min=1,max=10" hcl:"connection_timeout,attr"`

	// CustomOriginConfig for custom origins (non-S3)
	CustomOriginConfig *CustomOriginConfig `json:"custom_origin_config,omitempty" hcl:"custom_origin_config,attr"`

	// S3OriginConfig for S3 bucket origins
	S3OriginConfig *S3OriginConfig `json:"s3_origin_config,omitempty" hcl:"s3_origin_config,attr"`

	// CustomHeaders to include in requests to the origin
	CustomHeaders []OriginCustomHeader `json:"custom_headers,omitempty" hcl:"custom_headers,attr"`

	// OriginShield configuration
	OriginShield *OriginShield `json:"origin_shield,omitempty" hcl:"origin_shield,attr"`
}

// CustomOriginConfig represents custom origin settings.
type CustomOriginConfig struct {
	// HTTPPort for HTTP connections
	HTTPPort int `json:"http_port" hcl:"http_port,attr"`

	// HTTPSPort for HTTPS connections
	HTTPSPort int `json:"https_port" hcl:"https_port,attr"`

	// OriginProtocolPolicy controls protocol for origin requests
	// Valid values: "http-only" | "https-only" | "match-viewer"
	OriginProtocolPolicy string `json:"origin_protocol_policy" hcl:"origin_protocol_policy,attr"`

	// OriginSSLProtocols are the SSL/TLS protocols CloudFront can use
	OriginSSLProtocols []string `json:"origin_ssl_protocols,omitempty" hcl:"origin_ssl_protocols,attr"`

	// OriginReadTimeout in seconds (1-180)
	OriginReadTimeout *int `json:"origin_read_timeout,omitempty" validate:"min=1,max=180" hcl:"origin_read_timeout,attr"`

	// OriginKeepaliveTimeout in seconds (1-180)
	OriginKeepaliveTimeout *int `json:"origin_keepalive_timeout,omitempty" validate:"min=1,max=180" hcl:"origin_keepalive_timeout,attr"`
}

// S3OriginConfig represents S3 origin settings.
type S3OriginConfig struct {
	// OriginAccessIdentity for legacy S3 access (e.g., "origin-access-identity/cloudfront/ABCDEFG1234567")
	OriginAccessIdentity *string `json:"origin_access_identity,omitempty" hcl:"origin_access_identity,attr"`
}

// OriginCustomHeader represents a custom header to add to origin requests.
type OriginCustomHeader struct {
	Name  string `json:"name"  hcl:"name,attr"`
	Value string `json:"value" hcl:"value,attr"`
}

// OriginShield reduces load on your origin.
type OriginShield struct {
	Enabled            bool    `json:"enabled"                        hcl:"enabled,attr"`
	OriginShieldRegion *string `json:"origin_shield_region,omitempty" hcl:"origin_shield_region,attr"`
}

// OriginGroup represents a failover origin group.
type OriginGroup struct {
	OriginID         string   `json:"origin_id"         hcl:"origin_id,attr"`
	FailoverCriteria []int    `json:"failover_criteria" hcl:"failover_criteria,attr"`
	Members          []string `json:"members"           hcl:"members,attr"`
}

// OriginAccessControl represents origin access control (recommended over OAI).
type OriginAccessControl struct {
	Name            *string `json:"name,omitempty"   hcl:"name,attr"`
	Description     string  `json:"description"      hcl:"description,attr"`
	OriginType      string  `json:"origin_type"      hcl:"origin_type,attr"`
	SigningBehavior string  `json:"signing_behavior" hcl:"signing_behavior,attr"`
	SigningProtocol string  `json:"signing_protocol" hcl:"signing_protocol,attr"`
}

// CacheBehavior represents cache behavior configuration.
type CacheBehavior struct {
	// PathPattern for ordered cache behaviors (not used for default)
	PathPattern *string `json:"path_pattern,omitempty" hcl:"path_pattern,attr"`

	// TargetOriginID is the origin to route requests to
	TargetOriginID string `json:"target_origin_id" hcl:"target_origin_id,attr"`

	// ViewerProtocolPolicy controls HTTP/HTTPS for viewers
	// Valid values: "allow-all" | "https-only" | "redirect-to-https"
	ViewerProtocolPolicy string `json:"viewer_protocol_policy" hcl:"viewer_protocol_policy,attr"`

	// AllowedMethods are the HTTP methods CloudFront processes
	AllowedMethods []string `json:"allowed_methods,omitempty" hcl:"allowed_methods,attr"`

	// CachedMethods are the HTTP methods CloudFront caches
	CachedMethods []string `json:"cached_methods,omitempty" hcl:"cached_methods,attr"`

	// Compress enables automatic compression
	Compress *bool `json:"compress,omitempty" hcl:"compress,attr"`

	// CachePolicyID is the cache policy ID
	CachePolicyID *string `json:"cache_policy_id,omitempty" hcl:"cache_policy_id,attr"`

	// OriginRequestPolicyID is the origin request policy ID
	OriginRequestPolicyID *string `json:"origin_request_policy_id,omitempty" hcl:"origin_request_policy_id,attr"`

	// ResponseHeadersPolicyID is the response headers policy ID
	ResponseHeadersPolicyID *string `json:"response_headers_policy_id,omitempty" hcl:"response_headers_policy_id,attr"`

	// RealtimeLogConfigARN enables real-time logs
	RealtimeLogConfigARN *string `json:"realtime_log_config_arn,omitempty" hcl:"realtime_log_config_arn,attr"`

	// TrustedKeyGroups for signed URLs/cookies
	TrustedKeyGroups []string `json:"trusted_key_groups,omitempty" hcl:"trusted_key_groups,attr"`

	// TrustedSigners for signed URLs/cookies (legacy)
	TrustedSigners []string `json:"trusted_signers,omitempty" hcl:"trusted_signers,attr"`

	// LambdaFunctionAssociations for Lambda@Edge
	LambdaFunctionAssociations []LambdaFunctionAssociation `json:"lambda_function_associations,omitempty" hcl:"lambda_function_associations,attr"`

	// FunctionAssociations for CloudFront Functions
	FunctionAssociations []FunctionAssociation `json:"function_associations,omitempty" hcl:"function_associations,attr"`

	// MinTTL, DefaultTTL, MaxTTL for legacy cache settings (use cache policies instead)
	MinTTL     *int `json:"min_ttl,omitempty"     hcl:"min_ttl,attr"`
	DefaultTTL *int `json:"default_ttl,omitempty" hcl:"default_ttl,attr"`
	MaxTTL     *int `json:"max_ttl,omitempty"     hcl:"max_ttl,attr"`
}

// LambdaFunctionAssociation represents Lambda@Edge configuration.
type LambdaFunctionAssociation struct {
	EventType   string `json:"event_type"             hcl:"event_type,attr"`
	LambdaARN   string `json:"lambda_arn"             hcl:"lambda_arn,attr"`
	IncludeBody *bool  `json:"include_body,omitempty" hcl:"include_body,attr"`
}

// FunctionAssociation represents CloudFront Functions configuration.
type FunctionAssociation struct {
	EventType   string `json:"event_type"   hcl:"event_type,attr"`
	FunctionARN string `json:"function_arn" hcl:"function_arn,attr"`
}

// ViewerCertificate represents SSL/TLS certificate configuration.
type ViewerCertificate struct {
	// CloudFrontDefaultCertificate uses *.cloudfront.net certificate
	CloudFrontDefaultCertificate *bool `json:"cloudfront_default_certificate,omitempty" hcl:"cloudfront_default_certificate,attr"`

	// ACMCertificateARN is the ARN of the AWS Certificate Manager certificate
	ACMCertificateARN *string `json:"acm_certificate_arn,omitempty" hcl:"acm_certificate_arn,attr"`

	// IAMCertificateID is the IAM certificate identifier (legacy)
	IAMCertificateID *string `json:"iam_certificate_id,omitempty" hcl:"iam_certificate_id,attr"`

	// MinimumProtocolVersion is the minimum TLS version
	// Valid values: "TLSv1" | "TLSv1.1_2016" | "TLSv1.2_2018" | "TLSv1.2_2019" | "TLSv1.2_2021"
	MinimumProtocolVersion *string `json:"minimum_protocol_version,omitempty" hcl:"minimum_protocol_version,attr"`

	// SSLSupportMethod controls how CloudFront serves HTTPS
	// Valid values: "sni-only" | "vip" (vip has additional costs)
	SSLSupportMethod *string `json:"ssl_support_method,omitempty" hcl:"ssl_support_method,attr"`
}

// GeoRestriction represents geographic access restrictions.
type GeoRestriction struct {
	// RestrictionType controls access by geography
	// Valid values: "none" | "whitelist" | "blacklist"
	RestrictionType string `json:"restriction_type" hcl:"restriction_type,attr"`

	// Locations are ISO 3166-1-alpha-2 country codes
	Locations []string `json:"locations,omitempty" hcl:"locations,attr"`
}

// LoggingConfig represents access logging configuration.
type LoggingConfig struct {
	// Bucket is the S3 bucket for access logs
	Bucket string `json:"bucket" hcl:"bucket,attr"`

	// IncludeCookies includes cookies in logs
	IncludeCookies *bool `json:"include_cookies,omitempty" hcl:"include_cookies,attr"`

	// Prefix is the log file prefix
	Prefix *string `json:"prefix,omitempty" hcl:"prefix,attr"`
}

// CustomErrorResponse represents custom error page configuration.
type CustomErrorResponse struct {
	ErrorCode          int     `json:"error_code"                      hcl:"error_code,attr"`
	ResponseCode       *int    `json:"response_code,omitempty"         hcl:"response_code,attr"`
	ResponsePagePath   *string `json:"response_page_path,omitempty"    hcl:"response_page_path,attr"`
	ErrorCachingMinTTL *int    `json:"error_caching_min_ttl,omitempty" hcl:"error_caching_min_ttl,attr"`
}

// VPCOrigin represents VPC origin configuration.
type VPCOrigin struct {
	Name                 string             `json:"name"                   hcl:"name,attr"`
	ARN                  string             `json:"arn"                    hcl:"arn,attr"`
	HTTPPort             int                `json:"http_port"              hcl:"http_port,attr"`
	HTTPSPort            int                `json:"https_port"             hcl:"https_port,attr"`
	OriginProtocolPolicy string             `json:"origin_protocol_policy" hcl:"origin_protocol_policy,attr"`
	OriginSSLProtocols   OriginSSLProtocols `json:"origin_ssl_protocols"   hcl:"origin_ssl_protocols,attr"`
}

// OriginSSLProtocols represents SSL protocol configuration.
type OriginSSLProtocols struct {
	Items    []string `json:"items"    hcl:"items,attr"`
	Quantity int      `json:"quantity" hcl:"quantity,attr"`
}

// NewModule creates a new CloudFront module with sensible defaults.
func NewModule(comment string) *Module {
	source := "terraform-aws-modules/cloudfront/aws"
	version := "~> 3.0"
	create := true
	enabled := true
	httpVersion := "http2"
	waitForDeployment := true
	ipv6Enabled := true

	return &Module{
		Source:             source,
		Version:            version,
		Comment:            &comment,
		CreateDistribution: &create,
		Enabled:            &enabled,
		HTTPVersion:        &httpVersion,
		WaitForDeployment:  &waitForDeployment,
		IsIPv6Enabled:      &ipv6Enabled,
	}
}

// WithOrigin adds an origin configuration.
func (m *Module) WithOrigin(id string, origin Origin) *Module {
	if m.Origin == nil {
		m.Origin = make(map[string]Origin)
	}
	m.Origin[id] = origin
	return m
}

// WithS3Origin adds an S3 bucket origin with access control.
func (m *Module) WithS3Origin(id, bucketDomain, oaiID string) *Module {
	origin := Origin{
		DomainName: bucketDomain,
		OriginID:   id,
		S3OriginConfig: &S3OriginConfig{
			OriginAccessIdentity: &oaiID,
		},
	}
	return m.WithOrigin(id, origin)
}

// WithCustomOrigin adds a custom origin (ALB, API Gateway, etc.)
func (m *Module) WithCustomOrigin(id, domainName string, httpsOnly bool) *Module {
	protocol := "https-only"
	if !httpsOnly {
		protocol = "match-viewer"
	}

	origin := Origin{
		DomainName: domainName,
		OriginID:   id,
		CustomOriginConfig: &CustomOriginConfig{
			HTTPPort:             80,
			HTTPSPort:            443,
			OriginProtocolPolicy: protocol,
			OriginSSLProtocols:   []string{"TLSv1.2"},
		},
	}
	return m.WithOrigin(id, origin)
}

// WithDefaultCacheBehavior configures the default cache behavior.
func (m *Module) WithDefaultCacheBehavior(originID, viewerProtocol string) *Module {
	m.DefaultCacheBehavior = &CacheBehavior{
		TargetOriginID:       originID,
		ViewerProtocolPolicy: viewerProtocol,
		AllowedMethods:       []string{"GET", "HEAD", "OPTIONS"},
		CachedMethods:        []string{"GET", "HEAD"},
	}
	return m
}

// WithCertificate configures ACM certificate for custom domains.
func (m *Module) WithCertificate(acmCertARN, minTLSVersion string) *Module {
	sniOnly := "sni-only"
	m.ViewerCertificate = &ViewerCertificate{
		ACMCertificateARN:      &acmCertARN,
		MinimumProtocolVersion: &minTLSVersion,
		SSLSupportMethod:       &sniOnly,
	}
	return m
}

// WithAliases adds custom domain names (CNAMEs).
func (m *Module) WithAliases(aliases ...string) *Module {
	m.Aliases = aliases
	return m
}

// WithPriceClass sets the price class for edge location coverage.
func (m *Module) WithPriceClass(priceClass string) *Module {
	m.PriceClass = &priceClass
	return m
}

// WithLogging configures access logging to S3.
func (m *Module) WithLogging(bucket, prefix string, includeCookies bool) *Module {
	m.LoggingConfig = &LoggingConfig{
		Bucket:         bucket,
		Prefix:         &prefix,
		IncludeCookies: &includeCookies,
	}
	return m
}

// WithGeoRestriction configures geographic restrictions.
func (m *Module) WithGeoRestriction(restrictionType string, locations []string) *Module {
	m.GeoRestriction = &GeoRestriction{
		RestrictionType: restrictionType,
		Locations:       locations,
	}
	return m
}

// WithWAF associates a WAF web ACL.
func (m *Module) WithWAF(webACLID string) *Module {
	m.WebACLID = &webACLID
	return m
}

// WithOriginAccessControl creates origin access control for S3.
func (m *Module) WithOriginAccessControl(name, description string) *Module {
	create := true
	m.CreateOriginAccessControl = &create

	if m.OriginAccessControl == nil {
		m.OriginAccessControl = make(map[string]OriginAccessControl)
	}

	m.OriginAccessControl[name] = OriginAccessControl{
		Name:            &name,
		Description:     description,
		OriginType:      "s3",
		SigningBehavior: "always",
		SigningProtocol: "sigv4",
	}
	return m
}

// WithLambdaEdge adds Lambda@Edge function to default cache behavior.
func (m *Module) WithLambdaEdge(eventType, lambdaARN string) *Module {
	if m.DefaultCacheBehavior == nil {
		m.DefaultCacheBehavior = &CacheBehavior{}
	}

	m.DefaultCacheBehavior.LambdaFunctionAssociations = append(
		m.DefaultCacheBehavior.LambdaFunctionAssociations,
		LambdaFunctionAssociation{
			EventType: eventType,
			LambdaARN: lambdaARN,
		},
	)
	return m
}

// WithTags adds tags to the distribution.
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
	if m.Comment != nil {
		return *m.Comment
	}
	return "distribution"
}

// Configuration generates the HCL configuration for this module.
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
