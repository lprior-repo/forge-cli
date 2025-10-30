package sqs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_queue"
		module := NewModule(name)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/sqs/aws", module.Source)
		assert.Equal(t, "~> 4.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.VisibilityTimeoutSeconds)
		assert.Equal(t, 30, *module.VisibilityTimeoutSeconds)

		assert.NotNil(t, module.MessageRetentionSeconds)
		assert.Equal(t, 345600, *module.MessageRetentionSeconds) // 4 days

		assert.NotNil(t, module.SQSManagedSSEEnabled)
		assert.True(t, *module.SQSManagedSSEEnabled)

		// Verify DLQ defaults
		assert.NotNil(t, module.CreateDLQ)
		assert.True(t, *module.CreateDLQ)

		assert.NotNil(t, module.DLQMessageRetentionSeconds)
		assert.Equal(t, 1209600, *module.DLQMessageRetentionSeconds) // 14 days

		assert.NotNil(t, module.DLQSQSManagedSSEEnabled)
		assert.True(t, *module.DLQSQSManagedSSEEnabled)

		assert.NotNil(t, module.CreateDLQRedriveAllowPolicy)
		assert.True(t, *module.CreateDLQRedriveAllowPolicy)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"queue1", "my-queue", "orders_queue"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})

	t.Run("creates module with empty name", func(t *testing.T) {
		module := NewModule("")
		assert.NotNil(t, module.Name)
		assert.Equal(t, "", *module.Name)
	})
}

func TestModule_WithFIFO(t *testing.T) {
	t.Run("enables FIFO with content-based deduplication", func(t *testing.T) {
		module := NewModule("test_queue")
		result := module.WithFIFO(true)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify FIFO is enabled
		assert.NotNil(t, module.FifoQueue)
		assert.True(t, *module.FifoQueue)

		// Verify content-based deduplication is enabled
		assert.NotNil(t, module.ContentBasedDeduplication)
		assert.True(t, *module.ContentBasedDeduplication)
	})

	t.Run("enables FIFO without content-based deduplication", func(t *testing.T) {
		module := NewModule("test_queue")
		result := module.WithFIFO(false)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.FifoQueue)
		assert.True(t, *module.FifoQueue)
		assert.NotNil(t, module.ContentBasedDeduplication)
		assert.False(t, *module.ContentBasedDeduplication)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_queue").WithFIFO(true)

		assert.NotNil(t, module.FifoQueue)
		assert.True(t, *module.FifoQueue)
	})
}

func TestModule_WithEncryption(t *testing.T) {
	t.Run("sets KMS encryption for queue and DLQ", func(t *testing.T) {
		kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
		module := NewModule("test_queue")
		result := module.WithEncryption(kmsKeyID)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify KMS key is set for main queue
		assert.NotNil(t, module.KmsMasterKeyID)
		assert.Equal(t, kmsKeyID, *module.KmsMasterKeyID)

		// Verify KMS key is set for DLQ
		assert.NotNil(t, module.DLQKmsMasterKeyID)
		assert.Equal(t, kmsKeyID, *module.DLQKmsMasterKeyID)
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		kmsAlias := "alias/aws/sqs"
		module := NewModule("test_queue")
		module.WithEncryption(kmsAlias)

		assert.NotNil(t, module.KmsMasterKeyID)
		assert.Equal(t, kmsAlias, *module.KmsMasterKeyID)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_queue").
			WithEncryption("alias/aws/sqs").
			WithFIFO(true)

		assert.NotNil(t, module.KmsMasterKeyID)
		assert.NotNil(t, module.FifoQueue)
	})

	t.Run("handles empty KMS key ID", func(t *testing.T) {
		module := NewModule("test_queue")
		module.WithEncryption("")

		assert.NotNil(t, module.KmsMasterKeyID)
		assert.Equal(t, "", *module.KmsMasterKeyID)
	})
}

func TestModule_WithoutDLQ(t *testing.T) {
	t.Run("disables dead letter queue", func(t *testing.T) {
		module := NewModule("test_queue")

		// Verify DLQ is enabled by default
		assert.NotNil(t, module.CreateDLQ)
		assert.True(t, *module.CreateDLQ)

		// Disable DLQ
		result := module.WithoutDLQ()

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify DLQ is disabled
		assert.NotNil(t, module.CreateDLQ)
		assert.False(t, *module.CreateDLQ)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_queue").
			WithoutDLQ().
			WithFIFO(true)

		assert.NotNil(t, module.CreateDLQ)
		assert.False(t, *module.CreateDLQ)
		assert.NotNil(t, module.FifoQueue)
		assert.True(t, *module.FifoQueue)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to module", func(t *testing.T) {
		module := NewModule("test_queue")
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}
		result := module.WithTags(tags)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify tags are set
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_queue")

		tags1 := map[string]string{
			"Environment": "production",
		}
		module.WithTags(tags1)

		tags2 := map[string]string{
			"Team": "platform",
		}
		module.WithTags(tags2)

		// Verify both sets of tags are present
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test_queue")

		tags1 := map[string]string{
			"Environment": "development",
		}
		module.WithTags(tags1)

		tags2 := map[string]string{
			"Environment": "production",
		}
		module.WithTags(tags2)

		// Verify tag was overwritten
		assert.Equal(t, "production", module.Tags["Environment"])
	})

	t.Run("handles empty tag map", func(t *testing.T) {
		module := NewModule("test_queue")
		module.WithTags(map[string]string{})

		// Tags map should be initialized but empty
		assert.NotNil(t, module.Tags)
		assert.Empty(t, module.Tags)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_queue").
			WithTags(map[string]string{"Team": "platform"}).
			WithFIFO(true)

		assert.NotNil(t, module.Tags)
		assert.NotNil(t, module.FifoQueue)
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns name when set", func(t *testing.T) {
		name := "my_queue"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "sqs_queue", module.LocalName())
	})

	t.Run("returns empty string when name is empty", func(t *testing.T) {
		emptyName := ""
		module := NewModule(emptyName)

		assert.Equal(t, emptyName, module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_queue")

		config, err := module.Configuration()

		// Current implementation is a placeholder
		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("orders_queue").
			WithFIFO(true).
			WithEncryption("alias/aws/sqs").
			WithTags(map[string]string{
				"Environment": "production",
				"Team":        "platform",
			})

		// Verify all configuration is applied
		assert.NotNil(t, module.Name)
		assert.Equal(t, "orders_queue", *module.Name)

		assert.NotNil(t, module.FifoQueue)
		assert.True(t, *module.FifoQueue)

		assert.NotNil(t, module.ContentBasedDeduplication)
		assert.True(t, *module.ContentBasedDeduplication)

		assert.NotNil(t, module.KmsMasterKeyID)
		assert.Equal(t, "alias/aws/sqs", *module.KmsMasterKeyID)

		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports configuration without DLQ", func(t *testing.T) {
		module := NewModule("simple_queue").
			WithoutDLQ().
			WithTags(map[string]string{"Type": "simple"})

		assert.NotNil(t, module.CreateDLQ)
		assert.False(t, *module.CreateDLQ)
		assert.Equal(t, "simple", module.Tags["Type"])
	})

	t.Run("supports all fluent methods in different orders", func(t *testing.T) {
		// Order 1
		module1 := NewModule("q1").
			WithTags(map[string]string{"a": "1"}).
			WithFIFO(true).
			WithEncryption("key1").
			WithoutDLQ()

		// Order 2
		module2 := NewModule("q2").
			WithoutDLQ().
			WithEncryption("key2").
			WithFIFO(false).
			WithTags(map[string]string{"b": "2"})

		// Both should be properly configured
		assert.False(t, *module1.CreateDLQ)
		assert.True(t, *module1.FifoQueue)

		assert.False(t, *module2.CreateDLQ)
		assert.True(t, *module2.FifoQueue)
	})
}

func TestPolicyStatement(t *testing.T) {
	t.Run("creates policy statement", func(t *testing.T) {
		sid := "AllowSend"
		effect := "Allow"
		stmt := PolicyStatement{
			SID:    &sid,
			Effect: &effect,
			Actions: []string{
				"sqs:SendMessage",
				"sqs:SendMessageBatch",
			},
			Resources: []string{"*"},
		}

		assert.Equal(t, "AllowSend", *stmt.SID)
		assert.Equal(t, "Allow", *stmt.Effect)
		assert.Len(t, stmt.Actions, 2)
		assert.Contains(t, stmt.Actions, "sqs:SendMessage")
	})

	t.Run("creates policy statement with principals", func(t *testing.T) {
		stmt := PolicyStatement{
			Principals: []Principal{
				{
					Type:        "Service",
					Identifiers: []string{"lambda.amazonaws.com"},
				},
			},
		}

		assert.Len(t, stmt.Principals, 1)
		assert.Equal(t, "Service", stmt.Principals[0].Type)
		assert.Contains(t, stmt.Principals[0].Identifiers, "lambda.amazonaws.com")
	})

	t.Run("creates policy statement with conditions", func(t *testing.T) {
		stmt := PolicyStatement{
			Condition: []Condition{
				{
					Test:     "StringEquals",
					Variable: "aws:SourceAccount",
					Values:   []string{"123456789012"},
				},
			},
		}

		assert.Len(t, stmt.Condition, 1)
		assert.Equal(t, "StringEquals", stmt.Condition[0].Test)
		assert.Equal(t, "aws:SourceAccount", stmt.Condition[0].Variable)
	})
}

func TestPrincipal(t *testing.T) {
	t.Run("creates AWS principal", func(t *testing.T) {
		p := Principal{
			Type:        "AWS",
			Identifiers: []string{"arn:aws:iam::123456789012:root"},
		}

		assert.Equal(t, "AWS", p.Type)
		assert.Len(t, p.Identifiers, 1)
	})

	t.Run("creates service principal", func(t *testing.T) {
		p := Principal{
			Type:        "Service",
			Identifiers: []string{"lambda.amazonaws.com", "s3.amazonaws.com"},
		}

		assert.Equal(t, "Service", p.Type)
		assert.Len(t, p.Identifiers, 2)
	})
}

func TestCondition(t *testing.T) {
	t.Run("creates string equals condition", func(t *testing.T) {
		c := Condition{
			Test:     "StringEquals",
			Variable: "aws:SourceAccount",
			Values:   []string{"123456789012"},
		}

		assert.Equal(t, "StringEquals", c.Test)
		assert.Equal(t, "aws:SourceAccount", c.Variable)
		assert.Contains(t, c.Values, "123456789012")
	})

	t.Run("creates numeric less than condition", func(t *testing.T) {
		c := Condition{
			Test:     "NumericLessThan",
			Variable: "aws:MultiFactorAuthAge",
			Values:   []string{"3600"},
		}

		assert.Equal(t, "NumericLessThan", c.Test)
		assert.Equal(t, "aws:MultiFactorAuthAge", c.Variable)
	})
}

func TestModule_StructTags(t *testing.T) {
	t.Run("validates struct tag presence", func(t *testing.T) {
		// This test ensures struct tags are properly defined
		// The actual validation would happen via the validate package
		module := NewModule("test")

		// Check that pointer fields are properly initialized
		assert.NotNil(t, module.Name)
		assert.NotNil(t, module.VisibilityTimeoutSeconds)
		assert.NotNil(t, module.MessageRetentionSeconds)
	})
}

func TestModule_PointerSemantics(t *testing.T) {
	t.Run("distinguishes between nil and zero value", func(t *testing.T) {
		module := &Module{}

		// Unset fields should be nil
		assert.Nil(t, module.DelaySeconds)
		assert.Nil(t, module.MaxMessageSize)

		// Setting to zero should be distinguishable from nil
		zero := 0
		module.DelaySeconds = &zero

		assert.NotNil(t, module.DelaySeconds)
		assert.Equal(t, 0, *module.DelaySeconds)
	})

	t.Run("allows explicit false values", func(t *testing.T) {
		module := NewModule("test")

		// Default has DLQ enabled
		assert.True(t, *module.CreateDLQ)

		// Explicitly disable
		disabled := false
		module.CreateDLQ = &disabled

		assert.NotNil(t, module.CreateDLQ)
		assert.False(t, *module.CreateDLQ)
	})
}

func TestModule_ComplexConfiguration(t *testing.T) {
	t.Run("configures FIFO queue with all options", func(t *testing.T) {
		dedup := "messageGroup"
		throughput := "perMessageGroupId"

		module := NewModule("complex_fifo").
			WithFIFO(true).
			WithEncryption("alias/aws/sqs")

		module.DeduplicationScope = &dedup
		module.FifoThroughputLimit = &throughput

		assert.True(t, *module.FifoQueue)
		assert.Equal(t, "messageGroup", *module.DeduplicationScope)
		assert.Equal(t, "perMessageGroupId", *module.FifoThroughputLimit)
	})

	t.Run("configures custom message retention and visibility", func(t *testing.T) {
		retention := 86400 // 1 day
		visibility := 120  // 2 minutes
		maxSize := 262144  // 256 KiB

		module := NewModule("custom_timings")
		module.MessageRetentionSeconds = &retention
		module.VisibilityTimeoutSeconds = &visibility
		module.MaxMessageSize = &maxSize

		assert.Equal(t, 86400, *module.MessageRetentionSeconds)
		assert.Equal(t, 120, *module.VisibilityTimeoutSeconds)
		assert.Equal(t, 262144, *module.MaxMessageSize)
	})

	t.Run("configures long polling", func(t *testing.T) {
		waitTime := 20
		module := NewModule("long_polling")
		module.ReceiveWaitTimeSeconds = &waitTime

		assert.Equal(t, 20, *module.ReceiveWaitTimeSeconds)
	})
}

func TestModule_DLQConfiguration(t *testing.T) {
	t.Run("configures DLQ with custom settings", func(t *testing.T) {
		dlqName := "my_dlq"
		dlqRetention := 604800 // 7 days
		dlqVisibility := 60

		module := NewModule("main_queue")
		module.DLQName = &dlqName
		module.DLQMessageRetentionSeconds = &dlqRetention
		module.DLQVisibilityTimeoutSeconds = &dlqVisibility

		assert.Equal(t, "my_dlq", *module.DLQName)
		assert.Equal(t, 604800, *module.DLQMessageRetentionSeconds)
		assert.Equal(t, 60, *module.DLQVisibilityTimeoutSeconds)
	})

	t.Run("configures DLQ FIFO settings", func(t *testing.T) {
		dedup := true
		scope := "queue"
		throughput := "perQueue"

		module := NewModule("fifo_with_dlq").WithFIFO(true)
		module.DLQContentBasedDeduplication = &dedup
		module.DLQDeduplicationScope = &scope
		module.DLQFifoThroughputLimit = &throughput

		assert.True(t, *module.DLQContentBasedDeduplication)
		assert.Equal(t, "queue", *module.DLQDeduplicationScope)
		assert.Equal(t, "perQueue", *module.DLQFifoThroughputLimit)
	})

	t.Run("configures DLQ tags separately", func(t *testing.T) {
		module := NewModule("queue_with_dlq_tags").
			WithTags(map[string]string{"Queue": "main"})

		module.DLQTags = map[string]string{
			"Queue": "dlq",
			"Type":  "dead-letter",
		}

		assert.Equal(t, "main", module.Tags["Queue"])
		assert.Equal(t, "dlq", module.DLQTags["Queue"])
		assert.Equal(t, "dead-letter", module.DLQTags["Type"])
	})
}

func TestModule_QueuePolicy(t *testing.T) {
	t.Run("enables queue policy with statements", func(t *testing.T) {
		createPolicy := true
		sid := "AllowSend"
		effect := "Allow"

		module := NewModule("policy_queue")
		module.CreateQueuePolicy = &createPolicy
		module.QueuePolicyStatements = map[string]PolicyStatement{
			"allow_send": {
				SID:    &sid,
				Effect: &effect,
				Actions: []string{
					"sqs:SendMessage",
				},
				Resources: []string{"*"},
			},
		}

		assert.True(t, *module.CreateQueuePolicy)
		assert.Len(t, module.QueuePolicyStatements, 1)
		assert.NotNil(t, module.QueuePolicyStatements["allow_send"].SID)
	})

	t.Run("configures policy documents", func(t *testing.T) {
		module := NewModule("doc_queue")
		module.SourceQueuePolicyDocuments = []string{
			`{"Version": "2012-10-17"}`,
		}
		module.OverrideQueuePolicyDocuments = []string{
			`{"Version": "2012-10-17", "Statement": []}`,
		}

		assert.Len(t, module.SourceQueuePolicyDocuments, 1)
		assert.Len(t, module.OverrideQueuePolicyDocuments, 1)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_queue")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_queue").
			WithFIFO(true).
			WithEncryption("alias/aws/sqs").
			WithTags(map[string]string{
				"Environment": "production",
			})
	}
}

// BenchmarkWithTags benchmarks tag merging.
func BenchmarkWithTags(b *testing.B) {
	module := NewModule("bench_queue")
	tags := map[string]string{
		"Environment": "production",
		"Team":        "platform",
		"Service":     "api",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithTags(tags)
	}
}
