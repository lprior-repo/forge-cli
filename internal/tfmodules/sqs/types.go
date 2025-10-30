// Package sqs provides type-safe Terraform module definitions for terraform-aws-modules/sqs/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-sqs v4.0
package sqs

// Module represents the terraform-aws-modules/sqs/aws module.
// All fields use pointers to distinguish between "not set" (nil) and "set to zero value".
type Module struct {
	// Source is the Terraform module source
	Source string `json:"source" hcl:"source,attr"`

	// Version is the module version constraint
	Version string `json:"version,omitempty" hcl:"version,attr"`

	// Create determines whether to create SQS queue
	Create *bool `json:"create,omitempty" hcl:"create,attr"`

	// Region where the resource(s) will be managed
	Region *string `json:"region,omitempty" hcl:"region,attr"`

	// Tags to assign to all resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Queue Configuration
	// ================================

	// ContentBasedDeduplication enables content-based deduplication for FIFO queues
	ContentBasedDeduplication *bool `json:"content_based_deduplication,omitempty" hcl:"content_based_deduplication,attr"`

	// DeduplicationScope specifies whether message deduplication occurs at the message group or queue level
	// Valid values: "messageGroup" | "queue"
	DeduplicationScope *string `json:"deduplication_scope,omitempty" hcl:"deduplication_scope,attr"`

	// DelaySeconds is the time in seconds that the delivery of all messages will be delayed
	// Valid range: 0-900 (15 minutes)
	DelaySeconds *int `json:"delay_seconds,omitempty" validate:"min=0,max=900" hcl:"delay_seconds,attr"`

	// FifoQueue designates a FIFO queue
	FifoQueue *bool `json:"fifo_queue,omitempty" hcl:"fifo_queue,attr"`

	// FifoThroughputLimit specifies whether the FIFO queue throughput quota applies to entire queue or per message group
	// Valid values: "perQueue" | "perMessageGroupId"
	FifoThroughputLimit *string `json:"fifo_throughput_limit,omitempty" hcl:"fifo_throughput_limit,attr"`

	// KmsDataKeyReusePeriodSeconds is the length of time for which Amazon SQS can reuse a data key
	// Valid range: 60-86400 (1 minute to 24 hours)
	KmsDataKeyReusePeriodSeconds *int `json:"kms_data_key_reuse_period_seconds,omitempty" validate:"min=60,max=86400" hcl:"kms_data_key_reuse_period_seconds,attr"`

	// KmsMasterKeyID is the ID of an AWS-managed customer master key (CMK) or custom CMK
	KmsMasterKeyID *string `json:"kms_master_key_id,omitempty" hcl:"kms_master_key_id,attr"`

	// MaxMessageSize is the limit of how many bytes a message can contain
	// Valid range: 1024-1048576 (1 KiB to 1024 KiB)
	MaxMessageSize *int `json:"max_message_size,omitempty" validate:"min=1024,max=1048576" hcl:"max_message_size,attr"`

	// MessageRetentionSeconds is the number of seconds Amazon SQS retains a message
	// Valid range: 60-1209600 (1 minute to 14 days)
	MessageRetentionSeconds *int `json:"message_retention_seconds,omitempty" validate:"min=60,max=1209600" hcl:"message_retention_seconds,attr"`

	// Name is the human-readable name of the queue
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// UseNamePrefix determines whether name is used as a prefix
	UseNamePrefix *bool `json:"use_name_prefix,omitempty" hcl:"use_name_prefix,attr"`

	// ReceiveWaitTimeSeconds is the time for which a ReceiveMessage call will wait (long polling)
	// Valid range: 0-20 seconds
	ReceiveWaitTimeSeconds *int `json:"receive_wait_time_seconds,omitempty" validate:"min=0,max=20" hcl:"receive_wait_time_seconds,attr"`

	// RedriveAllowPolicy is the JSON policy to set up the Dead Letter Queue redrive permission
	RedriveAllowPolicy interface{} `json:"redrive_allow_policy,omitempty" hcl:"redrive_allow_policy,attr"`

	// RedrivePolicy is the JSON policy to set up the Dead Letter Queue
	RedrivePolicy interface{} `json:"redrive_policy,omitempty" hcl:"redrive_policy,attr"`

	// SQSManagedSSEEnabled enables server-side encryption with SQS-owned encryption keys
	SQSManagedSSEEnabled *bool `json:"sqs_managed_sse_enabled,omitempty" hcl:"sqs_managed_sse_enabled,attr"`

	// VisibilityTimeoutSeconds is the visibility timeout for the queue
	// Valid range: 0-43200 (12 hours)
	VisibilityTimeoutSeconds *int `json:"visibility_timeout_seconds,omitempty" validate:"min=0,max=43200" hcl:"visibility_timeout_seconds,attr"`

	// ================================
	// Queue Policy
	// ================================

	// CreateQueuePolicy determines whether to create SQS queue policy
	CreateQueuePolicy *bool `json:"create_queue_policy,omitempty" hcl:"create_queue_policy,attr"`

	// SourceQueuePolicyDocuments are IAM policy documents that are merged together
	SourceQueuePolicyDocuments []string `json:"source_queue_policy_documents,omitempty" hcl:"source_queue_policy_documents,attr"`

	// OverrideQueuePolicyDocuments are IAM policy documents that override statements with same sid
	OverrideQueuePolicyDocuments []string `json:"override_queue_policy_documents,omitempty" hcl:"override_queue_policy_documents,attr"`

	// QueuePolicyStatements is a map of IAM policy statements for custom permissions
	QueuePolicyStatements map[string]PolicyStatement `json:"queue_policy_statements,omitempty" hcl:"queue_policy_statements,attr"`

	// ================================
	// Dead Letter Queue
	// ================================

	// CreateDLQ determines whether to create SQS dead letter queue
	CreateDLQ *bool `json:"create_dlq,omitempty" hcl:"create_dlq,attr"`

	// DLQContentBasedDeduplication enables content-based deduplication for DLQ FIFO queues
	DLQContentBasedDeduplication *bool `json:"dlq_content_based_deduplication,omitempty" hcl:"dlq_content_based_deduplication,attr"`

	// DLQDeduplicationScope specifies whether DLQ message deduplication occurs at message group or queue level
	DLQDeduplicationScope *string `json:"dlq_deduplication_scope,omitempty" hcl:"dlq_deduplication_scope,attr"`

	// DLQDelaySeconds is the time in seconds that delivery will be delayed for DLQ
	DLQDelaySeconds *int `json:"dlq_delay_seconds,omitempty" validate:"min=0,max=900" hcl:"dlq_delay_seconds,attr"`

	// DLQKmsDataKeyReusePeriodSeconds is the data key reuse period for DLQ
	DLQKmsDataKeyReusePeriodSeconds *int `json:"dlq_kms_data_key_reuse_period_seconds,omitempty" validate:"min=60,max=86400" hcl:"dlq_kms_data_key_reuse_period_seconds,attr"`

	// DLQKmsMasterKeyID is the KMS key ID for DLQ encryption
	DLQKmsMasterKeyID *string `json:"dlq_kms_master_key_id,omitempty" hcl:"dlq_kms_master_key_id,attr"`

	// DLQMessageRetentionSeconds is the number of seconds DLQ retains a message
	DLQMessageRetentionSeconds *int `json:"dlq_message_retention_seconds,omitempty" validate:"min=60,max=1209600" hcl:"dlq_message_retention_seconds,attr"`

	// DLQName is the human-readable name of the dead letter queue
	DLQName *string `json:"dlq_name,omitempty" hcl:"dlq_name,attr"`

	// DLQReceiveWaitTimeSeconds is the long polling wait time for DLQ
	DLQReceiveWaitTimeSeconds *int `json:"dlq_receive_wait_time_seconds,omitempty" validate:"min=0,max=20" hcl:"dlq_receive_wait_time_seconds,attr"`

	// CreateDLQRedriveAllowPolicy determines whether to create a redrive allow policy for DLQ
	CreateDLQRedriveAllowPolicy *bool `json:"create_dlq_redrive_allow_policy,omitempty" hcl:"create_dlq_redrive_allow_policy,attr"`

	// DLQRedriveAllowPolicy is the JSON policy for DLQ redrive permission
	DLQRedriveAllowPolicy interface{} `json:"dlq_redrive_allow_policy,omitempty" hcl:"dlq_redrive_allow_policy,attr"`

	// DLQSQSManagedSSEEnabled enables SQS-managed encryption for DLQ
	DLQSQSManagedSSEEnabled *bool `json:"dlq_sqs_managed_sse_enabled,omitempty" hcl:"dlq_sqs_managed_sse_enabled,attr"`

	// DLQFifoThroughputLimit specifies DLQ FIFO throughput quota scope
	DLQFifoThroughputLimit *string `json:"dlq_fifo_throughput_limit,omitempty" hcl:"dlq_fifo_throughput_limit,attr"`

	// DLQVisibilityTimeoutSeconds is the visibility timeout for DLQ
	DLQVisibilityTimeoutSeconds *int `json:"dlq_visibility_timeout_seconds,omitempty" validate:"min=0,max=43200" hcl:"dlq_visibility_timeout_seconds,attr"`

	// DLQTags are additional tags to assign to the dead letter queue
	DLQTags map[string]string `json:"dlq_tags,omitempty" hcl:"dlq_tags,attr"`

	// ================================
	// Dead Letter Queue Policy
	// ================================

	// CreateDLQQueuePolicy determines whether to create DLQ queue policy
	CreateDLQQueuePolicy *bool `json:"create_dlq_queue_policy,omitempty" hcl:"create_dlq_queue_policy,attr"`

	// SourceDLQQueuePolicyDocuments are IAM policy documents merged for DLQ
	SourceDLQQueuePolicyDocuments []string `json:"source_dlq_queue_policy_documents,omitempty" hcl:"source_dlq_queue_policy_documents,attr"`

	// OverrideDLQQueuePolicyDocuments override DLQ policy statements with same sid
	OverrideDLQQueuePolicyDocuments []string `json:"override_dlq_queue_policy_documents,omitempty" hcl:"override_dlq_queue_policy_documents,attr"`

	// DLQQueuePolicyStatements is a map of IAM policy statements for DLQ
	DLQQueuePolicyStatements map[string]PolicyStatement `json:"dlq_queue_policy_statements,omitempty" hcl:"dlq_queue_policy_statements,attr"`
}

// PolicyStatement represents an IAM policy statement.
type PolicyStatement struct {
	// SID is the statement ID
	SID *string `json:"sid,omitempty" hcl:"sid,attr"`

	// Actions are the allowed actions
	Actions []string `json:"actions,omitempty" hcl:"actions,attr"`

	// NotActions are the denied actions
	NotActions []string `json:"not_actions,omitempty" hcl:"not_actions,attr"`

	// Effect is Allow or Deny
	Effect *string `json:"effect,omitempty" hcl:"effect,attr"`

	// Resources are the resources this statement applies to
	Resources []string `json:"resources,omitempty" hcl:"resources,attr"`

	// NotResources are the resources this statement does not apply to
	NotResources []string `json:"not_resources,omitempty" hcl:"not_resources,attr"`

	// Principals who can access the resource
	Principals []Principal `json:"principals,omitempty" hcl:"principals,block"`

	// NotPrincipals who cannot access the resource
	NotPrincipals []Principal `json:"not_principals,omitempty" hcl:"not_principals,block"`

	// Condition for conditional access
	Condition []Condition `json:"condition,omitempty" hcl:"condition,block"`
}

// Principal represents an IAM principal.
type Principal struct {
	// Type of principal (AWS, Service, etc.)
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

// NewModule creates a new SQS module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/sqs/aws"
	version := "~> 4.0"
	createDLQ := true
	sqsManagedSSE := true
	dlqSSE := true
	visibility := 30
	retention := 345600     // 4 days
	dlqRetention := 1209600 // 14 days

	return &Module{
		Source:  source,
		Version: version,
		Name:    &name,

		// Sensible defaults
		VisibilityTimeoutSeconds: &visibility,
		MessageRetentionSeconds:  &retention,
		SQSManagedSSEEnabled:     &sqsManagedSSE,

		// DLQ enabled by default
		CreateDLQ:                   &createDLQ,
		DLQMessageRetentionSeconds:  &dlqRetention,
		DLQSQSManagedSSEEnabled:     &dlqSSE,
		CreateDLQRedriveAllowPolicy: &createDLQ,
	}
}

// WithFIFO configures the queue as a FIFO queue.
func (m *Module) WithFIFO(contentBasedDedup bool) *Module {
	fifo := true
	m.FifoQueue = &fifo
	m.ContentBasedDeduplication = &contentBasedDedup
	return m
}

// WithEncryption configures KMS encryption.
func (m *Module) WithEncryption(kmsKeyID string) *Module {
	m.KmsMasterKeyID = &kmsKeyID
	m.DLQKmsMasterKeyID = &kmsKeyID
	return m
}

// WithoutDLQ disables the dead letter queue.
func (m *Module) WithoutDLQ() *Module {
	createDLQ := false
	m.CreateDLQ = &createDLQ
	return m
}

// WithTags adds tags to the queue.
func (m *Module) WithTags(tags map[string]string) *Module {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	for k, v := range tags {
		m.Tags[k] = v
	}
	return m
}

// This is extracted from the Name field if set, or defaults to "sqs_queue".
func (m *Module) LocalName() string {
	if m.Name != nil {
		return *m.Name
	}
	return "sqs_queue"
}

// This is a placeholder - full HCL generation will be implemented later.
func (m *Module) Configuration() (string, error) {
	// TODO: Implement full HCL generation using hclwrite or lingon's marshaling
	return "", nil
}
