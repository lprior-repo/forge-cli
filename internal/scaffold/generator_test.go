package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateProject tests that functions work without needing a generator struct.
func TestGenerateFunctional(t *testing.T) {
	t.Run("can call functions without generator", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Pure functional - no generator needed
		opts := &ProjectOptions{
			Name:   "test",
			Region: "us-east-1",
		}
		err := GenerateProject(tmpDir, opts)
		require.NoError(t, err)
	})
}

// TestGenerateProjectFiles tests project generation.
func TestGenerateProjectFiles(t *testing.T) {
	t.Run("generates complete project structure", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &ProjectOptions{
			Name:   "test-project",
			Region: "us-east-1",
		}

		err := GenerateProject(tmpDir, opts)
		require.NoError(t, err)

		// Verify files were created
		assert.FileExists(t, filepath.Join(tmpDir, "forge.hcl"))
		assert.FileExists(t, filepath.Join(tmpDir, ".gitignore"))
		assert.FileExists(t, filepath.Join(tmpDir, "README.md"))
	})

	t.Run("forge.hcl contains project name", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &ProjectOptions{
			Name:   "my-project",
			Region: "us-west-2",
		}

		err := GenerateProject(tmpDir, opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "forge.hcl"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "my-project")
		assert.Contains(t, string(content), "us-west-2")
	})
}

// TestGenerateGoStack tests Go stack generation.
func TestGenerateGoStack(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &StackOptions{
		Name:        "api",
		Runtime:     "provided.al2023",
		Description: "API Lambda",
	}

	err := GenerateStack(tmpDir, opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "api")

	// Verify Go files were created
	assert.FileExists(t, filepath.Join(stackDir, "main.go"))
	assert.FileExists(t, filepath.Join(stackDir, "go.mod"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
}

// TestGeneratePythonStack tests Python stack generation.
func TestGeneratePythonStack(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &StackOptions{
		Name:        "worker",
		Runtime:     "python3.13",
		Description: "Worker Lambda",
	}

	err := GenerateStack(tmpDir, opts)
	require.NoError(t, err)

	stackDir := filepath.Join(tmpDir, "worker")

	// Verify Python files were created
	assert.FileExists(t, filepath.Join(stackDir, "handler.py"))
	assert.FileExists(t, filepath.Join(stackDir, "requirements.txt"))
	assert.FileExists(t, filepath.Join(stackDir, "main.tf"))
	assert.FileExists(t, filepath.Join(stackDir, "stack.forge.hcl"))
}

// TestGenerateNodeStack tests Node.js stack generation.
func TestGenerateNodeStack(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &StackOptions{
		Name:        "frontend",
		Runtime:     "nodejs22.x",
		Description: "Frontend Lambda",
	}

	err := GenerateStack(tmpDir, opts)
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

// TestGenerateJavaStack tests Java stack generation.
func TestGenerateJavaStack(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &StackOptions{
		Name:        "service",
		Runtime:     "java21",
		Description: "Service Lambda",
	}

	err := GenerateStack(tmpDir, opts)
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

// TestGenerateStackUnsupportedRuntime tests error handling for unsupported runtimes.
func TestGenerateStackUnsupportedRuntime(t *testing.T) {
	tmpDir := t.TempDir()

	opts := &StackOptions{
		Name:        "test",
		Runtime:     "unsupported-runtime",
		Description: "Test",
	}

	err := GenerateStack(tmpDir, opts)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported runtime")
}

// TestTemplateData tests template data population.
func TestTemplateData(t *testing.T) {
	t.Run("stack template includes description", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &StackOptions{
			Name:        "test-stack",
			Runtime:     "provided.al2023",
			Description: "My Test Stack Description",
		}

		err := GenerateStack(tmpDir, opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "test-stack", "stack.forge.hcl"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "My Test Stack Description")
	})

	t.Run("go.mod contains stack name as module", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &StackOptions{
			Name:    "my-service",
			Runtime: "go1.x",
		}

		err := GenerateStack(tmpDir, opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "my-service", "go.mod"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "my-service")
	})
}

// TestMultipleStacks tests generating multiple stacks.
func TestMultipleStacks(t *testing.T) {
	tmpDir := t.TempDir()

	stacks := []StackOptions{
		{Name: "api", Runtime: "go1.x", Description: "API"},
		{Name: "worker", Runtime: "python3.13", Description: "Worker"},
		{Name: "frontend", Runtime: "nodejs22.x", Description: "Frontend"},
	}

	for _, opts := range stacks {
		opt := opts // capture loop variable
		err := GenerateStack(tmpDir, &opt)
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

// TestRuntimeVariants tests different runtime version variants.
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

			opts := &StackOptions{
				Name:    "test",
				Runtime: tt.runtime,
			}

			err := GenerateStack(tmpDir, opts)

			if tt.shouldError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.FileExists(t, filepath.Join(tmpDir, "test", tt.expectedFile))
			}
		})
	}
}

// TestTemplateRendering tests that templates render without errors.
func TestTemplateRendering(t *testing.T) {
	tmpDir := t.TempDir()

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

			err := GenerateStack(tmpDir, opts)
			require.NoError(t, err)

			// Verify main.tf was generated
			assert.FileExists(t, filepath.Join(tmpDir, "test-"+runtime, "main.tf"))
		})
	}
}

// TestGenerateProjectErrorPaths tests error handling in project generation.
func TestGenerateProjectErrorPaths(t *testing.T) {
	t.Run("fails on invalid project directory path", func(t *testing.T) {
		// Use a path that will fail to create (e.g., inside a file)
		tmpDir := t.TempDir()
		blockingFile := filepath.Join(tmpDir, "blocking")
		err := os.WriteFile(blockingFile, []byte("test"), 0o644)
		require.NoError(t, err)

		// Try to create a directory where a file exists
		opts := &ProjectOptions{
			Name:   "test",
			Region: "us-east-1",
		}

		err = GenerateProject(blockingFile, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create project directory")
	})

	t.Run("fails on read-only directory for forge.hcl", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create the directory but make it read-only after creation
		err := os.Mkdir(filepath.Join(tmpDir, "readonly"), 0o555)
		require.NoError(t, err)

		opts := &ProjectOptions{
			Name:   "test",
			Region: "us-east-1",
		}

		err = GenerateProject(filepath.Join(tmpDir, "readonly"), opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write")
	})

	t.Run("fails when gitignore cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		err := os.MkdirAll(projectDir, 0o755)
		require.NoError(t, err)

		opts := &ProjectOptions{
			Name:   "test",
			Region: "us-east-1",
		}

		// Write forge.hcl first
		forgeHCL := generateForgeHCL(opts)
		err = os.WriteFile(filepath.Join(projectDir, "forge.hcl"), []byte(forgeHCL), 0o644)
		require.NoError(t, err)

		// Make directory read-only after forge.hcl is written
		err = os.Chmod(projectDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(projectDir, 0o755) }() // Clean up

		err = GenerateProject(projectDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write .gitignore")
	})

	t.Run("fails when README cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		projectDir := filepath.Join(tmpDir, "project")
		err := os.MkdirAll(projectDir, 0o755)
		require.NoError(t, err)

		opts := &ProjectOptions{
			Name:   "test",
			Region: "us-east-1",
		}

		// Write forge.hcl and .gitignore first
		forgeHCL := generateForgeHCL(opts)
		err = os.WriteFile(filepath.Join(projectDir, "forge.hcl"), []byte(forgeHCL), 0o644)
		require.NoError(t, err)

		gitignore := generateGitignore()
		err = os.WriteFile(filepath.Join(projectDir, ".gitignore"), []byte(gitignore), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(projectDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(projectDir, 0o755) }() // Clean up

		err = GenerateProject(projectDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write README.md")
	})
}

// TestGenerateStackErrorPaths tests error handling in stack generation.
func TestGenerateStackErrorPaths(t *testing.T) {
	t.Run("fails when stack directory cannot be created", func(t *testing.T) {
		tmpDir := t.TempDir()
		blockingFile := filepath.Join(tmpDir, "blocking")
		err := os.WriteFile(blockingFile, []byte("test"), 0o644)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "blocking",
			Runtime: "go1.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create stack directory")
	})

	t.Run("fails when stack.forge.hcl cannot be written - Go", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o555) // Read-only
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write")
	})

	t.Run("fails when main.go cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		// Make directory read-only to prevent main.go write
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up permissions
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when handler.py cannot be written - Python", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when requirements.txt cannot be written - Python", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl and handler.py first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		handler := generatePythonHandler(&StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		})
		err = os.WriteFile(filepath.Join(stackDir, "handler.py"), []byte(handler), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when index.js cannot be written - Node", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when package.json cannot be written - Node", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl and index.js first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		index := generateNodeIndex(&StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "index.js"), []byte(index), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when Java directory structure cannot be created", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		// Make directory read-only to prevent creating src/
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when Handler.java cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		javaDir := filepath.Join(stackDir, "src", "main", "java", "com", "example")
		err := os.MkdirAll(javaDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		// Make Java directory read-only
		err = os.Chmod(javaDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(javaDir, 0o755)
	})

	t.Run("fails when pom.xml cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		javaDir := filepath.Join(stackDir, "src", "main", "java", "com", "example")
		err := os.MkdirAll(javaDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl and Handler.java first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		handler := generateJavaHandler(&StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		})
		err = os.WriteFile(filepath.Join(javaDir, "Handler.java"), []byte(handler), 0o644)
		require.NoError(t, err)

		// Make stack directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)

		// Clean up
		_ = os.Chmod(stackDir, 0o755)
	})

	t.Run("fails when go.mod cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl and main.go first
		stackHCL := generateStackHCL(&StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		mainGo := generateGoMain(&StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		})
		err = os.WriteFile(filepath.Join(stackDir, "main.go"), []byte(mainGo), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(stackDir, 0o755) }()

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		}

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write go.mod")
	})

	t.Run("fails when main.tf cannot be written - Go", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		// Write stack.forge.hcl, main.go, and go.mod first
		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "go1.x",
		}

		stackHCL := generateStackHCL(opts)
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		mainGo := generateGoMain(opts)
		err = os.WriteFile(filepath.Join(stackDir, "main.go"), []byte(mainGo), 0o644)
		require.NoError(t, err)

		goMod := generateGoMod(opts)
		err = os.WriteFile(filepath.Join(stackDir, "go.mod"), []byte(goMod), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(stackDir, 0o755) }()

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write main.tf")
	})

	t.Run("fails when main.tf cannot be written - Python", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "python3.13",
		}

		// Write all files except main.tf
		stackHCL := generateStackHCL(opts)
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		handler := generatePythonHandler(opts)
		err = os.WriteFile(filepath.Join(stackDir, "handler.py"), []byte(handler), 0o644)
		require.NoError(t, err)

		requirements := generateRequirementsTxt()
		err = os.WriteFile(filepath.Join(stackDir, "requirements.txt"), []byte(requirements), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(stackDir, 0o755) }()

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write main.tf")
	})

	t.Run("fails when main.tf cannot be written - Node", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		err := os.MkdirAll(stackDir, 0o755)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "nodejs22.x",
		}

		// Write all files except main.tf
		stackHCL := generateStackHCL(opts)
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		index := generateNodeIndex(opts)
		err = os.WriteFile(filepath.Join(stackDir, "index.js"), []byte(index), 0o644)
		require.NoError(t, err)

		pkg := generatePackageJson(opts)
		err = os.WriteFile(filepath.Join(stackDir, "package.json"), []byte(pkg), 0o644)
		require.NoError(t, err)

		// Make directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(stackDir, 0o755) }()

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write main.tf")
	})

	t.Run("fails when main.tf cannot be written - Java", func(t *testing.T) {
		tmpDir := t.TempDir()
		stackDir := filepath.Join(tmpDir, "test-stack")
		javaDir := filepath.Join(stackDir, "src", "main", "java", "com", "example")
		err := os.MkdirAll(javaDir, 0o755)
		require.NoError(t, err)

		opts := &StackOptions{
			Name:    "test-stack",
			Runtime: "java21",
		}

		// Write all files except main.tf
		stackHCL := generateStackHCL(opts)
		err = os.WriteFile(filepath.Join(stackDir, "stack.forge.hcl"), []byte(stackHCL), 0o644)
		require.NoError(t, err)

		handler := generateJavaHandler(opts)
		err = os.WriteFile(filepath.Join(javaDir, "Handler.java"), []byte(handler), 0o644)
		require.NoError(t, err)

		pom := generatePomXml(opts)
		err = os.WriteFile(filepath.Join(stackDir, "pom.xml"), []byte(pom), 0o644)
		require.NoError(t, err)

		// Make stack directory read-only
		err = os.Chmod(stackDir, 0o555)
		require.NoError(t, err)
		defer func() { _ = os.Chmod(stackDir, 0o755) }()

		err = GenerateStack(tmpDir, opts)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write main.tf")
	})
}

// TestEdgeCases tests edge cases in scaffold generation.
func TestEdgeCases(t *testing.T) {
	t.Run("empty description uses default", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &StackOptions{
			Name:        "test-stack",
			Runtime:     "go1.x",
			Description: "", // Empty description
		}

		err := GenerateStack(tmpDir, opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "test-stack", "stack.forge.hcl"))
		require.NoError(t, err)

		// Should contain default description: "{name} stack"
		assert.Contains(t, string(content), "test-stack stack")
	})

	t.Run("special characters in project name", func(t *testing.T) {
		tmpDir := t.TempDir()

		opts := &ProjectOptions{
			Name:   "my-awesome-project_v2",
			Region: "eu-west-1",
		}

		err := GenerateProject(tmpDir, opts)
		require.NoError(t, err)

		content, err := os.ReadFile(filepath.Join(tmpDir, "forge.hcl"))
		require.NoError(t, err)

		assert.Contains(t, string(content), "my-awesome-project_v2")
		assert.Contains(t, string(content), "eu-west-1")
	})

	t.Run("gitignore contains all expected patterns", func(t *testing.T) {
		gitignore := generateGitignore()

		expectedPatterns := []string{
			".terraform/",
			"*.tfstate",
			"*.tfstate.backup",
			".terraform.lock.hcl",
			"terraform.tfvars",
			"bin/",
			"dist/",
			"*.zip",
			".DS_Store",
		}

		for _, pattern := range expectedPatterns {
			assert.Contains(t, gitignore, pattern, "gitignore should contain %s", pattern)
		}
	})

	t.Run("README contains project name and basic instructions", func(t *testing.T) {
		opts := &ProjectOptions{
			Name:   "test-readme",
			Region: "us-east-1",
		}

		readme := generateReadme(opts)

		assert.Contains(t, readme, "test-readme")
		assert.Contains(t, readme, "forge deploy")
		assert.Contains(t, readme, "forge destroy")
		assert.Contains(t, readme, "forge version")
	})
}

// TestGenerateHelperFunctions tests individual generator helper functions.
func TestGenerateHelperFunctions(t *testing.T) {
	t.Run("generateForgeHCL produces valid HCL", func(t *testing.T) {
		opts := &ProjectOptions{
			Name:   "test-project",
			Region: "us-west-2",
		}

		hcl := generateForgeHCL(opts)

		assert.Contains(t, hcl, "project {")
		assert.Contains(t, hcl, "test-project")
		assert.Contains(t, hcl, "us-west-2")
	})

	t.Run("generateStackHCL with custom description", func(t *testing.T) {
		opts := &StackOptions{
			Name:        "api",
			Runtime:     "go1.x",
			Description: "Custom API description",
		}

		hcl := generateStackHCL(opts)

		assert.Contains(t, hcl, "stack {")
		assert.Contains(t, hcl, "api")
		assert.Contains(t, hcl, "go1.x")
		assert.Contains(t, hcl, "Custom API description")
	})

	t.Run("generateGoMain contains Lambda handler", func(t *testing.T) {
		opts := &StackOptions{Name: "test"}
		code := generateGoMain(opts)

		assert.Contains(t, code, "package main")
		assert.Contains(t, code, "lambda.Start(handler)")
		assert.Contains(t, code, "events.APIGatewayProxyRequest")
	})

	t.Run("generateGoMod contains module name", func(t *testing.T) {
		opts := &StackOptions{Name: "my-module"}
		mod := generateGoMod(opts)

		assert.Contains(t, mod, "module my-module")
		assert.Contains(t, mod, "go 1.21")
		assert.Contains(t, mod, "github.com/aws/aws-lambda-go")
	})

	t.Run("generateGoTerraform contains all resources", func(t *testing.T) {
		opts := &StackOptions{Name: "test", Runtime: "provided.al2023"}
		tf := generateGoTerraform(opts)

		assert.Contains(t, tf, "resource \"aws_lambda_function\" \"test\"")
		assert.Contains(t, tf, "resource \"aws_iam_role\" \"lambda\"")
		assert.Contains(t, tf, "resource \"aws_iam_role_policy_attachment\" \"lambda_basic\"")
		assert.Contains(t, tf, "output \"function_name\"")
		assert.Contains(t, tf, "output \"function_arn\"")
	})

	t.Run("generatePythonHandler contains handler function", func(t *testing.T) {
		opts := &StackOptions{Name: "test"}
		code := generatePythonHandler(opts)

		assert.Contains(t, code, "def handler(event, context):")
		assert.Contains(t, code, "import json")
		assert.Contains(t, code, "import logging")
	})

	t.Run("generateNodeIndex contains handler export", func(t *testing.T) {
		opts := &StackOptions{Name: "test"}
		code := generateNodeIndex(opts)

		assert.Contains(t, code, "exports.handler = async (event)")
		assert.Contains(t, code, "console.log")
	})

	t.Run("generatePackageJson contains correct metadata", func(t *testing.T) {
		opts := &StackOptions{
			Name:        "my-function",
			Description: "Test function",
		}
		pkg := generatePackageJson(opts)

		assert.Contains(t, pkg, "\"name\": \"my-function\"")
		assert.Contains(t, pkg, "\"description\": \"Test function\"")
		assert.Contains(t, pkg, "\"main\": \"index.js\"")
	})

	t.Run("generateJavaHandler contains handler class", func(t *testing.T) {
		opts := &StackOptions{Name: "test"}
		code := generateJavaHandler(opts)

		assert.Contains(t, code, "package com.example")
		assert.Contains(t, code, "public class Handler implements RequestHandler")
		assert.Contains(t, code, "handleRequest")
	})

	t.Run("generatePomXml contains correct dependencies", func(t *testing.T) {
		opts := &StackOptions{Name: "my-service"}
		pom := generatePomXml(opts)

		assert.Contains(t, pom, "<artifactId>my-service</artifactId>")
		assert.Contains(t, pom, "aws-lambda-java-core")
		assert.Contains(t, pom, "aws-lambda-java-events")
		assert.Contains(t, pom, "maven-shade-plugin")
	})
}
