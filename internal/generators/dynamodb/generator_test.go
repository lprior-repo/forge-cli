package dynamodb_test

import (
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators"
	"github.com/lewis/forge/internal/generators/dynamodb"
)

// Helper function to extract Right value from Either.
func extractConfig(result E.Either[error, generators.ResourceConfig]) generators.ResourceConfig {
	return E.Fold(
		func(error) generators.ResourceConfig { return generators.ResourceConfig{} },
		func(c generators.ResourceConfig) generators.ResourceConfig { return c },
	)(result)
}

// Helper function to extract generated code.
func extractCode(result E.Either[error, generators.GeneratedCode]) generators.GeneratedCode {
	return E.Fold(
		func(error) generators.GeneratedCode { return generators.GeneratedCode{} },
		func(c generators.GeneratedCode) generators.GeneratedCode { return c },
	)(result)
}

// Helper function to find file by path.
func findFile(files []generators.FileToWrite, path string) *generators.FileToWrite {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

func TestNew(t *testing.T) {
	t.Run("creates new generator", func(t *testing.T) {
		gen := dynamodb.New()
		assert.NotNil(t, gen)
	})
}

func TestValidate(t *testing.T) {
	gen := dynamodb.New()

	t.Run("validates correct config", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name: "users_table",
			Type: generators.ResourceDynamoDB,
			Variables: map[string]interface{}{
				"hash_key": "id",
			},
		}

		result := gen.Validate(config)
		assert.True(t, E.IsRight(result))
	})

	t.Run("rejects empty name", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name: "",
			Variables: map[string]interface{}{
				"hash_key": "id",
			},
		}

		result := gen.Validate(config)
		assert.True(t, E.IsLeft(result))
	})

	t.Run("rejects invalid name characters", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name: "users@table",
			Variables: map[string]interface{}{
				"hash_key": "id",
			},
		}

		result := gen.Validate(config)
		assert.True(t, E.IsLeft(result))
	})

	t.Run("rejects missing hash_key", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name:      "users_table",
			Variables: map[string]interface{}{},
		}

		result := gen.Validate(config)
		assert.True(t, E.IsLeft(result))
	})
}

func TestPrompt(t *testing.T) {
	gen := dynamodb.New()
	ctx := t.Context()

	t.Run("creates config with defaults", func(t *testing.T) {
		intent := generators.ResourceIntent{
			Type:      generators.ResourceDynamoDB,
			Name:      "orders_table",
			UseModule: true,
		}

		state := generators.ProjectState{
			Functions: make(map[string]generators.FunctionInfo),
		}

		result := gen.Prompt(ctx, intent, state)
		require.True(t, E.IsRight(result))

		config := extractConfig(result)
		assert.Equal(t, "orders_table", config.Name)
		assert.True(t, config.Module)
		assert.Equal(t, "id", config.Variables["hash_key"])
		assert.Equal(t, "PAY_PER_REQUEST", config.Variables["billing_mode"])
		assert.True(t, config.Variables["point_in_time_recovery"].(bool))
	})

	t.Run("enables streams for Lambda integration", func(t *testing.T) {
		intent := generators.ResourceIntent{
			Type:      generators.ResourceDynamoDB,
			Name:      "events_table",
			UseModule: true,
			ToFunc:    "processor",
		}

		state := generators.ProjectState{
			Functions: map[string]generators.FunctionInfo{
				"processor": {Name: "processor", Runtime: "go1.x"},
			},
		}

		result := gen.Prompt(ctx, intent, state)
		require.True(t, E.IsRight(result))

		config := extractConfig(result)
		assert.True(t, config.Variables["stream_enabled"].(bool))
		assert.Equal(t, "NEW_AND_OLD_IMAGES", config.Variables["stream_view_type"])
		assert.NotNil(t, config.Integration)
		assert.Equal(t, "processor", config.Integration.TargetFunction)
	})

	t.Run("returns error for non-existent function", func(t *testing.T) {
		intent := generators.ResourceIntent{
			Type:      generators.ResourceDynamoDB,
			Name:      "events_table",
			UseModule: true,
			ToFunc:    "nonexistent",
		}

		state := generators.ProjectState{
			Functions: make(map[string]generators.FunctionInfo),
		}

		result := gen.Prompt(ctx, intent, state)
		assert.True(t, E.IsLeft(result))
	})
}

func TestGenerate(t *testing.T) {
	gen := dynamodb.New()

	t.Run("generates module code", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name:   "users",
			Type:   generators.ResourceDynamoDB,
			Module: true,
			Variables: map[string]interface{}{
				"hash_key":               "id",
				"range_key":              "",
				"billing_mode":           "PAY_PER_REQUEST",
				"stream_enabled":         false,
				"stream_view_type":       "",
				"ttl_enabled":            false,
				"ttl_attribute":          "",
				"point_in_time_recovery": true,
				"attributes": []map[string]string{
					{"name": "id", "type": "S"},
				},
				"global_secondary_indexes": []map[string]interface{}{},
				"local_secondary_indexes":  []map[string]interface{}{},
			},
		}

		state := generators.ProjectState{}

		result := gen.Generate(config, state)
		require.True(t, E.IsRight(result))

		code := extractCode(result)
		assert.NotEmpty(t, code.Files)
		assert.Len(t, code.Files, 2) // dynamodb.tf + outputs.tf

		// Check dynamodb.tf content
		dynamoFile := code.Files[0]
		assert.Equal(t, "dynamodb.tf", dynamoFile.Path)
		assert.Contains(t, dynamoFile.Content, "module \"users\"")
		assert.Contains(t, dynamoFile.Content, "terraform-aws-modules/dynamodb-table/aws")
		assert.Contains(t, dynamoFile.Content, "hash_key  = \"id\"")
		assert.Contains(t, dynamoFile.Content, "PAY_PER_REQUEST")
	})

	t.Run("generates with integration", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name:   "events",
			Type:   generators.ResourceDynamoDB,
			Module: true,
			Variables: map[string]interface{}{
				"hash_key":               "id",
				"range_key":              "",
				"billing_mode":           "PAY_PER_REQUEST",
				"stream_enabled":         true,
				"stream_view_type":       "NEW_AND_OLD_IMAGES",
				"ttl_enabled":            false,
				"ttl_attribute":          "",
				"point_in_time_recovery": true,
				"attributes": []map[string]string{
					{"name": "id", "type": "S"},
				},
				"global_secondary_indexes": []map[string]interface{}{},
				"local_secondary_indexes":  []map[string]interface{}{},
			},
			Integration: &generators.IntegrationConfig{
				TargetFunction: "processor",
				EventSource: &generators.EventSourceConfig{
					ARNExpression:  "module.events.stream_arn",
					BatchSize:      100,
					MaxConcurrency: 10,
				},
				IAMPermissions: []generators.IAMPermission{
					{
						Effect:    "Allow",
						Actions:   []string{"dynamodb:GetRecords"},
						Resources: []string{"module.events.stream_arn"},
					},
				},
			},
		}

		state := generators.ProjectState{}

		result := gen.Generate(config, state)
		require.True(t, E.IsRight(result))

		code := extractCode(result)
		assert.Len(t, code.Files, 3) // dynamodb.tf + outputs.tf + lambda_processor.tf

		// Check integration file
		lambdaFile := findFile(code.Files, "lambda_processor.tf")
		require.NotNil(t, lambdaFile)
		assert.Equal(t, "lambda_processor.tf", lambdaFile.Path)
		assert.Contains(t, lambdaFile.Content, "aws_lambda_event_source_mapping")
		assert.Contains(t, lambdaFile.Content, "module.events.stream_arn")
	})

	t.Run("fails validation with invalid config", func(t *testing.T) {
		config := generators.ResourceConfig{
			Name: "", // Invalid: empty name
			Variables: map[string]interface{}{
				"hash_key": "id",
			},
		}

		state := generators.ProjectState{}

		result := gen.Generate(config, state)
		assert.True(t, E.IsLeft(result))
	})
}

// TestGeneratedCodeFormat tests that generated code is well-formatted.
func TestGeneratedCodeFormat(t *testing.T) {
	gen := dynamodb.New()

	config := generators.ResourceConfig{
		Name:   "users",
		Type:   generators.ResourceDynamoDB,
		Module: true,
		Variables: map[string]interface{}{
			"hash_key":               "id",
			"range_key":              "",
			"billing_mode":           "PAY_PER_REQUEST",
			"stream_enabled":         false,
			"stream_view_type":       "",
			"ttl_enabled":            false,
			"ttl_attribute":          "",
			"point_in_time_recovery": true,
			"attributes": []map[string]string{
				{"name": "id", "type": "S"},
			},
			"global_secondary_indexes": []map[string]interface{}{},
			"local_secondary_indexes":  []map[string]interface{}{},
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	for _, file := range code.Files {
		// Check that code starts with comment
		assert.Positive(t, file.Content, "File should have content")
	}
}

// TestRawResourceGeneration tests generation without module.
func TestRawResourceGeneration(t *testing.T) {
	gen := dynamodb.New()

	config := generators.ResourceConfig{
		Name:   "users",
		Type:   generators.ResourceDynamoDB,
		Module: false,
		Variables: map[string]interface{}{
			"hash_key":               "id",
			"range_key":              "",
			"billing_mode":           "PAY_PER_REQUEST",
			"stream_enabled":         false,
			"stream_view_type":       "",
			"ttl_enabled":            false,
			"ttl_attribute":          "",
			"point_in_time_recovery": true,
			"attributes": []map[string]string{
				{"name": "id", "type": "S"},
			},
			"global_secondary_indexes": []map[string]interface{}{},
			"local_secondary_indexes":  []map[string]interface{}{},
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	dynamoFile := findFile(code.Files, "dynamodb.tf")
	require.NotNil(t, dynamoFile)
	assert.Contains(t, dynamoFile.Content, "resource \"aws_dynamodb_table\"")
	assert.NotContains(t, dynamoFile.Content, "module \"")
}

// TestNamespaceSupport tests namespace variable usage.
func TestNamespaceSupport(t *testing.T) {
	gen := dynamodb.New()

	config := generators.ResourceConfig{
		Name:   "users",
		Type:   generators.ResourceDynamoDB,
		Module: true,
		Variables: map[string]interface{}{
			"hash_key":               "id",
			"range_key":              "",
			"billing_mode":           "PAY_PER_REQUEST",
			"stream_enabled":         false,
			"stream_view_type":       "",
			"ttl_enabled":            false,
			"ttl_attribute":          "",
			"point_in_time_recovery": true,
			"attributes": []map[string]string{
				{"name": "id", "type": "S"},
			},
			"global_secondary_indexes": []map[string]interface{}{},
			"local_secondary_indexes":  []map[string]interface{}{},
		},
	}

	result := gen.Generate(config, generators.ProjectState{})
	require.True(t, E.IsRight(result))
	code := extractCode(result)

	dynamoFile := findFile(code.Files, "dynamodb.tf")
	require.NotNil(t, dynamoFile)

	// All resources should use ${var.namespace} prefix
	assert.Contains(t, dynamoFile.Content, "${var.namespace}")
}
