package stepfunctions

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_state_machine"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/step-functions/aws", module.Source)
		assert.Equal(t, "~> 4.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.CreateRole)
		assert.True(t, *module.CreateRole)

		assert.NotNil(t, module.Type)
		assert.Equal(t, "STANDARD", *module.Type)

		assert.NotNil(t, module.AttachCloudwatchLogsPolicy)
		assert.True(t, *module.AttachCloudwatchLogsPolicy)

		assert.NotNil(t, module.Timeouts)
		assert.Equal(t, "5m", module.Timeouts["create"])
		assert.Equal(t, "5m", module.Timeouts["update"])
		assert.Equal(t, "5m", module.Timeouts["delete"])
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"workflow1", "order-processor", "data_pipeline"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})
}

func TestModule_WithDefinition(t *testing.T) {
	t.Run("sets state machine definition", func(t *testing.T) {
		definition := `{
			"Comment": "A Hello World example",
			"StartAt": "HelloWorld",
			"States": {
				"HelloWorld": {
					"Type": "Pass",
					"Result": "Hello World!",
					"End": true
				}
			}
		}`

		module := NewModule("test_state_machine")
		result := module.WithDefinition(definition)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Definition)
		assert.Contains(t, *module.Definition, "HelloWorld")
		assert.Contains(t, *module.Definition, "Pass")
	})

	t.Run("supports complex workflows", func(t *testing.T) {
		definition := `{
			"StartAt": "Choice",
			"States": {
				"Choice": {
					"Type": "Choice",
					"Choices": [
						{
							"Variable": "$.type",
							"StringEquals": "order",
							"Next": "ProcessOrder"
						}
					],
					"Default": "DefaultState"
				}
			}
		}`

		module := NewModule("test")
		module.WithDefinition(definition)

		assert.Contains(t, *module.Definition, "Choice")
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithDefinition(`{"StartAt": "Pass"}`).
			WithLogging("ALL", true)

		assert.NotNil(t, module.Definition)
		assert.NotNil(t, module.LoggingConfiguration)
	})
}

func TestModule_WithExpressType(t *testing.T) {
	t.Run("configures Express workflow type", func(t *testing.T) {
		module := NewModule("test_state_machine")
		result := module.WithExpressType()

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Type)
		assert.Equal(t, "EXPRESS", *module.Type)
	})

	t.Run("can be chained with other methods", func(t *testing.T) {
		module := NewModule("test").
			WithExpressType().
			WithTracing()

		assert.Equal(t, "EXPRESS", *module.Type)
		assert.NotNil(t, module.TracingConfiguration)
	})
}

func TestModule_WithLogging(t *testing.T) {
	t.Run("configures CloudWatch Logs with ALL level", func(t *testing.T) {
		module := NewModule("test_state_machine")
		result := module.WithLogging("ALL", true)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.LoggingConfiguration)
		assert.NotNil(t, module.LoggingConfiguration.Level)
		assert.Equal(t, "ALL", *module.LoggingConfiguration.Level)
		assert.NotNil(t, module.LoggingConfiguration.IncludeExecutionData)
		assert.True(t, *module.LoggingConfiguration.IncludeExecutionData)
		assert.NotNil(t, module.AttachCloudwatchLogsPolicy)
		assert.True(t, *module.AttachCloudwatchLogsPolicy)
	})

	t.Run("supports ERROR level", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("ERROR", false)

		assert.Equal(t, "ERROR", *module.LoggingConfiguration.Level)
		assert.False(t, *module.LoggingConfiguration.IncludeExecutionData)
	})

	t.Run("supports FATAL level", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("FATAL", true)

		assert.Equal(t, "FATAL", *module.LoggingConfiguration.Level)
	})

	t.Run("supports OFF level", func(t *testing.T) {
		module := NewModule("test")
		module.WithLogging("OFF", false)

		assert.Equal(t, "OFF", *module.LoggingConfiguration.Level)
	})
}

func TestModule_WithTracing(t *testing.T) {
	t.Run("enables X-Ray tracing", func(t *testing.T) {
		module := NewModule("test_state_machine")
		result := module.WithTracing()

		assert.Equal(t, module, result)
		assert.NotNil(t, module.TracingConfiguration)
		assert.NotNil(t, module.TracingConfiguration.Enabled)
		assert.True(t, *module.TracingConfiguration.Enabled)
		assert.NotNil(t, module.AttachXRayPolicy)
		assert.True(t, *module.AttachXRayPolicy)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithTracing().
			WithLogging("ALL", true)

		assert.True(t, *module.TracingConfiguration.Enabled)
		assert.NotNil(t, module.LoggingConfiguration)
	})
}

func TestModule_WithEncryption(t *testing.T) {
	t.Run("configures KMS encryption", func(t *testing.T) {
		kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345"
		reusePeriod := 300

		module := NewModule("test_state_machine")
		result := module.WithEncryption(kmsKeyID, reusePeriod)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.EncryptionConfiguration)
		assert.NotNil(t, module.EncryptionConfiguration.Type)
		assert.Equal(t, "CUSTOMER_MANAGED_KMS_KEY", *module.EncryptionConfiguration.Type)
		assert.NotNil(t, module.EncryptionConfiguration.KMSKeyID)
		assert.Equal(t, kmsKeyID, *module.EncryptionConfiguration.KMSKeyID)
		assert.NotNil(t, module.EncryptionConfiguration.KMSDataKeyReusePeriodSeconds)
		assert.Equal(t, reusePeriod, *module.EncryptionConfiguration.KMSDataKeyReusePeriodSeconds)
	})

	t.Run("supports minimum reuse period", func(t *testing.T) {
		module := NewModule("test")
		module.WithEncryption("kms-key", 60)

		assert.Equal(t, 60, *module.EncryptionConfiguration.KMSDataKeyReusePeriodSeconds)
	})

	t.Run("supports maximum reuse period", func(t *testing.T) {
		module := NewModule("test")
		module.WithEncryption("kms-key", 900)

		assert.Equal(t, 900, *module.EncryptionConfiguration.KMSDataKeyReusePeriodSeconds)
	})
}

func TestModule_WithLambdaIntegration(t *testing.T) {
	t.Run("configures Lambda function permissions", func(t *testing.T) {
		lambdaARNs := []string{
			"arn:aws:lambda:us-east-1:123456789012:function:step1",
			"arn:aws:lambda:us-east-1:123456789012:function:step2",
		}

		module := NewModule("test_state_machine")
		result := module.WithLambdaIntegration(lambdaARNs...)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.AttachPolicyForLambda)
		assert.True(t, *module.AttachPolicyForLambda)
		assert.NotNil(t, module.LambdaFunctionARNs)
		assert.Len(t, module.LambdaFunctionARNs, 2)
		assert.Contains(t, module.LambdaFunctionARNs, lambdaARNs[0])
		assert.Contains(t, module.LambdaFunctionARNs, lambdaARNs[1])
	})

	t.Run("appends Lambda ARNs when called multiple times", func(t *testing.T) {
		module := NewModule("test")

		module.WithLambdaIntegration("arn1")
		module.WithLambdaIntegration("arn2", "arn3")

		assert.Len(t, module.LambdaFunctionARNs, 3)
	})

	t.Run("supports single Lambda function", func(t *testing.T) {
		module := NewModule("test")
		module.WithLambdaIntegration("lambda-arn")

		assert.Len(t, module.LambdaFunctionARNs, 1)
		assert.Equal(t, "lambda-arn", module.LambdaFunctionARNs[0])
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the state machine", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_state_machine")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test")

		module.WithTags(map[string]string{"Key": "old"})
		module.WithTags(map[string]string{"Key": "new"})

		assert.Equal(t, "new", module.Tags["Key"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns state machine name when set", func(t *testing.T) {
		name := "my_workflow"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "state_machine", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_state_machine")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete STANDARD workflow configuration", func(t *testing.T) {
		definition := `{"StartAt": "HelloWorld", "States": {}}`

		module := NewModule("order-processor").
			WithDefinition(definition).
			WithLogging("ALL", true).
			WithTracing().
			WithEncryption("kms-key", 300).
			WithLambdaIntegration("lambda1", "lambda2").
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.Name)
		assert.Equal(t, "order-processor", *module.Name)
		assert.Equal(t, "STANDARD", *module.Type)
		assert.NotNil(t, module.Definition)
		assert.NotNil(t, module.LoggingConfiguration)
		assert.NotNil(t, module.TracingConfiguration)
		assert.NotNil(t, module.EncryptionConfiguration)
		assert.Len(t, module.LambdaFunctionARNs, 2)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports EXPRESS workflow configuration", func(t *testing.T) {
		definition := `{"StartAt": "Process", "States": {}}`

		module := NewModule("fast-processor").
			WithDefinition(definition).
			WithExpressType().
			WithLogging("ERROR", false).
			WithLambdaIntegration("lambda-arn")

		assert.Equal(t, "EXPRESS", *module.Type)
		assert.Equal(t, "ERROR", *module.LoggingConfiguration.Level)
	})
}

func TestEncryptionConfiguration(t *testing.T) {
	t.Run("creates AWS owned key configuration", func(t *testing.T) {
		encType := "AWS_OWNED_KEY"
		config := EncryptionConfiguration{
			Type: &encType,
		}

		assert.Equal(t, "AWS_OWNED_KEY", *config.Type)
		assert.Nil(t, config.KMSKeyID)
	})

	t.Run("creates customer managed key configuration", func(t *testing.T) {
		encType := "CUSTOMER_MANAGED_KMS_KEY"
		kmsKey := "arn:aws:kms:us-east-1:123456789012:key/12345"
		reusePeriod := 600

		config := EncryptionConfiguration{
			Type:                         &encType,
			KMSKeyID:                     &kmsKey,
			KMSDataKeyReusePeriodSeconds: &reusePeriod,
		}

		assert.Equal(t, "CUSTOMER_MANAGED_KMS_KEY", *config.Type)
		assert.Equal(t, kmsKey, *config.KMSKeyID)
		assert.Equal(t, 600, *config.KMSDataKeyReusePeriodSeconds)
	})
}

func TestLoggingConfiguration(t *testing.T) {
	t.Run("creates logging configuration with ALL level", func(t *testing.T) {
		level := "ALL"
		includeData := true
		logDest := "arn:aws:logs:us-east-1:123456789012:log-group:/aws/states"

		config := LoggingConfiguration{
			Level:                &level,
			IncludeExecutionData: &includeData,
			LogDestination:       &logDest,
		}

		assert.Equal(t, "ALL", *config.Level)
		assert.True(t, *config.IncludeExecutionData)
		assert.NotNil(t, config.LogDestination)
	})

	t.Run("creates minimal logging configuration", func(t *testing.T) {
		level := "ERROR"
		config := LoggingConfiguration{
			Level: &level,
		}

		assert.Equal(t, "ERROR", *config.Level)
		assert.Nil(t, config.IncludeExecutionData)
	})
}

func TestTracingConfiguration(t *testing.T) {
	t.Run("creates enabled tracing configuration", func(t *testing.T) {
		enabled := true
		config := TracingConfiguration{
			Enabled: &enabled,
		}

		assert.True(t, *config.Enabled)
	})

	t.Run("creates disabled tracing configuration", func(t *testing.T) {
		enabled := false
		config := TracingConfiguration{
			Enabled: &enabled,
		}

		assert.False(t, *config.Enabled)
	})
}

func TestPolicyStatement(t *testing.T) {
	t.Run("creates policy statement for Lambda invocation", func(t *testing.T) {
		effect := "Allow"
		stmt := PolicyStatement{
			Effect: &effect,
			Actions: []string{
				"lambda:InvokeFunction",
			},
			Resources: []string{
				"arn:aws:lambda:us-east-1:123456789012:function:*",
			},
		}

		assert.Equal(t, "Allow", *stmt.Effect)
		assert.Len(t, stmt.Actions, 1)
		assert.Contains(t, stmt.Actions, "lambda:InvokeFunction")
	})

	t.Run("creates policy statement for DynamoDB access", func(t *testing.T) {
		effect := "Allow"
		stmt := PolicyStatement{
			Effect: &effect,
			Actions: []string{
				"dynamodb:GetItem",
				"dynamodb:PutItem",
			},
			Resources: []string{
				"arn:aws:dynamodb:us-east-1:123456789012:table/*",
			},
		}

		assert.Len(t, stmt.Actions, 2)
	})
}

func TestModule_TypeDefaults(t *testing.T) {
	t.Run("defaults to STANDARD workflow type", func(t *testing.T) {
		module := NewModule("test")

		assert.Equal(t, "STANDARD", *module.Type)
	})

	t.Run("can override to EXPRESS type", func(t *testing.T) {
		module := NewModule("test").WithExpressType()

		assert.Equal(t, "EXPRESS", *module.Type)
	})
}

func TestModule_RoleDefaults(t *testing.T) {
	t.Run("creates IAM role by default", func(t *testing.T) {
		module := NewModule("test")

		assert.True(t, *module.CreateRole)
	})

	t.Run("attaches CloudWatch Logs policy by default", func(t *testing.T) {
		module := NewModule("test")

		assert.True(t, *module.AttachCloudwatchLogsPolicy)
	})
}

// BenchmarkNewModule benchmarks module creation
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_state_machine")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls
func BenchmarkFluentAPI(b *testing.B) {
	definition := `{"StartAt": "HelloWorld", "States": {}}`

	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_state_machine").
			WithDefinition(definition).
			WithLogging("ALL", true).
			WithTracing().
			WithLambdaIntegration("lambda-arn").
			WithTags(map[string]string{"Environment": "production"})
	}
}
