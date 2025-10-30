package tfmodules_test

import (
	"fmt"
	"testing"

	"github.com/lewis/forge/internal/tfmodules"
	"github.com/lewis/forge/internal/tfmodules/sqs"
)

// ExampleSQSModule demonstrates using type-safe SQS module
func ExampleSQSModule() {
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

// ExampleStack demonstrates composing multiple modules
func ExampleStack() {
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

// ExampleModuleOutputs demonstrates using module outputs in other resources
func ExampleModuleOutputs() {
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

// ExampleModuleComparison shows Lingon-style vs map[string]interface{}
func ExampleModuleComparison() {
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
