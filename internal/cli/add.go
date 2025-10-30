// Package cli provides command-line interface for forge commands.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/generators"
	"github.com/lewis/forge/internal/generators/dynamodb"
	"github.com/lewis/forge/internal/generators/s3"
	"github.com/lewis/forge/internal/generators/sns"
	"github.com/lewis/forge/internal/generators/sqs"
	"github.com/spf13/cobra"
)

// NewAddCmd creates the 'add' command
func NewAddCmd() *cobra.Command {
	var (
		addToFunc   string
		addRaw      bool
		addNoModule bool
	)

	addCmd := &cobra.Command{
	Use:   "add <resource-type> <name>",
	Short: "Add AWS resources with generated Terraform code",
	Long: `
╭──────────────────────────────────────────────────────────────╮
│  ➕ Forge Add Resource                                      │
╰──────────────────────────────────────────────────────────────╯

Generate production-ready Terraform code for AWS resources.
Smart defaults, best practices, and Lambda integrations built-in.

📦 Supported Resource Types:
  sqs          - SQS queue with DLQ, encryption, monitoring
  dynamodb     - DynamoDB table with streams and backup
  sns          - SNS topic with subscriptions
  s3           - S3 bucket with versioning and encryption

🎯 What You Get:
  • Production-ready Terraform modules
  • AWS best practices applied automatically
  • Optional Lambda function integration
  • Encryption and monitoring configured
  • IAM policies generated

🚀 Examples:

  # Add standalone SQS queue
  forge add sqs orders-queue
    → Creates SQS queue with DLQ
    → Configures encryption at rest
    → Sets up CloudWatch alarms

  # Add SQS queue with Lambda trigger
  forge add sqs orders-queue --to=processor
    → Links queue to Lambda function
    → Generates IAM permissions
    → Configures batch settings

  # Use raw Terraform resources (no modules)
  forge add sqs orders-queue --raw

💡 Pro Tips:
  • Generated code is fully editable
  • Uses Terraform modules by default for simplicity
  • Use --raw for maximum control
  • Review generated code before applying

📁 Output:
  infra/
  ├── sqs_orders_queue.tf      # Generated resource
  └── sqs_orders_queue_iam.tf  # Generated IAM policies
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd, args, addToFunc, addRaw, addNoModule)
		},
	}

	addCmd.Flags().StringVar(&addToFunc, "to", "", "Target Lambda function for integration")
	addCmd.Flags().BoolVar(&addRaw, "raw", false, "Generate raw Terraform resources instead of modules")
	addCmd.Flags().BoolVar(&addNoModule, "no-module", false, "Alias for --raw")

	return addCmd
}

// runAdd executes the add command (I/O ACTION)
func runAdd(cmd *cobra.Command, args []string, toFunc string, raw bool, noModule bool) error {
	ctx := cmd.Context()
	resourceType := args[0]
	resourceName := args[1]

	// Determine module preference
	useModule := !raw && !noModule

	// Create resource intent
	intent := generators.ResourceIntent{
		Type:      generators.ResourceType(resourceType),
		Name:      resourceName,
		ToFunc:    toFunc,
		UseModule: useModule,
		Flags:     make(map[string]string),
	}

	// Get project root (working directory)
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Discover existing project state
	fmt.Println("🔍 Discovering project resources...")

	// Get generator for resource type
	registry := createGeneratorRegistry()
	generator, ok := registry.Get(intent.Type)
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", intent.Type)
	}

	infraDir := filepath.Join(projectRoot, "infra")

	// Chain all operations - automatic error short-circuiting
	writtenResult := E.Chain(func(state generators.ProjectState) E.Either[error, generators.WrittenFiles] {
		// Prompt for configuration (with defaults for MVP)
		fmt.Printf("📋 Configuring %s '%s'...\n", intent.Type, intent.Name)
		return E.Chain(func(config generators.ResourceConfig) E.Either[error, generators.WrittenFiles] {
			// Generate Terraform code
			fmt.Println("🔨 Generating Terraform code...")
			return E.Chain(func(code generators.GeneratedCode) E.Either[error, generators.WrittenFiles] {
				// Write files to disk
				fmt.Println("📝 Writing files...")
				return writeGeneratedFiles(code, infraDir)
			})(generator.Generate(config, state))
		})(generator.Prompt(ctx, intent, state))
	})(discoverProjectState(projectRoot))

	// Handle final result - report success or return error
	return E.Fold(
		func(e error) error {
			// Error case - return error
			return e
		},
		func(written generators.WrittenFiles) error {
			// Success case - report results
			fmt.Println("\n✅ Successfully added", intent.Type, intent.Name)
			if len(written.Created) > 0 {
				fmt.Println("\nCreated files:")
				for _, file := range written.Created {
					fmt.Printf("  + %s\n", file)
				}
			}
			if len(written.Updated) > 0 {
				fmt.Println("\nUpdated files:")
				for _, file := range written.Updated {
					fmt.Printf("  ~ %s\n", file)
				}
			}

			// Next steps
			fmt.Println("\nNext steps:")
			fmt.Println("  1. Review generated Terraform in infra/")
			fmt.Println("  2. Run: terraform init")
			fmt.Println("  3. Run: terraform plan")
			fmt.Println("  4. Run: terraform apply")

			return nil
		},
	)(writtenResult)
}

// createGeneratorRegistry creates registry with all generators (PURE)
func createGeneratorRegistry() *generators.Registry {
	registry := generators.NewRegistry()
	registry.Register(generators.ResourceSQS, sqs.New())
	registry.Register(generators.ResourceDynamoDB, dynamodb.New())
	registry.Register(generators.ResourceSNS, sns.New())
	registry.Register(generators.ResourceS3, s3.New())
	return registry
}

// discoverProjectState scans project for existing resources (I/O ACTION)
func discoverProjectState(projectRoot string) E.Either[error, generators.ProjectState] {
	infraDir := filepath.Join(projectRoot, "infra")

	// Check if infra directory exists
	if _, err := os.Stat(infraDir); os.IsNotExist(err) {
		return E.Left[generators.ProjectState](
			fmt.Errorf("infra/ directory not found - run 'forge new' first"),
		)
	}

	// For MVP, return minimal state
	// Phase 2: implement full Terraform parsing
	state := generators.ProjectState{
		ProjectRoot: projectRoot,
		Functions:   make(map[string]generators.FunctionInfo),
		Queues:      make(map[string]generators.QueueInfo),
		Tables:      make(map[string]generators.TableInfo),
		APIs:        make(map[string]generators.APIInfo),
		Topics:      make(map[string]generators.TopicInfo),
		InfraFiles:  []string{},
	}

	// Scan for .tf files
	entries, err := os.ReadDir(infraDir)
	if err != nil {
		return E.Left[generators.ProjectState](
			fmt.Errorf("failed to read infra directory: %w", err),
		)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".tf" {
			state.InfraFiles = append(state.InfraFiles, filepath.Join(infraDir, entry.Name()))
		}
	}

	// TODO: Parse .tf files to populate Functions, Queues, etc.
	// For MVP, just check if files exist

	return E.Right[error](state)
}

// writeGeneratedFiles persists code to disk (I/O ACTION)
func writeGeneratedFiles(code generators.GeneratedCode, infraDir string) E.Either[error, generators.WrittenFiles] {
	written := generators.WrittenFiles{
		Created: []string{},
		Updated: []string{},
		Skipped: []string{},
	}

	// Ensure infra directory exists
	if err := os.MkdirAll(infraDir, 0755); err != nil {
		return E.Left[generators.WrittenFiles](
			fmt.Errorf("failed to create infra directory: %w", err),
		)
	}

	// Write each file
	for _, file := range code.Files {
		filePath := filepath.Join(infraDir, file.Path)

		switch file.Mode {
		case generators.WriteModeCreate:
			// Create new file (error if exists)
			if _, err := os.Stat(filePath); err == nil {
				written.Skipped = append(written.Skipped, file.Path)
				continue
			}
			if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
				return E.Left[generators.WrittenFiles](
					fmt.Errorf("failed to create %s: %w", file.Path, err),
				)
			}
			written.Created = append(written.Created, file.Path)

		case generators.WriteModeAppend:
			// Append to existing file or create new
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return E.Left[generators.WrittenFiles](
					fmt.Errorf("failed to open %s: %w", file.Path, err),
				)
			}
			defer f.Close()

			// Check if file existed before
			stat, _ := os.Stat(filePath)
			existed := stat != nil && stat.Size() > 0

			if _, err := f.WriteString("\n" + file.Content); err != nil {
				return E.Left[generators.WrittenFiles](
					fmt.Errorf("failed to append to %s: %w", file.Path, err),
				)
			}

			if existed {
				written.Updated = append(written.Updated, file.Path)
			} else {
				written.Created = append(written.Created, file.Path)
			}

		case generators.WriteModeUpdate:
			// Update existing resource in file (requires parsing)
			// For MVP, treat as append
			// TODO: Implement smart update in Phase 2
			f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return E.Left[generators.WrittenFiles](
					fmt.Errorf("failed to open %s: %w", file.Path, err),
				)
			}
			defer f.Close()

			if _, err := f.WriteString("\n" + file.Content); err != nil {
				return E.Left[generators.WrittenFiles](
					fmt.Errorf("failed to write to %s: %w", file.Path, err),
				)
			}
			written.Updated = append(written.Updated, file.Path)
		}
	}

	return E.Right[error](written)
}
