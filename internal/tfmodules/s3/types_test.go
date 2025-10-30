package s3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test-bucket"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/s3-bucket/aws", module.Source)
		assert.Equal(t, "~> 4.0", module.Version)
		assert.NotNil(t, module.Bucket)
		assert.Equal(t, name, *module.Bucket)

		// Verify secure defaults
		assert.NotNil(t, module.CreateBucket)
		assert.True(t, *module.CreateBucket)

		assert.NotNil(t, module.BlockPublicACLs)
		assert.True(t, *module.BlockPublicACLs)

		assert.NotNil(t, module.BlockPublicPolicy)
		assert.True(t, *module.BlockPublicPolicy)

		assert.NotNil(t, module.IgnorePublicACLs)
		assert.True(t, *module.IgnorePublicACLs)

		assert.NotNil(t, module.RestrictPublicBuckets)
		assert.True(t, *module.RestrictPublicBuckets)

		assert.NotNil(t, module.ObjectOwnership)
		assert.Equal(t, "BucketOwnerEnforced", *module.ObjectOwnership)

		// Verify versioning enabled by default
		assert.NotNil(t, module.Versioning)
		assert.Equal(t, "true", module.Versioning["enabled"])

		// Verify encryption enabled by default
		assert.NotNil(t, module.ServerSideEncryptionConfiguration)
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"data-bucket", "logs", "assets-123"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Bucket)
			assert.Equal(t, name, *module.Bucket)
		}
	})
}

func TestModule_WithVersioning(t *testing.T) {
	t.Run("enables versioning", func(t *testing.T) {
		module := NewModule("test-bucket")
		result := module.WithVersioning(true)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Versioning)
		assert.Equal(t, "true", module.Versioning["enabled"])
	})

	t.Run("disables versioning", func(t *testing.T) {
		module := NewModule("test-bucket")
		module.WithVersioning(false)

		assert.NotNil(t, module.Versioning)
		assert.Equal(t, "false", module.Versioning["enabled"])
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithVersioning(true).
			WithEncryption("kms-key")

		assert.Equal(t, "true", module.Versioning["enabled"])
		assert.NotNil(t, module.ServerSideEncryptionConfiguration)
	})
}

func TestModule_WithEncryption(t *testing.T) {
	t.Run("configures KMS encryption", func(t *testing.T) {
		kmsKeyARN := "arn:aws:kms:us-east-1:123456789012:key/12345"

		module := NewModule("test-bucket")
		result := module.WithEncryption(kmsKeyARN)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.ServerSideEncryptionConfiguration)

		config := module.ServerSideEncryptionConfiguration
		rule := config["rule"].(map[string]interface{})
		sseConfig := rule["apply_server_side_encryption_by_default"].(map[string]interface{})

		assert.Equal(t, "aws:kms", sseConfig["sse_algorithm"])
		assert.Equal(t, kmsKeyARN, sseConfig["kms_master_key_id"])
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		module := NewModule("test-bucket")
		module.WithEncryption("alias/aws/s3")

		config := module.ServerSideEncryptionConfiguration
		rule := config["rule"].(map[string]interface{})
		sseConfig := rule["apply_server_side_encryption_by_default"].(map[string]interface{})

		assert.Equal(t, "alias/aws/s3", sseConfig["kms_master_key_id"])
	})
}

func TestModule_WithPublicAccess(t *testing.T) {
	t.Run("allows public access by removing blocks", func(t *testing.T) {
		module := NewModule("test-bucket")

		// Verify blocks are enabled by default
		assert.True(t, *module.BlockPublicACLs)

		result := module.WithPublicAccess()

		assert.Equal(t, module, result)
		assert.NotNil(t, module.BlockPublicACLs)
		assert.False(t, *module.BlockPublicACLs)
		assert.False(t, *module.BlockPublicPolicy)
		assert.False(t, *module.IgnorePublicACLs)
		assert.False(t, *module.RestrictPublicBuckets)
	})

	t.Run("can be chained with other methods", func(t *testing.T) {
		module := NewModule("test").
			WithPublicAccess().
			WithVersioning(false)

		assert.False(t, *module.BlockPublicACLs)
		assert.Equal(t, "false", module.Versioning["enabled"])
	})
}

func TestModule_WithWebsite(t *testing.T) {
	t.Run("configures static website hosting", func(t *testing.T) {
		module := NewModule("test-bucket")
		result := module.WithWebsite("index.html", "error.html")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Website)
		assert.Equal(t, "index.html", module.Website["index_document"])
		assert.Equal(t, "error.html", module.Website["error_document"])
	})

	t.Run("supports custom document names", func(t *testing.T) {
		module := NewModule("test-bucket")
		module.WithWebsite("home.html", "404.html")

		assert.Equal(t, "home.html", module.Website["index_document"])
		assert.Equal(t, "404.html", module.Website["error_document"])
	})
}

func TestModule_WithLogging(t *testing.T) {
	t.Run("configures access logging", func(t *testing.T) {
		targetBucket := "logs-bucket"
		targetPrefix := "access-logs/"

		module := NewModule("test-bucket")
		result := module.WithLogging(targetBucket, targetPrefix)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Logging)
		assert.Equal(t, targetBucket, module.Logging["target_bucket"])
		assert.Equal(t, targetPrefix, module.Logging["target_prefix"])
	})

	t.Run("supports empty prefix", func(t *testing.T) {
		module := NewModule("test-bucket")
		module.WithLogging("logs-bucket", "")

		assert.Equal(t, "", module.Logging["target_prefix"])
	})
}

func TestModule_WithCORS(t *testing.T) {
	t.Run("adds CORS rules", func(t *testing.T) {
		allowedOrigins := []string{"https://example.com"}
		allowedMethods := []string{"GET", "POST"}
		allowedHeaders := []string{"*"}

		module := NewModule("test-bucket")
		result := module.WithCORS(allowedOrigins, allowedMethods, allowedHeaders)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.CORSRule)
		assert.Len(t, module.CORSRule, 1)

		corsRule := module.CORSRule[0].(map[string]interface{})
		assert.Equal(t, allowedOrigins, corsRule["allowed_origins"])
		assert.Equal(t, allowedMethods, corsRule["allowed_methods"])
		assert.Equal(t, allowedHeaders, corsRule["allowed_headers"])
	})

	t.Run("supports wildcard CORS", func(t *testing.T) {
		module := NewModule("test-bucket")
		module.WithCORS([]string{"*"}, []string{"*"}, []string{"*"})

		corsRule := module.CORSRule[0].(map[string]interface{})
		assert.Equal(t, []string{"*"}, corsRule["allowed_origins"])
	})
}

func TestModule_WithLifecycleRule(t *testing.T) {
	t.Run("adds lifecycle rule", func(t *testing.T) {
		rule := map[string]interface{}{
			"id":     "archive-old-versions",
			"status": "Enabled",
			"noncurrent_version_transitions": []map[string]interface{}{
				{
					"days":          30,
					"storage_class": "GLACIER",
				},
			},
		}

		module := NewModule("test-bucket")
		result := module.WithLifecycleRule(rule)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.LifecycleRule)
		assert.Len(t, module.LifecycleRule, 1)
		assert.Equal(t, rule, module.LifecycleRule[0])
	})

	t.Run("adds multiple lifecycle rules", func(t *testing.T) {
		module := NewModule("test-bucket")

		rule1 := map[string]interface{}{"id": "rule1"}
		module.WithLifecycleRule(rule1)

		rule2 := map[string]interface{}{"id": "rule2"}
		module.WithLifecycleRule(rule2)

		assert.Len(t, module.LifecycleRule, 2)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the bucket", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test-bucket")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test-bucket")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns bucket name when set", func(t *testing.T) {
		name := "my-bucket"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "s3_bucket", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test-bucket")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("assets").
			WithVersioning(true).
			WithEncryption("alias/aws/s3").
			WithLogging("logs-bucket", "assets/").
			WithCORS([]string{"*"}, []string{"GET"}, []string{"*"}).
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.Bucket)
		assert.Equal(t, "assets", *module.Bucket)
		assert.Equal(t, "true", module.Versioning["enabled"])
		assert.NotNil(t, module.ServerSideEncryptionConfiguration)
		assert.NotNil(t, module.Logging)
		assert.NotNil(t, module.CORSRule)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports website hosting configuration", func(t *testing.T) {
		module := NewModule("website").
			WithVersioning(false).
			WithPublicAccess().
			WithWebsite("index.html", "error.html")

		assert.False(t, *module.BlockPublicACLs)
		assert.NotNil(t, module.Website)
	})

	t.Run("supports lifecycle rules", func(t *testing.T) {
		rule1 := map[string]interface{}{"id": "archive"}
		rule2 := map[string]interface{}{"id": "expire"}

		module := NewModule("data").
			WithLifecycleRule(rule1).
			WithLifecycleRule(rule2)

		assert.Len(t, module.LifecycleRule, 2)
	})
}

func TestGrant(t *testing.T) {
	t.Run("creates grant with canonical user", func(t *testing.T) {
		userID := "123456789012"
		grant := Grant{
			ID:          &userID,
			Type:        "CanonicalUser",
			Permissions: []string{"READ", "WRITE"},
		}

		assert.Equal(t, userID, *grant.ID)
		assert.Equal(t, "CanonicalUser", grant.Type)
		assert.Len(t, grant.Permissions, 2)
	})

	t.Run("creates grant with predefined group", func(t *testing.T) {
		uri := "http://acs.amazonaws.com/groups/global/AllUsers"
		grant := Grant{
			Type:        "Group",
			URI:         &uri,
			Permissions: []string{"READ"},
		}

		assert.Equal(t, "Group", grant.Type)
		assert.Equal(t, uri, *grant.URI)
	})
}

func TestModule_EncryptionDefaults(t *testing.T) {
	t.Run("default encryption uses AES256", func(t *testing.T) {
		module := NewModule("test-bucket")

		config := module.ServerSideEncryptionConfiguration
		rule := config["rule"].(map[string]interface{})
		sseConfig := rule["apply_server_side_encryption_by_default"].(map[string]interface{})

		assert.Equal(t, "AES256", sseConfig["sse_algorithm"])
	})
}

func TestModule_SecurityDefaults(t *testing.T) {
	t.Run("blocks all public access by default", func(t *testing.T) {
		module := NewModule("test-bucket")

		assert.True(t, *module.BlockPublicACLs)
		assert.True(t, *module.BlockPublicPolicy)
		assert.True(t, *module.IgnorePublicACLs)
		assert.True(t, *module.RestrictPublicBuckets)
	})

	t.Run("enforces bucket owner object ownership", func(t *testing.T) {
		module := NewModule("test-bucket")

		assert.Equal(t, "BucketOwnerEnforced", *module.ObjectOwnership)
	})
}

// BenchmarkNewModule benchmarks module creation
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench-bucket")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench-bucket").
			WithVersioning(true).
			WithEncryption("alias/aws/s3").
			WithCORS([]string{"*"}, []string{"GET"}, []string{"*"}).
			WithTags(map[string]string{"Environment": "production"})
	}
}
