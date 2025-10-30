// Package secretsmanager provides type-safe Terraform module definitions for terraform-aws-modules/secrets-manager/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-secrets-manager v1.0
package secretsmanager

// Module represents the terraform-aws-modules/secrets-manager/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create determines whether resources will be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to add to all resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Secret Configuration
	// ================================

	// Name is the friendly name of the secret
	// Can contain: uppercase/lowercase letters, digits, /_+=.@-
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// NamePrefix creates a unique name beginning with the prefix
	NamePrefix *string `json:"name_prefix,omitempty" hcl:"name_prefix,attr"`

	// Description of the secret
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// KMSKeyID is the ARN or ID of the AWS KMS key for encryption
	// Defaults to aws/secretsmanager if not specified
	KMSKeyID *string `json:"kms_key_id,omitempty" hcl:"kms_key_id,attr"`

	// RecoveryWindowInDays is the number of days before deletion (0-30)
	// 0 = force deletion without recovery, 7-30 = recovery window, default = 30
	RecoveryWindowInDays *int `json:"recovery_window_in_days,omitempty" validate:"min=0,max=30" hcl:"recovery_window_in_days,attr"`

	// ForceOverwriteReplicaSecret overwrites secret with same name in destination region
	ForceOverwriteReplicaSecret *bool `json:"force_overwrite_replica_secret,omitempty" hcl:"force_overwrite_replica_secret,attr"`

	// ================================
	// Secret Value
	// ================================

	// SecretString is text data to encrypt and store
	SecretString *string `json:"secret_string,omitempty" hcl:"secret_string,attr"`

	// SecretBinary is binary data to encrypt and store (base64 encoded)
	SecretBinary *string `json:"secret_binary,omitempty" hcl:"secret_binary,attr"`

	// IgnoreSecretChanges ignores external changes to secret_string or secret_binary
	IgnoreSecretChanges *bool `json:"ignore_secret_changes,omitempty" hcl:"ignore_secret_changes,attr"`

	// ================================
	// Replication
	// ================================

	// Replica configures secret replication to other regions
	Replica map[string]Replica `json:"replica,omitempty" hcl:"replica,attr"`

	// ================================
	// Resource Policy
	// ================================

	// CreatePolicy determines whether a resource policy will be created
	CreatePolicy *bool `json:"create_policy,omitempty" hcl:"create_policy,attr"`

	// SourcePolicyDocuments are IAM policy documents merged together
	SourcePolicyDocuments []string `json:"source_policy_documents,omitempty" hcl:"source_policy_documents,attr"`

	// OverridePolicyDocuments override statements with same sid
	OverridePolicyDocuments []string `json:"override_policy_documents,omitempty" hcl:"override_policy_documents,attr"`

	// PolicyStatements is a map of IAM policy statements
	PolicyStatements map[string]PolicyStatement `json:"policy_statements,omitempty" hcl:"policy_statements,attr"`

	// BlockPublicPolicy validates policy to prevent broad access
	BlockPublicPolicy *bool `json:"block_public_policy,omitempty" hcl:"block_public_policy,attr"`

	// ================================
	// Rotation
	// ================================

	// EnableRotation enables automatic rotation
	EnableRotation *bool `json:"enable_rotation,omitempty" hcl:"enable_rotation,attr"`

	// RotationLambdaARN is the Lambda function ARN for rotation
	RotationLambdaARN *string `json:"rotation_lambda_arn,omitempty" hcl:"rotation_lambda_arn,attr"`

	// RotationRules configures rotation schedule
	RotationRules *RotationRules `json:"rotation_rules,omitempty" hcl:"rotation_rules,attr"`

	// RotateImmediately rotates the secret immediately upon creation
	RotateImmediately *bool `json:"rotate_immediately,omitempty" hcl:"rotate_immediately,attr"`
}

// Replica represents secret replication configuration.
type Replica struct {
	// Region is the destination region (defaults to map key if not set)
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// KMSKeyID is the KMS key for encryption in replica region
	KMSKeyID *string `json:"kms_key_id,omitempty" hcl:"kms_key_id,attr"`
}

// PolicyStatement represents an IAM policy statement.
type PolicyStatement struct {
	// SID is the statement ID
	SID *string `json:"sid,omitempty" hcl:"sid,attr"`

	// Effect is Allow or Deny
	Effect *string `json:"effect,omitempty" hcl:"effect,attr"`

	// Actions are the allowed actions
	Actions []string `json:"actions,omitempty" hcl:"actions,attr"`

	// NotActions are the denied actions
	NotActions []string `json:"not_actions,omitempty" hcl:"not_actions,attr"`

	// Resources are the resources this statement applies to
	Resources []string `json:"resources,omitempty" hcl:"resources,attr"`

	// NotResources are the resources this statement does not apply to
	NotResources []string `json:"not_resources,omitempty" hcl:"not_resources,attr"`

	// Principals who can access the secret
	Principals []Principal `json:"principals,omitempty" hcl:"principals,block"`

	// NotPrincipals who cannot access the secret
	NotPrincipals []Principal `json:"not_principals,omitempty" hcl:"not_principals,block"`

	// Condition for conditional access
	Condition []Condition `json:"condition,omitempty" hcl:"condition,block"`
}

// Principal represents an IAM principal.
type Principal struct {
	// Type of principal (AWS, Service, Federated, etc.)
	Type string `json:"type" hcl:"type,attr"`

	// Identifiers (ARNs, account IDs, etc.)
	Identifiers []string `json:"identifiers" hcl:"identifiers,attr"`
}

// Condition represents an IAM condition.
type Condition struct {
	// Test is the condition operator
	Test string `json:"test" hcl:"test,attr"`

	// Variable is the condition key
	Variable string `json:"variable" hcl:"variable,attr"`

	// Values are the condition values
	Values []string `json:"values" hcl:"values,attr"`
}

// RotationRules represents automatic rotation configuration.
type RotationRules struct {
	// AutomaticallyAfterDays is the number of days between rotations (1-365)
	AutomaticallyAfterDays *int `json:"automatically_after_days,omitempty" validate:"min=1,max=365" hcl:"automatically_after_days,attr"`

	// Duration is the rotation window duration (e.g., "3h", "2h30m")
	Duration *string `json:"duration,omitempty" hcl:"duration,attr"`

	// ScheduleExpression is a cron expression for rotation schedule
	ScheduleExpression *string `json:"schedule_expression,omitempty" hcl:"schedule_expression,attr"`
}

// NewModule creates a new Secrets Manager module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/secrets-manager/aws"
	version := "~> 1.0"
	create := true
	recoveryWindow := 30
	blockPublic := true

	return &Module{
		Source:               source,
		Version:              version,
		Name:                 &name,
		Create:               &create,
		RecoveryWindowInDays: &recoveryWindow,
		BlockPublicPolicy:    &blockPublic,
	}
}

// WithSecretString sets the secret value as a string.
func (m *Module) WithSecretString(value string) *Module {
	m.SecretString = &value
	return m
}

// WithSecretJSON sets the secret value as JSON (common for multiple values).
func (m *Module) WithSecretJSON(jsonString string) *Module {
	m.SecretString = &jsonString
	return m
}

// WithKMSKey configures customer-managed KMS encryption.
func (m *Module) WithKMSKey(kmsKeyID string) *Module {
	m.KMSKeyID = &kmsKeyID
	return m
}

// WithRecoveryWindow sets the recovery window (0 for immediate deletion, 7-30 for recovery).
func (m *Module) WithRecoveryWindow(days int) *Module {
	m.RecoveryWindowInDays = &days
	return m
}

// WithReplication adds replication to another region.
func (m *Module) WithReplication(region, kmsKeyID string) *Module {
	if m.Replica == nil {
		m.Replica = make(map[string]Replica)
	}
	m.Replica[region] = Replica{
		Region:   &region,
		KMSKeyID: &kmsKeyID,
	}
	return m
}

// WithRotation enables automatic secret rotation.
func (m *Module) WithRotation(lambdaARN string, daysInterval int) *Module {
	enabled := true
	m.EnableRotation = &enabled
	m.RotationLambdaARN = &lambdaARN
	m.RotationRules = &RotationRules{
		AutomaticallyAfterDays: &daysInterval,
	}
	return m
}

// WithPolicy adds a resource policy statement.
func (m *Module) WithPolicy(statementID string, statement PolicyStatement) *Module {
	createPolicy := true
	m.CreatePolicy = &createPolicy
	if m.PolicyStatements == nil {
		m.PolicyStatements = make(map[string]PolicyStatement)
	}
	m.PolicyStatements[statementID] = statement
	return m
}

// WithTags adds tags to the secret.
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
	return "secret"
}

// Configuration generates the HCL configuration for this module.
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
