// Package generators provides resource generation for forge add command.
// It implements a registry pattern for pluggable resource generators following
// functional programming principles with pure generation logic and I/O at edges.
package generators

import (
	"context"

	E "github.com/IBM/fp-go/either"
)

type (
	// ResourceType identifies the kind of resource to generate.
	ResourceType string

	// ResourceIntent represents user's intent to add a resource (PURE DATA).
	ResourceIntent struct {
		Type      ResourceType      // What kind of resource
		Name      string            // Resource name
		ToFunc    string            // Target Lambda function (for integrations)
		UseModule bool              // Use serverless.tf module vs raw resources
		Flags     map[string]string // Additional CLI flags
	}

	// ProjectState represents current project resources (PURE DATA).
	ProjectState struct {
		ProjectRoot string                  // Absolute path to project root
		Functions   map[string]FunctionInfo // Existing Lambda functions
		Queues      map[string]QueueInfo    // Existing SQS queues
		Tables      map[string]TableInfo    // Existing DynamoDB tables
		APIs        map[string]APIInfo      // Existing API Gateways
		Topics      map[string]TopicInfo    // Existing SNS topics
		InfraFiles  []string                // Paths to .tf files
	}

	// FunctionInfo describes an existing Lambda function.
	FunctionInfo struct {
		Name       string // Function name
		Runtime    string // Runtime (go1.x, python3.13, etc.)
		SourcePath string // Path to function source code
		Handler    string // Handler name
		TFResource string // Terraform resource name
	}

	// QueueInfo describes an existing SQS queue.
	QueueInfo struct {
		Name       string // Queue name
		URL        string // Queue URL (if known)
		ARN        string // Queue ARN (if known)
		TFResource string // Terraform resource/module name
	}

	// TableInfo describes an existing DynamoDB table.
	TableInfo struct {
		Name       string // Table name
		ARN        string // Table ARN (if known)
		TFResource string // Terraform resource/module name
	}

	// APIInfo describes an existing API Gateway.
	APIInfo struct {
		Name       string // API name
		Type       string // HTTP, REST, or WebSocket
		TFResource string // Terraform resource/module name
	}

	// TopicInfo describes an existing SNS topic.
	TopicInfo struct {
		Name       string // Topic name
		ARN        string // Topic ARN (if known)
		TFResource string // Terraform resource/module name
	}

	// ResourceConfig contains configuration for resource generation (PURE DATA).
	ResourceConfig struct {
		Type        ResourceType           // Resource type
		Name        string                 // Resource name
		Module      bool                   // Use module vs raw resources
		Variables   map[string]interface{} // Configuration variables
		Integration *IntegrationConfig     // Optional integration config
	}

	// IntegrationConfig defines how to wire resources together (PURE DATA).
	IntegrationConfig struct {
		TargetFunction string             // Lambda function to integrate with
		EventSource    *EventSourceConfig // Event source mapping config
		IAMPermissions []IAMPermission    // Required IAM permissions
		EnvVars        map[string]string  // Environment variables to add
	}

	// EventSourceConfig for Lambda event source mappings (PURE DATA).
	EventSourceConfig struct {
		ARNExpression         string // Terraform expression for source ARN
		BatchSize             int    // Batch size for events
		MaxBatchingWindowSecs int    // Maximum batching window
		MaxConcurrency        int    // Maximum concurrent invocations
	}

	// IAMPermission defines an IAM policy statement (PURE DATA).
	IAMPermission struct {
		Effect    string   // Allow or Deny
		Actions   []string // IAM actions (e.g., sqs:ReceiveMessage)
		Resources []string // Terraform resource references
	}

	// GeneratedCode represents generated Terraform code (PURE DATA).
	GeneratedCode struct {
		Resources   string        // Resource definitions
		Variables   string        // Variable definitions
		Outputs     string        // Output definitions
		ModuleCalls string        // Module invocations
		Files       []FileToWrite // Files to write
	}

	// FileToWrite specifies a file to create/update (PURE DATA).
	FileToWrite struct {
		Path    string    // Relative path from infra/
		Content string    // File content
		Mode    WriteMode // How to write (create, append, update)
	}

	// WriteMode determines how to write files.
	WriteMode string

	// WrittenFiles tracks what was written (PURE DATA).
	WrittenFiles struct {
		Created []string // Newly created files
		Updated []string // Modified files
		Skipped []string // Skipped (already exist)
	}

	// Generator defines the interface for resource generators.
	Generator interface {
		// Prompt gathers configuration from user (I/O ACTION)
		Prompt(ctx context.Context, intent ResourceIntent, state ProjectState) E.Either[error, ResourceConfig]

		// Generate creates Terraform code from configuration (PURE CALCULATION)
		Generate(config ResourceConfig, state ProjectState) E.Either[error, GeneratedCode]

		// Validate checks if configuration is valid (PURE CALCULATION)
		Validate(config ResourceConfig) E.Either[error, ResourceConfig]
	}

	// Registry is an immutable map of resource type to generator.
	// Uses copy-on-write semantics for thread-safety and functional purity.
	Registry map[ResourceType]Generator
)

const (
	ResourceLambda        ResourceType = "lambda"
	ResourceSQS           ResourceType = "sqs"
	ResourceSNS           ResourceType = "sns"
	ResourceDynamoDB      ResourceType = "dynamodb"
	ResourceAPIGateway    ResourceType = "apigw"
	ResourceEventBridge   ResourceType = "eventbridge"
	ResourceStepFunctions ResourceType = "sfn"
	ResourceS3            ResourceType = "s3"
	ResourceCognito       ResourceType = "cognito"
)

const (
	WriteModeCreate WriteMode = "create" // Create new file (error if exists)
	WriteModeAppend WriteMode = "append" // Append to existing file
	WriteModeUpdate WriteMode = "update" // Update existing resource in file
)

// NewRegistry creates an empty registry.
func NewRegistry() Registry {
	return make(Registry)
}

// Register returns a NEW registry with the generator added (copy-on-write).
// Does not mutate the original registry - pure functional operation.
func (r Registry) Register(resourceType ResourceType, generator Generator) Registry {
	// Create new registry with capacity for existing + 1
	newRegistry := make(Registry, len(r)+1)

	// Copy all existing generators
	for k, v := range r {
		newRegistry[k] = v
	}

	// Add new generator
	newRegistry[resourceType] = generator

	return newRegistry
}

// Get retrieves a generator for a resource type.
// Returns (generator, true) if found, (nil, false) if not found.
func (r Registry) Get(resourceType ResourceType) (Generator, bool) {
	gen, ok := r[resourceType]
	return gen, ok
}

// DiscoverFunc scans project to find existing resources (I/O ACTION).
type DiscoverFunc func(projectRoot string) E.Either[error, ProjectState]

// PromptFunc interactively gathers configuration (I/O ACTION).
type PromptFunc func(ctx context.Context, intent ResourceIntent, state ProjectState) E.Either[error, ResourceConfig]

// GenerateFunc creates Terraform code (PURE CALCULATION).
type GenerateFunc func(config ResourceConfig, state ProjectState) E.Either[error, GeneratedCode]

// WriteFunc persists generated code to disk (I/O ACTION).
type WriteFunc func(code GeneratedCode, infraDir string) E.Either[error, WrittenFiles]
