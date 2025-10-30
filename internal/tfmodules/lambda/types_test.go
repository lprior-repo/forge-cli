package lambda

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_function"
		module := NewModule(name)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/lambda/aws", module.Source)
		assert.Equal(t, "~> 7.0", module.Version)
		assert.NotNil(t, module.FunctionName)
		assert.Equal(t, name, *module.FunctionName)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.CreateFunction)
		assert.True(t, *module.CreateFunction)

		assert.NotNil(t, module.CreatePackage)
		assert.True(t, *module.CreatePackage)

		assert.NotNil(t, module.CreateRole)
		assert.True(t, *module.CreateRole)

		assert.NotNil(t, module.MemorySize)
		assert.Equal(t, 128, *module.MemorySize)

		assert.NotNil(t, module.Timeout)
		assert.Equal(t, 3, *module.Timeout)

		assert.NotNil(t, module.EphemeralStorageSize)
		assert.Equal(t, 512, *module.EphemeralStorageSize)

		assert.NotNil(t, module.PackageType)
		assert.Equal(t, "Zip", *module.PackageType)

		assert.NotNil(t, module.AttachCloudwatchLogsPolicy)
		assert.True(t, *module.AttachCloudwatchLogsPolicy)

		assert.NotNil(t, module.Timeouts)
		assert.Equal(t, "10m", module.Timeouts["create"])
		assert.Equal(t, "10m", module.Timeouts["update"])
		assert.Equal(t, "10m", module.Timeouts["delete"])
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"api", "worker", "data-processor"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.FunctionName)
			assert.Equal(t, name, *module.FunctionName)
		}
	})

	t.Run("creates module with empty name", func(t *testing.T) {
		module := NewModule("")
		assert.NotNil(t, module.FunctionName)
		assert.Equal(t, "", *module.FunctionName)
	})
}

func TestModule_WithRuntime(t *testing.T) {
	t.Run("sets runtime and handler", func(t *testing.T) {
		module := NewModule("test_function")
		result := module.WithRuntime("python3.13", "handler.main")

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		assert.NotNil(t, module.Runtime)
		assert.Equal(t, "python3.13", *module.Runtime)

		assert.NotNil(t, module.Handler)
		assert.Equal(t, "handler.main", *module.Handler)
	})

	t.Run("supports Go runtime", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithRuntime("provided.al2023", "bootstrap")

		assert.NotNil(t, module.Runtime)
		assert.Equal(t, "provided.al2023", *module.Runtime)
		assert.Equal(t, "bootstrap", *module.Handler)
	})

	t.Run("supports Node.js runtime", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithRuntime("nodejs20.x", "index.handler")

		assert.NotNil(t, module.Runtime)
		assert.Equal(t, "nodejs20.x", *module.Runtime)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithRuntime("python3.13", "main.handler").
			WithMemoryAndTimeout(512, 30)

		assert.NotNil(t, module.Runtime)
		assert.NotNil(t, module.MemorySize)
	})
}

func TestModule_WithMemoryAndTimeout(t *testing.T) {
	t.Run("configures memory and timeout", func(t *testing.T) {
		module := NewModule("test_function")
		result := module.WithMemoryAndTimeout(512, 30)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.MemorySize)
		assert.Equal(t, 512, *module.MemorySize)
		assert.NotNil(t, module.Timeout)
		assert.Equal(t, 30, *module.Timeout)
	})

	t.Run("supports maximum values", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithMemoryAndTimeout(10240, 900)

		assert.Equal(t, 10240, *module.MemorySize)
		assert.Equal(t, 900, *module.Timeout)
	})

	t.Run("supports minimum values", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithMemoryAndTimeout(128, 1)

		assert.Equal(t, 128, *module.MemorySize)
		assert.Equal(t, 1, *module.Timeout)
	})
}

func TestModule_WithVPC(t *testing.T) {
	t.Run("configures VPC settings", func(t *testing.T) {
		subnetIDs := []string{"subnet-1", "subnet-2"}
		securityGroupIDs := []string{"sg-1", "sg-2"}

		module := NewModule("test_function")
		result := module.WithVPC(subnetIDs, securityGroupIDs)

		assert.Equal(t, module, result)
		assert.Equal(t, subnetIDs, module.VPCSubnetIDs)
		assert.Equal(t, securityGroupIDs, module.VPCSecurityGroupIDs)
		assert.NotNil(t, module.AttachNetworkPolicy)
		assert.True(t, *module.AttachNetworkPolicy)
	})

	t.Run("handles empty subnet IDs", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithVPC([]string{}, []string{"sg-1"})

		assert.Empty(t, module.VPCSubnetIDs)
		assert.NotEmpty(t, module.VPCSecurityGroupIDs)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithVPC([]string{"subnet-1"}, []string{"sg-1"}).
			WithRuntime("python3.13", "main.handler")

		assert.NotEmpty(t, module.VPCSubnetIDs)
		assert.NotNil(t, module.Runtime)
	})
}

func TestModule_WithEnvironment(t *testing.T) {
	t.Run("sets environment variables", func(t *testing.T) {
		envVars := map[string]string{
			"TABLE_NAME": "users",
			"REGION":     "us-east-1",
		}

		module := NewModule("test_function")
		result := module.WithEnvironment(envVars)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.EnvironmentVariables)
		assert.Equal(t, "users", module.EnvironmentVariables["TABLE_NAME"])
		assert.Equal(t, "us-east-1", module.EnvironmentVariables["REGION"])
	})

	t.Run("merges environment variables when called multiple times", func(t *testing.T) {
		module := NewModule("test_function")

		env1 := map[string]string{"KEY1": "value1"}
		module.WithEnvironment(env1)

		env2 := map[string]string{"KEY2": "value2"}
		module.WithEnvironment(env2)

		assert.Equal(t, "value1", module.EnvironmentVariables["KEY1"])
		assert.Equal(t, "value2", module.EnvironmentVariables["KEY2"])
	})

	t.Run("overwrites existing variables with same key", func(t *testing.T) {
		module := NewModule("test_function")

		env1 := map[string]string{"KEY": "old"}
		module.WithEnvironment(env1)

		env2 := map[string]string{"KEY": "new"}
		module.WithEnvironment(env2)

		assert.Equal(t, "new", module.EnvironmentVariables["KEY"])
	})

	t.Run("handles empty environment map", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithEnvironment(map[string]string{})

		assert.NotNil(t, module.EnvironmentVariables)
		assert.Empty(t, module.EnvironmentVariables)
	})
}

func TestModule_WithTracing(t *testing.T) {
	t.Run("enables X-Ray tracing with Active mode", func(t *testing.T) {
		module := NewModule("test_function")
		result := module.WithTracing("Active")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.TracingMode)
		assert.Equal(t, "Active", *module.TracingMode)
		assert.NotNil(t, module.AttachTracingPolicy)
		assert.True(t, *module.AttachTracingPolicy)
	})

	t.Run("enables PassThrough mode", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithTracing("PassThrough")

		assert.NotNil(t, module.TracingMode)
		assert.Equal(t, "PassThrough", *module.TracingMode)
	})
}

func TestModule_WithLayers(t *testing.T) {
	t.Run("adds Lambda layers", func(t *testing.T) {
		layers := []string{
			"arn:aws:lambda:us-east-1:123456789012:layer:my-layer:1",
			"arn:aws:lambda:us-east-1:123456789012:layer:another-layer:2",
		}

		module := NewModule("test_function")
		result := module.WithLayers(layers...)

		assert.Equal(t, module, result)
		assert.Len(t, module.Layers, 2)
		assert.Contains(t, module.Layers, layers[0])
		assert.Contains(t, module.Layers, layers[1])
	})

	t.Run("appends layers when called multiple times", func(t *testing.T) {
		module := NewModule("test_function")

		module.WithLayers("layer1")
		module.WithLayers("layer2", "layer3")

		assert.Len(t, module.Layers, 3)
	})

	t.Run("handles single layer", func(t *testing.T) {
		module := NewModule("test_function")
		module.WithLayers("layer1")

		assert.Len(t, module.Layers, 1)
		assert.Equal(t, "layer1", module.Layers[0])
	})
}

func TestModule_WithFunctionURL(t *testing.T) {
	t.Run("enables Function URL with IAM auth", func(t *testing.T) {
		module := NewModule("test_function")
		result := module.WithFunctionURL("AWS_IAM", nil)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CreateLambdaFunctionURL)
		assert.True(t, *module.CreateLambdaFunctionURL)
		assert.NotNil(t, module.AuthorizationType)
		assert.Equal(t, "AWS_IAM", *module.AuthorizationType)
		assert.Nil(t, module.CORS)
	})

	t.Run("enables Function URL with NONE auth and CORS", func(t *testing.T) {
		maxAge := 3600
		cors := &CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST"},
			MaxAge:       &maxAge,
		}

		module := NewModule("test_function")
		module.WithFunctionURL("NONE", cors)

		assert.NotNil(t, module.CreateLambdaFunctionURL)
		assert.True(t, *module.CreateLambdaFunctionURL)
		assert.Equal(t, "NONE", *module.AuthorizationType)
		assert.NotNil(t, module.CORS)
		assert.Equal(t, []string{"*"}, module.CORS.AllowOrigins)
		assert.Equal(t, 3600, *module.CORS.MaxAge)
	})
}

func TestModule_WithDeadLetterQueue(t *testing.T) {
	t.Run("configures DLQ", func(t *testing.T) {
		dlqARN := "arn:aws:sqs:us-east-1:123456789012:my-dlq"

		module := NewModule("test_function")
		result := module.WithDeadLetterQueue(dlqARN)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.DeadLetterTargetARN)
		assert.Equal(t, dlqARN, *module.DeadLetterTargetARN)
		assert.NotNil(t, module.AttachDeadLetterPolicy)
		assert.True(t, *module.AttachDeadLetterPolicy)
	})

	t.Run("supports SNS DLQ", func(t *testing.T) {
		dlqARN := "arn:aws:sns:us-east-1:123456789012:my-dlq"
		module := NewModule("test_function")
		module.WithDeadLetterQueue(dlqARN)

		assert.Equal(t, dlqARN, *module.DeadLetterTargetARN)
	})
}

func TestModule_WithEventSourceMapping(t *testing.T) {
	t.Run("adds event source mapping", func(t *testing.T) {
		batchSize := 10
		enabled := true
		mapping := EventSourceMapping{
			EventSourceARN: "arn:aws:sqs:us-east-1:123456789012:my-queue",
			BatchSize:      &batchSize,
			Enabled:        &enabled,
		}

		module := NewModule("test_function")
		result := module.WithEventSourceMapping("sqs", mapping)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.EventSourceMapping)
		assert.Len(t, module.EventSourceMapping, 1)
		assert.Equal(t, mapping.EventSourceARN, module.EventSourceMapping["sqs"].EventSourceARN)
		assert.Equal(t, 10, *module.EventSourceMapping["sqs"].BatchSize)
	})

	t.Run("adds multiple event source mappings", func(t *testing.T) {
		module := NewModule("test_function")

		sqs := EventSourceMapping{EventSourceARN: "sqs-arn"}
		module.WithEventSourceMapping("sqs", sqs)

		dynamodb := EventSourceMapping{EventSourceARN: "dynamodb-arn"}
		module.WithEventSourceMapping("dynamodb", dynamodb)

		assert.Len(t, module.EventSourceMapping, 2)
		assert.Contains(t, module.EventSourceMapping, "sqs")
		assert.Contains(t, module.EventSourceMapping, "dynamodb")
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the function", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_function")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_function")

		tags1 := map[string]string{"Environment": "production"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Team": "platform"}
		module.WithTags(tags2)

		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns function name when set", func(t *testing.T) {
		name := "my_function"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "lambda_function", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_function")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("api").
			WithRuntime("python3.13", "handler.main").
			WithMemoryAndTimeout(512, 30).
			WithEnvironment(map[string]string{"ENV": "prod"}).
			WithTracing("Active").
			WithLayers("layer1", "layer2").
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.FunctionName)
		assert.Equal(t, "api", *module.FunctionName)
		assert.Equal(t, "python3.13", *module.Runtime)
		assert.Equal(t, 512, *module.MemorySize)
		assert.Equal(t, "Active", *module.TracingMode)
		assert.Len(t, module.Layers, 2)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports VPC configuration with Function URL", func(t *testing.T) {
		module := NewModule("api").
			WithVPC([]string{"subnet-1"}, []string{"sg-1"}).
			WithFunctionURL("AWS_IAM", nil)

		assert.NotEmpty(t, module.VPCSubnetIDs)
		assert.True(t, *module.CreateLambdaFunctionURL)
	})
}

func TestCORSConfig(t *testing.T) {
	t.Run("creates CORS configuration", func(t *testing.T) {
		allowCreds := true
		maxAge := 3600
		cors := CORSConfig{
			AllowCredentials: &allowCreds,
			AllowOrigins:     []string{"https://example.com"},
			AllowMethods:     []string{"GET", "POST"},
			AllowHeaders:     []string{"Content-Type", "Authorization"},
			ExposeHeaders:    []string{"X-Custom-Header"},
			MaxAge:           &maxAge,
		}

		assert.True(t, *cors.AllowCredentials)
		assert.Contains(t, cors.AllowOrigins, "https://example.com")
		assert.Len(t, cors.AllowMethods, 2)
		assert.Equal(t, 3600, *cors.MaxAge)
	})
}

func TestEventSourceMapping(t *testing.T) {
	t.Run("creates SQS event source mapping", func(t *testing.T) {
		batchSize := 10
		batchWindow := 5
		enabled := true
		mapping := EventSourceMapping{
			EventSourceARN:                 "arn:aws:sqs:us-east-1:123456789012:queue",
			BatchSize:                      &batchSize,
			MaximumBatchingWindowInSeconds: &batchWindow,
			Enabled:                        &enabled,
		}

		assert.Equal(t, "arn:aws:sqs:us-east-1:123456789012:queue", mapping.EventSourceARN)
		assert.Equal(t, 10, *mapping.BatchSize)
		assert.Equal(t, 5, *mapping.MaximumBatchingWindowInSeconds)
		assert.True(t, *mapping.Enabled)
	})

	t.Run("creates DynamoDB Streams mapping with starting position", func(t *testing.T) {
		position := "TRIM_HORIZON"
		batchSize := 100
		mapping := EventSourceMapping{
			EventSourceARN:   "arn:aws:dynamodb:us-east-1:123456789012:table/mytable/stream",
			StartingPosition: &position,
			BatchSize:        &batchSize,
		}

		assert.Equal(t, "TRIM_HORIZON", *mapping.StartingPosition)
		assert.Equal(t, 100, *mapping.BatchSize)
	})
}

func TestAllowedTrigger(t *testing.T) {
	t.Run("creates S3 trigger", func(t *testing.T) {
		sourceARN := "arn:aws:s3:::my-bucket"
		trigger := AllowedTrigger{
			Service:   "s3.amazonaws.com",
			SourceARN: &sourceARN,
		}

		assert.Equal(t, "s3.amazonaws.com", trigger.Service)
		assert.Equal(t, sourceARN, *trigger.SourceARN)
	})

	t.Run("creates API Gateway trigger", func(t *testing.T) {
		principal := "apigateway.amazonaws.com"
		trigger := AllowedTrigger{
			Service:   "apigateway.amazonaws.com",
			Principal: &principal,
		}

		assert.Equal(t, "apigateway.amazonaws.com", trigger.Service)
		assert.NotNil(t, trigger.Principal)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_function")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_function").
			WithRuntime("python3.13", "handler.main").
			WithMemoryAndTimeout(512, 30).
			WithTracing("Active").
			WithEnvironment(map[string]string{"ENV": "prod"})
	}
}
