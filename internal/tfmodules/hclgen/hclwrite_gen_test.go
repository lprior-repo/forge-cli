package hclgen_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/tfmodules/hclgen"
)

// TestToHCLWrite_BasicModule tests basic module generation with hclwrite.
func TestToHCLWrite_BasicModule(t *testing.T) {
	type SimpleModule struct {
		FunctionName *string `hcl:"function_name,attr"`
		Runtime      *string `hcl:"runtime,attr"`
	}

	functionName := "test-function"
	runtime := "python3.13"

	module := &SimpleModule{
		FunctionName: &functionName,
		Runtime:      &runtime,
	}

	result, err := hclgen.ToHCLWrite("test_lambda", "terraform-aws-modules/lambda/aws", "7.16.0", module)
	require.NoError(t, err)

	// Check that result contains expected content
	assert.Contains(t, result, `module "test_lambda"`)
	assert.Contains(t, result, `source`)
	assert.Contains(t, result, `terraform-aws-modules/lambda/aws`)
	assert.Contains(t, result, `version`)
	assert.Contains(t, result, `7.16.0`)
	assert.Contains(t, result, `function_name`)
	assert.Contains(t, result, `test-function`)
	assert.Contains(t, result, `runtime`)
	assert.Contains(t, result, `python3.13`)
}

// TestToHCLWrite_NilFields tests that nil pointer fields are omitted.
func TestToHCLWrite_NilFields(t *testing.T) {
	type ModuleWithNils struct {
		Required *string `hcl:"required,attr"`
		Optional *string `hcl:"optional,attr"`
	}

	required := "value"
	module := &ModuleWithNils{
		Required: &required,
		Optional: nil, // Should be omitted
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "required")
	assert.NotContains(t, result, "optional") // Nil fields are skipped
}

// TestToHCLWrite_EmptySlices tests that empty slices are omitted.
func TestToHCLWrite_EmptySlices(t *testing.T) {
	type ModuleWithSlice struct {
		Tags []string `hcl:"tags,attr"`
	}

	module := &ModuleWithSlice{
		Tags: []string{}, // Empty - should be omitted
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.NotContains(t, result, "tags")
}

// TestToHCLWrite_EmptyMaps tests that empty maps are omitted.
func TestToHCLWrite_EmptyMaps(t *testing.T) {
	type ModuleWithMap struct {
		Config map[string]string `hcl:"config,attr"`
	}

	module := &ModuleWithMap{
		Config: map[string]string{}, // Empty - should be omitted
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.NotContains(t, result, "config")
}

// TestToHCLWrite_Numbers tests number type handling.
func TestToHCLWrite_Numbers(t *testing.T) {
	type ModuleWithNumbers struct {
		Timeout    *int     `hcl:"timeout,attr"`
		MemorySize *int     `hcl:"memory_size,attr"`
		RateLimit  *float64 `hcl:"rate_limit,attr"`
	}

	timeout := 30
	memorySize := 1024
	rateLimit := 100.5
	module := &ModuleWithNumbers{
		Timeout:    &timeout,
		MemorySize: &memorySize,
		RateLimit:  &rateLimit,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "timeout")
	assert.Contains(t, result, "memory_size")
	assert.Contains(t, result, "rate_limit")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "1024")
	assert.Contains(t, result, "100.5")
}

// TestToHCLWrite_Booleans tests boolean type handling.
func TestToHCLWrite_Booleans(t *testing.T) {
	type ModuleWithBools struct {
		Enabled  *bool `hcl:"enabled,attr"`
		Disabled *bool `hcl:"disabled,attr"`
	}

	enabled := true
	disabled := false
	module := &ModuleWithBools{
		Enabled:  &enabled,
		Disabled: &disabled,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "enabled")
	assert.Contains(t, result, "disabled")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "false")
}

// TestToHCLWrite_StringSlice tests string array handling.
func TestToHCLWrite_StringSlice(t *testing.T) {
	type ModuleWithSlice struct {
		Actions []string `hcl:"actions,attr"`
	}

	module := &ModuleWithSlice{
		Actions: []string{"s3:GetObject", "s3:PutObject"},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "actions")
	assert.Contains(t, result, "s3:GetObject")
	assert.Contains(t, result, "s3:PutObject")
}

// TestToHCLWrite_StringMap tests string map handling.
func TestToHCLWrite_StringMap(t *testing.T) {
	type ModuleWithMap struct {
		Tags map[string]string `hcl:"tags,attr"`
	}

	module := &ModuleWithMap{
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "platform",
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "tags")
	assert.Contains(t, result, "Environment")
	assert.Contains(t, result, "production")
	assert.Contains(t, result, "Team")
	assert.Contains(t, result, "platform")
}

// TestToHCLWrite_TerraformReferences tests Terraform reference handling.
func TestToHCLWrite_TerraformReferences(t *testing.T) {
	type ModuleWithRefs struct {
		TableName   *string `hcl:"table_name,attr"`
		BucketName  *string `hcl:"bucket_name,attr"`
		VarRef      *string `hcl:"var_ref,attr"`
		DataRef     *string `hcl:"data_ref,attr"`
		LocalRef    *string `hcl:"local_ref,attr"`
		ResourceRef *string `hcl:"resource_ref,attr"`
	}

	tableName := "${module.dynamodb.table_name}"
	bucketName := "module.s3.bucket_id"
	varRef := "var.namespace"
	dataRef := "data.aws_caller_identity.current.account_id"
	localRef := "local.region"
	resourceRef := "resource.aws_iam_role.lambda.arn"

	module := &ModuleWithRefs{
		TableName:   &tableName,
		BucketName:  &bucketName,
		VarRef:      &varRef,
		DataRef:     &dataRef,
		LocalRef:    &localRef,
		ResourceRef: &resourceRef,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	// Terraform references should be unquoted
	assert.Contains(t, result, "table_name")
	assert.Contains(t, result, "module.dynamodb.table_name")
	assert.NotContains(t, result, `"module.dynamodb.table_name"`)

	assert.Contains(t, result, "bucket_name")
	assert.Contains(t, result, "module.s3.bucket_id")

	assert.Contains(t, result, "var_ref")
	assert.Contains(t, result, "var.namespace")

	assert.Contains(t, result, "data_ref")
	assert.Contains(t, result, "data.aws_caller_identity.current.account_id")

	assert.Contains(t, result, "local_ref")
	assert.Contains(t, result, "local.region")

	assert.Contains(t, result, "resource_ref")
	assert.Contains(t, result, "resource.aws_iam_role.lambda.arn")
}

// TestToHCLWrite_SliceWithReferences tests slices containing Terraform references.
func TestToHCLWrite_SliceWithReferences(t *testing.T) {
	type ModuleWithRefSlice struct {
		Layers []string `hcl:"layers,attr"`
	}

	module := &ModuleWithRefSlice{
		Layers: []string{
			"arn:aws:lambda:us-east-1:123:layer:test",
			"${module.layer.arn}",
			"module.other_layer.arn",
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "layers")
	// Static ARN should be quoted
	assert.Contains(t, result, `"arn:aws:lambda:us-east-1:123:layer:test"`)
	// References should be unquoted
	assert.Contains(t, result, "module.layer.arn")
	assert.Contains(t, result, "module.other_layer.arn")
	// Should not double-quote references
	assert.NotContains(t, result, `"module.layer.arn"`)
}

// TestToHCLWrite_MapWithReferences tests maps containing Terraform references.
func TestToHCLWrite_MapWithReferences(t *testing.T) {
	type ModuleWithRefMap struct {
		Environment map[string]string `hcl:"environment,attr"`
	}

	module := &ModuleWithRefMap{
		Environment: map[string]string{
			"TABLE_NAME": "${module.dynamodb.table_name}",
			"BUCKET":     "module.s3.bucket_id",
			"REGION":     "us-east-1",
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "environment")
	assert.Contains(t, result, "TABLE_NAME")
	assert.Contains(t, result, "module.dynamodb.table_name")
	assert.Contains(t, result, "BUCKET")
	assert.Contains(t, result, "module.s3.bucket_id")
	assert.Contains(t, result, "REGION")
	assert.Contains(t, result, `"us-east-1"`) // Static value should be quoted
	// References should not be quoted
	assert.NotContains(t, result, `"module.dynamodb.table_name"`)
	assert.NotContains(t, result, `"module.s3.bucket_id"`)
}

// TestToHCLWrite_NestedStruct tests nested struct as attribute.
func TestToHCLWrite_NestedStruct(t *testing.T) {
	type CORS struct {
		AllowOrigins []string `hcl:"allow_origins,attr"`
		AllowMethods []string `hcl:"allow_methods,attr"`
		MaxAge       *int     `hcl:"max_age,attr"`
	}

	type ModuleWithNested struct {
		FunctionName *string `hcl:"function_name,attr"`
		CORS         *CORS   `hcl:"cors,attr"`
	}

	functionName := "test"
	maxAge := 3600
	module := &ModuleWithNested{
		FunctionName: &functionName,
		CORS: &CORS{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST"},
			MaxAge:       &maxAge,
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.Contains(t, result, "cors")
	assert.Contains(t, result, "allow_origins")
	assert.Contains(t, result, "allow_methods")
	assert.Contains(t, result, "max_age")
}

// TestToHCLWrite_NestedBlock tests nested block generation.
func TestToHCLWrite_NestedBlock(t *testing.T) {
	type Environment struct {
		Variables map[string]string `hcl:"variables,attr"`
	}

	type ModuleWithBlock struct {
		FunctionName *string      `hcl:"function_name,attr"`
		Environment  *Environment `hcl:"environment,block"`
	}

	functionName := "test"
	module := &ModuleWithBlock{
		FunctionName: &functionName,
		Environment: &Environment{
			Variables: map[string]string{
				"KEY": "value",
			},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.Contains(t, result, "environment {")
	assert.Contains(t, result, "variables")
	assert.Contains(t, result, "KEY")
	assert.Contains(t, result, "value")
}

// TestToHCLWrite_SliceOfStructsBlock tests repeated blocks from slice.
func TestToHCLWrite_SliceOfStructsBlock(t *testing.T) {
	type Attribute struct {
		Name string `hcl:"name,attr"`
		Type string `hcl:"type,attr"`
	}

	type ModuleWithSliceBlock struct {
		TableName  *string     `hcl:"table_name,attr"`
		Attributes []Attribute `hcl:"attribute,block"`
	}

	tableName := "users"
	module := &ModuleWithSliceBlock{
		TableName: &tableName,
		Attributes: []Attribute{
			{Name: "id", Type: "S"},
			{Name: "email", Type: "S"},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "table_name")
	assert.Contains(t, result, "attribute {")
	assert.Contains(t, result, `name = "id"`)
	assert.Contains(t, result, `type = "S"`)
	assert.Contains(t, result, `name = "email"`)

	// Count occurrences of "attribute {" - should be 2
	count := strings.Count(result, "attribute {")
	assert.Equal(t, 2, count, "Should have 2 attribute blocks")
}

// TestToHCLWrite_MapBlocks tests blocks from map with labels.
func TestToHCLWrite_MapBlocks(t *testing.T) {
	type PolicyStatement struct {
		Effect  *string  `hcl:"effect,attr"`
		Actions []string `hcl:"actions,attr"`
	}

	type ModuleWithMapBlocks struct {
		FunctionName     *string                    `hcl:"function_name,attr"`
		PolicyStatements map[string]PolicyStatement `hcl:"policy_statement,block"`
	}

	functionName := "test"
	effect := "Allow"
	module := &ModuleWithMapBlocks{
		FunctionName: &functionName,
		PolicyStatements: map[string]PolicyStatement{
			"dynamodb": {
				Effect:  &effect,
				Actions: []string{"dynamodb:GetItem", "dynamodb:PutItem"},
			},
			"s3": {
				Effect:  &effect,
				Actions: []string{"s3:GetObject"},
			},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.Contains(t, result, `policy_statement "dynamodb"`)
	assert.Contains(t, result, `policy_statement "s3"`)
	assert.Contains(t, result, "dynamodb:GetItem")
	assert.Contains(t, result, "s3:GetObject")
}

// TestToHCLWrite_DeterministicOrdering tests that output is deterministic.
func TestToHCLWrite_DeterministicOrdering(t *testing.T) {
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

	module := &Module{
		Zebra:   &zebra,
		Alpha:   &alpha,
		Beta:    &beta,
		Charlie: &charlie,
	}

	hcl1, err1 := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err1)

	hcl2, err2 := hclgen.ToHCLWrite("test", "source", "", module)
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

// TestToHCLWrite_SkipsSpecialFields tests that Source, Version, Region are skipped.
func TestToHCLWrite_SkipsSpecialFields(t *testing.T) {
	type ModuleWithSpecialFields struct {
		Source  string  `hcl:"source,attr"`
		Version string  `hcl:"version,attr"`
		Region  *string `hcl:"region,attr"`
		Name    *string `hcl:"name,attr"`
	}

	name := "test"
	region := "us-east-1"
	module := &ModuleWithSpecialFields{
		Source:  "should-be-skipped",
		Version: "should-be-skipped",
		Region:  &region,
		Name:    &name,
	}

	hcl, err := hclgen.ToHCLWrite("test", "actual-source", "actual-version", module)
	require.NoError(t, err)

	// Should use provided source/version, not from struct
	assert.Contains(t, hcl, `"actual-source"`)
	assert.Contains(t, hcl, `"actual-version"`)
	assert.NotContains(t, hcl, "should-be-skipped")

	// Region should be skipped (special field)
	assert.NotContains(t, hcl, "region")

	// Name should be included
	assert.Contains(t, hcl, "name")
	assert.Contains(t, hcl, `"test"`)
}

// TestToHCLWrite_WithoutVersion tests module without version constraint.
func TestToHCLWrite_WithoutVersion(t *testing.T) {
	type SimpleModule struct {
		Name *string `hcl:"name,attr"`
	}

	name := "test"
	module := &SimpleModule{
		Name: &name,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, `source = "source"`)
	assert.NotContains(t, result, "version")
}

// TestToHCLWrite_ComplexRealWorld tests a complex real-world module.
func TestToHCLWrite_ComplexRealWorld(t *testing.T) {
	type CORSConfig struct {
		AllowOrigins []string `hcl:"allow_origins,attr"`
		AllowMethods []string `hcl:"allow_methods,attr"`
		MaxAge       *int     `hcl:"max_age,attr"`
	}

	type LambdaModule struct {
		FunctionName         *string           `hcl:"function_name,attr"`
		Runtime              *string           `hcl:"runtime,attr"`
		Handler              *string           `hcl:"handler,attr"`
		Timeout              *int              `hcl:"timeout,attr"`
		MemorySize           *int              `hcl:"memory_size,attr"`
		EnvironmentVariables map[string]string `hcl:"environment_variables,attr"`
		Layers               []string          `hcl:"layers,attr"`
		Tags                 map[string]string `hcl:"tags,attr"`
		CORS                 *CORSConfig       `hcl:"cors,attr"`
	}

	functionName := "api-handler"
	runtime := "python3.13"
	handler := "app.handler"
	timeout := 30
	memory := 1024
	maxAge := 3600

	module := &LambdaModule{
		FunctionName: &functionName,
		Runtime:      &runtime,
		Handler:      &handler,
		Timeout:      &timeout,
		MemorySize:   &memory,
		EnvironmentVariables: map[string]string{
			"TABLE_NAME":  "${module.dynamodb.table_name}",
			"BUCKET_NAME": "module.s3.bucket_id",
			"REGION":      "us-east-1",
		},
		Layers: []string{
			"arn:aws:lambda:us-east-1:123:layer:base",
			"${module.utils_layer.arn}",
		},
		Tags: map[string]string{
			"Environment": "production",
			"Team":        "platform",
		},
		CORS: &CORSConfig{
			AllowOrigins: []string{"https://example.com"},
			AllowMethods: []string{"GET", "POST", "PUT"},
			MaxAge:       &maxAge,
		},
	}

	result, err := hclgen.ToHCLWrite("api_lambda", "terraform-aws-modules/lambda/aws", "~> 7.0", module)
	require.NoError(t, err)

	// Verify key components
	assert.Contains(t, result, `module "api_lambda"`)
	assert.Contains(t, result, `"terraform-aws-modules/lambda/aws"`)
	assert.Contains(t, result, `"~> 7.0"`)
	assert.Contains(t, result, `"api-handler"`)
	assert.Contains(t, result, `"python3.13"`)
	assert.Contains(t, result, `"app.handler"`)
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "1024")

	// Environment variables with references
	assert.Contains(t, result, "environment_variables")
	assert.Contains(t, result, "TABLE_NAME")
	assert.Contains(t, result, "module.dynamodb.table_name")
	assert.Contains(t, result, "REGION")
	assert.Contains(t, result, `"us-east-1"`)

	// Layers with references
	assert.Contains(t, result, "layers")
	assert.Contains(t, result, `"arn:aws:lambda:us-east-1:123:layer:base"`)
	assert.Contains(t, result, "module.utils_layer.arn")

	// Tags
	assert.Contains(t, result, "tags")
	assert.Contains(t, result, "Environment")
	assert.Contains(t, result, `"production"`)
	assert.Contains(t, result, "Team")
	assert.Contains(t, result, `"platform"`)

	// CORS nested struct
	assert.Contains(t, result, "cors")
	assert.Contains(t, result, "allow_origins")
	assert.Contains(t, result, "allow_methods")
	assert.Contains(t, result, "3600")
}

// TestToHCLWrite_ErrorCases tests error handling.
func TestToHCLWrite_ErrorCases(t *testing.T) {
	// Non-struct input
	_, err := hclgen.ToHCLWrite("test", "source", "", "not a struct")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "expected struct")
}

// TestToHCLWrite_InterfaceMapValues tests map[string]interface{} handling.
func TestToHCLWrite_InterfaceMapValues(t *testing.T) {
	type ModuleWithInterfaceMap struct {
		Config map[string]interface{} `hcl:"config,attr"`
	}

	module := &ModuleWithInterfaceMap{
		Config: map[string]interface{}{
			"enabled": true,
			"timeout": 30,
			"name":    "test",
			"ref":     "${var.namespace}",
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "config")
	assert.Contains(t, result, "enabled")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "timeout")
	assert.Contains(t, result, "30")
	assert.Contains(t, result, "name")
	assert.Contains(t, result, `"test"`)
	assert.Contains(t, result, "var.namespace")
	assert.NotContains(t, result, `"var.namespace"`)
}

// TestToHCLWrite_PointerToStruct tests pointer to struct input.
func TestToHCLWrite_PointerToStruct(t *testing.T) {
	type SimpleModule struct {
		Name *string `hcl:"name,attr"`
	}

	name := "test"
	module := &SimpleModule{
		Name: &name,
	}

	// Should work with pointer
	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, `"test"`)
}

// TestToHCLWrite_ValueStruct tests value struct input.
func TestToHCLWrite_ValueStruct(t *testing.T) {
	type SimpleModule struct {
		Name *string `hcl:"name,attr"`
	}

	name := "test"
	module := SimpleModule{
		Name: &name,
	}

	// Should work with value
	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)
	assert.Contains(t, result, "name")
	assert.Contains(t, result, `"test"`)
}

// TestToHCLWrite_SkipsHiddenTag tests that hcl:"-" skips field.
func TestToHCLWrite_SkipsHiddenTag(t *testing.T) {
	type ModuleWithHidden struct {
		Visible *string `hcl:"visible,attr"`
		Hidden  *string `hcl:"-"`
	}

	visible := "yes"
	hidden := "no"
	module := &ModuleWithHidden{
		Visible: &visible,
		Hidden:  &hidden,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "visible")
	assert.NotContains(t, result, "Hidden")
	assert.NotContains(t, result, "no")
}

// TestToHCLWrite_SkipsNoTag tests that fields without hcl tag are skipped.
func TestToHCLWrite_SkipsNoTag(t *testing.T) {
	type ModuleWithNoTag struct {
		WithTag *string `hcl:"with_tag,attr"`
		NoTag   *string // No hcl tag
	}

	withTag := "yes"
	noTag := "no"
	module := &ModuleWithNoTag{
		WithTag: &withTag,
		NoTag:   &noTag,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "with_tag")
	assert.NotContains(t, result, "NoTag")
	assert.NotContains(t, result, "no")
}

// TestToHCLWrite_Uint tests unsigned integer handling.
func TestToHCLWrite_Uint(t *testing.T) {
	type ModuleWithUint struct {
		Port *uint16 `hcl:"port,attr"`
	}

	var port uint16 = 8080
	module := &ModuleWithUint{
		Port: &port,
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "port")
	assert.Contains(t, result, "8080")
}

// TestToHCLWrite_NilBlockPointer tests that nil block pointers are skipped.
func TestToHCLWrite_NilBlockPointer(t *testing.T) {
	type Environment struct {
		Variables map[string]string `hcl:"variables,attr"`
	}

	type ModuleWithNilBlock struct {
		FunctionName *string      `hcl:"function_name,attr"`
		Environment  *Environment `hcl:"environment,block"`
	}

	functionName := "test"
	module := &ModuleWithNilBlock{
		FunctionName: &functionName,
		Environment:  nil, // Nil block should be skipped
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.NotContains(t, result, "environment")
}

// TestToHCLWrite_MapKeySorting tests that map keys are sorted for determinism.
func TestToHCLWrite_MapKeySorting(t *testing.T) {
	type ModuleWithMap struct {
		Tags map[string]string `hcl:"tags,attr"`
	}

	module := &ModuleWithMap{
		Tags: map[string]string{
			"Zebra":   "z",
			"Alpha":   "a",
			"Charlie": "c",
			"Beta":    "b",
		},
	}

	result1, err1 := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err1)

	result2, err2 := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err2)

	// Results should be identical (deterministic)
	assert.Equal(t, result1, result2)

	// Keys should appear in alphabetical order
	alphaPos := strings.Index(result1, "Alpha")
	betaPos := strings.Index(result1, "Beta")
	charliePos := strings.Index(result1, "Charlie")
	zebraPos := strings.Index(result1, "Zebra")

	assert.Less(t, alphaPos, betaPos)
	assert.Less(t, betaPos, charliePos)
	assert.Less(t, charliePos, zebraPos)
}

// TestToHCLWrite_BlockMapKeySorting tests that block map keys are sorted.
func TestToHCLWrite_BlockMapKeySorting(t *testing.T) {
	type Statement struct {
		Effect *string `hcl:"effect,attr"`
	}

	type ModuleWithBlockMap struct {
		Statements map[string]Statement `hcl:"statement,block"`
	}

	effect := "Allow"
	module := &ModuleWithBlockMap{
		Statements: map[string]Statement{
			"zebra":   {Effect: &effect},
			"alpha":   {Effect: &effect},
			"charlie": {Effect: &effect},
		},
	}

	result1, err1 := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err1)

	result2, err2 := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err2)

	// Results should be identical (deterministic)
	assert.Equal(t, result1, result2)

	// Block labels should appear in alphabetical order
	alphaPos := strings.Index(result1, `statement "alpha"`)
	charliePos := strings.Index(result1, `statement "charlie"`)
	zebraPos := strings.Index(result1, `statement "zebra"`)

	assert.Less(t, alphaPos, charliePos)
	assert.Less(t, charliePos, zebraPos)
}

// TestToHCLWrite_PointerToStructBlock tests pointer blocks.
func TestToHCLWrite_PointerToStructBlock(t *testing.T) {
	type Config struct {
		Value *string `hcl:"value,attr"`
	}

	type ModuleWithPtrBlock struct {
		Name   *string `hcl:"name,attr"`
		Config *Config `hcl:"config,block"`
	}

	name := "test"
	value := "data"
	module := &ModuleWithPtrBlock{
		Name:   &name,
		Config: &Config{Value: &value},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)
	
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "config {")
	assert.Contains(t, result, "value")
	assert.Contains(t, result, `"data"`)
}

// TestToHCLWrite_SliceOfPointerBlocks tests slice of pointer structs as blocks.
func TestToHCLWrite_SliceOfPointerBlocks(t *testing.T) {
	type Rule struct {
		Name *string `hcl:"name,attr"`
	}

	type ModuleWithSliceOfPtrs struct {
		FunctionName *string `hcl:"function_name,attr"`
		Rules        []*Rule `hcl:"rule,block"`
	}

	functionName := "test"
	rule1Name := "rule1"
	rule2Name := "rule2"
	module := &ModuleWithSliceOfPtrs{
		FunctionName: &functionName,
		Rules: []*Rule{
			{Name: &rule1Name},
			{Name: &rule2Name},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.Contains(t, result, "rule {")
	assert.Contains(t, result, `"rule1"`)
	assert.Contains(t, result, `"rule2"`)
}

// TestToHCLWrite_MapOfPointerBlocks tests map of pointer structs as blocks.
func TestToHCLWrite_MapOfPointerBlocks(t *testing.T) {
	type Statement struct {
		Effect *string `hcl:"effect,attr"`
	}

	type ModuleWithMapOfPtrs struct {
		FunctionName *string              `hcl:"function_name,attr"`
		Statements   map[string]*Statement `hcl:"statement,block"`
	}

	functionName := "test"
	effect := "Allow"
	module := &ModuleWithMapOfPtrs{
		FunctionName: &functionName,
		Statements: map[string]*Statement{
			"s3": {Effect: &effect},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "function_name")
	assert.Contains(t, result, `statement "s3"`)
	assert.Contains(t, result, "effect")
	assert.Contains(t, result, `"Allow"`)
}

// TestToHCLWrite_StructAttribute tests struct as attribute (not block).
func TestToHCLWrite_StructAttribute(t *testing.T) {
	type Metadata struct {
		Version *int    `hcl:"version,attr"`
		Author  *string `hcl:"author,attr"`
	}

	type ModuleWithStructAttr struct {
		Name     *string   `hcl:"name,attr"`
		Metadata *Metadata `hcl:"metadata,attr"`
	}

	name := "test"
	version := 1
	author := "team"
	module := &ModuleWithStructAttr{
		Name: &name,
		Metadata: &Metadata{
			Version: &version,
			Author:  &author,
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "name")
	assert.Contains(t, result, "metadata")
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "author")
	assert.Contains(t, result, `"team"`)
}

// TestToHCLWrite_NilStructPointerAttribute tests nil struct pointer as attribute.
func TestToHCLWrite_NilStructPointerAttribute(t *testing.T) {
	type Config struct {
		Value *string `hcl:"value,attr"`
	}

	type ModuleWithNilStruct struct {
		Name   *string `hcl:"name,attr"`
		Config *Config `hcl:"config,attr"`
	}

	name := "test"
	module := &ModuleWithNilStruct{
		Name:   &name,
		Config: nil, // nil struct pointer
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "name")
	assert.NotContains(t, result, "config")
}

// TestToHCLWrite_IntSlice tests slice of integers.
func TestToHCLWrite_IntSlice(t *testing.T) {
	type ModuleWithIntSlice struct {
		Ports []int `hcl:"ports,attr"`
	}

	module := &ModuleWithIntSlice{
		Ports: []int{8080, 8443, 9000},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "ports")
	assert.Contains(t, result, "8080")
	assert.Contains(t, result, "8443")
	assert.Contains(t, result, "9000")
}

// TestToHCLWrite_MixedMapValues tests map with mixed value types.
func TestToHCLWrite_MixedMapValues(t *testing.T) {
	type ModuleWithMixedMap struct {
		Config map[string]interface{} `hcl:"config,attr"`
	}

	module := &ModuleWithMixedMap{
		Config: map[string]interface{}{
			"enabled": true,
			"count":   5,
			"rate":    1.5,
			"name":    "test",
			"ref":     "var.namespace",
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "config")
	assert.Contains(t, result, "enabled")
	assert.Contains(t, result, "true")
	assert.Contains(t, result, "count")
	assert.Contains(t, result, "5")
	assert.Contains(t, result, "rate")
	assert.Contains(t, result, "1.5")
	assert.Contains(t, result, "name")
	assert.Contains(t, result, `"test"`)
	assert.Contains(t, result, "var.namespace")
}

// TestToHCLWrite_InterfaceValue tests interface{} field with various types.
func TestToHCLWrite_InterfaceValue(t *testing.T) {
	type ModuleWithInterface struct {
		Value interface{} `hcl:"value,attr"`
	}

	// Test with string
	module1 := &ModuleWithInterface{Value: "test"}
	result1, err1 := hclgen.ToHCLWrite("test", "source", "", module1)
	require.NoError(t, err1)
	assert.Contains(t, result1, `"test"`)

	// Test with int
	module2 := &ModuleWithInterface{Value: 42}
	result2, err2 := hclgen.ToHCLWrite("test", "source", "", module2)
	require.NoError(t, err2)
	assert.Contains(t, result2, "42")

	// Test with bool
	module3 := &ModuleWithInterface{Value: true}
	result3, err3 := hclgen.ToHCLWrite("test", "source", "", module3)
	require.NoError(t, err3)
	assert.Contains(t, result3, "true")

	// Test with nil
	module4 := &ModuleWithInterface{Value: nil}
	result4, err4 := hclgen.ToHCLWrite("test", "source", "", module4)
	require.NoError(t, err4)
	assert.NotContains(t, result4, "value")
}

// TestToHCLWrite_EmptyStruct tests empty struct.
func TestToHCLWrite_EmptyStruct(t *testing.T) {
	type EmptyStruct struct{}

	type ModuleWithEmpty struct {
		Name   *string      `hcl:"name,attr"`
		Config *EmptyStruct `hcl:"config,attr"`
	}

	name := "test"
	module := &ModuleWithEmpty{
		Name:   &name,
		Config: &EmptyStruct{},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "name")
	// Empty struct should produce empty object or be omitted
}

// TestToHCLWrite_NestedEmptySlice tests nested structures with empty slices.
func TestToHCLWrite_NestedEmptySlice(t *testing.T) {
	type Inner struct {
		Items []string `hcl:"items,attr"`
	}

	type ModuleWithNested struct {
		Name  *string `hcl:"name,attr"`
		Inner *Inner  `hcl:"inner,attr"`
	}

	name := "test"
	module := &ModuleWithNested{
		Name: &name,
		Inner: &Inner{
			Items: []string{}, // Empty slice
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "name")
	// Inner with empty slice might be omitted or show empty list
}

// TestToHCLWrite_SliceOfStructsAttribute tests slice of structs as attribute.
func TestToHCLWrite_SliceOfStructsAttribute(t *testing.T) {
	type Item struct {
		Name  string `hcl:"name,attr"`
		Value int    `hcl:"value,attr"`
	}

	type ModuleWithSliceAttr struct {
		Name  *string `hcl:"name,attr"`
		Items []Item  `hcl:"items,attr"`
	}

	name := "test"
	module := &ModuleWithSliceAttr{
		Name: &name,
		Items: []Item{
			{Name: "item1", Value: 1},
			{Name: "item2", Value: 2},
		},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "name")
	assert.Contains(t, result, "items")
	assert.Contains(t, result, "item1")
	assert.Contains(t, result, "item2")
}

// TestToHCLWrite_RefInSliceNonString tests non-string slices (should not scan for refs).
func TestToHCLWrite_RefInSliceNonString(t *testing.T) {
	type ModuleWithNonStringSlice struct {
		Counts []int `hcl:"counts,attr"`
	}

	module := &ModuleWithNonStringSlice{
		Counts: []int{1, 2, 3},
	}

	result, err := hclgen.ToHCLWrite("test", "source", "", module)
	require.NoError(t, err)

	assert.Contains(t, result, "counts")
	assert.Contains(t, result, "1")
	assert.Contains(t, result, "2")
	assert.Contains(t, result, "3")
}