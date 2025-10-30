package dynamodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModule(t *testing.T) {
	t.Run("creates module with sensible defaults", func(t *testing.T) {
		name := "test_table"
		module := NewModule(name)

		require.NotNil(t, module)
		assert.Equal(t, "terraform-aws-modules/dynamodb-table/aws", module.Source)
		assert.Equal(t, "~> 4.0", module.Version)
		assert.NotNil(t, module.Name)
		assert.Equal(t, name, *module.Name)

		// Verify sensible defaults
		assert.NotNil(t, module.CreateTable)
		assert.True(t, *module.CreateTable)

		assert.NotNil(t, module.BillingMode)
		assert.Equal(t, "PAY_PER_REQUEST", *module.BillingMode)

		assert.NotNil(t, module.PointInTimeRecoveryEnabled)
		assert.True(t, *module.PointInTimeRecoveryEnabled)

		assert.NotNil(t, module.ServerSideEncryptionEnabled)
		assert.True(t, *module.ServerSideEncryptionEnabled)

		assert.NotNil(t, module.DeletionProtectionEnabled)
		assert.True(t, *module.DeletionProtectionEnabled)

		assert.NotNil(t, module.Timeouts)
		assert.Equal(t, "10m", module.Timeouts["create"])
		assert.Equal(t, "60m", module.Timeouts["update"])
		assert.Equal(t, "10m", module.Timeouts["delete"])
	})

	t.Run("creates module with different names", func(t *testing.T) {
		names := []string{"users", "orders", "inventory-items"}
		for _, name := range names {
			module := NewModule(name)
			assert.NotNil(t, module.Name)
			assert.Equal(t, name, *module.Name)
		}
	})
}

func TestModule_WithHashKey(t *testing.T) {
	t.Run("sets partition key", func(t *testing.T) {
		module := NewModule("test_table")
		result := module.WithHashKey("userId", "S")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.HashKey)
		assert.Equal(t, "userId", *module.HashKey)
		assert.Len(t, module.Attributes, 1)
		assert.Equal(t, "userId", module.Attributes[0].Name)
		assert.Equal(t, "S", module.Attributes[0].Type)
	})

	t.Run("supports number type", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithHashKey("id", "N")

		assert.Equal(t, "N", module.Attributes[0].Type)
	})

	t.Run("supports binary type", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithHashKey("binaryKey", "B")

		assert.Equal(t, "B", module.Attributes[0].Type)
	})

	t.Run("supports method chaining", func(t *testing.T) {
		module := NewModule("test").
			WithHashKey("userId", "S").
			WithRangeKey("timestamp", "N")

		assert.NotNil(t, module.HashKey)
		assert.NotNil(t, module.RangeKey)
	})
}

func TestModule_WithRangeKey(t *testing.T) {
	t.Run("sets sort key", func(t *testing.T) {
		module := NewModule("test_table")
		result := module.WithRangeKey("timestamp", "N")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.RangeKey)
		assert.Equal(t, "timestamp", *module.RangeKey)
		assert.Len(t, module.Attributes, 1)
		assert.Equal(t, "timestamp", module.Attributes[0].Name)
		assert.Equal(t, "N", module.Attributes[0].Type)
	})

	t.Run("adds to existing attributes", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithHashKey("userId", "S")
		module.WithRangeKey("timestamp", "N")

		assert.Len(t, module.Attributes, 2)
	})
}

func TestModule_WithStreams(t *testing.T) {
	t.Run("enables DynamoDB Streams with NEW_IMAGE", func(t *testing.T) {
		module := NewModule("test_table")
		result := module.WithStreams("NEW_IMAGE")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.StreamEnabled)
		assert.True(t, *module.StreamEnabled)
		assert.NotNil(t, module.StreamViewType)
		assert.Equal(t, "NEW_IMAGE", *module.StreamViewType)
	})

	t.Run("supports KEYS_ONLY view type", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithStreams("KEYS_ONLY")

		assert.Equal(t, "KEYS_ONLY", *module.StreamViewType)
	})

	t.Run("supports NEW_AND_OLD_IMAGES view type", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithStreams("NEW_AND_OLD_IMAGES")

		assert.Equal(t, "NEW_AND_OLD_IMAGES", *module.StreamViewType)
	})
}

func TestModule_WithGSI(t *testing.T) {
	t.Run("adds Global Secondary Index", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:           "email-index",
			HashKey:        "email",
			ProjectionType: "ALL",
		}

		module := NewModule("test_table")
		result := module.WithGSI(gsi)

		assert.Equal(t, module, result)
		assert.Len(t, module.GlobalSecondaryIndexes, 1)
		assert.Equal(t, "email-index", module.GlobalSecondaryIndexes[0].Name)
		assert.Equal(t, "email", module.GlobalSecondaryIndexes[0].HashKey)
		assert.Equal(t, "ALL", module.GlobalSecondaryIndexes[0].ProjectionType)
	})

	t.Run("adds multiple GSIs", func(t *testing.T) {
		module := NewModule("test_table")

		gsi1 := GlobalSecondaryIndex{Name: "gsi1", HashKey: "attr1", ProjectionType: "ALL"}
		module.WithGSI(gsi1)

		gsi2 := GlobalSecondaryIndex{Name: "gsi2", HashKey: "attr2", ProjectionType: "KEYS_ONLY"}
		module.WithGSI(gsi2)

		assert.Len(t, module.GlobalSecondaryIndexes, 2)
	})

	t.Run("supports GSI with range key", func(t *testing.T) {
		rangeKey := "timestamp"
		gsi := GlobalSecondaryIndex{
			Name:           "composite-index",
			HashKey:        "userId",
			RangeKey:       &rangeKey,
			ProjectionType: "ALL",
		}

		module := NewModule("test_table")
		module.WithGSI(gsi)

		assert.NotNil(t, module.GlobalSecondaryIndexes[0].RangeKey)
		assert.Equal(t, "timestamp", *module.GlobalSecondaryIndexes[0].RangeKey)
	})

	t.Run("supports GSI with INCLUDE projection", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:             "include-index",
			HashKey:          "status",
			ProjectionType:   "INCLUDE",
			NonKeyAttributes: []string{"name", "email"},
		}

		module := NewModule("test_table")
		module.WithGSI(gsi)

		assert.Equal(t, "INCLUDE", module.GlobalSecondaryIndexes[0].ProjectionType)
		assert.Len(t, module.GlobalSecondaryIndexes[0].NonKeyAttributes, 2)
	})
}

func TestModule_WithTTL(t *testing.T) {
	t.Run("enables Time To Live", func(t *testing.T) {
		module := NewModule("test_table")
		result := module.WithTTL("expiresAt")

		assert.Equal(t, module, result)
		assert.NotNil(t, module.TTLEnabled)
		assert.True(t, *module.TTLEnabled)
		assert.NotNil(t, module.TTLAttributeName)
		assert.Equal(t, "expiresAt", *module.TTLAttributeName)
	})

	t.Run("supports different attribute names", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithTTL("ttl")

		assert.Equal(t, "ttl", *module.TTLAttributeName)
	})
}

func TestModule_WithEncryption(t *testing.T) {
	t.Run("configures KMS encryption", func(t *testing.T) {
		kmsKeyARN := "arn:aws:kms:us-east-1:123456789012:key/12345"

		module := NewModule("test_table")
		result := module.WithEncryption(kmsKeyARN)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.ServerSideEncryptionEnabled)
		assert.True(t, *module.ServerSideEncryptionEnabled)
		assert.NotNil(t, module.ServerSideEncryptionKMSKeyARN)
		assert.Equal(t, kmsKeyARN, *module.ServerSideEncryptionKMSKeyARN)
	})

	t.Run("supports KMS alias", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithEncryption("alias/aws/dynamodb")

		assert.Equal(t, "alias/aws/dynamodb", *module.ServerSideEncryptionKMSKeyARN)
	})
}

func TestModule_WithProvisioned(t *testing.T) {
	t.Run("configures PROVISIONED billing mode", func(t *testing.T) {
		module := NewModule("test_table")
		result := module.WithProvisioned(5, 10)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.BillingMode)
		assert.Equal(t, "PROVISIONED", *module.BillingMode)
		assert.NotNil(t, module.ReadCapacity)
		assert.Equal(t, 5, *module.ReadCapacity)
		assert.NotNil(t, module.WriteCapacity)
		assert.Equal(t, 10, *module.WriteCapacity)
	})

	t.Run("supports minimum capacity", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithProvisioned(1, 1)

		assert.Equal(t, 1, *module.ReadCapacity)
		assert.Equal(t, 1, *module.WriteCapacity)
	})

	t.Run("supports high capacity", func(t *testing.T) {
		module := NewModule("test_table")
		module.WithProvisioned(1000, 1000)

		assert.Equal(t, 1000, *module.ReadCapacity)
	})
}

func TestModule_WithTags(t *testing.T) {
	t.Run("adds tags to the table", func(t *testing.T) {
		tags := map[string]string{
			"Environment": "production",
			"Team":        "platform",
		}

		module := NewModule("test_table")
		result := module.WithTags(tags)

		assert.Equal(t, module, result)
		assert.NotNil(t, module.Tags)
		assert.Equal(t, "production", module.Tags["Environment"])
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("merges tags when called multiple times", func(t *testing.T) {
		module := NewModule("test_table")

		module.WithTags(map[string]string{"Key1": "value1"})
		module.WithTags(map[string]string{"Key2": "value2"})

		assert.Equal(t, "value1", module.Tags["Key1"])
		assert.Equal(t, "value2", module.Tags["Key2"])
	})
}

func TestModule_LocalName(t *testing.T) {
	t.Run("returns table name when set", func(t *testing.T) {
		name := "my_table"
		module := NewModule(name)

		assert.Equal(t, name, module.LocalName())
	})

	t.Run("returns default when name is nil", func(t *testing.T) {
		module := &Module{}

		assert.Equal(t, "dynamodb_table", module.LocalName())
	})
}

func TestModule_Configuration(t *testing.T) {
	t.Run("returns empty string and nil error as placeholder", func(t *testing.T) {
		module := NewModule("test_table")

		config, err := module.Configuration()

		require.NoError(t, err)
		assert.Empty(t, config)
	})
}

func TestModule_FluentAPI(t *testing.T) {
	t.Run("supports complete fluent configuration", func(t *testing.T) {
		module := NewModule("users").
			WithHashKey("userId", "S").
			WithRangeKey("timestamp", "N").
			WithStreams("NEW_AND_OLD_IMAGES").
			WithTTL("expiresAt").
			WithEncryption("alias/aws/dynamodb").
			WithTags(map[string]string{"Team": "platform"})

		assert.NotNil(t, module.Name)
		assert.Equal(t, "users", *module.Name)
		assert.Equal(t, "userId", *module.HashKey)
		assert.Equal(t, "timestamp", *module.RangeKey)
		assert.True(t, *module.StreamEnabled)
		assert.True(t, *module.TTLEnabled)
		assert.Equal(t, "platform", module.Tags["Team"])
	})

	t.Run("supports GSI configuration", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:           "email-index",
			HashKey:        "email",
			ProjectionType: "ALL",
		}

		module := NewModule("users").
			WithHashKey("userId", "S").
			WithGSI(gsi)

		assert.Len(t, module.GlobalSecondaryIndexes, 1)
	})

	t.Run("supports provisioned billing", func(t *testing.T) {
		module := NewModule("users").
			WithHashKey("userId", "S").
			WithProvisioned(10, 20)

		assert.Equal(t, "PROVISIONED", *module.BillingMode)
		assert.Equal(t, 10, *module.ReadCapacity)
	})
}

func TestAttribute(t *testing.T) {
	t.Run("creates string attribute", func(t *testing.T) {
		attr := Attribute{
			Name: "userId",
			Type: "S",
		}

		assert.Equal(t, "userId", attr.Name)
		assert.Equal(t, "S", attr.Type)
	})

	t.Run("creates number attribute", func(t *testing.T) {
		attr := Attribute{
			Name: "count",
			Type: "N",
		}

		assert.Equal(t, "N", attr.Type)
	})

	t.Run("creates binary attribute", func(t *testing.T) {
		attr := Attribute{
			Name: "data",
			Type: "B",
		}

		assert.Equal(t, "B", attr.Type)
	})
}

func TestGlobalSecondaryIndex(t *testing.T) {
	t.Run("creates GSI with ALL projection", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:           "status-index",
			HashKey:        "status",
			ProjectionType: "ALL",
		}

		assert.Equal(t, "status-index", gsi.Name)
		assert.Equal(t, "ALL", gsi.ProjectionType)
		assert.Nil(t, gsi.RangeKey)
	})

	t.Run("creates GSI with KEYS_ONLY projection", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:           "keys-index",
			HashKey:        "pk",
			ProjectionType: "KEYS_ONLY",
		}

		assert.Equal(t, "KEYS_ONLY", gsi.ProjectionType)
	})

	t.Run("creates GSI with INCLUDE projection", func(t *testing.T) {
		gsi := GlobalSecondaryIndex{
			Name:             "include-index",
			HashKey:          "gsiPk",
			ProjectionType:   "INCLUDE",
			NonKeyAttributes: []string{"attr1", "attr2", "attr3"},
		}

		assert.Equal(t, "INCLUDE", gsi.ProjectionType)
		assert.Len(t, gsi.NonKeyAttributes, 3)
	})

	t.Run("creates GSI with provisioned capacity", func(t *testing.T) {
		readCap := 5
		writeCap := 10

		gsi := GlobalSecondaryIndex{
			Name:           "provisioned-index",
			HashKey:        "pk",
			ProjectionType: "ALL",
			ReadCapacity:   &readCap,
			WriteCapacity:  &writeCap,
		}

		assert.Equal(t, 5, *gsi.ReadCapacity)
		assert.Equal(t, 10, *gsi.WriteCapacity)
	})
}

func TestLocalSecondaryIndex(t *testing.T) {
	t.Run("creates LSI", func(t *testing.T) {
		lsi := LocalSecondaryIndex{
			Name:           "local-index",
			RangeKey:       "sortKey",
			ProjectionType: "ALL",
		}

		assert.Equal(t, "local-index", lsi.Name)
		assert.Equal(t, "sortKey", lsi.RangeKey)
		assert.Equal(t, "ALL", lsi.ProjectionType)
	})
}

func TestReplicaRegion(t *testing.T) {
	t.Run("creates replica configuration", func(t *testing.T) {
		kmsKey := "arn:aws:kms:us-west-2:123456789012:key/12345"
		pitr := true

		replica := ReplicaRegion{
			RegionName:                 "us-west-2",
			KMSKeyARN:                  &kmsKey,
			PointInTimeRecoveryEnabled: &pitr,
		}

		assert.Equal(t, "us-west-2", replica.RegionName)
		assert.Equal(t, kmsKey, *replica.KMSKeyARN)
		assert.True(t, *replica.PointInTimeRecoveryEnabled)
	})
}

// BenchmarkNewModule benchmarks module creation.
func BenchmarkNewModule(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_table")
	}
}

// BenchmarkFluentAPI benchmarks fluent API calls.
func BenchmarkFluentAPI(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewModule("bench_table").
			WithHashKey("userId", "S").
			WithRangeKey("timestamp", "N").
			WithStreams("NEW_AND_OLD_IMAGES").
			WithTags(map[string]string{"Environment": "production"})
	}
}
