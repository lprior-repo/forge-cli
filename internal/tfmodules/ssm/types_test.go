package ssm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "/app/config/test"
		module := NewModule(name)

		// Verify basic properties
		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/ssm-parameter/aws", module.Source)
		assert.Equal(t, "~> 1.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.Create)
		assert.True(t, *module.Create)

		assert.NotNil(t, module.Type)
		assert.Equal(t, "String", *module.Type)

		assert.NotNil(t, module.Tier)
		assert.Equal(t, "Standard", *module.Tier)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"/app/db/host", "/config/api-key", "/secrets/password"}
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

func TestModule_WithValue(t *testing.T) {
	t.Run("sets string value", func(t *testing.T) {
		value := "my-parameter-value"
		module := NewModule("/app/config/test")
		result := module.WithValue(value)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify value is set
		assert.NotNil(t, module.Value)
		assert.Equal(t, value, *module.Value)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/app/config/test").
			WithValue("test-value")

		assert.NotNil(t, module.Value)
		assert.Equal(t, "test-value", *module.Value)
	})

	t.Run("handles various value types", func(t *testing.T) {
		tests := []string{
			"simple-string",
			"value with spaces",
			"123456",
			"https://example.com",
		}

		for _, val := range tests {
			module := NewModule("/test").WithValue(val)
			assert.Equal(t, val, *module.Value)
		}
	})
}

func TestModule_WithStringList(t *testing.T) {
	t.Run("sets string list value", func(t *testing.T) {
		values := []string{"value1", "value2", "value3"}
		module := NewModule("/app/config/list")
		result := module.WithStringList(values)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify type is set to StringList
		assert.NotNil(t, module.Type)
		assert.Equal(t, "StringList", *module.Type)

		// Verify values are set
		assert.NotNil(t, module.Values)
		assert.Equal(t, values, module.Values)
		assert.Len(t, module.Values, 3)
	})

	t.Run("supports empty list", func(t *testing.T) {
		module := NewModule("/test").WithStringList([]string{})

		assert.NotNil(t, module.Type)
		assert.Equal(t, "StringList", *module.Type)
		assert.Empty(t, module.Values)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/test").
			WithStringList([]string{"a", "b", "c"})

		assert.Equal(t, "StringList", *module.Type)
		assert.Len(t, module.Values, 3)
	})
}

func TestModule_WithSecureString(t *testing.T) {
	t.Run("sets secure string with KMS key", func(t *testing.T) {
		value := "secret-password"
		kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
		module := NewModule("/app/secrets/password")
		result := module.WithSecureString(value, kmsKeyID)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify type is set to SecureString
		assert.NotNil(t, module.Type)
		assert.Equal(t, "SecureString", *module.Type)

		// Verify secure type flag
		assert.NotNil(t, module.SecureType)
		assert.True(t, *module.SecureType)

		// Verify value is set
		assert.NotNil(t, module.Value)
		assert.Equal(t, value, *module.Value)

		// Verify KMS key is set
		assert.NotNil(t, module.KeyID)
		assert.Equal(t, kmsKeyID, *module.KeyID)
	})

	t.Run("sets secure string without KMS key", func(t *testing.T) {
		value := "secret-password"
		module := NewModule("/app/secrets/password")
		module.WithSecureString(value, "")

		// Should use default AWS managed key
		assert.Equal(t, "SecureString", *module.Type)
		assert.True(t, *module.SecureType)
		assert.Equal(t, value, *module.Value)
		assert.Nil(t, module.KeyID) // No custom KMS key
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		kmsAlias := "alias/my-parameter-key"
		module := NewModule("/test").WithSecureString("secret", kmsAlias)

		assert.NotNil(t, module.KeyID)
		assert.Equal(t, kmsAlias, *module.KeyID)
	})
}

func TestModule_WithAdvancedTier(t *testing.T) {
	t.Run("sets tier to Advanced", func(t *testing.T) {
		module := NewModule("/app/config/large")
		result := module.WithAdvancedTier()

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify tier is set to Advanced
		assert.NotNil(t, module.Tier)
		assert.Equal(t, "Advanced", *module.Tier)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/test").
			WithAdvancedTier().
			WithValue("large value")

		assert.Equal(t, "Advanced", *module.Tier)
		assert.NotNil(t, module.Value)
	})
}

func TestModule_WithIntelligentTiering(t *testing.T) {
	t.Run("sets tier to Intelligent-Tiering", func(t *testing.T) {
		module := NewModule("/app/config/auto")
		result := module.WithIntelligentTiering()

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify tier is set to Intelligent-Tiering
		assert.NotNil(t, module.Tier)
		assert.Equal(t, "Intelligent-Tiering", *module.Tier)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/test").
			WithIntelligentTiering().
			WithValue("value")

		assert.Equal(t, "Intelligent-Tiering", *module.Tier)
	})
}

func TestModule_WithValidation(t *testing.T) {
	t.Run("adds regex validation pattern", func(t *testing.T) {
		pattern := "^[a-zA-Z0-9]+$"
		module := NewModule("/app/config/validated")
		result := module.WithValidation(pattern)

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify pattern is set
		assert.NotNil(t, module.AllowedPattern)
		assert.Equal(t, pattern, *module.AllowedPattern)
	})

	t.Run("supports various regex patterns", func(t *testing.T) {
		tests := []struct {
			name    string
			pattern string
		}{
			{"alphanumeric", "^[a-zA-Z0-9]+$"},
			{"email", `^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$`},
			{"url", `^https?://.*$`},
			{"ip", `^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				module := NewModule("/test").WithValidation(tt.pattern)
				assert.Equal(t, tt.pattern, *module.AllowedPattern)
			})
		}
	})
}

func TestModule_WithAMIDataType(t *testing.T) {
	t.Run("sets data type to AMI", func(t *testing.T) {
		module := NewModule("/app/ami/latest")
		result := module.WithAMIDataType()

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify data type is set
		assert.NotNil(t, module.DataType)
		assert.Equal(t, "aws:ec2:image", *module.DataType)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/test").
			WithAMIDataType().
			WithValue("ami-12345678")

		assert.Equal(t, "aws:ec2:image", *module.DataType)
		assert.NotNil(t, module.Value)
	})
}

func TestModule_WithIgnoreChanges(t *testing.T) {
	t.Run("enables ignore changes", func(t *testing.T) {
		module := NewModule("/app/config/external")
		result := module.WithIgnoreChanges()

		// Verify method returns module for chaining
		assert.Equal(t, module, result)

		// Verify ignore changes is enabled
		assert.NotNil(t, module.IgnoreValueChanges)
		assert.True(t, *module.IgnoreValueChanges)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("/test").
			WithIgnoreChanges().
			WithValue("value")

		assert.True(t, *module.IgnoreValueChanges)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to module", func(t *testing.T) {
		module := NewModule("/app/config/test")
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
		module := NewModule("/test")

		tags1 := map[string]string{"Environment": "production"}
		module.WithTags(tags1)

		tags2 := map[string]string{"Team": "platform"}
		module.WithTags(tags2)

		// Verify both sets of tags are present
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("overwrites existing tags with same key", func(t *testing.T) {
		module := NewModule("/test")

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
		name := "/app/config/test"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "parameter", module.LocalName())
	})

	t.Run("returns empty string when name is empty", func(t *testing.T) {
		emptyName := ""
		module := NewModule(emptyName)

		assert.Equal(t, emptyName, module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("/test")

		config, err := module.Configuration()

		// Current implementation is a placeholder
		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("/app/config/db-host").
			WithValue("localhost:5432").
			WithValidation(`^\w+:\d+$`).
			WithTags(map[string]string{
				"Environment": "production",
				"Team":        "platform",
			})

		// Verify all configuration is applied
		assert.NotNil(t, module.Name)
		assert.Equal(t, "/app/config/db-host", *module.Name)

		assert.NotNil(t, module.Value)
		assert.Equal(t, "localhost:5432", *module.Value)

		assert.NotNil(t, module.AllowedPattern)

		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports secure string fluent API", func(t *testing.T) {
		module := NewModule("/app/secrets/api-key").
			WithSecureString("secret-key-123", "alias/app-key").
			WithAdvancedTier().
			WithIgnoreChanges().
			WithTags(map[string]string{"Type": "secret"})

		assert.Equal(t, "SecureString", *module.Type)
		assert.True(t, *module.SecureType)
		assert.Equal(t, "Advanced", *module.Tier)
		assert.True(t, *module.IgnoreValueChanges)
		assert.Equal(t, "secret", module.Tags["Type"])
	})

	t.Run("supports string list fluent API", func(t *testing.T) {
		module := NewModule("/app/config/allowed-ips").
			WithStringList([]string{"192.168.1.1", "192.168.1.2"}).
			WithValidation(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`).
			WithTags(map[string]string{"Type": "list"})

		assert.Equal(t, "StringList", *module.Type)
		assert.Len(t, module.Values, 2)
		assert.Equal(t, "list", module.Tags["Type"])
	})

	t.Run("supports AMI parameter fluent API", func(t *testing.T) {
		module := NewModule("/app/ami/latest-ubuntu").
			WithValue("ami-12345678").
			WithAMIDataType().
			WithTags(map[string]string{"OS": "ubuntu"})

		assert.Equal(t, "aws:ec2:image", *module.DataType)
		assert.Equal(t, "ami-12345678", *module.Value)
		assert.Equal(t, "ubuntu", module.Tags["OS"])
	})
}

func TestModule_TierCombinations(t *testing.T) {
	t.Run("standard tier with small value", func(t *testing.T) {
		module := NewModule("/test").
			WithValue("small-value")

		// Default tier should be Standard
		assert.Equal(t, "Standard", *module.Tier)
	})

	t.Run("advanced tier for large values", func(t *testing.T) {
		module := NewModule("/test").
			WithAdvancedTier().
			WithValue("large-value-over-4kb")

		assert.Equal(t, "Advanced", *module.Tier)
	})

	t.Run("intelligent tiering for variable usage", func(t *testing.T) {
		module := NewModule("/test").
			WithIntelligentTiering().
			WithValue("variable-usage-value")

		assert.Equal(t, "Intelligent-Tiering", *module.Tier)
	})
}

func TestModule_TypeCombinations(t *testing.T) {
	t.Run("string type with validation", func(t *testing.T) {
		module := NewModule("/test").
			WithValue("test-value").
			WithValidation("^test-.*$")

		assert.Equal(t, "String", *module.Type)
		assert.NotNil(t, module.AllowedPattern)
	})

	t.Run("secure string with KMS and advanced tier", func(t *testing.T) {
		module := NewModule("/test").
			WithSecureString("secret", "alias/key").
			WithAdvancedTier()

		assert.Equal(t, "SecureString", *module.Type)
		assert.Equal(t, "Advanced", *module.Tier)
		assert.NotNil(t, module.KeyID)
	})

	t.Run("string list with standard tier", func(t *testing.T) {
		module := NewModule("/test").
			WithStringList([]string{"a", "b", "c"})

		assert.Equal(t, "StringList", *module.Type)
		assert.Equal(t, "Standard", *module.Tier)
	})
}

// BenchmarkNewModule benchmarks module creation
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("/bench/parameter")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("/bench/parameter").
			WithValue("test-value").
			WithValidation("^.*$").
			WithTags(map[string]string{"Environment": "production"})
	}
}

// BenchmarkWithSecureString benchmarks secure string creation
func BenchmarkWithSecureString(b *testing.B) {
	module := NewModule("/bench/secret")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithSecureString("secret-value", "alias/key")
	}
}

// BenchmarkWithStringList benchmarks string list creation
func BenchmarkWithStringList(b *testing.B) {
	module := NewModule("/bench/list")
	values := []string{"value1", "value2", "value3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		module.WithStringList(values)
	}
}
