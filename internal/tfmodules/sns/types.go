// Package sns provides type-safe Terraform module definitions for terraform-aws-modules/sns/aws.
// Generated from https://github.com/terraform-aws-modules/terraform-aws-sns v6.0
package sns

import (
	"github.com/lewis/forge/internal/tfmodules/hclgen"
)

// Module represents the terraform-aws-modules/sns/aws module.
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

	// Tags to assign to all resources
	Tags map[string]string `json:"tags,omitempty" hcl:"tags,attr"`

	// ================================
	// Topic Configuration
	// ================================

	// Name is the name of the SNS topic to create
	Name *string `json:"name,omitempty" hcl:"name,attr"`

	// UseNamePrefix determines whether name is used as a prefix
	UseNamePrefix *bool `json:"use_name_prefix,omitempty" hcl:"use_name_prefix,attr"`

	// DisplayName is the display name for the SNS topic
	DisplayName *string `json:"display_name,omitempty" hcl:"display_name,attr"`

	// FifoTopic indicates whether to create a FIFO (first-in-first-out) topic
	FifoTopic *bool `json:"fifo_topic,omitempty" hcl:"fifo_topic,attr"`

	// ContentBasedDeduplication enables content-based deduplication for FIFO topics
	ContentBasedDeduplication *bool `json:"content_based_deduplication,omitempty" hcl:"content_based_deduplication,attr"`

	// FifoThroughputScope enables higher throughput for FIFO topics
	// Valid values: "Topic" | "MessageGroup"
	FifoThroughputScope *string `json:"fifo_throughput_scope,omitempty" hcl:"fifo_throughput_scope,attr"`

	// ArchivePolicy is the message archive policy for FIFO topics
	ArchivePolicy *string `json:"archive_policy,omitempty" hcl:"archive_policy,attr"`

	// ================================
	// Delivery & Tracing
	// ================================

	// DeliveryPolicy is the SNS delivery policy
	DeliveryPolicy *string `json:"delivery_policy,omitempty" hcl:"delivery_policy,attr"`

	// TracingConfig is the tracing mode of an Amazon SNS topic
	// Valid values: "PassThrough" | "Active"
	TracingConfig *string `json:"tracing_config,omitempty" hcl:"tracing_config,attr"`

	// SignatureVersion is the hashing algorithm used while creating the signature
	// Valid values: 1 (SHA1) | 2 (SHA256)
	SignatureVersion *int `json:"signature_version,omitempty" hcl:"signature_version,attr"`

	// ================================
	// Encryption
	// ================================

	// KMSMasterKeyID is the ID of an AWS-managed customer master key (CMK)
	KMSMasterKeyID *string `json:"kms_master_key_id,omitempty" hcl:"kms_master_key_id,attr"`

	// ================================
	// Feedback
	// ================================

	// ApplicationFeedback configures application feedback settings
	ApplicationFeedback *FeedbackConfig `json:"application_feedback,omitempty" hcl:"application_feedback,attr"`

	// FirehoseFeedback configures Kinesis Data Firehose feedback settings
	FirehoseFeedback *FeedbackConfig `json:"firehose_feedback,omitempty" hcl:"firehose_feedback,attr"`

	// HTTPFeedback configures HTTP feedback settings
	HTTPFeedback *FeedbackConfig `json:"http_feedback,omitempty" hcl:"http_feedback,attr"`

	// LambdaFeedback configures Lambda feedback settings
	LambdaFeedback *FeedbackConfig `json:"lambda_feedback,omitempty" hcl:"lambda_feedback,attr"`

	// SQSFeedback configures SQS feedback settings
	SQSFeedback *FeedbackConfig `json:"sqs_feedback,omitempty" hcl:"sqs_feedback,attr"`

	// ================================
	// Topic Policy
	// ================================

	// CreateTopicPolicy determines whether an SNS topic policy is created
	CreateTopicPolicy *bool `json:"create_topic_policy,omitempty" hcl:"create_topic_policy,attr"`

	// TopicPolicy is an externally created fully-formed AWS policy as JSON
	TopicPolicy *string `json:"topic_policy,omitempty" hcl:"topic_policy,attr"`

	// EnableDefaultTopicPolicy specifies whether to enable the default topic policy
	EnableDefaultTopicPolicy *bool `json:"enable_default_topic_policy,omitempty" hcl:"enable_default_topic_policy,attr"`

	// SourceTopicPolicyDocuments are IAM policy documents merged together
	SourceTopicPolicyDocuments []string `json:"source_topic_policy_documents,omitempty" hcl:"source_topic_policy_documents,attr"`

	// OverrideTopicPolicyDocuments override statements with same sid
	OverrideTopicPolicyDocuments []string `json:"override_topic_policy_documents,omitempty" hcl:"override_topic_policy_documents,attr"`

	// TopicPolicyStatements is a map of IAM policy statements for custom permissions
	TopicPolicyStatements map[string]PolicyStatement `json:"topic_policy_statements,omitempty" hcl:"topic_policy_statements,attr"`

	// ================================
	// Subscriptions
	// ================================

	// CreateSubscription determines whether SNS subscriptions are created
	CreateSubscription *bool `json:"create_subscription,omitempty" hcl:"create_subscription,attr"`

	// Subscriptions is a map of subscription definitions to create
	Subscriptions map[string]Subscription `json:"subscriptions,omitempty" hcl:"subscriptions,attr"`

	// ================================
	// Data Protection
	// ================================

	// DataProtectionPolicy is the data protection policy JSON
	DataProtectionPolicy *string `json:"data_protection_policy,omitempty" hcl:"data_protection_policy,attr"`
}

// FeedbackConfig represents delivery feedback configuration.
type FeedbackConfig struct {
	// FailureRoleARN is the IAM role ARN for failure feedback
	FailureRoleARN *string `json:"failure_role_arn,omitempty" hcl:"failure_role_arn,attr"`

	// SuccessRoleARN is the IAM role ARN for success feedback
	SuccessRoleARN *string `json:"success_role_arn,omitempty" hcl:"success_role_arn,attr"`

	// SuccessSampleRate is the percentage of successful messages to sample
	SuccessSampleRate *int `json:"success_sample_rate,omitempty" hcl:"success_sample_rate,attr"`
}

// Subscription represents an SNS subscription.
type Subscription struct {
	// Protocol is the subscription protocol (sqs, lambda, email, etc.)
	Protocol string `json:"protocol" hcl:"protocol,attr"`

	// Endpoint is the subscription endpoint (ARN, email, URL, etc.)
	Endpoint string `json:"endpoint" hcl:"endpoint,attr"`

	// ConfirmationTimeoutInMinutes is the timeout for subscription confirmation
	ConfirmationTimeoutInMinutes *int `json:"confirmation_timeout_in_minutes,omitempty" hcl:"confirmation_timeout_in_minutes,attr"`

	// DeliveryPolicy is the JSON delivery policy for the subscription
	DeliveryPolicy *string `json:"delivery_policy,omitempty" hcl:"delivery_policy,attr"`

	// EndpointAutoConfirms indicates if the endpoint auto-confirms
	EndpointAutoConfirms *bool `json:"endpoint_auto_confirms,omitempty" hcl:"endpoint_auto_confirms,attr"`

	// FilterPolicy is the JSON filter policy for the subscription
	FilterPolicy *string `json:"filter_policy,omitempty" hcl:"filter_policy,attr"`

	// FilterPolicyScope is the scope of the filter policy
	// Valid values: "MessageAttributes" | "MessageBody"
	FilterPolicyScope *string `json:"filter_policy_scope,omitempty" hcl:"filter_policy_scope,attr"`

	// RawMessageDelivery enables raw message delivery
	RawMessageDelivery *bool `json:"raw_message_delivery,omitempty" hcl:"raw_message_delivery,attr"`

	// RedrivePolicy is the JSON redrive policy for the subscription
	RedrivePolicy *string `json:"redrive_policy,omitempty" hcl:"redrive_policy,attr"`

	// ReplayPolicy is the JSON replay policy for the subscription
	ReplayPolicy *string `json:"replay_policy,omitempty" hcl:"replay_policy,attr"`

	// SubscriptionRoleARN is the IAM role ARN for the subscription
	SubscriptionRoleARN *string `json:"subscription_role_arn,omitempty" hcl:"subscription_role_arn,attr"`
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

// NewModule creates a new SNS module with sensible defaults.
func NewModule(name string) *Module {
	source := "terraform-aws-modules/sns/aws"
	version := "~> 6.0"
	create := true
	createPolicy := true
	enableDefaultPolicy := true
	createSub := true

	return &Module{
		Source:                   source,
		Version:                  version,
		Name:                     &name,
		Create:                   &create,
		CreateTopicPolicy:        &createPolicy,
		EnableDefaultTopicPolicy: &enableDefaultPolicy,
		CreateSubscription:       &createSub,
	}
}

// WithFIFO configures the topic as a FIFO topic.
func (m *Module) WithFIFO(contentBasedDedup bool) *Module {
	fifo := true
	m.FifoTopic = &fifo
	m.ContentBasedDeduplication = &contentBasedDedup
	return m
}

// WithEncryption configures KMS encryption.
func (m *Module) WithEncryption(kmsKeyID string) *Module {
	m.KMSMasterKeyID = &kmsKeyID
	return m
}

// WithSubscription adds a subscription to the topic.
func (m *Module) WithSubscription(name string, sub Subscription) *Module {
	if m.Subscriptions == nil {
		m.Subscriptions = make(map[string]Subscription)
	}
	m.Subscriptions[name] = sub
	return m
}

// WithLambdaSubscription adds a Lambda subscription.
func (m *Module) WithLambdaSubscription(name, lambdaARN string) *Module {
	sub := Subscription{
		Protocol: "lambda",
		Endpoint: lambdaARN,
	}
	return m.WithSubscription(name, sub)
}

// WithSQSSubscription adds an SQS subscription.
func (m *Module) WithSQSSubscription(name, queueARN string, rawMessageDelivery bool) *Module {
	sub := Subscription{
		Protocol:           "sqs",
		Endpoint:           queueARN,
		RawMessageDelivery: &rawMessageDelivery,
	}
	return m.WithSubscription(name, sub)
}

// WithTags adds tags to the topic.
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
	return "sns_topic"
}

// Configuration generates the HCL configuration for this module.
// PURE: Same module configuration always produces the same HCL output.
func (m *Module) Configuration() (string, error) {
	return hclgen.ToHCL(m.LocalName(), m.Source, m.Version, m)
}
