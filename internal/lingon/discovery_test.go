package lingon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscoverFunctions(t *testing.T) {
	t.Run("discovers Go functions", func(t *testing.T) {
		// Create temp directory structure
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		helloDir := filepath.Join(handlersDir, "hello")

		require.NoError(t, os.MkdirAll(helloDir, 0755))

		// Create go.mod
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "go.mod"), []byte("module test"), 0644))

		// Create handler
		require.NoError(t, os.WriteFile(filepath.Join(helloDir, "main.go"), []byte("package main"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "hello")
		assert.Equal(t, "provided.al2023", functions["hello"].Runtime)
		assert.Equal(t, "bootstrap", functions["hello"].Handler)
	})

	t.Run("discovers Node.js functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		apiDir := filepath.Join(handlersDir, "api")

		require.NoError(t, os.MkdirAll(apiDir, 0755))

		// Create package.json
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))

		// Create handler
		require.NoError(t, os.WriteFile(filepath.Join(apiDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "api")
		assert.Equal(t, "nodejs20.x", functions["api"].Runtime)
		assert.Equal(t, "index.handler", functions["api"].Handler)
	})

	t.Run("discovers Python functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		workerDir := filepath.Join(handlersDir, "worker")

		require.NoError(t, os.MkdirAll(workerDir, 0755))

		// Create requirements.txt
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "requirements.txt"), []byte("boto3"), 0644))

		// Create handler
		require.NoError(t, os.WriteFile(filepath.Join(workerDir, "main.py"), []byte("def handler(event, context): pass"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "worker")
		assert.Equal(t, "python3.11", functions["worker"].Runtime)
		assert.Equal(t, "main.handler", functions["worker"].Handler)
	})

	t.Run("discovers Java functions with pom.xml", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		javaDir := filepath.Join(handlersDir, "process")

		require.NoError(t, os.MkdirAll(javaDir, 0755))

		// Create pom.xml
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "pom.xml"), []byte("<project></project>"), 0644))

		// Create handler
		require.NoError(t, os.WriteFile(filepath.Join(javaDir, "Handler.java"), []byte("public class Handler {}"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "process")
		assert.Equal(t, "java17", functions["process"].Runtime)
		assert.Equal(t, "Handler::handleRequest", functions["process"].Handler)
	})

	t.Run("discovers Java functions with build.gradle", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		javaDir := filepath.Join(handlersDir, "gradle-fn")

		require.NoError(t, os.MkdirAll(javaDir, 0755))

		// Create build.gradle
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "build.gradle"), []byte("plugins { id 'java' }"), 0644))

		// Create handler
		require.NoError(t, os.WriteFile(filepath.Join(javaDir, "Main.java"), []byte("public class Main {}"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "gradle-fn")
		assert.Equal(t, "java17", functions["gradle-fn"].Runtime)
	})

	t.Run("discovers multiple functions", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")

		require.NoError(t, os.MkdirAll(handlersDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))

		// Create multiple handlers
		for _, name := range []string{"api", "worker", "processor"} {
			dir := filepath.Join(handlersDir, name)
			require.NoError(t, os.MkdirAll(dir, 0755))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "index.js"), []byte("exports.handler = async () => {}"), 0644))
		}

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 3)
		assert.Contains(t, functions, "api")
		assert.Contains(t, functions, "worker")
		assert.Contains(t, functions, "processor")
	})

	t.Run("applies function metadata from function.forge.json", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		helloDir := filepath.Join(handlersDir, "hello")

		require.NoError(t, os.MkdirAll(helloDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(helloDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		// Create metadata file
		metadata := FunctionMetadata{
			Timeout: 60,
			Memory:  512,
		}
		metadata.HTTP.Method = "POST"
		metadata.HTTP.Path = "/api/hello"

		metadataJSON, err := json.Marshal(metadata)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(helloDir, "function.forge.json"), metadataJSON, 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Equal(t, 60, functions["hello"].Timeout)
		assert.Equal(t, 512, functions["hello"].MemorySize)
		assert.NotNil(t, functions["hello"].HTTPRouting)
		assert.Equal(t, "POST", functions["hello"].HTTPRouting.Method)
		assert.Equal(t, "/api/hello", functions["hello"].HTTPRouting.Path)
	})

	t.Run("skips directories without handler files", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		emptyDir := filepath.Join(handlersDir, "empty")
		validDir := filepath.Join(handlersDir, "valid")

		require.NoError(t, os.MkdirAll(emptyDir, 0755))
		require.NoError(t, os.MkdirAll(validDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))

		// Only create handler in valid directory
		require.NoError(t, os.WriteFile(filepath.Join(validDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		functions, err := DiscoverFunctions(tmpDir)

		require.NoError(t, err)
		assert.Len(t, functions, 1)
		assert.Contains(t, functions, "valid")
		assert.NotContains(t, functions, "empty")
	})

	t.Run("fails when handlers directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := DiscoverFunctions(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "handlers directory not found")
	})

	t.Run("fails when no runtime can be detected", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")

		require.NoError(t, os.MkdirAll(handlersDir, 0755))

		_, err := DiscoverFunctions(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to detect project runtime")
	})

	t.Run("fails when no handler functions found", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")

		require.NoError(t, os.MkdirAll(handlersDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))

		_, err := DiscoverFunctions(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no handler functions found")
	})

	t.Run("fails when metadata JSON is invalid", func(t *testing.T) {
		tmpDir := t.TempDir()
		srcDir := filepath.Join(tmpDir, "src")
		handlersDir := filepath.Join(srcDir, "handlers")
		helloDir := filepath.Join(handlersDir, "hello")

		require.NoError(t, os.MkdirAll(helloDir, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "package.json"), []byte("{}"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(helloDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		// Create invalid JSON metadata
		require.NoError(t, os.WriteFile(filepath.Join(helloDir, "function.forge.json"), []byte("{ invalid json }"), 0644))

		_, err := DiscoverFunctions(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load metadata")
	})
}

func TestDetectProjectRuntime(t *testing.T) {
	t.Run("detects Go runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "provided.al2023", runtime)
	})

	t.Run("detects Node.js runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "nodejs20.x", runtime)
	})

	t.Run("detects Python runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "requirements.txt"), []byte("boto3"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "python3.11", runtime)
	})

	t.Run("detects Java runtime with pom.xml", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "pom.xml"), []byte("<project></project>"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "java17", runtime)
	})

	t.Run("detects Java runtime with build.gradle", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "build.gradle"), []byte("plugins { id 'java' }"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "java17", runtime)
	})

	t.Run("fails when no runtime detected", func(t *testing.T) {
		tmpDir := t.TempDir()

		_, err := detectProjectRuntime(tmpDir)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not detect project language")
	})

	t.Run("prefers Go over other runtimes", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create multiple manifest files
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test"), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte("{}"), 0644))

		runtime, err := detectProjectRuntime(tmpDir)

		require.NoError(t, err)
		assert.Equal(t, "provided.al2023", runtime) // Go takes precedence
	})
}

func TestGetHandlerFile(t *testing.T) {
	t.Run("finds main.go for Go runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644))

		handlerFile := getHandlerFile(tmpDir, "provided.al2023")

		assert.Equal(t, "main.go", handlerFile)
	})

	t.Run("finds index.js for Node.js runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte("exports.handler = async () => {}"), 0644))

		handlerFile := getHandlerFile(tmpDir, "nodejs20.x")

		assert.Equal(t, "index.js", handlerFile)
	})

	t.Run("finds index.ts for TypeScript", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.ts"), []byte("export const handler = async () => {}"), 0644))

		handlerFile := getHandlerFile(tmpDir, "nodejs20.x")

		assert.Equal(t, "index.ts", handlerFile)
	})

	t.Run("finds main.py for Python runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "main.py"), []byte("def handler(event, context): pass"), 0644))

		handlerFile := getHandlerFile(tmpDir, "python3.11")

		assert.Equal(t, "main.py", handlerFile)
	})

	t.Run("finds Handler.java for Java runtime", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Handler.java"), []byte("public class Handler {}"), 0644))

		handlerFile := getHandlerFile(tmpDir, "java17")

		assert.Equal(t, "Handler.java", handlerFile)
	})

	t.Run("returns empty string when no handler file found", func(t *testing.T) {
		tmpDir := t.TempDir()

		handlerFile := getHandlerFile(tmpDir, "nodejs20.x")

		assert.Equal(t, "", handlerFile)
	})

	t.Run("prefers index.js over handler.js for Node.js", func(t *testing.T) {
		tmpDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "index.js"), []byte(""), 0644))
		require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "handler.js"), []byte(""), 0644))

		handlerFile := getHandlerFile(tmpDir, "nodejs20.x")

		assert.Equal(t, "index.js", handlerFile)
	})
}

func TestGetHandlerName(t *testing.T) {
	testCases := []struct {
		runtime  string
		expected string
	}{
		{"provided.al2023", "bootstrap"},
		{"go1.x", "bootstrap"},
		{"nodejs18.x", "index.handler"},
		{"nodejs20.x", "index.handler"},
		{"python3.9", "main.handler"},
		{"python3.11", "main.handler"},
		{"java11", "Handler::handleRequest"},
		{"java17", "Handler::handleRequest"},
		{"unknown", "index.handler"}, // default
	}

	for _, tc := range testCases {
		t.Run(tc.runtime, func(t *testing.T) {
			handler := getHandlerName(tc.runtime)
			assert.Equal(t, tc.expected, handler)
		})
	}
}

func TestLoadFunctionMetadata(t *testing.T) {
	t.Run("loads valid metadata", func(t *testing.T) {
		tmpDir := t.TempDir()
		metadataPath := filepath.Join(tmpDir, "function.forge.json")

		metadata := FunctionMetadata{
			Timeout: 120,
			Memory:  1024,
		}
		metadata.HTTP.Method = "GET"
		metadata.HTTP.Path = "/test"

		data, err := json.Marshal(metadata)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(metadataPath, data, 0644))

		loaded, err := loadFunctionMetadata(metadataPath)

		require.NoError(t, err)
		assert.Equal(t, 120, loaded.Timeout)
		assert.Equal(t, 1024, loaded.Memory)
		assert.Equal(t, "GET", loaded.HTTP.Method)
		assert.Equal(t, "/test", loaded.HTTP.Path)
	})

	t.Run("fails on invalid JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		metadataPath := filepath.Join(tmpDir, "function.forge.json")

		require.NoError(t, os.WriteFile(metadataPath, []byte("{ invalid }"), 0644))

		_, err := loadFunctionMetadata(metadataPath)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse JSON")
	})

	t.Run("fails when file does not exist", func(t *testing.T) {
		_, err := loadFunctionMetadata("/nonexistent/path.json")

		assert.Error(t, err)
	})
}

func TestApplyMetadata(t *testing.T) {
	t.Run("applies timeout from metadata", func(t *testing.T) {
		config := FunctionConfig{
			Timeout: 30,
		}
		metadata := &FunctionMetadata{
			Timeout: 120,
		}

		applyMetadata(&config, metadata, "test")

		assert.Equal(t, 120, config.Timeout)
	})

	t.Run("does not override timeout when metadata is zero", func(t *testing.T) {
		config := FunctionConfig{
			Timeout: 30,
		}
		metadata := &FunctionMetadata{
			Timeout: 0,
		}

		applyMetadata(&config, metadata, "test")

		assert.Equal(t, 30, config.Timeout)
	})

	t.Run("applies memory from metadata", func(t *testing.T) {
		config := FunctionConfig{
			MemorySize: 256,
		}
		metadata := &FunctionMetadata{
			Memory: 512,
		}

		applyMetadata(&config, metadata, "test")

		assert.Equal(t, 512, config.MemorySize)
	})

	t.Run("applies HTTP routing from metadata", func(t *testing.T) {
		config := FunctionConfig{}
		metadata := &FunctionMetadata{}
		metadata.HTTP.Method = "POST"
		metadata.HTTP.Path = "/api/users"

		applyMetadata(&config, metadata, "test")

		assert.NotNil(t, config.HTTPRouting)
		assert.Equal(t, "POST", config.HTTPRouting.Method)
		assert.Equal(t, "/api/users", config.HTTPRouting.Path)
	})

	t.Run("does not apply HTTP routing when method or path is empty", func(t *testing.T) {
		config := FunctionConfig{}
		metadata := &FunctionMetadata{}
		metadata.HTTP.Method = "GET"
		metadata.HTTP.Path = "" // Empty path

		applyMetadata(&config, metadata, "test")

		assert.Nil(t, config.HTTPRouting)
	})

	t.Run("applies all metadata fields", func(t *testing.T) {
		config := FunctionConfig{
			Timeout:    30,
			MemorySize: 256,
		}
		metadata := &FunctionMetadata{
			Timeout: 60,
			Memory:  1024,
		}
		metadata.HTTP.Method = "PUT"
		metadata.HTTP.Path = "/api/resource"

		applyMetadata(&config, metadata, "test")

		assert.Equal(t, 60, config.Timeout)
		assert.Equal(t, 1024, config.MemorySize)
		assert.NotNil(t, config.HTTPRouting)
		assert.Equal(t, "PUT", config.HTTPRouting.Method)
		assert.Equal(t, "/api/resource", config.HTTPRouting.Path)
	})
}
