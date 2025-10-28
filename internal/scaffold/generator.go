package scaffold

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/*
var templatesFS embed.FS

// Generator generates project scaffolding
type Generator struct {
	projectRoot string
	tmpl        *template.Template
}

// NewGenerator creates a new scaffold generator
func NewGenerator(projectRoot string) (*Generator, error) {
	// Parse all embedded templates
	tmpl, err := template.New("").
		Funcs(templateFuncs()).
		ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Generator{
		projectRoot: projectRoot,
		tmpl:        tmpl,
	}, nil
}

// ProjectOptions configures project generation
type ProjectOptions struct {
	Name   string
	Region string
}

// StackOptions configures stack generation
type StackOptions struct {
	Name        string
	Runtime     string
	Description string
}

// GenerateProject creates a new forge project structure
func (g *Generator) GenerateProject(opts *ProjectOptions) error {
	// Create project directory
	if err := os.MkdirAll(g.projectRoot, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate forge.hcl
	if err := g.renderTemplate("forge.hcl.tmpl", filepath.Join(g.projectRoot, "forge.hcl"), opts); err != nil {
		return err
	}

	// Generate .gitignore
	if err := g.renderTemplate("gitignore.tmpl", filepath.Join(g.projectRoot, ".gitignore"), opts); err != nil {
		return err
	}

	// Generate README.md
	if err := g.renderTemplate("README.md.tmpl", filepath.Join(g.projectRoot, "README.md"), opts); err != nil {
		return err
	}

	return nil
}

// GenerateStack creates a new stack with the specified runtime
func (g *Generator) GenerateStack(opts *StackOptions) error {
	stackDir := filepath.Join(g.projectRoot, opts.Name)

	// Create stack directory
	if err := os.MkdirAll(stackDir, 0755); err != nil {
		return fmt.Errorf("failed to create stack directory: %w", err)
	}

	// Generate stack.forge.hcl
	stackHCLPath := filepath.Join(stackDir, "stack.forge.hcl")
	if err := g.renderTemplate("stack.forge.hcl.tmpl", stackHCLPath, opts); err != nil {
		return err
	}

	// Generate runtime-specific files
	switch {
	case strings.HasPrefix(opts.Runtime, "go"), strings.HasPrefix(opts.Runtime, "provided"):
		return g.generateGoStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "python"):
		return g.generatePythonStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "nodejs"):
		return g.generateNodeStack(stackDir, opts)
	case strings.HasPrefix(opts.Runtime, "java"):
		return g.generateJavaStack(stackDir, opts)
	default:
		return fmt.Errorf("unsupported runtime: %s", opts.Runtime)
	}
}

// generateGoStack creates Go-specific files
func (g *Generator) generateGoStack(stackDir string, opts *StackOptions) error {
	// Create main.go
	mainPath := filepath.Join(stackDir, "main.go")
	if err := g.renderTemplate("go_main.go.tmpl", mainPath, opts); err != nil {
		return err
	}

	// Create go.mod
	modPath := filepath.Join(stackDir, "go.mod")
	if err := g.renderTemplate("go.mod.tmpl", modPath, opts); err != nil {
		return err
	}

	// Create main.tf
	tfPath := filepath.Join(stackDir, "main.tf")
	if err := g.renderTemplate("go_main.tf.tmpl", tfPath, opts); err != nil {
		return err
	}

	return nil
}

// generatePythonStack creates Python-specific files
func (g *Generator) generatePythonStack(stackDir string, opts *StackOptions) error {
	// Create handler.py
	handlerPath := filepath.Join(stackDir, "handler.py")
	if err := g.renderTemplate("python_handler.py.tmpl", handlerPath, opts); err != nil {
		return err
	}

	// Create requirements.txt
	reqPath := filepath.Join(stackDir, "requirements.txt")
	if err := g.renderTemplate("requirements.txt.tmpl", reqPath, opts); err != nil {
		return err
	}

	// Create main.tf
	tfPath := filepath.Join(stackDir, "main.tf")
	if err := g.renderTemplate("python_main.tf.tmpl", tfPath, opts); err != nil {
		return err
	}

	return nil
}

// generateNodeStack creates Node.js-specific files
func (g *Generator) generateNodeStack(stackDir string, opts *StackOptions) error {
	// Create index.js
	indexPath := filepath.Join(stackDir, "index.js")
	if err := g.renderTemplate("node_index.js.tmpl", indexPath, opts); err != nil {
		return err
	}

	// Create package.json
	pkgPath := filepath.Join(stackDir, "package.json")
	if err := g.renderTemplate("package.json.tmpl", pkgPath, opts); err != nil {
		return err
	}

	// Create main.tf
	tfPath := filepath.Join(stackDir, "main.tf")
	if err := g.renderTemplate("node_main.tf.tmpl", tfPath, opts); err != nil {
		return err
	}

	return nil
}

// generateJavaStack creates Java-specific files
func (g *Generator) generateJavaStack(stackDir string, opts *StackOptions) error {
	// Create src/main/java/com/example directory structure
	javaDir := filepath.Join(stackDir, "src", "main", "java", "com", "example")
	if err := os.MkdirAll(javaDir, 0755); err != nil {
		return fmt.Errorf("failed to create Java source directory: %w", err)
	}

	// Create Handler.java
	handlerPath := filepath.Join(javaDir, "Handler.java")
	if err := g.renderTemplate("java_handler.java.tmpl", handlerPath, opts); err != nil {
		return err
	}

	// Create pom.xml
	pomPath := filepath.Join(stackDir, "pom.xml")
	if err := g.renderTemplate("pom.xml.tmpl", pomPath, opts); err != nil {
		return err
	}

	// Create main.tf
	tfPath := filepath.Join(stackDir, "main.tf")
	if err := g.renderTemplate("java_main.tf.tmpl", tfPath, opts); err != nil {
		return err
	}

	return nil
}

// renderTemplate renders a template to a file
func (g *Generator) renderTemplate(templateName, outputPath string, data interface{}) error {
	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", outputPath, err)
	}
	defer f.Close()

	// Execute template
	tmplPath := "templates/" + templateName
	if err := g.tmpl.ExecuteTemplate(f, templateName, data); err != nil {
		return fmt.Errorf("failed to render %s: %w", tmplPath, err)
	}

	return nil
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"replace": strings.ReplaceAll,
	}
}
