package lingon

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FunctionMetadata represents the optional function.forge.json file.
type FunctionMetadata struct {
	HTTP struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	} `json:"http"`
	Timeout int `json:"timeout"`
	Memory  int `json:"memory"`
}

// DiscoverFunctions scans src/handlers and returns function configs.
func DiscoverFunctions(projectRoot string) (map[string]FunctionConfig, error) {
	srcDir := filepath.Join(projectRoot, "src")
	handlersDir := filepath.Join(srcDir, "handlers")

	// Check if handlers directory exists
	if _, err := os.Stat(handlersDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("handlers directory not found: %s", handlersDir)
	}

	// Detect project language
	runtime, err := detectProjectRuntime(srcDir)
	if err != nil {
		return nil, fmt.Errorf("failed to detect project runtime: %w", err)
	}

	handlers := make(map[string]FunctionConfig)

	// Scan src/handlers/*
	entries, err := os.ReadDir(handlersDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read handlers directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		handlerDir := filepath.Join(handlersDir, name)

		// Check if handler file exists
		handlerFile := getHandlerFile(handlerDir, runtime)
		if handlerFile == "" {
			// Skip directories without handler files
			continue
		}

		// Create default config
		config := FunctionConfig{
			Handler:     getHandlerName(runtime),
			Runtime:     runtime,
			Timeout:     30,
			MemorySize:  256,
			Description: name + " Lambda function",
			Source: SourceConfig{
				Path: filepath.Join("src", "handlers", name),
			},
		}

		// Load function.forge.json if exists
		metadataPath := filepath.Join(handlerDir, "function.forge.json")
		if _, err := os.Stat(metadataPath); err == nil {
			metadata, err := loadFunctionMetadata(metadataPath)
			if err != nil {
				return nil, fmt.Errorf("failed to load metadata for %s: %w", name, err)
			}
			applyMetadata(&config, metadata, name)
		}

		handlers[name] = config
	}

	if len(handlers) == 0 {
		return nil, fmt.Errorf("no handler functions found in %s", handlersDir)
	}

	return handlers, nil
}

// detectProjectRuntime detects the project's language from manifest files.
func detectProjectRuntime(srcDir string) (string, error) {
	// Check for Go
	if _, err := os.Stat(filepath.Join(srcDir, "go.mod")); err == nil {
		return "provided.al2023", nil // Go 1.x with custom runtime
	}

	// Check for Node.js
	if _, err := os.Stat(filepath.Join(srcDir, "package.json")); err == nil {
		return "nodejs20.x", nil
	}

	// Check for Python
	if _, err := os.Stat(filepath.Join(srcDir, "requirements.txt")); err == nil {
		return "python3.11", nil
	}

	// Check for Java
	if _, err := os.Stat(filepath.Join(srcDir, "pom.xml")); err == nil {
		return "java17", nil
	}
	if _, err := os.Stat(filepath.Join(srcDir, "build.gradle")); err == nil {
		return "java17", nil
	}

	return "", errors.New("could not detect project language (no go.mod, package.json, requirements.txt, or build files found)")
}

// getHandlerFile returns the handler file name if it exists.
func getHandlerFile(dir, runtime string) string {
	candidates := []string{}

	switch {
	case strings.HasPrefix(runtime, "go") || runtime == "provided.al2023":
		candidates = []string{"main.go"}
	case strings.HasPrefix(runtime, "nodejs"):
		candidates = []string{"index.js", "index.ts", "handler.js", "handler.ts"}
	case strings.HasPrefix(runtime, "python"):
		candidates = []string{"main.py", "handler.py", "lambda_function.py"}
	case strings.HasPrefix(runtime, "java"):
		candidates = []string{"Handler.java", "Main.java"}
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
			return candidate
		}
	}

	return ""
}

// getHandlerName returns the default handler name for the runtime.
func getHandlerName(runtime string) string {
	switch {
	case strings.HasPrefix(runtime, "go") || runtime == "provided.al2023":
		return "bootstrap"
	case strings.HasPrefix(runtime, "nodejs"):
		return "index.handler"
	case strings.HasPrefix(runtime, "python"):
		return "main.handler"
	case strings.HasPrefix(runtime, "java"):
		return "Handler::handleRequest"
	default:
		return "index.handler"
	}
}

// loadFunctionMetadata reads and parses function.forge.json.
func loadFunctionMetadata(path string) (*FunctionMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var metadata FunctionMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &metadata, nil
}

// applyMetadata applies the metadata to the function config.
func applyMetadata(config *FunctionConfig, metadata *FunctionMetadata, functionName string) {
	// Apply timeout if specified
	if metadata.Timeout > 0 {
		config.Timeout = metadata.Timeout
	}

	// Apply memory if specified
	if metadata.Memory > 0 {
		config.MemorySize = metadata.Memory
	}

	// Apply HTTP routing if specified
	if metadata.HTTP.Method != "" && metadata.HTTP.Path != "" {
		config.HTTPRouting = &HTTPRoutingConfig{
			Path:   metadata.HTTP.Path,
			Method: metadata.HTTP.Method,
		}
	}
}
