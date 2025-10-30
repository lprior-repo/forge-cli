package sns

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_topic"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/sns/aws", module.Source)
		assert.Equal(t, "~> 6.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.CreateTopicPolicy)
		assert.True(t, *module.CreateTopicPolicy)

		assert.NotNil(t, module.EnableDefaultTopicPolicy)
		assert.True(t, *module.EnableDefaultTopicPolicy)

		assert.NotNil(t, module.CreateSubscription)
		assert.True(t, *module.CreateSubscription)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"notifications", "alerts", "events-topic"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})
}

func TestModule_WithFIFO(t *testing.T) {
	t.Run("enables FIFO with content-based deduplication", func(t *testing.T) {
		module := NewModule("test_topic")
		result := module.WithFIFO(true)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.FifoTopic)
		assert.True(t, *module.FifoTopic)
		assert.NotNil(t, module.ContentBasedDeduplication)
		assert.True(t, *module.ContentBasedDeduplication)
	})

	t.Run("enables FIFO without content-based deduplication", func(t *testing.T) {
		module := NewModule("test_topic")
		result := module.WithFIFO(false)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.FifoTopic)
		assert.True(t, *module.FifoTopic)
		assert.NotNil(t, module.ContentBasedDeduplication)
		assert.False(t, *module.ContentBasedDeduplication)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_topic").
			WithFIFO(true).
			WithEncryption("alias/aws/sns")

		assert.NotNil(t, module.FifoTopic)
		assert.NotNil(t, module.KMSMasterKeyID)
	})
}

func TestModule_WithEncryption(t *testing.T) {
	t.Run("configures KMS encryption", func(t *testing.T) {
		kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345"

		module := NewModule("test_topic")
		result := module.WithEncryption(kmsKeyID)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.KMSMasterKeyID)
		assert.Equal(t, kmsKeyID, *module.KMSMasterKeyID)
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		module := NewModule("test_topic")
		module.WithEncryption("alias/aws/sns")

		assert.Equal(t, "alias/aws/sns", *module.KMSMasterKeyID)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithEncryption("alias/aws/sns").
			WithFIFO(false)

		assert.NotNil(t, module.KMSMasterKeyID)
		assert.NotNil(t, module.FifoTopic)
	})
}

func TestModule_WithSubscription(t *testing.T) {
	t.Run("adds a subscription", func(t *testing.T) {
		sub := Subscription{
			Protocol: "sqs",
			Endpoint: "arn:aws:sqs:us-east-1:123456789012:queue",
		}

		module := NewModule("test_topic")
		result := module.WithSubscription("queue_sub", sub)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Subscriptions)
		assert.Len(t, module.Subscriptions, 1)
		assert.Equal(t, "sqs", module.Subscriptions["queue_sub"].Protocol)
		assert.Equal(t, sub.Endpoint, module.Subscriptions["queue_sub"].Endpoint)
	})

	t.Run("adds multiple subscriptions", func(t *testing.T) {
		module := NewModule("test_topic")

		sub1 := Subscription{Protocol: "sqs", Endpoint: "sqs-arn"}
		module.WithSubscription("sqs", sub1)

		sub2 := Subscription{Protocol: "lambda", Endpoint: "lambda-arn"}
		module.WithSubscription("lambda", sub2)

		assert.Len(t, module.Subscriptions, 2)
		assert.Contains(t, module.Subscriptions, "sqs")
		assert.Contains(t, module.Subscriptions, "lambda")
	})

	t.Run("supports email subscription", func(t *testing.T) {
		module := NewModule("test_topic")
		sub := Subscription{
			Protocol: "email",
			Endpoint: "user@example.com",
		}
		module.WithSubscription("email", sub)

		assert.Equal(t, "email", module.Subscriptions["email"].Protocol)
	})
}

func TestModule_WithLambdaSubscription(t *testing.T) {
	t.Run("adds Lambda subscription", func(t *testing.T) {
		lambdaARN := "arn:aws:lambda:us-east-1:123456789012:function:processor"

		module := NewModule("test_topic")
		result := module.WithLambdaSubscription("lambda_sub", lambdaARN)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Subscriptions)
		assert.Len(t, module.Subscriptions, 1)

		sub := module.Subscriptions["lambda_sub"]
		assert.Equal(t, "lambda", sub.Protocol)
		assert.Equal(t, lambdaARN, sub.Endpoint)
	})

	t.Run("supports multiple Lambda subscriptions", func(t *testing.T) {
		module := NewModule("test_topic")

		module.WithLambdaSubscription("lambda1", "arn1")
		module.WithLambdaSubscription("lambda2", "arn2")

		assert.Len(t, module.Subscriptions, 2)
	})
}

func TestModule_WithSQSSubscription(t *testing.T) {
	t.Run("adds SQS subscription with raw message delivery", func(t *testing.T) {
		queueARN := "arn:aws:sqs:us-east-1:123456789012:queue"

		module := NewModule("test_topic")
		result := module.WithSQSSubscription("sqs_sub", queueARN, true)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Subscriptions)

		sub := module.Subscriptions["sqs_sub"]
		assert.Equal(t, "sqs", sub.Protocol)
		assert.Equal(t, queueARN, sub.Endpoint)
		assert.NotNil(t, sub.RawMessageDelivery)
		assert.True(t, *sub.RawMessageDelivery)
	})

	t.Run("adds SQS subscription without raw message delivery", func(t *testing.T) {
		module := NewModule("test_topic")
		module.WithSQSSubscription("sqs_sub", "queue-arn", false)

		sub := module.Subscriptions["sqs_sub"]
		assert.NotNil(t, sub.RawMessageDelivery)
		assert.False(t, *sub.RawMessageDelivery)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the topic", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_topic")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_topic")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test_topic")

		module.WithTags(map[string]string{"Key": "old"})
		module.WithTags(map[string]string{"Key": "new"})

		assert.Equal(t, "new", module.Tags["Key"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns topic name when set", func(t *testing.T) {
		name := "my_topic"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "sns_topic", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_topic")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("notifications").
			WithFIFO(true).
			WithEncryption("alias/aws/sns").
			WithLambdaSubscription("lambda", "lambda-arn").
			WithSQSSubscription("sqs", "sqs-arn", true).
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.Name)
		assert.Equal(t, "notifications", *module.Name)
		assert.True(t, *module.FifoTopic)
		assert.NotNil(t, module.KMSMasterKeyID)
		assert.Len(t, module.Subscriptions, 2)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports standard topic configuration", func(t *testing.T) {
		module := NewModule("alerts").
			WithEncryption("kms-key").
			WithLambdaSubscription("processor", "arn")

		assert.Nil(t, module.FifoTopic)
		assert.NotNil(t, module.KMSMasterKeyID)
	})
}

func TestFeedbackConfig(t *testing.T) {
	t.Run("creates feedback configuration", func(t *testing.T) {
		failureRole := "arn:aws:iam::123456789012:role/failure"
		successRole := "arn:aws:iam::123456789012:role/success"
		sampleRate := 100

		feedback := FeedbackConfig{
			FailureRoleARN:    &failureRole,
			SuccessRoleARN:    &successRole,
			SuccessSampleRate: &sampleRate,
		}

		assert.Equal(t, failureRole, *feedback.FailureRoleARN)
		assert.Equal(t, successRole, *feedback.SuccessRoleARN)
		assert.Equal(t, 100, *feedback.SuccessSampleRate)
	})
}

func TestSubscription(t *testing.T) {
	t.Run("creates SQS subscription with all options", func(t *testing.T) {
		rawDelivery := true
		filterPolicy := `{"type": ["order"]}`
		filterScope := "MessageBody"
		timeout := 5

		sub := Subscription{
			Protocol:                     "sqs",
			Endpoint:                     "arn:aws:sqs:us-east-1:123456789012:queue",
			RawMessageDelivery:           &rawDelivery,
			FilterPolicy:                 &filterPolicy,
			FilterPolicyScope:            &filterScope,
			ConfirmationTimeoutInMinutes: &timeout,
		}

		assert.Equal(t, "sqs", sub.Protocol)
		assert.True(t, *sub.RawMessageDelivery)
		assert.NotNil(t, sub.FilterPolicy)
		assert.Equal(t, "MessageBody", *sub.FilterPolicyScope)
	})

	t.Run("creates Lambda subscription with redrive policy", func(t *testing.T) {
		redrivePolicy := `{"deadLetterTargetArn": "arn:aws:sqs:us-east-1:123456789012:dlq"}`

		sub := Subscription{
			Protocol:      "lambda",
			Endpoint:      "lambda-arn",
			RedrivePolicy: &redrivePolicy,
		}

		assert.Equal(t, "lambda", sub.Protocol)
		assert.NotNil(t, sub.RedrivePolicy)
	})

	t.Run("creates email subscription", func(t *testing.T) {
		autoConfirm := true

		sub := Subscription{
			Protocol:             "email",
			Endpoint:             "user@example.com",
			EndpointAutoConfirms: &autoConfirm,
		}

		assert.Equal(t, "email", sub.Protocol)
		assert.True(t, *sub.EndpointAutoConfirms)
	})
}

func TestPolicyStatement(t *testing.T) {
	t.Run("creates policy statement with principals", func(t *testing.T) {
		sid := "AllowPublish"
		effect := "Allow"

		stmt := PolicyStatement{
			SID:    &sid,
			Effect: &effect,
			Actions: []string{
				"SNS:Publish",
			},
			Resources: []string{"*"},
			Principals: []Principal{
				{
					Type:        "AWS",
					Identifiers: []string{"arn:aws:iam::123456789012:root"},
				},
			},
		}

		assert.Equal(t, "AllowPublish", *stmt.SID)
		assert.Equal(t, "Allow", *stmt.Effect)
		assert.Len(t, stmt.Actions, 1)
		assert.Len(t, stmt.Principals, 1)
	})

	t.Run("creates policy statement with conditions", func(t *testing.T) {
		stmt := PolicyStatement{
			Actions:   []string{"SNS:Subscribe"},
			Resources: []string{"*"},
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
			Identifiers: []string{"lambda.amazonaws.com"},
		}

		assert.Equal(t, "Service", p.Type)
		assert.Contains(t, p.Identifiers, "lambda.amazonaws.com")
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

	t.Run("creates string like condition", func(t *testing.T) {
		c := Condition{
			Test:     "StringLike",
			Variable: "aws:PrincipalOrgID",
			Values:   []string{"o-*"},
		}

		assert.Equal(t, "StringLike", c.Test)
	})
}

func TestModule_PointerSemantics(t *testing.T) {
	t.Run("distinguishes between nil and false", func(t *testing.T) {
		module := &Module{}

		// Unset fields should be nil
		assert.Nil(t, module.FifoTopic)
		assert.Nil(t, module.ContentBasedDeduplication)

		// Setting to false should be distinguishable from nil
		fifo := false
		module.FifoTopic = &fifo

		assert.NotNil(t, module.FifoTopic)
		assert.False(t, *module.FifoTopic)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_topic")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_topic").
			WithFIFO(true).
			WithEncryption("alias/aws/sns").
			WithLambdaSubscription("lambda", "lambda-arn").
			WithTags(map[string]string{"Environment": "production"})
	}
}
