package secretsmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_secret"
		module := NewModule(name)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/secrets-manager/aws", module.Source)
		assert.Equal(t, "~> 1.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.RecoveryWindowInDays)
		assert.Equal(t, 30, *module.RecoveryWindowInDays)

		assert.NotNil(t, module.BlockPublicPolicy)
		assert.True(t, *module.BlockPublicPolicy)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"secret1", "my-secret", "db_password"}
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

func TestModule_WithSecretString(t *testing.T) {
	t.Run("sets secret string value", func(t *testing.T) {
		value := "my-secret-value"
		module := NewModule("test_secret")
		result := module.WithSecretString(value)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify secret string is set
		assert.NotNil(t, module.SecretString)
		assert.Equal(t, value, *module.SecretString)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_secret").
			WithSecretString("password123")

		assert.NotNil(t, module.SecretString)
		assert.Equal(t, "password123", *module.SecretString)
	})
}

func TestModule_WithSecretJSON(t *testing.T) {
	t.Run("sets secret JSON value", func(t *testing.T) {
		jsonValue := `{"username": "admin", "password": "secret"}`
		module := NewModule("test_secret")
		result := module.WithSecretJSON(jsonValue)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify secret string is set with JSON
		assert.NotNil(t, module.SecretString)
		assert.JSONEq(t, jsonValue, *module.SecretString)
		assert.Contains(t, *module.SecretString, "username")
	})

	t.Run("handles complex JSON structures", func(t *testing.T) {
		jsonValue := `{"db": {"host": "localhost", "port": 5432}, "creds": {"user": "admin"}}`
		module := NewModule("test_secret").WithSecretJSON(jsonValue)

		assert.Contains(t, *module.SecretString, "db")
		assert.Contains(t, *module.SecretString, "creds")
	})
}

func TestModule_WithKMSKey(t *testing.T) {
	t.Run("sets KMS key for encryption", func(t *testing.T) {
		kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
		module := NewModule("test_secret")
		result := module.WithKMSKey(kmsKeyID)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify KMS key is set
		assert.NotNil(t, module.KMSKeyID)
		assert.Equal(t, kmsKeyID, *module.KMSKeyID)
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		kmsAlias := "alias/my-secret-key"
		module := NewModule("test_secret").WithKMSKey(kmsAlias)

		assert.NotNil(t, module.KMSKeyID)
		assert.Equal(t, kmsAlias, *module.KMSKeyID)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test_secret").
			WithKMSKey("alias/my-key").
			WithSecretString("value")

		assert.NotNil(t, module.KMSKeyID)
		assert.NotNil(t, module.SecretString)
	})
}

func TestModule_WithRecoveryWindow(t *testing.T) {
	t.Run("sets recovery window", func(t *testing.T) {
		module := NewModule("test_secret")
		result := module.WithRecoveryWindow(7)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify recovery window is set
		assert.NotNil(t, module.RecoveryWindowInDays)
		assert.Equal(t, 7, *module.RecoveryWindowInDays)
	})

	t.Run("supports immediate deletion with zero", func(t *testing.T) {
		module := NewModule("test_secret").WithRecoveryWindow(0)

		assert.NotNil(t, module.RecoveryWindowInDays)
		assert.Equal(t, 0, *module.RecoveryWindowInDays)
	})

	t.Run("supports various recovery window values", func(t *testing.T) {
		tests := []int{0, 7, 14, 21, 30}
		for _, days := range tests {
			module := NewModule("test_secret").WithRecoveryWindow(days)
			assert.Equal(t, days, *module.RecoveryWindowInDays)
		}
	})
}

func TestModule_WithReplication(t *testing.T) {
	t.Run("adds replication to another region", func(t *testing.T) {
		region := "us-west-2"
		kmsKeyID := "arn:aws:kms:us-west-2:123456789012:key/abcd"
		module := NewModule("test_secret")
		result := module.WithReplication(region, kmsKeyID)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify replication is configured
		assert.NotNil(t, module.Replica)
		assert.Len(t, module.Replica, 1)
		assert.NotNil(t, module.Replica[region].Region)
		assert.Equal(t, region, *module.Replica[region].Region)
		assert.NotNil(t, module.Replica[region].KMSKeyID)
		assert.Equal(t, kmsKeyID, *module.Replica[region].KMSKeyID)
	})

	t.Run("adds multiple replicas", func(t *testing.T) {
		module := NewModule("test_secret")
		module.WithReplication("us-west-2", "key1")
		module.WithReplication("eu-west-1", "key2")

		assert.Len(t, module.Replica, 2)
		assert.NotNil(t, module.Replica["us-west-2"])
		assert.NotNil(t, module.Replica["eu-west-1"])
	})
}

func TestModule_WithRotation(t *testing.T) {
	t.Run("enables rotation with Lambda function", func(t *testing.T) {
		lambdaARN := "arn:aws:lambda:us-east-1:123456789012:function:rotate-secret"
		daysInterval := 30
		module := NewModule("test_secret")
		result := module.WithRotation(lambdaARN, daysInterval)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify rotation is enabled
		assert.NotNil(t, module.EnableRotation)
		assert.True(t, *module.EnableRotation)

		assert.NotNil(t, module.RotationLambdaARN)
		assert.Equal(t, lambdaARN, *module.RotationLambdaARN)

		assert.NotNil(t, module.RotationRules)
		assert.NotNil(t, module.RotationRules.AutomaticallyAfterDays)
		assert.Equal(t, daysInterval, *module.RotationRules.AutomaticallyAfterDays)
	})

	t.Run("supports different rotation intervals", func(t *testing.T) {
		tests := []int{1, 7, 30, 90, 365}
		for _, days := range tests {
			module := NewModule("test_secret").WithRotation("arn:aws:lambda:...", days)
			assert.Equal(t, days, *module.RotationRules.AutomaticallyAfterDays)
		}
	})
}

func TestModule_WithPolicy(t *testing.T) {
	t.Run("adds resource policy statement", func(t *testing.T) {
		sid := "AllowRead"
		effect := "Allow"
		statement := PolicyStatement{
			SID:    &sid,
			Effect: &effect,
			Actions: []string{
				"secretsmanager:GetSecretValue",
			},
			Resources: []string{"*"},
		}
		module := NewModule("test_secret")
		result := module.WithPolicy("allow_read", statement)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify policy is created
		assert.NotNil(t, module.CreatePolicy)
		assert.True(t, *module.CreatePolicy)

		assert.NotNil(t, module.PolicyStatements)
		assert.Len(t, module.PolicyStatements, 1)
		assert.NotNil(t, module.PolicyStatements["allow_read"].SID)
		assert.Equal(t, "AllowRead", *module.PolicyStatements["allow_read"].SID)
	})

	t.Run("adds multiple policy statements", func(t *testing.T) {
		module := NewModule("test_secret")

		sid1 := "AllowRead"
		effect1 := "Allow"
		module.WithPolicy("allow_read", PolicyStatement{
			SID:    &sid1,
			Effect: &effect1,
		})

		sid2 := "DenyWrite"
		effect2 := "Deny"
		module.WithPolicy("deny_write", PolicyStatement{
			SID:    &sid2,
			Effect: &effect2,
		})

		assert.Len(t, module.PolicyStatements, 2)
		assert.Equal(t, "AllowRead", *module.PolicyStatements["allow_read"].SID)
		assert.Equal(t, "DenyWrite", *module.PolicyStatements["deny_write"].SID)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to module", func(t *testing.T) {
		module := NewModule("test_secret")
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
		module := NewModule("test_secret")

		tags1 := map[string]string{"Environment": "production"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Team": "platform"}
		module.WithTags(tags2)

		// Verify both sets of tags are present
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("test_secret")

		tags1 := map[string]string{"Environment": "development"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Environment": "production"}
		module.WithTags(tags2)

		// Verify tag was overwritten
		assert.Equal(t, "production", module.Tags["Environment"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns name when set", func(t *testing.T) {
		name := "my_secret"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "secret", module.LocalName())
	})

	t.Run("returns empty string when name is empty", func(t *testing.T) {
		emptyName := ""
		module := NewModule(emptyName)

		assert.Equal(t, emptyName, module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_secret")

		config, err := module.Configuration()

		// Current implementation is a placeholder
		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		jsonValue := `{"username": "admin", "password": "secret"}`
		module := NewModule("db_credentials").
			WithSecretJSON(jsonValue).
			WithKMSKey("alias/my-key").
			WithRecoveryWindow(7).
			WithReplication("us-west-2", "key1").
			WithTags(map[string]string{"Team": "platform"})

		// Verify all configuration is applied
		assert.NotNil(t, module.Name)
		assert.Equal(t, "db_credentials", *module.Name)

		assert.NotNil(t, module.SecretString)
		assert.Contains(t, *module.SecretString, "username")

		assert.NotNil(t, module.KMSKeyID)
		assert.Equal(t, "alias/my-key", *module.KMSKeyID)

		assert.NotNil(t, module.RecoveryWindowInDays)
		assert.Equal(t, 7, *module.RecoveryWindowInDays)

		assert.Len(t, module.Replica, 1)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports rotation fluent API", func(t *testing.T) {
		module := NewModule("rotated_secret").
			WithSecretString("initial-value").
			WithRotation("arn:aws:lambda:...", 30).
			WithTags(map[string]string{"Rotation": "enabled"})

		assert.True(t, *module.EnableRotation)
		assert.Equal(t, 30, *module.RotationRules.AutomaticallyAfterDays)
		assert.Equal(t, "enabled", module.Tags["Rotation"])
	})
}

func TestPolicyStatement(t *testing.T) {
	t.Run("creates policy statement", func(t *testing.T) {
		sid := "AllowRead"
		effect := "Allow"
		stmt := PolicyStatement{
			SID:    &sid,
			Effect: &effect,
			Actions: []string{
				"secretsmanager:GetSecretValue",
				"secretsmanager:DescribeSecret",
			},
			Resources: []string{"*"},
		}

		assert.Equal(t, "AllowRead", *stmt.SID)
		assert.Equal(t, "Allow", *stmt.Effect)
		assert.Len(t, stmt.Actions, 2)
		assert.Contains(t, stmt.Actions, "secretsmanager:GetSecretValue")
	})

	t.Run("creates policy statement with principals", func(t *testing.T) {
		stmt := PolicyStatement{
			Principals: []Principal{
				{
					Type:        "AWS",
					Identifiers: []string{"arn:aws:iam::123456789012:root"},
				},
			},
		}

		assert.Len(t, stmt.Principals, 1)
		assert.Equal(t, "AWS", stmt.Principals[0].Type)
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
	})
}

func TestReplica(t *testing.T) {
	t.Run("creates replica configuration", func(t *testing.T) {
		region := "eu-west-1"
		kmsKey := "arn:aws:kms:eu-west-1:123456789012:key/abcd"
		replica := Replica{
			Region:   &region,
			KMSKeyID: &kmsKey,
		}

		assert.Equal(t, "eu-west-1", *replica.Region)
		assert.Contains(t, *replica.KMSKeyID, "eu-west-1")
	})
}

func TestRotationRules(t *testing.T) {
	t.Run("creates rotation rules with days interval", func(t *testing.T) {
		days := 30
		rules := RotationRules{
			AutomaticallyAfterDays: &days,
		}

		assert.Equal(t, 30, *rules.AutomaticallyAfterDays)
	})

	t.Run("creates rotation rules with duration", func(t *testing.T) {
		duration := "3h"
		rules := RotationRules{
			Duration: &duration,
		}

		assert.Equal(t, "3h", *rules.Duration)
	})

	t.Run("creates rotation rules with schedule expression", func(t *testing.T) {
		schedule := "cron(0 0 * * ? *)"
		rules := RotationRules{
			ScheduleExpression: &schedule,
		}

		assert.Equal(t, "cron(0 0 * * ? *)", *rules.ScheduleExpression)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_secret")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_secret").
			WithSecretJSON(`{"key": "value"}`).
			WithKMSKey("alias/key").
			WithTags(map[string]string{"Environment": "production"})
	}
}

// BenchmarkWithReplication benchmarks replication addition.
func BenchmarkWithReplication(b *testing.B) {
	module := NewModule("bench_secret")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithReplication("us-west-2", "key1")
	}
}
