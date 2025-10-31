package hclgen_test

import (
	"strings"
	"testing"

	"github.com/lewis/forge/internal/tfmodules/hclgen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToHCL_SimpleModule tests basic module generation.
func TestToHCL_SimpleModule(t *testing.T) {
	type SimpleModule struct {
		Name    *string           `hcl:"name,attr"`
		Enabled *bool             `hcl:"enabled,attr"`
		Tags    map[string]string `hcl:"tags,attr"`
	}

	name := "test-module"
	enabled := true
	module := SimpleModule{
		Name:    &name,
		Enabled: &enabled,
		Tags: map[string]string{
			"Environment": "test",
			"ManagedBy":   "forge",
		},
	}

	hcl, err := hclgen.ToHCL("my_module", "terraform-aws-modules/test/aws", "~> 1.0", module)
	require.NoError(t, err)

	// Verify structure
	assert.Contains(t, hcl, `module "my_module" {`)
	assert.Contains(t, hcl, `source  = "terraform-aws-modules/test/aws"`)
	assert.Contains(t, hcl, `version = "~> 1.0"`)
	assert.Contains(t, hcl, `name = "test-module"`)
	assert.Contains(t, hcl, `enabled = true`)
	assert.Contains(t, hcl, `tags = {`)
	assert.Contains(t, hcl, `Environment = "test"`)
	assert.Contains(t, hcl, `ManagedBy = "forge"`)
}

// TestToHCL_WithNilFields tests that nil pointer fields are skipped.
func TestToHCL_WithNilFields(t *testing.T) {
	type ModuleWithOptionals struct {
		Name     *string `hcl:"name,attr"`
		Optional *string `hcl:"optional,attr"`
		Required *bool   `hcl:"required,attr"`
	}

	name := "test"
	required := true
	module := ModuleWithOptionals{
		Name:     &name,
		Optional: nil, // Should be skipped
		Required: &required,
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `name = "test"`)
	assert.Contains(t, hcl, `required = true`)
	assert.NotContains(t, hcl, "optional") // Nil fields are skipped
}

// TestToHCL_WithTerraformReferences tests Terraform variable references.
func TestToHCL_WithTerraformReferences(t *testing.T) {
	type ModuleWithRefs struct {
		BucketName *string `hcl:"bucket_name,attr"`
		TableARN   *string `hcl:"table_arn,attr"`
	}

	bucketName := "${var.namespace}my-bucket"
	tableARN := "${module.dynamodb.table_arn}"
	module := ModuleWithRefs{
		BucketName: &bucketName,
		TableARN:   &tableARN,
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	// Terraform references should not be quoted
	assert.Contains(t, hcl, `bucket_name = ${var.namespace}my-bucket`)
	assert.Contains(t, hcl, `table_arn = ${module.dynamodb.table_arn}`)
	assert.NotContains(t, hcl, `"${`) // Should not have quotes around references
}

// TestToHCL_WithNestedStructs tests nested struct conversion.
func TestToHCL_WithNestedStructs(t *testing.T) {
	type Attribute struct {
		Name string `hcl:"name,attr"`
		Type string `hcl:"type,attr"`
	}

	type ModuleWithNested struct {
		TableName  *string     `hcl:"table_name,attr"`
		Attributes []Attribute `hcl:"attribute,block"`
	}

	tableName := "users"
	module := ModuleWithNested{
		TableName: &tableName,
		Attributes: []Attribute{
			{Name: "id", Type: "S"},
			{Name: "email", Type: "S"},
		},
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `table_name = "users"`)
	assert.Contains(t, hcl, `attribute {`)
	assert.Contains(t, hcl, `name = "id"`)
	assert.Contains(t, hcl, `type = "S"`)

	// Count occurrences of "attribute {" - should be 2
	count := strings.Count(hcl, "attribute {")
	assert.Equal(t, 2, count, "Should have 2 attribute blocks")
}

// TestToHCL_WithIntegerFields tests integer field conversion.
func TestToHCL_WithIntegerFields(t *testing.T) {
	type ModuleWithInts struct {
		Timeout    *int `hcl:"timeout,attr"`
		MemorySize *int `hcl:"memory_size,attr"`
	}

	timeout := 30
	memory := 512
	module := ModuleWithInts{
		Timeout:    &timeout,
		MemorySize: &memory,
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `timeout = 30`)
	assert.Contains(t, hcl, `memory_size = 512`)
	assert.NotContains(t, hcl, `"30"`) // Numbers should not be quoted
}

// TestToHCL_WithEmptySlicesAndMaps tests that empty collections are skipped.
func TestToHCL_WithEmptySlicesAndMaps(t *testing.T) {
	type ModuleWithCollections struct {
		Name       *string           `hcl:"name,attr"`
		EmptyMap   map[string]string `hcl:"empty_map,attr"`
		EmptySlice []string          `hcl:"empty_slice,attr"`
	}

	name := "test"
	module := ModuleWithCollections{
		Name:       &name,
		EmptyMap:   map[string]string{},
		EmptySlice: []string{},
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `name = "test"`)
	assert.NotContains(t, hcl, "empty_map")   // Empty maps are skipped
	assert.NotContains(t, hcl, "empty_slice") // Empty slices are skipped
}

// TestToHCL_WithStringSlice tests string array conversion.
func TestToHCL_WithStringSlice(t *testing.T) {
	type ModuleWithStrings struct {
		Actions []string `hcl:"actions,attr"`
	}

	module := ModuleWithStrings{
		Actions: []string{"s3:GetObject", "s3:PutObject", "s3:DeleteObject"},
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `actions = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]`)
}

// TestToHCL_WithNestedMaps tests nested map conversion.
func TestToHCL_WithNestedMaps(t *testing.T) {
	type ModuleWithNestedMap struct {
		Config map[string]interface{} `hcl:"config,attr"`
	}

	module := ModuleWithNestedMap{
		Config: map[string]interface{}{
			"enabled": true,
			"timeout": 30,
			"nested": map[string]interface{}{
				"key": "value",
			},
		},
	}

	hcl, err := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, hcl, `config = {`)
	assert.Contains(t, hcl, `enabled = true`)
	assert.Contains(t, hcl, `timeout = 30`)
}

// TestToHCL_DeterministicOrdering tests that output is deterministic.
func TestToHCL_DeterministicOrdering(t *testing.T) {
	type Module struct {
		Zebra   *string `hcl:"zebra,attr"`
		Alpha   *string `hcl:"alpha,attr"`
		Beta    *string `hcl:"beta,attr"`
		Charlie *string `hcl:"charlie,attr"`
	}

	zebra := "z"
	alpha := "a"
	beta := "b"
	charlie := "c"

	module := Module{
		Zebra:   &zebra,
		Alpha:   &alpha,
		Beta:    &beta,
		Charlie: &charlie,
	}

	hcl1, err1 := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err1)

	hcl2, err2 := hclgen.ToHCL("test", "source", "", module)
	require.NoError(t, err2)

	// Should be identical - deterministic output
	assert.Equal(t, hcl1, hcl2)

	// Should be in alphabetical order
	alphaPos := strings.Index(hcl1, "alpha")
	betaPos := strings.Index(hcl1, "beta")
	charliePos := strings.Index(hcl1, "charlie")
	zebraPos := strings.Index(hcl1, "zebra")

	assert.Less(t, alphaPos, betaPos, "alpha should come before beta")
	assert.Less(t, betaPos, charliePos, "beta should come before charlie")
	assert.Less(t, charliePos, zebraPos, "charlie should come before zebra")
}

// TestToHCL_SkipsSpecialFields tests that Source, Version, Region are skipped.
func TestToHCL_SkipsSpecialFields(t *testing.T) {
	type ModuleWithSpecialFields struct {
		Source  string  `hcl:"source,attr"`
		Version string  `hcl:"version,attr"`
		Region  *string `hcl:"region,attr"`
		Name    *string `hcl:"name,attr"`
	}

	name := "test"
	region := "us-east-1"
	module := ModuleWithSpecialFields{
		Source:  "should-be-skipped",
		Version: "should-be-skipped",
		Region:  &region,
		Name:    &name,
	}

	hcl, err := hclgen.ToHCL("test", "actual-source", "actual-version", module)
	require.NoError(t, err)

	// Should use provided source/version, not from struct
	assert.Contains(t, hcl, `source  = "actual-source"`)
	assert.Contains(t, hcl, `version = "actual-version"`)
	assert.NotContains(t, hcl, "should-be-skipped")

	// Region should be skipped (special field)
	assert.NotContains(t, hcl, "region")

	// Name should be included
	assert.Contains(t, hcl, `name = "test"`)
}

// TestToHCL_ComplexRealWorldExample tests a complex real-world module.
func TestToHCL_ComplexRealWorldExample(t *testing.T) {
	type Attribute struct {
		Name string `hcl:"name,attr"`
		Type string `hcl:"type,attr"`
	}

	type GSI struct {
		Name           string   `hcl:"name,attr"`
		HashKey        string   `hcl:"hash_key,attr"`
		RangeKey       *string  `hcl:"range_key,attr"`
		ProjectionType string   `hcl:"projection_type,attr"`
		ReadCapacity   *int     `hcl:"read_capacity,attr"`
		WriteCapacity  *int     `hcl:"write_capacity,attr"`
	}

	type DynamoDBModule struct {
		Name                        *string            `hcl:"name,attr"`
		BillingMode                 *string            `hcl:"billing_mode,attr"`
		HashKey                     *string            `hcl:"hash_key,attr"`
		RangeKey                    *string            `hcl:"range_key,attr"`
		Attributes                  []Attribute        `hcl:"attributes,attr"`
		GlobalSecondaryIndexes      []GSI              `hcl:"global_secondary_indexes,attr"`
		StreamEnabled               *bool              `hcl:"stream_enabled,attr"`
		StreamViewType              *string            `hcl:"stream_view_type,attr"`
		PointInTimeRecoveryEnabled  *bool              `hcl:"point_in_time_recovery_enabled,attr"`
		Tags                        map[string]string  `hcl:"tags,attr"`
	}

	name := "${var.namespace}users-table"
	billingMode := "PAY_PER_REQUEST"
	hashKey := "userId"
	rangeKey := "timestamp"
	streamEnabled := true
	streamViewType := "NEW_AND_OLD_IMAGES"
	pitrEnabled := true
	rangeKeyGSI := "status"

	module := DynamoDBModule{
		Name:        &name,
		BillingMode: &billingMode,
		HashKey:     &hashKey,
		RangeKey:    &rangeKey,
		Attributes: []Attribute{
			{Name: "userId", Type: "S"},
			{Name: "timestamp", Type: "N"},
			{Name: "status", Type: "S"},
		},
		GlobalSecondaryIndexes: []GSI{
			{
				Name:           "status-index",
				HashKey:        "status",
				RangeKey:       &rangeKeyGSI,
				ProjectionType: "ALL",
			},
		},
		StreamEnabled:              &streamEnabled,
		StreamViewType:             &streamViewType,
		PointInTimeRecoveryEnabled: &pitrEnabled,
		Tags: map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		},
	}

	hcl, err := hclgen.ToHCL("users_table", "terraform-aws-modules/dynamodb-table/aws", "~> 4.0", module)
	require.NoError(t, err)

	// Verify key components
	assert.Contains(t, hcl, `module "users_table" {`)
	assert.Contains(t, hcl, `source  = "terraform-aws-modules/dynamodb-table/aws"`)
	assert.Contains(t, hcl, `version = "~> 4.0"`)
	assert.Contains(t, hcl, `name = ${var.namespace}users-table`) // No quotes for Terraform refs
	assert.Contains(t, hcl, `billing_mode = "PAY_PER_REQUEST"`)
	assert.Contains(t, hcl, `hash_key = "userId"`)
	assert.Contains(t, hcl, `range_key = "timestamp"`)
	assert.Contains(t, hcl, `stream_enabled = true`)
	assert.Contains(t, hcl, `stream_view_type = "NEW_AND_OLD_IMAGES"`)
	assert.Contains(t, hcl, `point_in_time_recovery_enabled = true`)
	assert.Contains(t, hcl, `Environment = "production"`)
	assert.Contains(t, hcl, `ManagedBy = "forge"`)

	// Verify structure (attributes should be an array of objects)
	assert.Contains(t, hcl, `attributes = [`)
	assert.Contains(t, hcl, `name = "userId"`)
	assert.Contains(t, hcl, `type = "S"`)

	// GSIs should be an array of objects
	assert.Contains(t, hcl, `global_secondary_indexes = [`)
	assert.Contains(t, hcl, `name = "status-index"`)
	assert.Contains(t, hcl, `projection_type = "ALL"`)
}
