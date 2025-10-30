package tfmodules_test

import (
	"fmt"
	"testing"

	"github.com/lewis/forge/internal/tfmodules"
	"github.com/lewis/forge/internal/tfmodules/dynamodb"
	"github.com/lewis/forge/internal/tfmodules/s3"
	"github.com/lewis/forge/internal/tfmodules/sns"
	"github.com/lewis/forge/internal/tfmodules/sqs"
)

// TestExampleSQSModule demonstrates using type-safe SQS module
func TestExampleSQSModule(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create an SQS queue with type safety
	queue := sqs.NewModule("orders_queue")

	// Configure with fluent API
	queue.WithFIFO(true).
		WithEncryption("arn:aws:kms:us-east-1:123456789012:key/12345").
		WithTags(map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		})

	// Access is fully type-safe - no string maps!
	fmt.Printf("Queue name: %s\n", *queue.Name)
	fmt.Printf("FIFO enabled: %v\n", *queue.FifoQueue)
	fmt.Printf("Content dedup: %v\n", *queue.ContentBasedDeduplication)

	// Output references are type-safe
	queueARN := tfmodules.NewOutput(queue, "queue_arn")
	fmt.Printf("Reference: %s\n", queueARN.Ref())

	// Output:
	// Queue name: orders_queue
	// FIFO enabled: true
	// Content dedup: true
	// Reference: module.orders_queue.queue_arn
}

// TestExampleStack demonstrates composing multiple modules
func TestExampleStack(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create a stack
	stack := tfmodules.NewStack("my-app")

	// Add SQS queue
	queue := sqs.NewModule("orders_queue")
	queue.WithTags(map[string]string{
		"Application": "my-app",
	})
	stack.AddModule(queue)

	// Add another queue
	dlq := sqs.NewModule("failed_orders")
	dlq.WithoutDLQ() // This one doesn't need its own DLQ
	stack.AddModule(dlq)

	// Validate all modules
	if err := stack.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Stack %s has %d modules\n", stack.Name, len(stack.Modules))

	// Output:
	// Stack my-app has 2 modules
}

// TestExampleModuleOutputs demonstrates using module outputs in other resources
func TestExampleModuleOutputs(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create SQS queue
	queue := sqs.NewModule("orders_queue")

	// Get type-safe output references
	queueURL := tfmodules.NewOutput(queue, "queue_url")
	queueARN := tfmodules.NewOutput(queue, "queue_arn")

	// These can be used in Lambda environment variables, IAM policies, etc.
	fmt.Printf("Queue URL reference: %s\n", queueURL.Ref())
	fmt.Printf("Queue ARN reference: %s\n", queueARN.Ref())

	// Output:
	// Queue URL reference: module.orders_queue.queue_url
	// Queue ARN reference: module.orders_queue.queue_arn
}

// TestModuleTypeSystem demonstrates compile-time type safety
func TestModuleTypeSystem(t *testing.T) {
	// All configuration is strongly typed
	queue := sqs.NewModule("test_queue")

	// These are all checked at compile time:
	visibility := 60
	queue.VisibilityTimeoutSeconds = &visibility

	retention := 86400
	queue.MessageRetentionSeconds = &retention

	fifo := true
	queue.FifoQueue = &fifo

	// Invalid values would be caught by validation
	// (not at compile time, but immediately when set)

	// Verify configuration
	if queue.VisibilityTimeoutSeconds == nil {
		t.Error("VisibilityTimeoutSeconds should be set")
	}

	if *queue.VisibilityTimeoutSeconds != 60 {
		t.Errorf("Expected 60, got %d", *queue.VisibilityTimeoutSeconds)
	}
}

// TestModuleBuilderPattern demonstrates fluent API
func TestModuleBuilderPattern(t *testing.T) {
	// Fluent API for common configurations
	queue := sqs.NewModule("my_queue").
		WithFIFO(true).
		WithEncryption("alias/aws/sqs").
		WithTags(map[string]string{
			"Team": "platform",
		})

	// Verify FIFO was set
	if queue.FifoQueue == nil || !*queue.FifoQueue {
		t.Error("FIFO should be enabled")
	}

	// Verify encryption was set
	if queue.KmsMasterKeyID == nil {
		t.Error("KMS key should be set")
	}

	// Verify tags
	if queue.Tags["Team"] != "platform" {
		t.Error("Tags should be set")
	}
}

// TestModuleDefaults demonstrates sensible defaults
func TestModuleDefaults(t *testing.T) {
	queue := sqs.NewModule("test_queue")

	// NewModule sets sensible defaults
	if queue.VisibilityTimeoutSeconds == nil {
		t.Error("Should have default visibility timeout")
	}

	if *queue.VisibilityTimeoutSeconds != 30 {
		t.Errorf("Default visibility should be 30, got %d", *queue.VisibilityTimeoutSeconds)
	}

	if queue.MessageRetentionSeconds == nil {
		t.Error("Should have default message retention")
	}

	if *queue.MessageRetentionSeconds != 345600 {
		t.Errorf("Default retention should be 4 days (345600), got %d", *queue.MessageRetentionSeconds)
	}

	// DLQ enabled by default
	if queue.CreateDLQ == nil || !*queue.CreateDLQ {
		t.Error("DLQ should be enabled by default")
	}
}

// BenchmarkModuleCreation benchmarks module creation
func BenchmarkModuleCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		queue := sqs.NewModule("bench_queue")
		queue.WithFIFO(true)
		queue.WithTags(map[string]string{"env": "bench"})
	}
}

// TestExampleModuleComparison shows Lingon-style vs map[string]interface{}
func TestExampleModuleComparison(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Old approach (Phase 1): map[string]interface{}
	// Variables: map[string]interface{}{
	//     "visibility_timeout_seconds": 30,
	//     "message_retention_seconds":  345600,
	//     "create_dlq":                 true,
	// }
	// ❌ No compile-time safety
	// ❌ No IDE autocomplete
	// ❌ Easy to make typos

	// New approach (Phase 3): Lingon-style types
	queue := sqs.NewModule("orders_queue")
	visibility := 30
	retention := 345600
	createDLQ := true

	queue.VisibilityTimeoutSeconds = &visibility
	queue.MessageRetentionSeconds = &retention
	queue.CreateDLQ = &createDLQ

	// ✅ Compile-time type checking
	// ✅ Full IDE autocomplete
	// ✅ Self-documenting with struct tags
	// ✅ Validation rules in struct tags

	fmt.Printf("Queue configured: %s\n", *queue.Name)

	// Output:
	// Queue configured: orders_queue
}

// TestExampleDynamoDBModule demonstrates using type-safe DynamoDB module
func TestExampleDynamoDBModule(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create a DynamoDB table with type safety
	table := dynamodb.NewModule("users")

	// Configure with fluent API
	table.WithHashKey("id", "S").
		WithRangeKey("timestamp", "N").
		WithStreams("NEW_AND_OLD_IMAGES").
		WithTTL("expiresAt").
		WithTags(map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		})

	// Add Global Secondary Index
	gsi := dynamodb.GlobalSecondaryIndex{
		Name:           "email-index",
		HashKey:        "email",
		ProjectionType: "ALL",
	}
	table.WithGSI(gsi)

	// Type-safe output references
	tableARN := tfmodules.NewOutput(table, "table_arn")
	streamARN := tfmodules.NewOutput(table, "stream_arn")

	fmt.Printf("Table name: %s\n", *table.Name)
	fmt.Printf("Streams enabled: %v\n", *table.StreamEnabled)
	fmt.Printf("Stream view type: %s\n", *table.StreamViewType)
	fmt.Printf("Table ARN: %s\n", tableARN.Ref())
	fmt.Printf("Stream ARN: %s\n", streamARN.Ref())
}

// TestExampleSNSModule demonstrates using type-safe SNS module
func TestExampleSNSModule(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create an SNS topic with type safety
	topic := sns.NewModule("notifications")

	// Configure with fluent API
	topic.WithFIFO(true).
		WithEncryption("alias/aws/sns").
		WithLambdaSubscription("lambda_sub", "arn:aws:lambda:us-east-1:123456789012:function:processor").
		WithSQSSubscription("sqs_sub", "arn:aws:sqs:us-east-1:123456789012:queue", true).
		WithTags(map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		})

	// Type-safe output references
	topicARN := tfmodules.NewOutput(topic, "topic_arn")

	fmt.Printf("Topic name: %s\n", *topic.Name)
	fmt.Printf("FIFO enabled: %v\n", *topic.FifoTopic)
	fmt.Printf("Subscriptions: %d\n", len(topic.Subscriptions))
	fmt.Printf("Topic ARN: %s\n", topicARN.Ref())
}

// TestExampleS3Module demonstrates using type-safe S3 module
func TestExampleS3Module(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create an S3 bucket with type safety
	bucket := s3.NewModule("my-data-bucket")

	// Configure with fluent API
	bucket.WithVersioning(true).
		WithEncryption("arn:aws:kms:us-east-1:123456789012:key/12345").
		WithLogging("logs-bucket", "my-data-bucket/").
		WithCORS(
			[]string{"https://example.com"},
			[]string{"GET", "PUT", "POST"},
			[]string{"*"},
		).
		WithTags(map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		})

	// Type-safe output references
	bucketID := tfmodules.NewOutput(bucket, "s3_bucket_id")
	bucketARN := tfmodules.NewOutput(bucket, "s3_bucket_arn")

	fmt.Printf("Bucket name: %s\n", *bucket.Bucket)
	fmt.Printf("Versioning: %v\n", bucket.Versioning["enabled"])
	fmt.Printf("Public access blocked: %v\n", *bucket.BlockPublicACLs)
	fmt.Printf("Bucket ID: %s\n", bucketID.Ref())
	fmt.Printf("Bucket ARN: %s\n", bucketARN.Ref())
}

// TestExampleMultiResourceStack demonstrates composing different resource types
func TestExampleMultiResourceStack(t *testing.T) {
	t.Skip("Documentation example - not a real test")
	// Create a complete serverless stack
	stack := tfmodules.NewStack("serverless-app")

	// Add S3 bucket for data
	bucket := s3.NewModule("app-data")
	bucket.WithVersioning(true).WithEncryption("alias/aws/s3")
	stack.AddModule(bucket)

	// Add DynamoDB table
	table := dynamodb.NewModule("app-data")
	table.WithHashKey("id", "S").WithStreams("NEW_AND_OLD_IMAGES")
	stack.AddModule(table)

	// Add SNS topic for notifications
	topic := sns.NewModule("app-notifications")
	topic.WithEncryption("alias/aws/sns")
	stack.AddModule(topic)

	// Add SQS queue for async processing
	queue := sqs.NewModule("app-jobs")
	queue.WithEncryption("alias/aws/sqs")
	stack.AddModule(queue)

	// Validate all modules
	if err := stack.Validate(); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		return
	}

	fmt.Printf("Stack %s has %d modules\n", stack.Name, len(stack.Modules))

	// Output:
	// Stack serverless-app has 4 modules
}
