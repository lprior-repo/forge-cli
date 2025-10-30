// Package tfmodules provides type-safe Terraform module definitions compatible with Lingon.
// This extends Lingon to support terraform-aws-modules with full type safety and compile-time validation.
package tfmodules

import (
	"fmt"

	"github.com/golingon/lingon/pkg/terra"
)

// Module is the base interface that all Terraform modules must implement.
// This is compatible with Lingon's terra.Resource interface.
type Module interface {
	// LocalName returns the Terraform local name for this module instance
	LocalName() string

	// Configuration returns the module configuration as HCL
	Configuration() (string, error)
}

// BaseModule provides common functionality for all modules
type BaseModule struct {
	// localName is the Terraform local identifier
	localName string

	// source is the module source (e.g., "terraform-aws-modules/sqs/aws")
	source string

	// version is the module version constraint
	version string
}

// NewBaseModule creates a new base module
func NewBaseModule(localName, source, version string) BaseModule {
	return BaseModule{
		localName: localName,
		source:    source,
		version:   version,
	}
}

// LocalName returns the Terraform local name
func (b BaseModule) LocalName() string {
	return b.localName
}

// Source returns the module source
func (b BaseModule) Source() string {
	return b.source
}

// Version returns the module version
func (b BaseModule) Version() string {
	return b.version
}

// Output represents a Terraform module output reference.
// This allows type-safe references to module outputs in other resources.
type Output struct {
	// module is the module this output comes from
	module Module

	// attribute is the output attribute name
	attribute string
}

// NewOutput creates a new output reference
func NewOutput(module Module, attribute string) Output {
	return Output{
		module:    module,
		attribute: attribute,
	}
}

// Ref returns the Terraform reference string for this output.
// Example: "module.my_queue.queue_arn"
func (o Output) Ref() terra.StringValue {
	ref := terra.NewReference(fmt.Sprintf("module.%s.%s", o.module.LocalName(), o.attribute))
	return terra.ReferenceAsString(ref)
}

// String returns the string representation
func (o Output) String() string {
	return string(o.Ref())
}

// ModuleCall is a helper to generate module block HCL
type ModuleCall struct {
	Name    string
	Source  string
	Version string
	Args    map[string]interface{}
}

// ToHCL converts the module call to HCL string
func (m ModuleCall) ToHCL() string {
	hcl := fmt.Sprintf("module \"%s\" {\n", m.Name)
	hcl += fmt.Sprintf("  source  = \"%s\"\n", m.Source)
	if m.Version != "" {
		hcl += fmt.Sprintf("  version = \"%s\"\n", m.Version)
	}
	hcl += "\n"

	// TODO: Convert Args to proper HCL
	// For now, this is a simplified version
	// Full implementation would use hclwrite or lingon's HCL marshaling

	hcl += "}\n"
	return hcl
}

// Validator provides validation for module configurations
type Validator interface {
	// Validate checks if the module configuration is valid
	Validate() error
}

// WithValidation is a helper that adds validation to module operations
func WithValidation(m Module) error {
	if v, ok := m.(Validator); ok {
		return v.Validate()
	}
	return nil
}

// Stack represents a collection of modules that can be deployed together.
// This is compatible with Lingon's stack concept.
type Stack struct {
	// Name is the stack identifier
	Name string

	// Modules are all modules in this stack
	Modules []Module

	// Dependencies tracks module dependencies
	Dependencies map[string][]string
}

// NewStack creates a new module stack
func NewStack(name string) *Stack {
	return &Stack{
		Name:         name,
		Modules:      []Module{},
		Dependencies: make(map[string][]string),
	}
}

// AddModule adds a module to the stack
func (s *Stack) AddModule(m Module) {
	s.Modules = append(s.Modules, m)
}

// AddDependency declares that one module depends on another
func (s *Stack) AddDependency(dependent, dependsOn string) {
	if s.Dependencies[dependent] == nil {
		s.Dependencies[dependent] = []string{}
	}
	s.Dependencies[dependent] = append(s.Dependencies[dependent], dependsOn)
}

// ToHCL generates HCL for all modules in the stack
func (s *Stack) ToHCL() (string, error) {
	var hcl string

	for _, m := range s.Modules {
		config, err := m.Configuration()
		if err != nil {
			return "", fmt.Errorf("failed to generate config for %s: %w", m.LocalName(), err)
		}
		hcl += config + "\n\n"
	}

	return hcl, nil
}

// Validate validates all modules in the stack
func (s *Stack) Validate() error {
	for _, m := range s.Modules {
		if err := WithValidation(m); err != nil {
			return fmt.Errorf("validation failed for %s: %w", m.LocalName(), err)
		}
	}
	return nil
}
