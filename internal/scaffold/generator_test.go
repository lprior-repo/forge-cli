package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewGenerator tests generator creation
func TestNewGenerator(t *testing.T) {
	t.Run("creates generator successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(tmpDir)

		require.NoError(t, err)
		assert.NotNil(t, gen)
		assert.Equal(t, tmpDir, gen.projectRoot)
	})
}

// TestGenerateProject tests project generation
func TestGenerateProject(t *testing.T) {
	t.Run("generates complete project structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(tmpDir)
		require.NoError(t, err)

		opts := &ProjectOptions{
			Name:   "test-project",
			Region: "us-east-1",
		}

		err = gen.GenerateProject(opts)
		require.NoError(t, err)

		// Verify files were created
		assert.FileExists(t, filepath.Join(tmpDir, "forge.hcl"))
		assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
		assert.FileExists(t, filepath.Join(tmpDir, "README.md"))
	})

	t.Run("forge.hcl contains project name", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(tmpDir)
		require.NoError(t, err)

		opts := &ProjectOptions{
			Name:   "my-project",
			Region: "us-west-2",
		}

		err = gen.GenerateProject(opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "forge.hcl"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "my-project")
		assert.Contains(t, string(content), "us-west-2")
	})
}

// TestGenerateGoStack tests Go stack generation
func TestGenerateGoStack(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	opts := &StackOptions{
		Name:        "api",
		Runtime:     "provided.al2023",
		Description: "API Lambda",
	}

	err = gen.GenerateStack(opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "api")

	// Verify Go files were created
	assert.FileExists(t, filepath.Join(stackDir, "main.go"))
	assert.FileExists(t, filepath.Join(stackDir, "go.mod"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
}

// TestGeneratePythonStack tests Python stack generation
func TestGeneratePythonStack(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	opts := &StackOptions{
		Name:        "worker",
		Runtime:     "python3.13",
		Description: "Worker Lambda",
	}

	err = gen.GenerateStack(opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "worker")

	// Verify Python files were created
	assert.FileExists(t, filepath.Join(stackDir, "handler.py"))
	assert.FileExists(t, filepath.Join(stackDir, "requirements.txt"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
}

// TestGenerateNodeStack tests Node.js stack generation
func TestGenerateNodeStack(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	opts := &StackOptions{
		Name:        "frontend",
		Runtime:     "nodejs22.x",
		Description: "Frontend Lambda",
	}

	err = gen.GenerateStack(opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "frontend")

	// Verify Node files were created
	assert.FileExists(t, filepath.Join(stackDir, "index.js"))
	assert.FileExists(t, filepath.Join(stackDir, "package.json"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))

	// Verify package.json contains correct name
	content, err := os.ReadFile(filepath.Join(stackDir, "package.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "frontend")
}

// TestGenerateJavaStack tests Java stack generation
func TestGenerateJavaStack(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	opts := &StackOptions{
		Name:        "service",
		Runtime:     "java21",
		Description: "Service Lambda",
	}

	err = gen.GenerateStack(opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "service")

	// Verify Java files and directory structure were created
	assert.DirExists(t, filepath.Join(stackDir, "src", "main", "java", "com", "example"))
	assert.FileExists(t, filepath.Join(stackDir, "src", "main", "java", "com", "example", "Handler.java"))
	assert.FileExists(t, filepath.Join(stackDir, "pom.xml"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))

	// Verify pom.xml contains correct artifactId
	content, err := os.ReadFile(filepath.Join(stackDir, "pom.xml"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "service")
}

// TestGenerateStackUnsupportedRuntime tests error handling for unsupported runtimes
func TestGenerateStackUnsupportedRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	opts := &StackOptions{
		Name:        "test",
		Runtime:     "unsupported-runtime",
		Description: "Test",
	}

	err = gen.GenerateStack(opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported runtime")
}

// TestTemplateData tests template data population
func TestTemplateData(t *testing.T) {
	t.Run("stack template includes description", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(tmpDir)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:        "test-stack",
			Runtime:     "provided.al2023",
			Description: "My Test Stack Description",
		}

		err = gen.GenerateStack(opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "test-stack", "stack.forge.hcl"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "My Test Stack Description")
	})

	t.Run("go.mod contains stack name as module", func(t *testing.T) {
		tmpDir := t.TempDir()
		gen, err := NewGenerator(tmpDir)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "my-service",
			Runtime: "go1.x",
		}

		err = gen.GenerateStack(opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "my-service", "go.mod"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "my-service")
	})
}

// TestMultipleStacks tests generating multiple stacks
func TestMultipleStacks(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	stacks := []StackOptions{
		{Name: "api", Runtime: "go1.x", Description: "API"},
		{Name: "worker", Runtime: "python3.13", Description: "Worker"},
		{Name: "frontend", Runtime: "nodejs22.x", Description: "Frontend"},
	}

	for _, opts := range stacks {
		opt := opts // capture loop variable
		err := gen.GenerateStack(&opt)
		require.NoError(t, err)
	}

	// Verify all stack directories exist
	assert.DirExists(t, filepath.Join(tmpDir, "api"))
	assert.DirExists(t, filepath.Join(tmpDir, "worker"))
	assert.DirExists(t, filepath.Join(tmpDir, "frontend"))

	// Verify each has stack.forge.hcl
	assert.FileExists(t, filepath.Join(tmpDir, "api", "stack.forge.hcl"))
	assert.FileExists(t, filepath.Join(tmpDir, "worker", "stack.forge.hcl"))
	assert.FileExists(t, filepath.Join(tmpDir, "frontend", "stack.forge.hcl"))
}

// TestRuntimeVariants tests different runtime version variants
func TestRuntimeVariants(t *testing.T) {
	tests := []struct {
		runtime      string
		expectedFile string
		shouldError  bool
	}{
		{"go1.x", "main.go", false},
		{"provided.al2", "main.go", false},
		{"provided.al2023", "main.go", false},
		{"python3.9", "handler.py", false},
		{"python3.13", "handler.py", false},
		{"nodejs18.x", "index.js", false},
		{"nodejs22.x", "index.js", false},
		{"java11", "pom.xml", false},
		{"java21", "pom.xml", false},
		{"ruby2.7", "", true}, // Unsupported
	}

	for _, tt := range tests {
		t.Run(tt.runtime, func(t *testing.T) {
			tmpDir := t.TempDir()
			gen, err := NewGenerator(tmpDir)
			require.NoError(t, err)

			opts := &StackOptions{
				Name:    "test",
				Runtime: tt.runtime,
			}

			err = gen.GenerateStack(opts)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.FileExists(t, filepath.Join(tmpDir, "test", tt.expectedFile))
			}
		})
	}
}

// TestTemplateRendering tests that templates render without errors
func TestTemplateRendering(t *testing.T) {
	tmpDir := t.TempDir()
	gen, err := NewGenerator(tmpDir)
	require.NoError(t, err)

	// Test each template type
	runtimes := []string{
		"provided.al2023",
		"python3.13",
		"nodejs22.x",
		"java21",
	}

	for _, runtime := range runtimes {
		t.Run(runtime, func(t *testing.T) {
			opts := &StackOptions{
				Name:        "test-" + runtime,
				Runtime:     runtime,
				Description: "Test stack for " + runtime,
			}

			err := gen.GenerateStack(opts)
			require.NoError(t, err)

			// Verify main.tf was generated
			assert.FileExists(t, filepath.Join(tmpDir, "test-"+runtime, "main.tf"))
		})
	}
}
