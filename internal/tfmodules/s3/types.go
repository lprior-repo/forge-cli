// Package s3 provides type-safe Terraform module definitions for terraform-aws-modules/s3-bucket/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-s3-bucket v4.0
package s3

// Module represents the terraform-aws-modules/s3-bucket/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// CreateBucket controls if S3 bucket should be created
	CreateBucket *bool `json:"create_bucket,omitempty" hcl:"create_bucket,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to the bucket
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Bucket Configuration
	// ================================

	// Bucket is the name of the bucket (omit for random name)
	Bucket *string `json:"bucket,omitempty" hcl:"bucket,attr"`

	// BucketPrefix creates a unique bucket name with this prefix
	BucketPrefix *string `json:"bucket_prefix,omitempty" hcl:"bucket_prefix,attr"`

	// ACL is the canned ACL to apply (conflicts with Grant)
	ACL *string `json:"acl,omitempty" hcl:"acl,attr"`

	// Policy is a valid bucket policy JSON document
	Policy *string `json:"policy,omitempty" hcl:"policy,attr"`

	// ForceDestroy allows deletion of all objects so bucket can be destroyed
	ForceDestroy *bool `json:"force_destroy,omitempty" hcl:"force_destroy,attr"`

	// ExpectedBucketOwner is the account ID of the expected bucket owner
	ExpectedBucketOwner *string `json:"expected_bucket_owner,omitempty" hcl:"expected_bucket_owner,attr"`

	// ================================
	// Directory Bucket (Express One Zone)
	// ================================

	// IsDirectoryBucket indicates if this is a directory bucket
	IsDirectoryBucket *bool `json:"is_directory_bucket,omitempty" hcl:"is_directory_bucket,attr"`

	// Type is the bucket type (Directory for directory buckets)
	Type *string `json:"type,omitempty" hcl:"type,attr"`

	// DataRedundancy specifies data redundancy (SingleAvailabilityZone)
	DataRedundancy *string `json:"data_redundancy,omitempty" hcl:"data_redundancy,attr"`

	// AvailabilityZoneID is the Availability Zone ID or Local Zone ID
	AvailabilityZoneID *string `json:"availability_zone_id,omitempty" hcl:"availability_zone_id,attr"`

	// LocationType is the location type (AvailabilityZone or LocalZone)
	LocationType *string `json:"location_type,omitempty" hcl:"location_type,attr"`

	// ================================
	// Access Control
	// ================================

	// BlockPublicACLs blocks public ACLs for this bucket
	BlockPublicACLs *bool `json:"block_public_acls,omitempty" hcl:"block_public_acls,attr"`

	// BlockPublicPolicy blocks public bucket policies
	BlockPublicPolicy *bool `json:"block_public_policy,omitempty" hcl:"block_public_policy,attr"`

	// IgnorePublicACLs causes S3 to ignore public ACLs
	IgnorePublicACLs *bool `json:"ignore_public_acls,omitempty" hcl:"ignore_public_acls,attr"`

	// RestrictPublicBuckets restricts public bucket policies
	RestrictPublicBuckets *bool `json:"restrict_public_buckets,omitempty" hcl:"restrict_public_buckets,attr"`

	// SkipDestroyPublicAccessBlock skips destroying public access block config
	SkipDestroyPublicAccessBlock *bool `json:"skip_destroy_public_access_block,omitempty" hcl:"skip_destroy_public_access_block,attr"`

	// ControlObjectOwnership manages S3 Bucket Ownership Controls
	ControlObjectOwnership *bool `json:"control_object_ownership,omitempty" hcl:"control_object_ownership,attr"`

	// ObjectOwnership determines object ownership
	// Valid values: "BucketOwnerEnforced" | "BucketOwnerPreferred" | "ObjectWriter"
	ObjectOwnership *string `json:"object_ownership,omitempty" hcl:"object_ownership,attr"`

	// Grant is an ACL policy grant (conflicts with ACL)
	Grant []Grant `json:"grant,omitempty" hcl:"grant,attr"`

	// Owner is the bucket owner's display name and ID (conflicts with ACL)
	Owner map[string]string `json:"owner,omitempty" hcl:"owner,attr"`

	// ================================
	// Bucket Policies
	// ================================

	// AttachPolicy attaches custom bucket policy
	AttachPolicy *bool `json:"attach_policy,omitempty" hcl:"attach_policy,attr"`

	// AttachPublicPolicy attaches public bucket policy
	AttachPublicPolicy *bool `json:"attach_public_policy,omitempty" hcl:"attach_public_policy,attr"`

	// AttachELBLogDeliveryPolicy attaches ELB log delivery policy
	AttachELBLogDeliveryPolicy *bool `json:"attach_elb_log_delivery_policy,omitempty" hcl:"attach_elb_log_delivery_policy,attr"`

	// AttachLBLogDeliveryPolicy attaches ALB/NLB log delivery policy
	AttachLBLogDeliveryPolicy *bool `json:"attach_lb_log_delivery_policy,omitempty" hcl:"attach_lb_log_delivery_policy,attr"`

	// AttachAccessLogDeliveryPolicy attaches S3 access log delivery policy
	AttachAccessLogDeliveryPolicy *bool `json:"attach_access_log_delivery_policy,omitempty" hcl:"attach_access_log_delivery_policy,attr"`

	// AttachCloudTrailLogDeliveryPolicy attaches CloudTrail log delivery policy
	AttachCloudTrailLogDeliveryPolicy *bool `json:"attach_cloudtrail_log_delivery_policy,omitempty" hcl:"attach_cloudtrail_log_delivery_policy,attr"`

	// AttachWAFLogDeliveryPolicy attaches WAF log delivery policy
	AttachWAFLogDeliveryPolicy *bool `json:"attach_waf_log_delivery_policy,omitempty" hcl:"attach_waf_log_delivery_policy,attr"`

	// AttachDenyInsecureTransportPolicy denies non-SSL transport
	AttachDenyInsecureTransportPolicy *bool `json:"attach_deny_insecure_transport_policy,omitempty" hcl:"attach_deny_insecure_transport_policy,attr"`

	// AttachRequireLatestTLSPolicy requires latest TLS version
	AttachRequireLatestTLSPolicy *bool `json:"attach_require_latest_tls_policy,omitempty" hcl:"attach_require_latest_tls_policy,attr"`

	// AttachInventoryDestinationPolicy attaches inventory destination policy
	AttachInventoryDestinationPolicy *bool `json:"attach_inventory_destination_policy,omitempty" hcl:"attach_inventory_destination_policy,attr"`

	// AttachAnalyticsDestinationPolicy attaches analytics destination policy
	AttachAnalyticsDestinationPolicy *bool `json:"attach_analytics_destination_policy,omitempty" hcl:"attach_analytics_destination_policy,attr"`

	// AttachDenyIncorrectEncryptionHeaders denies incorrect encryption headers
	AttachDenyIncorrectEncryptionHeaders *bool `json:"attach_deny_incorrect_encryption_headers,omitempty" hcl:"attach_deny_incorrect_encryption_headers,attr"`

	// AttachDenyIncorrectKMSKeySSE denies incorrect KMS key SSE
	AttachDenyIncorrectKMSKeySSE *bool `json:"attach_deny_incorrect_kms_key_sse,omitempty" hcl:"attach_deny_incorrect_kms_key_sse,attr"`

	// AllowedKMSKeyARN is the ARN of allowed KMS key
	AllowedKMSKeyARN *string `json:"allowed_kms_key_arn,omitempty" hcl:"allowed_kms_key_arn,attr"`

	// AttachDenyUnencryptedObjectUploads denies unencrypted uploads
	AttachDenyUnencryptedObjectUploads *bool `json:"attach_deny_unencrypted_object_uploads,omitempty" hcl:"attach_deny_unencrypted_object_uploads,attr"`

	// AttachDenySSECEncryptedObjectUploads denies SSEC encrypted uploads
	AttachDenySSECEncryptedObjectUploads *bool `json:"attach_deny_ssec_encrypted_object_uploads,omitempty" hcl:"attach_deny_ssec_encrypted_object_uploads,attr"`

	// ================================
	// Access Logging
	// ================================

	// Logging contains access bucket logging configuration
	Logging map[string]interface{} `json:"logging,omitempty" hcl:"logging,attr"`

	// AccessLogDeliveryPolicySourceBuckets are S3 bucket ARNs allowed to deliver logs
	AccessLogDeliveryPolicySourceBuckets []string `json:"access_log_delivery_policy_source_buckets,omitempty" hcl:"access_log_delivery_policy_source_buckets,attr"`

	// AccessLogDeliveryPolicySourceAccounts are AWS Account IDs allowed to deliver logs
	AccessLogDeliveryPolicySourceAccounts []string `json:"access_log_delivery_policy_source_accounts,omitempty" hcl:"access_log_delivery_policy_source_accounts,attr"`

	// AccessLogDeliveryPolicySourceOrganizations are AWS Organization IDs allowed to deliver logs
	AccessLogDeliveryPolicySourceOrganizations []string `json:"access_log_delivery_policy_source_organizations,omitempty" hcl:"access_log_delivery_policy_source_organizations,attr"`

	// LBLogDeliveryPolicySourceOrganizations are AWS Organization IDs allowed to deliver ALB/NLB logs
	LBLogDeliveryPolicySourceOrganizations []string `json:"lb_log_delivery_policy_source_organizations,omitempty" hcl:"lb_log_delivery_policy_source_organizations,attr"`

	// ================================
	// Versioning & Lifecycle
	// ================================

	// Versioning contains versioning configuration
	Versioning map[string]string `json:"versioning,omitempty" hcl:"versioning,attr"`

	// LifecycleRule contains object lifecycle management configuration
	LifecycleRule []interface{} `json:"lifecycle_rule,omitempty" hcl:"lifecycle_rule,attr"`

	// TransitionDefaultMinimumObjectSize is the default minimum object size behavior
	// Valid values: "all_storage_classes_128K" | "varies_by_storage_class"
	TransitionDefaultMinimumObjectSize *string `json:"transition_default_minimum_object_size,omitempty" hcl:"transition_default_minimum_object_size,attr"`

	// ================================
	// Encryption
	// ================================

	// ServerSideEncryptionConfiguration contains server-side encryption config
	ServerSideEncryptionConfiguration map[string]interface{} `json:"server_side_encryption_configuration,omitempty" hcl:"server_side_encryption_configuration,attr"`

	// ================================
	// Website Hosting
	// ================================

	// Website contains static website hosting or redirect configuration
	Website map[string]string `json:"website,omitempty" hcl:"website,attr"`

	// ================================
	// CORS
	// ================================

	// CORSRule contains Cross-Origin Resource Sharing rules
	CORSRule []interface{} `json:"cors_rule,omitempty" hcl:"cors_rule,attr"`

	// ================================
	// Replication
	// ================================

	// ReplicationConfiguration contains cross-region replication config
	ReplicationConfiguration map[string]interface{} `json:"replication_configuration,omitempty" hcl:"replication_configuration,attr"`

	// ================================
	// Object Lock
	// ================================

	// ObjectLockEnabled enables Object Lock configuration
	ObjectLockEnabled *bool `json:"object_lock_enabled,omitempty" hcl:"object_lock_enabled,attr"`

	// ObjectLockConfiguration contains S3 object locking configuration
	ObjectLockConfiguration map[string]interface{} `json:"object_lock_configuration,omitempty" hcl:"object_lock_configuration,attr"`

	// ================================
	// Performance
	// ================================

	// AccelerationStatus sets the accelerate configuration
	// Valid values: "Enabled" | "Suspended"
	AccelerationStatus *string `json:"acceleration_status,omitempty" hcl:"acceleration_status,attr"`

	// RequestPayer specifies who bears data transfer costs
	// Valid values: "BucketOwner" | "Requester"
	RequestPayer *string `json:"request_payer,omitempty" hcl:"request_payer,attr"`

	// ================================
	// Intelligent Tiering
	// ================================

	// IntelligentTiering contains intelligent tiering configuration
	IntelligentTiering map[string]interface{} `json:"intelligent_tiering,omitempty" hcl:"intelligent_tiering,attr"`

	// ================================
	// Metrics & Analytics
	// ================================

	// MetricConfiguration contains bucket metric configuration
	MetricConfiguration []interface{} `json:"metric_configuration,omitempty" hcl:"metric_configuration,attr"`

	// InventoryConfiguration contains S3 inventory configuration
	InventoryConfiguration map[string]interface{} `json:"inventory_configuration,omitempty" hcl:"inventory_configuration,attr"`

	// InventorySourceAccountID is the inventory source account ID
	InventorySourceAccountID *string `json:"inventory_source_account_id,omitempty" hcl:"inventory_source_account_id,attr"`

	// InventorySourceBucketARN is the inventory source bucket ARN
	InventorySourceBucketARN *string `json:"inventory_source_bucket_arn,omitempty" hcl:"inventory_source_bucket_arn,attr"`

	// InventorySelfSourceDestination indicates if source is also destination
	InventorySelfSourceDestination *bool `json:"inventory_self_source_destination,omitempty" hcl:"inventory_self_source_destination,attr"`

	// AnalyticsConfiguration contains bucket analytics configuration
	AnalyticsConfiguration map[string]interface{} `json:"analytics_configuration,omitempty" hcl:"analytics_configuration,attr"`

	// AnalyticsSourceAccountID is the analytics source account ID
	AnalyticsSourceAccountID *string `json:"analytics_source_account_id,omitempty" hcl:"analytics_source_account_id,attr"`

	// AnalyticsSourceBucketARN is the analytics source bucket ARN
	AnalyticsSourceBucketARN *string `json:"analytics_source_bucket_arn,omitempty" hcl:"analytics_source_bucket_arn,attr"`

	// AnalyticsSelfSourceDestination indicates if source is also destination
	AnalyticsSelfSourceDestination *bool `json:"analytics_self_source_destination,omitempty" hcl:"analytics_self_source_destination,attr"`

	// ================================
	// Directory Bucket Metadata
	// ================================

	// CreateMetadataConfiguration creates metadata configuration resource
	CreateMetadataConfiguration *bool `json:"create_metadata_configuration,omitempty" hcl:"create_metadata_configuration,attr"`

	// MetadataInventoryTableConfigurationState is the inventory table state
	// Valid values: "ENABLED" | "DISABLED"
	MetadataInventoryTableConfigurationState *string `json:"metadata_inventory_table_configuration_state,omitempty" hcl:"metadata_inventory_table_configuration_state,attr"`

	// MetadataEncryptionConfiguration is the encryption configuration block
	MetadataEncryptionConfiguration map[string]interface{} `json:"metadata_encryption_configuration,omitempty" hcl:"metadata_encryption_configuration,attr"`

	// MetadataJournalTableRecordExpirationDays is the number of days to retain journal records
	MetadataJournalTableRecordExpirationDays *int `json:"metadata_journal_table_record_expiration_days,omitempty" hcl:"metadata_journal_table_record_expiration_days,attr"`

	// MetadataJournalTableRecordExpiration is the journal record expiration state
	// Valid values: "ENABLED" | "DISABLED"
	MetadataJournalTableRecordExpiration *string `json:"metadata_journal_table_record_expiration,omitempty" hcl:"metadata_journal_table_record_expiration,attr"`
}

// Grant represents an ACL policy grant
type Grant struct {
	// ID is the canonical user ID
	ID *string `json:"id,omitempty" hcl:"id,attr"`

	// Type is the grantee type
	Type string `json:"type" hcl:"type,attr"`

	// Permissions are the access permissions
	Permissions []string `json:"permissions" hcl:"permissions,attr"`

	// URI is the URI for predefined groups
	URI *string `json:"uri,omitempty" hcl:"uri,attr"`
}

// NewModule creates a new S3 module with sensible defaults
func NewModule(name string) *Module {
	source := "terraform-aws-modules/s3-bucket/aws"
	version := "~> 4.0"
	create := true
	blockPublicACLs := true
	blockPublicPolicy := true
	ignorePublicACLs := true
	restrictPublicBuckets := true
	objectOwnership := "BucketOwnerEnforced"

	return &Module{
		Source:       source,
		Version:      version,
		Bucket:       &name,
		CreateBucket: &create,

		// Secure defaults
		BlockPublicACLs:       &blockPublicACLs,
		BlockPublicPolicy:     &blockPublicPolicy,
		IgnorePublicACLs:      &ignorePublicACLs,
		RestrictPublicBuckets: &restrictPublicBuckets,
		ObjectOwnership:       &objectOwnership,

		// Versioning enabled by default
		Versioning: map[string]string{
			"enabled": "true",
		},

		// Server-side encryption enabled by default
		ServerSideEncryptionConfiguration: map[string]interface{}{
			"rule": map[string]interface{}{
				"apply_server_side_encryption_by_default": map[string]interface{}{
					"sse_algorithm": "AES256",
				},
			},
		},
	}
}

// WithVersioning configures versioning
func (m *Module) WithVersioning(enabled bool) *Module {
	if enabled {
		m.Versioning = map[string]string{"enabled": "true"}
	} else {
		m.Versioning = map[string]string{"enabled": "false"}
	}
	return m
}

// WithEncryption configures KMS encryption
func (m *Module) WithEncryption(kmsKeyARN string) *Module {
	m.ServerSideEncryptionConfiguration = map[string]interface{}{
		"rule": map[string]interface{}{
			"apply_server_side_encryption_by_default": map[string]interface{}{
				"sse_algorithm":     "aws:kms",
				"kms_master_key_id": kmsKeyARN,
			},
		},
	}
	return m
}

// WithPublicAccess allows public access (removes blocks)
func (m *Module) WithPublicAccess() *Module {
	blockFalse := false
	m.BlockPublicACLs = &blockFalse
	m.BlockPublicPolicy = &blockFalse
	m.IgnorePublicACLs = &blockFalse
	m.RestrictPublicBuckets = &blockFalse
	return m
}

// WithWebsite configures static website hosting
func (m *Module) WithWebsite(indexDocument, errorDocument string) *Module {
	m.Website = map[string]string{
		"index_document": indexDocument,
		"error_document": errorDocument,
	}
	return m
}

// WithLogging configures access logging
func (m *Module) WithLogging(targetBucket, targetPrefix string) *Module {
	m.Logging = map[string]interface{}{
		"target_bucket": targetBucket,
		"target_prefix": targetPrefix,
	}
	return m
}

// WithCORS adds CORS rules
func (m *Module) WithCORS(allowedOrigins, allowedMethods, allowedHeaders []string) *Module {
	m.CORSRule = []interface{}{
		map[string]interface{}{
			"allowed_origins": allowedOrigins,
			"allowed_methods": allowedMethods,
			"allowed_headers": allowedHeaders,
		},
	}
	return m
}

// WithLifecycleRule adds lifecycle rules
func (m *Module) WithLifecycleRule(rule map[string]interface{}) *Module {
	m.LifecycleRule = append(m.LifecycleRule, rule)
	return m
}

// WithTags adds tags to the bucket
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
	if m.Bucket != nil {
		return *m.Bucket
	}
	return "s3_bucket"
}

// Configuration generates the HCL configuration for this module
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
