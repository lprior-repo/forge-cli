// Package stepfunctions provides type-safe Terraform module definitions for terraform-aws-modules/step-functions/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-step-functions v4.0
package stepfunctions

// Module represents the terraform-aws-modules/step-functions/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create controls whether Step Function resource should be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// CreateRole controls whether IAM role should be created
	CreateRole *bool `json:"create_role,omitempty" hcl:"create_role,attr"`

	// UseExistingRole uses an existing IAM role
	UseExistingRole *bool `json:"use_existing_role,omitempty" hcl:"use_existing_role,attr"`

	// UseExistingCloudwatchLogGroup uses an existing CloudWatch log group
	UseExistingCloudwatchLogGroup *bool `json:"use_existing_cloudwatch_log_group,omitempty" hcl:"use_existing_cloudwatch_log_group,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to the Step Function
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// State Machine Configuration
	// ================================

	// Name is the name of the Step Function
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Definition is the Amazon States Language definition
	Definition *string `json:"definition,omitempty" hcl:"definition,attr"`

	// Type determines whether a Standard or Express state machine is created
	// Valid values: "STANDARD" | "EXPRESS"
	Type *string `json:"type,omitempty" hcl:"type,attr"`

	// RoleARN is the IAM role ARN to use
	RoleARN *string `json:"role_arn,omitempty" hcl:"role_arn,attr"`

	// Publish determines whether to set a version when created
	Publish *bool `json:"publish,omitempty" hcl:"publish,attr"`

	// ================================
	// Encryption
	// ================================

	// EncryptionConfiguration encrypts data in the State Machine
	EncryptionConfiguration *EncryptionConfiguration `json:"encryption_configuration,omitempty" hcl:"encryption_configuration,attr"`

	// ================================
	// Logging Configuration
	// ================================

	// LoggingConfiguration defines execution history logging
	LoggingConfiguration *LoggingConfiguration `json:"logging_configuration,omitempty" hcl:"logging_configuration,attr"`

	// CloudwatchLogGroupName is the CloudWatch Logs group name
	CloudwatchLogGroupName *string `json:"cloudwatch_log_group_name,omitempty" hcl:"cloudwatch_log_group_name,attr"`

	// CloudwatchLogGroupRetentionInDays is the log retention period
	CloudwatchLogGroupRetentionInDays *int `json:"cloudwatch_log_group_retention_in_days,omitempty" hcl:"cloudwatch_log_group_retention_in_days,attr"`

	// CloudwatchLogGroupKMSKeyID is the KMS key for log encryption
	CloudwatchLogGroupKMSKeyID *string `json:"cloudwatch_log_group_kms_key_id,omitempty" hcl:"cloudwatch_log_group_kms_key_id,attr"`

	// ================================
	// Tracing
	// ================================

	// TracingConfiguration enables X-Ray tracing
	TracingConfiguration *TracingConfiguration `json:"tracing_configuration,omitempty" hcl:"tracing_configuration,attr"`

	// AttachPolicyForLambda attaches policies for Lambda integration
	AttachPolicyForLambda *bool `json:"attach_policy_for_lambda,omitempty" hcl:"attach_policy_for_lambda,attr"`

	// LambdaFunctionARNs are Lambda function ARNs to allow invocation
	LambdaFunctionARNs []string `json:"lambda_function_arns,omitempty" hcl:"lambda_function_arns,attr"`

	// AttachPolicyJSONs attaches custom JSON policies
	AttachPolicyJSONs *bool `json:"attach_policy_jsons,omitempty" hcl:"attach_policy_jsons,attr"`

	// PolicyJSONs is a list of custom IAM policy JSON documents
	PolicyJSONs []string `json:"policy_jsons,omitempty" hcl:"policy_jsons,attr"`

	// AttachPolicies attaches managed policy ARNs
	AttachPolicies *bool `json:"attach_policies,omitempty" hcl:"attach_policies,attr"`

	// Policies is a list of managed policy ARNs
	Policies []string `json:"policies,omitempty" hcl:"policies,attr"`

	// AttachPolicyStatements attaches IAM policy statements
	AttachPolicyStatements *bool `json:"attach_policy_statements,omitempty" hcl:"attach_policy_statements,attr"`

	// PolicyStatements is a list of IAM policy statements
	PolicyStatements []PolicyStatement `json:"policy_statements,omitempty" hcl:"policy_statements,attr"`

	// AttachCloudwatchLogsPolicy attaches CloudWatch Logs policy
	AttachCloudwatchLogsPolicy *bool `json:"attach_cloudwatch_logs_policy,omitempty" hcl:"attach_cloudwatch_logs_policy,attr"`

	// AttachXRayPolicy attaches X-Ray tracing policy
	AttachXRayPolicy *bool `json:"attach_xray_policy,omitempty" hcl:"attach_xray_policy,attr"`

	// Timeouts for Terraform resource management
	Timeouts map[string]string `json:"sfn_state_machine_timeouts,omitempty" hcl:"sfn_state_machine_timeouts,attr"`
}

// EncryptionConfiguration represents encryption settings.
type EncryptionConfiguration struct {
	// Type is the encryption type
	// Valid values: "AWS_OWNED_KEY" | "CUSTOMER_MANAGED_KMS_KEY"
	Type *string `json:"type,omitempty" hcl:"type,attr"`

	// KMSKeyID is the KMS key ID for customer-managed keys
	KMSKeyID *string `json:"kms_key_id,omitempty" hcl:"kms_key_id,attr"`

	// KMSDataKeyReusePeriodSeconds is the data key reuse period (60-900)
	KMSDataKeyReusePeriodSeconds *int `json:"kms_data_key_reuse_period_seconds,omitempty" hcl:"kms_data_key_reuse_period_seconds,attr"`
}

// LoggingConfiguration represents logging settings.
type LoggingConfiguration struct {
	// Level is the logging level
	// Valid values: "ALL" | "ERROR" | "FATAL" | "OFF"
	Level *string `json:"level,omitempty" hcl:"level,attr"`

	// IncludeExecutionData includes execution data in logs
	IncludeExecutionData *bool `json:"include_execution_data,omitempty" hcl:"include_execution_data,attr"`

	// LogDestination is the CloudWatch Logs ARN
	LogDestination *string `json:"log_destination,omitempty" hcl:"log_destination,attr"`
}

// TracingConfiguration represents X-Ray tracing settings.
type TracingConfiguration struct {
	// Enabled indicates if X-Ray tracing is enabled
	Enabled *bool `json:"enabled,omitempty" hcl:"enabled,attr"`
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

// NewModule creates a new Step Functions module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/step-functions/aws"
	version := "~> 4.0"
	create := true
	createRole := true
	sfnType := "STANDARD"
	attachLogs := true

	return &Module{
		Source:                     source,
		Version:                    version,
		Name:                       &name,
		Create:                     &create,
		CreateRole:                 &createRole,
		Type:                       &sfnType,
		AttachCloudwatchLogsPolicy: &attachLogs,
		Timeouts: map[string]string{
			"create": "5m",
			"update": "5m",
			"delete": "5m",
		},
	}
}

// WithDefinition sets the state machine definition.
func (m *Module) WithDefinition(definition string) *Module {
	m.Definition = &definition
	return m
}

// WithExpressType configures as Express workflow.
func (m *Module) WithExpressType() *Module {
	expressType := "EXPRESS"
	m.Type = &expressType
	return m
}

// WithLogging configures CloudWatch Logs.
func (m *Module) WithLogging(level string, includeExecutionData bool) *Module {
	m.LoggingConfiguration = &LoggingConfiguration{
		Level:                &level,
		IncludeExecutionData: &includeExecutionData,
	}
	attachLogs := true
	m.AttachCloudwatchLogsPolicy = &attachLogs
	return m
}

// WithTracing enables X-Ray tracing.
func (m *Module) WithTracing() *Module {
	enabled := true
	m.TracingConfiguration = &TracingConfiguration{
		Enabled: &enabled,
	}
	attachXRay := true
	m.AttachXRayPolicy = &attachXRay
	return m
}

// WithEncryption configures KMS encryption.
func (m *Module) WithEncryption(kmsKeyID string, reusePeriod int) *Module {
	encType := "CUSTOMER_MANAGED_KMS_KEY"
	m.EncryptionConfiguration = &EncryptionConfiguration{
		Type:                         &encType,
		KMSKeyID:                     &kmsKeyID,
		KMSDataKeyReusePeriodSeconds: &reusePeriod,
	}
	return m
}

// WithLambdaIntegration configures Lambda function permissions.
func (m *Module) WithLambdaIntegration(lambdaARNs ...string) *Module {
	attachLambda := true
	m.AttachPolicyForLambda = &attachLambda
	m.LambdaFunctionARNs = append(m.LambdaFunctionARNs, lambdaARNs...)
	return m
}

// WithTags adds tags to the state machine.
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
	return "state_machine"
}

// Configuration generates the HCL configuration for this module.
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
