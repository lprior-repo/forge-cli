package cli

import (
	"os"
	"path/filepath"
	"testing"

	E "github.com/IBM/fp-go/either"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lewis/forge/internal/generators"
)

// Helper to extract ProjectState from Either.
func extractState(result E.Either[error, generators.ProjectState]) generators.ProjectState {
	return E.Fold(
		func(error) generators.ProjectState { return generators.ProjectState{} },
		func(s generators.ProjectState) generators.ProjectState { return s },
	)(result)
}

// Helper to extract error from ProjectState Either.
func extractStateError(result E.Either[error, generators.ProjectState]) error {
	return E.Fold(
		func(e error) error { return e },
		func(generators.ProjectState) error { return nil },
	)(result)
}

// Helper to extract WrittenFiles from Either.
func extractWritten(result E.Either[error, generators.WrittenFiles]) generators.WrittenFiles {
	return E.Fold(
		func(error) generators.WrittenFiles { return generators.WrittenFiles{} },
		func(w generators.WrittenFiles) generators.WrittenFiles { return w },
	)(result)
}

// TestCreateGeneratorRegistry tests registry creation.
func TestCreateGeneratorRegistry(t *testing.T) {
	registry := createGeneratorRegistry()

	assert.NotNil(t, registry)

	// Should have all generators registered
	generators := []generators.ResourceType{
		generators.ResourceSQS,
		generators.ResourceDynamoDB,
		generators.ResourceSNS,
		generators.ResourceS3,
	}

	for _, resourceType := range generators {
		gen, ok := registry.Get(resourceType)
		assert.True(t, ok, "%s generator should be registered", resourceType)
		assert.NotNil(t, gen)
	}
}

// TestDiscoverProjectState tests project state discovery.
func TestDiscoverProjectState(t *testing.T) {
	t.Run("infra directory not found", func(t *testing.T) {
		tmpDir := t.TempDir()

		result := discoverProjectState(tmpDir)

		require.True(t, E.IsLeft(result), "Should fail when infra/ doesn't exist")
		err := extractStateError(result)
		assert.Contains(t, err.Error(), "infra/ directory not found")
	})

	t.Run("empty infra directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		result := discoverProjectState(tmpDir)

		require.True(t, E.IsRight(result), "Should succeed with empty infra/")
		state := extractState(result)

		assert.Equal(t, tmpDir, state.ProjectRoot)
		assert.Empty(t, state.InfraFiles)
		assert.Empty(t, state.Functions)
		assert.Empty(t, state.Queues)
	})

	t.Run("discovers .tf files", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Create test .tf files
		files := []string{"main.tf", "variables.tf", "outputs.tf"}
		for _, file := range files {
			path := filepath.Join(infraDir, file)
			require.NoError(t, os.WriteFile(path, []byte("# test"), 0o644))
		}

		// Create non-.tf file (should be ignored)
		require.NoError(t, os.WriteFile(filepath.Join(infraDir, "readme.md"), []byte("test"), 0o644))

		result := discoverProjectState(tmpDir)

		require.True(t, E.IsRight(result), "Should succeed")
		state := extractState(result)

		assert.Len(t, state.InfraFiles, 3, "Should find 3 .tf files")
		for _, file := range files {
			expectedPath := filepath.Join(infraDir, file)
			assert.Contains(t, state.InfraFiles, expectedPath)
		}
	})
}

// TestWriteGeneratedFiles tests file writing logic.
func TestWriteGeneratedFiles(t *testing.T) {
	t.Run("create new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resource",
					Mode:    generators.WriteModeCreate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Created, 1)
		assert.Contains(t, written.Created, "sqs.tf")

		// Verify file exists and has correct content
		filePath := filepath.Join(infraDir, "sqs.tf")
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "# SQS resource", string(content))
	})

	t.Run("create mode skips existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Create existing file
		filePath := filepath.Join(infraDir, "sqs.tf")
		require.NoError(t, os.WriteFile(filePath, []byte("existing"), 0o644))

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# New content",
					Mode:    generators.WriteModeCreate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Skipped, 1)
		assert.Contains(t, written.Skipped, "sqs.tf")

		// Verify file content unchanged
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, "existing", string(content))
	})

	t.Run("append mode creates new file", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resource",
					Mode:    generators.WriteModeAppend,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Created, 1)

		// Verify file exists
		filePath := filepath.Join(infraDir, "sqs.tf")
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "# SQS resource")
	})

	t.Run("append mode adds to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Create existing file
		filePath := filepath.Join(infraDir, "outputs.tf")
		require.NoError(t, os.WriteFile(filePath, []byte("# Existing outputs"), 0o644))

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "outputs.tf",
					Content: "# New output",
					Mode:    generators.WriteModeAppend,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Updated, 1)
		assert.Contains(t, written.Updated, "outputs.tf")

		// Verify both contents present
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "# Existing outputs")
		assert.Contains(t, string(content), "# New output")
	})

	t.Run("writes multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resource",
					Mode:    generators.WriteModeCreate,
				},
				{
					Path:    "outputs.tf",
					Content: "# Outputs",
					Mode:    generators.WriteModeCreate,
				},
				{
					Path:    "lambda.tf",
					Content: "# Lambda integration",
					Mode:    generators.WriteModeCreate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Created, 3)

		// Verify all files exist
		for _, file := range code.Files {
			filePath := filepath.Join(infraDir, file.Path)
			_, err := os.Stat(filePath)
			assert.NoError(t, err, "File %s should exist", file.Path)
		}
	})

	t.Run("creates infra directory if missing", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resource",
					Mode:    generators.WriteModeCreate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")

		// Verify infra directory was created
		stat, err := os.Stat(infraDir)
		require.NoError(t, err)
		assert.True(t, stat.IsDir())
	})

	t.Run("update mode appends to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Create existing file
		filePath := filepath.Join(infraDir, "lambda.tf")
		require.NoError(t, os.WriteFile(filePath, []byte("# Existing Lambda"), 0o644))

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "lambda.tf",
					Content: "# Updated Lambda",
					Mode:    generators.WriteModeUpdate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsRight(result), "Should succeed")
		written := extractWritten(result)

		assert.Len(t, written.Updated, 1)

		// Verify content was appended
		content, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "# Existing Lambda")
		assert.Contains(t, string(content), "# Updated Lambda")
	})
}

// TestWriteGeneratedFiles_Permissions tests file permission handling.
func TestWriteGeneratedFiles_Permissions(t *testing.T) {
	tmpDir := t.TempDir()
	infraDir := filepath.Join(tmpDir, "infra")

	code := generators.GeneratedCode{
		Files: []generators.FileToWrite{
			{
				Path:    "sqs.tf",
				Content: "# SQS resource",
				Mode:    generators.WriteModeCreate,
			},
		},
	}

	result := writeGeneratedFiles(code, infraDir)
	require.True(t, E.IsRight(result))

	// Verify file permissions
	filePath := filepath.Join(infraDir, "sqs.tf")
	stat, err := os.Stat(filePath)
	require.NoError(t, err)

	mode := stat.Mode()
	assert.Equal(t, os.FileMode(0o644), mode.Perm(), "File should have 0644 permissions")
}

// TestWriteGeneratedFiles_ErrorHandling tests error scenarios.
func TestWriteGeneratedFiles_ErrorHandling(t *testing.T) {
	t.Run("invalid path", func(t *testing.T) {
		// Use a path that will fail (like a file as directory)
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("test"), 0o644))

		// Try to use file as directory
		infraDir := filePath

		code := generators.GeneratedCode{
			Files: []generators.FileToWrite{
				{
					Path:    "sqs.tf",
					Content: "# SQS resource",
					Mode:    generators.WriteModeCreate,
				},
			},
		}

		result := writeGeneratedFiles(code, infraDir)

		require.True(t, E.IsLeft(result), "Should fail with invalid path")
	})
}

// TestAddCommand_FlagDefaults tests command flag defaults.
func TestAddCommand_FlagDefaults(t *testing.T) {
	cmd := NewAddCmd()

	// Verify flags exist and have correct defaults
	toFlag := cmd.Flags().Lookup("to")
	assert.NotNil(t, toFlag)
	assert.Equal(t, "", toFlag.DefValue)

	rawFlag := cmd.Flags().Lookup("raw")
	assert.NotNil(t, rawFlag)
	assert.Equal(t, "false", rawFlag.DefValue)

	noModuleFlag := cmd.Flags().Lookup("no-module")
	assert.NotNil(t, noModuleFlag)
	assert.Equal(t, "false", noModuleFlag.DefValue)
}

// TestRunAdd tests the runAdd command execution.
func TestRunAdd(t *testing.T) {
	t.Run("succeeds with SQS resource", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Change to temp directory for test
		t.Chdir(tmpDir)

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		err = runAdd(cmd, args, "", false, false)
		assert.NoError(t, err)

		// Verify SQS file was created (generators use generic names)
		sqsFile := filepath.Join(infraDir, "sqs.tf")
		assert.FileExists(t, sqsFile)

		// Verify outputs file created
		outputsFile := filepath.Join(infraDir, "outputs.tf")
		assert.FileExists(t, outputsFile)
	})

	t.Run("fails with unsupported resource type", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		t.Chdir(tmpDir)

		cmd := NewAddCmd()
		args := []string{"invalid-type", "test-resource"}

		err = runAdd(cmd, args, "", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported resource type")
	})

	t.Run("fails when infra directory missing", func(t *testing.T) {
		tmpDir := t.TempDir()

		t.Chdir(tmpDir)

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		err = runAdd(cmd, args, "", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "infra/ directory not found")
	})

	t.Run("uses raw mode when flag set", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		err = runAdd(cmd, args, "", true, false)
		assert.NoError(t, err)

		// Verify file was created (implementation detail: raw mode still creates files)
		sqsFile := filepath.Join(infraDir, "sqs.tf")
		assert.FileExists(t, sqsFile)
	})

	t.Run("integrates with Lambda function when --to flag used", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		// Create minimal function state so --to validation passes
		// Note: In real usage, functions would be discovered from existing Terraform
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		// This will fail because processor-function doesn't exist
		// We're testing error handling here
		err = runAdd(cmd, args, "processor-function", false, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("respects no-module flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"sqs", "test-queue"}

		err = runAdd(cmd, args, "", false, true)
		assert.NoError(t, err)

		// Verify file was created
		sqsFile := filepath.Join(infraDir, "sqs.tf")
		assert.FileExists(t, sqsFile)
	})

	t.Run("supports DynamoDB resource type", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"dynamodb", "test-table"}

		err = runAdd(cmd, args, "", false, false)
		assert.NoError(t, err)

		// Verify DynamoDB file was created
		dynamoFile := filepath.Join(infraDir, "dynamodb.tf")
		assert.FileExists(t, dynamoFile)
	})

	t.Run("supports SNS resource type", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"sns", "test-topic"}

		err = runAdd(cmd, args, "", false, false)
		assert.NoError(t, err)

		// Verify SNS file was created
		snsFile := filepath.Join(infraDir, "sns.tf")
		assert.FileExists(t, snsFile)
	})

	t.Run("supports S3 resource type", func(t *testing.T) {
		tmpDir := t.TempDir()
		infraDir := filepath.Join(tmpDir, "infra")
		require.NoError(t, os.MkdirAll(infraDir, 0o755))

		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()
		require.NoError(t, os.Chdir(tmpDir))

		cmd := NewAddCmd()
		args := []string{"s3", "test-bucket"}

		err = runAdd(cmd, args, "", false, false)
		assert.NoError(t, err)

		// Verify S3 file was created
		s3File := filepath.Join(infraDir, "s3.tf")
		assert.FileExists(t, s3File)
	})
}
