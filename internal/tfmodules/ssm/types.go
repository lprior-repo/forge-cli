// Package ssm provides type-safe Terraform module definitions for terraform-aws-modules/ssm-parameter/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-ssm-parameter v1.0
package ssm

// Module represents the terraform-aws-modules/ssm-parameter/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create determines whether SSM Parameter should be created
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// IgnoreValueChanges ignores changes in parameter value
	IgnoreValueChanges *bool `json:"ignore_value_changes,omitempty" hcl:"ignore_value_changes,attr"`

	// SecureType indicates if the value should be treated as secure
	SecureType *bool `json:"secure_type,omitempty" hcl:"secure_type,attr"`

	// Tags to assign to the parameter
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Parameter Configuration
	// ================================

	// Name of the SSM parameter
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// Description of the parameter
	Description *string `json:"description,omitempty" hcl:"description,attr"`

	// Type of the parameter
	// Valid values: "String" | "StringList" | "SecureString"
	Type *string `json:"type,omitempty" hcl:"type,attr"`

	// Tier to assign to the parameter
	// Valid values: "Standard" | "Advanced" | "Intelligent-Tiering"
	// Note: Downgrading Advanced to Standard recreates the resource
	Tier *string `json:"tier,omitempty" hcl:"tier,attr"`

	// Value of the parameter (for String and SecureString types)
	Value *string `json:"value,omitempty" hcl:"value,attr"`

	// Values is a list for StringList type (will be JSON encoded)
	Values []string `json:"values,omitempty" hcl:"values,attr"`

	// KeyID is the KMS key ID or ARN for encrypting SecureString parameters
	KeyID *string `json:"key_id,omitempty" hcl:"key_id,attr"`

	// AllowedPattern is a regex used to validate the parameter value
	AllowedPattern *string `json:"allowed_pattern,omitempty" hcl:"allowed_pattern,attr"`

	// DataType of the parameter
	// Valid values: "text" | "aws:ssm:integration" | "aws:ec2:image"
	DataType *string `json:"data_type,omitempty" hcl:"data_type,attr"`

	// Overwrite an existing parameter
	// Defaults to false during create, true for subsequent operations
	Overwrite *bool `json:"overwrite,omitempty" hcl:"overwrite,attr"`
}

// NewModule creates a new SSM Parameter module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/ssm-parameter/aws"
	version := "~> 1.0"
	create := true
	paramType := "String"
	tier := "Standard"

	return &Module{
		Source:  source,
		Version: version,
		Name:    &name,
		Create:  &create,
		Type:    &paramType,
		Tier:    &tier,
	}
}

// WithValue sets a string value
func (m *Module) WithValue(value string) *Module {
	m.Value = &value
	return m
}

// WithStringList sets multiple values (StringList type)
func (m *Module) WithStringList(values []string) *Module {
	stringListType := "StringList"
	m.Type = &stringListType
	m.Values = values
	return m
}

// WithSecureString creates a SecureString parameter with KMS encryption
func (m *Module) WithSecureString(value, kmsKeyID string) *Module {
	secureType := "SecureString"
	secure := true
	m.Type = &secureType
	m.SecureType = &secure
	m.Value = &value
	if kmsKeyID != "" {
		m.KeyID = &kmsKeyID
	}
	return m
}

// WithAdvancedTier configures the parameter as Advanced tier
// Required for parameters > 4 KB or high throughput
func (m *Module) WithAdvancedTier() *Module {
	tier := "Advanced"
	m.Tier = &tier
	return m
}

// WithIntelligentTiering enables intelligent tiering
func (m *Module) WithIntelligentTiering() *Module {
	tier := "Intelligent-Tiering"
	m.Tier = &tier
	return m
}

// WithValidation adds a regex pattern for value validation
func (m *Module) WithValidation(pattern string) *Module {
	m.AllowedPattern = &pattern
	return m
}

// WithAMIDataType configures the parameter for AMI IDs
func (m *Module) WithAMIDataType() *Module {
	dataType := "aws:ec2:image"
	m.DataType = &dataType
	return m
}

// WithIgnoreChanges ignores external changes to the parameter value
func (m *Module) WithIgnoreChanges() *Module {
	ignore := true
	m.IgnoreValueChanges = &ignore
	return m
}

// WithTags adds tags to the parameter
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
	return "parameter"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
