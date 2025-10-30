// Package appconfig provides type-safe Terraform module definitions for terraform-aws-modules/appconfig/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-appconfig v2.0
package appconfig

// Module represents the terraform-aws-modules/appconfig/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create determines whether resources are created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// Tags to apply to all resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Application Configuration
	// ================================

	// Name of the application (1-64 characters)
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Description of the application (max 1024 characters)
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// ================================
	// Environments
	// ================================

	// Environments is a map of environment configurations
	Environments map[string]Environment `json:"environments,omitempty" hcl:"environments,attr"`

	// ================================
	// Configuration Profile
	// ================================

	// ConfigProfileName is the configuration profile name (1-64 characters)
	ConfigProfileName *string `json:"config_profile_name,omitempty" hcl:"config_profile_name,attr"`

	// ConfigProfileDescription describes the configuration profile (max 1024 characters)
	ConfigProfileDescription *string `json:"config_profile_description,omitempty" hcl:"config_profile_description,attr"`

	// ConfigProfileType is the type of configurations
	// Valid values: "AWS.AppConfig.FeatureFlags" | "AWS.Freeform"
	ConfigProfileType *string `json:"config_profile_type,omitempty" hcl:"config_profile_type,attr"`

	// ConfigProfileLocationURI locates the configuration
	// Can be: "hosted", SSM document, SSM parameter, or S3 object
	ConfigProfileLocationURI *string `json:"config_profile_location_uri,omitempty" hcl:"config_profile_location_uri,attr"`

	// ConfigProfileRetrievalRoleARN is the IAM role for configuration retrieval
	ConfigProfileRetrievalRoleARN *string `json:"config_profile_retrieval_role_arn,omitempty" hcl:"config_profile_retrieval_role_arn,attr"`

	// ConfigProfileValidator validates configuration (max 2)
	ConfigProfileValidator []Validator `json:"config_profile_validator,omitempty" hcl:"config_profile_validator,attr"`

	// ConfigProfileTags are additional tags for the configuration profile
	ConfigProfileTags map[string]string `json:"config_profile_tags,omitempty" hcl:"config_profile_tags,attr"`

	// ================================
	// Retrieval Role
	// ================================

	// CreateRetrievalRole creates IAM role for configuration retrieval
	CreateRetrievalRole *bool `json:"create_retrieval_role,omitempty" hcl:"create_retrieval_role,attr"`

	// RetrievalRoleName is the configuration retrieval role name
	RetrievalRoleName *string `json:"retrieval_role_name,omitempty" hcl:"retrieval_role_name,attr"`

	// RetrievalRoleUseNamePrefix uses name-prefix strategy
	RetrievalRoleUseNamePrefix *bool `json:"retrieval_role_use_name_prefix,omitempty" hcl:"retrieval_role_use_name_prefix,attr"`

	// RetrievalRoleDescription describes the retrieval role
	RetrievalRoleDescription *string `json:"retrieval_role_description,omitempty" hcl:"retrieval_role_description,attr"`

	// RetrievalRolePath is the IAM path for the retrieval role
	RetrievalRolePath *string `json:"retrieval_role_path,omitempty" hcl:"retrieval_role_path,attr"`

	// RetrievalRolePermissionsBoundary sets permissions boundary
	RetrievalRolePermissionsBoundary *string `json:"retrieval_role_permissions_boundary,omitempty" hcl:"retrieval_role_permissions_boundary,attr"`

	// SSMParameterConfigurationARN is the SSM parameter ARN
	SSMParameterConfigurationARN *string `json:"ssm_parameter_configuration_arn,omitempty" hcl:"ssm_parameter_configuration_arn,attr"`

	// SSMDocumentConfigurationARN is the SSM document ARN
	SSMDocumentConfigurationARN *string `json:"ssm_document_configuration_arn,omitempty" hcl:"ssm_document_configuration_arn,attr"`

	// S3ConfigurationSourceARN is the S3 object ARN
	S3ConfigurationSourceARN *string `json:"s3_configuration_source_arn,omitempty" hcl:"s3_configuration_source_arn,attr"`

	// SecretsManagerSecretARN is the Secrets Manager secret ARN
	SecretsManagerSecretARN *string `json:"secrets_manager_secret_arn,omitempty" hcl:"secrets_manager_secret_arn,attr"`

	// ================================
	// Deployment Strategy
	// ================================

	// CreateDeploymentStrategy creates a deployment strategy
	CreateDeploymentStrategy *bool `json:"create_deployment_strategy,omitempty" hcl:"create_deployment_strategy,attr"`

	// DeploymentStrategyName is the deployment strategy name
	DeploymentStrategyName *string `json:"deployment_strategy_name,omitempty" hcl:"deployment_strategy_name,attr"`

	// DeploymentStrategyDescription describes the deployment strategy
	DeploymentStrategyDescription *string `json:"deployment_strategy_description,omitempty" hcl:"deployment_strategy_description,attr"`

	// DeploymentDurationInMinutes is the deployment duration (0-1440)
	DeploymentDurationInMinutes *int `json:"deployment_duration_in_minutes,omitempty" hcl:"deployment_duration_in_minutes,attr" validate:"min=0,max=1440"`

	// GrowthFactor is the percentage of targets to receive deployment (1-100)
	GrowthFactor *float64 `json:"growth_factor,omitempty" hcl:"growth_factor,attr" validate:"min=1,max=100"`

	// GrowthType is the growth type
	// Valid values: "LINEAR" | "EXPONENTIAL"
	GrowthType *string `json:"growth_type,omitempty" hcl:"growth_type,attr"`

	// FinalBakeTimeInMinutes is the bake time after deployment (0-1440)
	FinalBakeTimeInMinutes *int `json:"final_bake_time_in_minutes,omitempty" hcl:"final_bake_time_in_minutes,attr" validate:"min=0,max=1440"`

	// ReplicateTo replicates configuration
	// Valid values: "NONE" | "SSM_DOCUMENT"
	ReplicateTo *string `json:"replicate_to,omitempty" hcl:"replicate_to,attr"`

	// ================================
	// Hosted Configuration Version
	// ================================

	// CreateHostedConfigurationVersion creates hosted configuration version
	CreateHostedConfigurationVersion *bool `json:"create_hosted_configuration_version,omitempty" hcl:"create_hosted_configuration_version,attr"`

	// HostedConfigurationVersionContent is the configuration content
	HostedConfigurationVersionContent *string `json:"hosted_configuration_version_content,omitempty" hcl:"hosted_configuration_version_content,attr"`

	// HostedConfigurationVersionContentType is the content type
	HostedConfigurationVersionContentType *string `json:"hosted_configuration_version_content_type,omitempty" hcl:"hosted_configuration_version_content_type,attr"`

	// HostedConfigurationVersionDescription describes the version
	HostedConfigurationVersionDescription *string `json:"hosted_configuration_version_description,omitempty" hcl:"hosted_configuration_version_description,attr"`

	// ================================
	// Extension
	// ================================

	// CreateExtension creates an AppConfig extension
	CreateExtension *bool `json:"create_extension,omitempty" hcl:"create_extension,attr"`

	// ExtensionName is the extension name
	ExtensionName *string `json:"extension_name,omitempty" hcl:"extension_name,attr"`

	// ExtensionDescription describes the extension
	ExtensionDescription *string `json:"extension_description,omitempty" hcl:"extension_description,attr"`

	// ExtensionActions defines extension action points
	ExtensionActions map[string][]ExtensionAction `json:"extension_actions,omitempty" hcl:"extension_actions,attr"`

	// ExtensionParameters are extension parameters
	ExtensionParameters map[string]ExtensionParameter `json:"extension_parameters,omitempty" hcl:"extension_parameters,attr"`
}

// Environment represents an AppConfig environment
type Environment struct {
	// Name of the environment
	Name string `json:"name" hcl:"name,attr"`

	// Description of the environment
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// Monitors are CloudWatch alarms to monitor during deployment
	Monitors []Monitor `json:"monitors,omitempty" hcl:"monitors,attr"`

	// Tags for the environment
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`
}

// Monitor represents a CloudWatch alarm monitor
type Monitor struct {
	// AlarmARN is the CloudWatch alarm ARN
	AlarmARN string `json:"alarm_arn" hcl:"alarm_arn,attr"`

	// AlarmRoleARN is the IAM role ARN to invoke the alarm
	AlarmRoleARN *string `json:"alarm_role_arn,omitempty" hcl:"alarm_role_arn,attr"`
}

// Validator represents a configuration validator
type Validator struct {
	// Type is the validator type
	// Valid values: "JSON_SCHEMA" | "LAMBDA"
	Type string `json:"type" hcl:"type,attr"`

	// Content is the validator content (JSON schema or Lambda ARN)
	Content string `json:"content" hcl:"content,attr"`
}

// ExtensionAction represents an extension action
type ExtensionAction struct {
	// Name of the action
	Name string `json:"name" hcl:"name,attr"`

	// URI of the action (Lambda ARN, SNS topic, SQS queue, EventBridge)
	URI string `json:"uri" hcl:"uri,attr"`

	// RoleARN is the IAM role ARN for the action
	RoleARN *string `json:"role_arn,omitempty" hcl:"role_arn,attr"`

	// Description of the action
	Description *string `json:"description,omitempty" hcl:"description,attr"`
}

// ExtensionParameter represents an extension parameter
type ExtensionParameter struct {
	// Required indicates if the parameter is required
	Required *bool `json:"required,omitempty" hcl:"required,attr"`

	// Description of the parameter
	Description *string `json:"description,omitempty" hcl:"description,attr"`
}

// NewModule creates a new AppConfig module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/appconfig/aws"
	version := "~> 2.0"
	create := true
	locationURI := "hosted"

	return &Module{
		Source:                   source,
		Version:                  version,
		Name:                     &name,
		Create:                   &create,
		ConfigProfileLocationURI: &locationURI,
	}
}

// WithEnvironment adds an environment
func (m *Module) WithEnvironment(name string, env Environment) *Module {
	if m.Environments == nil {
		m.Environments = make(map[string]Environment)
	}
	m.Environments[name] = env
	return m
}

// WithFeatureFlags configures as a feature flag configuration
func (m *Module) WithFeatureFlags(content string) *Module {
	profileType := "AWS.AppConfig.FeatureFlags"
	contentType := "application/json"
	createHosted := true

	m.ConfigProfileType = &profileType
	m.CreateHostedConfigurationVersion = &createHosted
	m.HostedConfigurationVersionContent = &content
	m.HostedConfigurationVersionContentType = &contentType
	return m
}

// WithFreeformConfig configures as a freeform configuration
func (m *Module) WithFreeformConfig(content, contentType string) *Module {
	profileType := "AWS.Freeform"
	createHosted := true

	m.ConfigProfileType = &profileType
	m.CreateHostedConfigurationVersion = &createHosted
	m.HostedConfigurationVersionContent = &content
	m.HostedConfigurationVersionContentType = &contentType
	return m
}

// WithDeploymentStrategy adds a deployment strategy
func (m *Module) WithDeploymentStrategy(durationMin int, growthFactor float64, bakeTimeMin int) *Module {
	create := true
	growthType := "LINEAR"

	m.CreateDeploymentStrategy = &create
	m.DeploymentDurationInMinutes = &durationMin
	m.GrowthFactor = &growthFactor
	m.GrowthType = &growthType
	m.FinalBakeTimeInMinutes = &bakeTimeMin
	return m
}

// WithValidator adds a configuration validator
func (m *Module) WithValidator(validatorType, content string) *Module {
	m.ConfigProfileValidator = append(m.ConfigProfileValidator, Validator{
		Type:    validatorType,
		Content: content,
	})
	return m
}

// WithTags adds tags to the application
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
	return "appconfig"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
