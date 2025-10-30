// Package dynamodb provides type-safe Terraform module definitions for terraform-aws-modules/dynamodb-table/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-dynamodb-table v4.0
package dynamodb

// Module represents the terraform-aws-modules/dynamodb-table/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// CreateTable controls if DynamoDB table and associated resources are created
	CreateTable *bool `json:"create_table,omitempty" hcl:"create_table,attr"`

	// Region where this resource will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to all resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Table Configuration
	// ================================

	// Name of the DynamoDB table
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// BillingMode controls how you are billed for read/write throughput
	// Valid values: "PROVISIONED" | "PAY_PER_REQUEST"
	BillingMode *string `json:"billing_mode,omitempty" hcl:"billing_mode,attr"`

	// TableClass is the storage class of the table
	// Valid values: "STANDARD" | "STANDARD_INFREQUENT_ACCESS"
	TableClass *string `json:"table_class,omitempty" hcl:"table_class,attr"`

	// DeletionProtectionEnabled enables deletion protection for table
	DeletionProtectionEnabled *bool `json:"deletion_protection_enabled,omitempty" hcl:"deletion_protection_enabled,attr"`

	// ================================
	// Attributes & Keys
	// ================================

	// Attributes is a list of attribute definitions
	// Each attribute has: name (string) and type ("S"|"N"|"B")
	Attributes []Attribute `json:"attributes,omitempty" hcl:"attributes,attr"`

	// HashKey is the attribute to use as the hash (partition) key
	HashKey *string `json:"hash_key,omitempty" hcl:"hash_key,attr"`

	// RangeKey is the attribute to use as the range (sort) key
	RangeKey *string `json:"range_key,omitempty" hcl:"range_key,attr"`

	// ================================
	// Capacity & Scaling
	// ================================

	// WriteCapacity is the number of write units (PROVISIONED mode only)
	WriteCapacity *int `json:"write_capacity,omitempty" hcl:"write_capacity,attr"`

	// ReadCapacity is the number of read units (PROVISIONED mode only)
	ReadCapacity *int `json:"read_capacity,omitempty" hcl:"read_capacity,attr"`

	// OnDemandThroughput sets max read/write units for on-demand tables
	OnDemandThroughput map[string]interface{} `json:"on_demand_throughput,omitempty" hcl:"on_demand_throughput,attr"`

	// WarmThroughput sets warm read/write units
	WarmThroughput map[string]interface{} `json:"warm_throughput,omitempty" hcl:"warm_throughput,attr"`

	// AutoscalingEnabled determines whether to enable autoscaling
	AutoscalingEnabled *bool `json:"autoscaling_enabled,omitempty" hcl:"autoscaling_enabled,attr"`

	// AutoscalingDefaults provides default autoscaling settings
	AutoscalingDefaults map[string]string `json:"autoscaling_defaults,omitempty" hcl:"autoscaling_defaults,attr"`

	// AutoscalingRead configures read autoscaling (max_capacity required)
	AutoscalingRead map[string]string `json:"autoscaling_read,omitempty" hcl:"autoscaling_read,attr"`

	// AutoscalingWrite configures write autoscaling (max_capacity required)
	AutoscalingWrite map[string]string `json:"autoscaling_write,omitempty" hcl:"autoscaling_write,attr"`

	// AutoscalingIndexes configures index-specific autoscaling
	AutoscalingIndexes map[string]map[string]string `json:"autoscaling_indexes,omitempty" hcl:"autoscaling_indexes,attr"`

	// ================================
	// Indexes
	// ================================

	// GlobalSecondaryIndexes describes GSIs for the table
	GlobalSecondaryIndexes []GlobalSecondaryIndex `json:"global_secondary_indexes,omitempty" hcl:"global_secondary_indexes,attr"`

	// LocalSecondaryIndexes describes LSIs on the table (creation only)
	LocalSecondaryIndexes []LocalSecondaryIndex `json:"local_secondary_indexes,omitempty" hcl:"local_secondary_indexes,attr"`

	// IgnoreChangesGlobalSecondaryIndex ignores lifecycle changes to GSIs
	IgnoreChangesGlobalSecondaryIndex *bool `json:"ignore_changes_global_secondary_index,omitempty" hcl:"ignore_changes_global_secondary_index,attr"`

	// ================================
	// Streams
	// ================================

	// StreamEnabled indicates whether Streams are enabled
	StreamEnabled *bool `json:"stream_enabled,omitempty" hcl:"stream_enabled,attr"`

	// StreamViewType determines what information is written to the stream
	// Valid values: "KEYS_ONLY" | "NEW_IMAGE" | "OLD_IMAGE" | "NEW_AND_OLD_IMAGES"
	StreamViewType *string `json:"stream_view_type,omitempty" hcl:"stream_view_type,attr"`

	// ================================
	// Backup & Recovery
	// ================================

	// PointInTimeRecoveryEnabled enables point-in-time recovery
	PointInTimeRecoveryEnabled *bool `json:"point_in_time_recovery_enabled,omitempty" hcl:"point_in_time_recovery_enabled,attr"`

	// PointInTimeRecoveryPeriodInDays is the number of days for continuous backups
	PointInTimeRecoveryPeriodInDays *int `json:"point_in_time_recovery_period_in_days,omitempty" hcl:"point_in_time_recovery_period_in_days,attr"`

	// RestoreDateTime is the point-in-time recovery point to restore
	RestoreDateTime *string `json:"restore_date_time,omitempty" hcl:"restore_date_time,attr"`

	// RestoreSourceName is the name of the table to restore from
	RestoreSourceName *string `json:"restore_source_name,omitempty" hcl:"restore_source_name,attr"`

	// RestoreSourceTableARN is the ARN of the source table (cross-region restores)
	RestoreSourceTableARN *string `json:"restore_source_table_arn,omitempty" hcl:"restore_source_table_arn,attr"`

	// RestoreToLatestTime restores to the most recent recovery point
	RestoreToLatestTime *bool `json:"restore_to_latest_time,omitempty" hcl:"restore_to_latest_time,attr"`

	// ================================
	// TTL
	// ================================

	// TTLEnabled indicates whether TTL is enabled
	TTLEnabled *bool `json:"ttl_enabled,omitempty" hcl:"ttl_enabled,attr"`

	// TTLAttributeName is the attribute to store the TTL timestamp
	TTLAttributeName *string `json:"ttl_attribute_name,omitempty" hcl:"ttl_attribute_name,attr"`

	// ================================
	// Encryption
	// ================================

	// ServerSideEncryptionEnabled enables encryption at rest
	ServerSideEncryptionEnabled *bool `json:"server_side_encryption_enabled,omitempty" hcl:"server_side_encryption_enabled,attr"`

	// ServerSideEncryptionKMSKeyARN is the ARN of the KMS key for encryption
	ServerSideEncryptionKMSKeyARN *string `json:"server_side_encryption_kms_key_arn,omitempty" hcl:"server_side_encryption_kms_key_arn,attr"`

	// ================================
	// Global Tables
	// ================================

	// ReplicaRegions are region names for creating replicas
	ReplicaRegions []ReplicaRegion `json:"replica_regions,omitempty" hcl:"replica_regions,attr"`

	// ================================
	// Import & Policy
	// ================================

	// ImportTable configures importing S3 data into a new table
	ImportTable map[string]interface{} `json:"import_table,omitempty" hcl:"import_table,attr"`

	// ResourcePolicy is the JSON definition of the resource-based policy
	ResourcePolicy *string `json:"resource_policy,omitempty" hcl:"resource_policy,attr"`

	// ================================
	// Timeouts
	// ================================

	// Timeouts for Terraform resource management
	Timeouts map[string]string `json:"timeouts,omitempty" hcl:"timeouts,attr"`
}

// Attribute represents a DynamoDB table attribute
type Attribute struct {
	// Name of the attribute
	Name string `json:"name" hcl:"name,attr"`

	// Type of the attribute (S, N, or B)
	Type string `json:"type" hcl:"type,attr"`
}

// GlobalSecondaryIndex represents a GSI configuration
type GlobalSecondaryIndex struct {
	// Name of the GSI
	Name string `json:"name" hcl:"name,attr"`

	// HashKey for the GSI
	HashKey string `json:"hash_key" hcl:"hash_key,attr"`

	// RangeKey for the GSI (optional)
	RangeKey *string `json:"range_key,omitempty" hcl:"range_key,attr"`

	// ProjectionType (ALL, KEYS_ONLY, INCLUDE)
	ProjectionType string `json:"projection_type" hcl:"projection_type,attr"`

	// NonKeyAttributes for INCLUDE projection
	NonKeyAttributes []string `json:"non_key_attributes,omitempty" hcl:"non_key_attributes,attr"`

	// WriteCapacity for PROVISIONED mode
	WriteCapacity *int `json:"write_capacity,omitempty" hcl:"write_capacity,attr"`

	// ReadCapacity for PROVISIONED mode
	ReadCapacity *int `json:"read_capacity,omitempty" hcl:"read_capacity,attr"`
}

// LocalSecondaryIndex represents an LSI configuration
type LocalSecondaryIndex struct {
	// Name of the LSI
	Name string `json:"name" hcl:"name,attr"`

	// RangeKey for the LSI
	RangeKey string `json:"range_key" hcl:"range_key,attr"`

	// ProjectionType (ALL, KEYS_ONLY, INCLUDE)
	ProjectionType string `json:"projection_type" hcl:"projection_type,attr"`

	// NonKeyAttributes for INCLUDE projection
	NonKeyAttributes []string `json:"non_key_attributes,omitempty" hcl:"non_key_attributes,attr"`
}

// ReplicaRegion represents a global table replica configuration
type ReplicaRegion struct {
	// RegionName for the replica
	RegionName string `json:"region_name" hcl:"region_name,attr"`

	// KMSKeyARN for replica encryption
	KMSKeyARN *string `json:"kms_key_arn,omitempty" hcl:"kms_key_arn,attr"`

	// PropagateTagsToStreams propagates tags to replica streams
	PropagateTagsToStreams *bool `json:"propagate_tags,omitempty" hcl:"propagate_tags,attr"`

	// PointInTimeRecoveryEnabled for the replica
	PointInTimeRecoveryEnabled *bool `json:"point_in_time_recovery,omitempty" hcl:"point_in_time_recovery,attr"`
}

// NewModule creates a new DynamoDB module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/dynamodb-table/aws"
	version := "~> 4.0"
	billingMode := "PAY_PER_REQUEST"
	pitrEnabled := true
	create := true

	return &Module{
		Source:      source,
		Version:     version,
		Name:        &name,
		CreateTable: &create,

		// Sensible defaults
		BillingMode:                 &billingMode,
		PointInTimeRecoveryEnabled:  &pitrEnabled,
		ServerSideEncryptionEnabled: &pitrEnabled,
		DeletionProtectionEnabled:   &pitrEnabled,
		Timeouts: map[string]string{
			"create": "10m",
			"update": "60m",
			"delete": "10m",
		},
	}
}

// WithHashKey sets the partition key
func (m *Module) WithHashKey(name, attrType string) *Module {
	m.HashKey = &name
	m.Attributes = append(m.Attributes, Attribute{Name: name, Type: attrType})
	return m
}

// WithRangeKey sets the sort key
func (m *Module) WithRangeKey(name, attrType string) *Module {
	m.RangeKey = &name
	m.Attributes = append(m.Attributes, Attribute{Name: name, Type: attrType})
	return m
}

// WithStreams enables DynamoDB Streams
func (m *Module) WithStreams(viewType string) *Module {
	enabled := true
	m.StreamEnabled = &enabled
	m.StreamViewType = &viewType
	return m
}

// WithGSI adds a Global Secondary Index
func (m *Module) WithGSI(gsi GlobalSecondaryIndex) *Module {
	m.GlobalSecondaryIndexes = append(m.GlobalSecondaryIndexes, gsi)
	return m
}

// WithTTL enables Time To Live
func (m *Module) WithTTL(attributeName string) *Module {
	enabled := true
	m.TTLEnabled = &enabled
	m.TTLAttributeName = &attributeName
	return m
}

// WithEncryption configures KMS encryption
func (m *Module) WithEncryption(kmsKeyARN string) *Module {
	enabled := true
	m.ServerSideEncryptionEnabled = &enabled
	m.ServerSideEncryptionKMSKeyARN = &kmsKeyARN
	return m
}

// WithProvisioned configures PROVISIONED billing mode
func (m *Module) WithProvisioned(readCapacity, writeCapacity int) *Module {
	mode := "PROVISIONED"
	m.BillingMode = &mode
	m.ReadCapacity = &readCapacity
	m.WriteCapacity = &writeCapacity
	return m
}

// WithTags adds tags to the table
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
	return "dynamodb_table"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
