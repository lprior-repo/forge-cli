package tfmodules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockModule is a test implementation of the Module interface
type MockModule struct {
	name   string
	config string
	err    error
}

func (m *MockModule) LocalName() string {
	return m.name
}

func (m *MockModule) Configuration() (string, error) {
	return m.config, m.err
}

// MockValidator is a test module that implements Validator
type MockValidator struct {
	MockModule
	validateErr error
}

func (m *MockValidator) Validate() error {
	return m.validateErr
}

func TestNewBaseModule(t *testing.T) {
	t.Run("creates base module with all parameters", func(t *testing.T) {
		localName := "my_module"
		source := "terraform-aws-modules/sqs/aws"
		version := "~> 4.0"

		base := NewBaseModule(localName, source, version)

		assert.Equal(t, localName, base.LocalName())
		assert.Equal(t, source, base.Source())
		assert.Equal(t, version, base.Version())
	})

	t.Run("creates base module with empty values", func(t *testing.T) {
		base := NewBaseModule("", "", "")

		assert.Empty(t, base.LocalName())
		assert.Empty(t, base.Source())
		assert.Empty(t, base.Version())
	})

	t.Run("creates base module with special characters", func(t *testing.T) {
		localName := "my-module_v2"
		source := "git::https://github.com/user/repo.git//modules/sqs"
		version := ">= 1.0, < 2.0"

		base := NewBaseModule(localName, source, version)

		assert.Equal(t, localName, base.LocalName())
		assert.Equal(t, source, base.Source())
		assert.Equal(t, version, base.Version())
	})
}

func TestBaseModule_LocalName(t *testing.T) {
	t.Run("returns configured local name", func(t *testing.T) {
		base := NewBaseModule("test_module", "source", "version")
		assert.Equal(t, "test_module", base.LocalName())
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		base := BaseModule{}
		assert.Empty(t, base.LocalName())
	})
}

func TestBaseModule_Source(t *testing.T) {
	t.Run("returns configured source", func(t *testing.T) {
		source := "terraform-aws-modules/vpc/aws"
		base := NewBaseModule("vpc", source, "~> 3.0")

		assert.Equal(t, source, base.Source())
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		base := BaseModule{}
		assert.Empty(t, base.Source())
	})
}

func TestBaseModule_Version(t *testing.T) {
	t.Run("returns configured version", func(t *testing.T) {
		version := "~> 4.0"
		base := NewBaseModule("module", "source", version)

		assert.Equal(t, version, base.Version())
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		base := BaseModule{}
		assert.Empty(t, base.Version())
	})
}

func TestNewOutput(t *testing.T) {
	t.Run("creates output with module and attribute", func(t *testing.T) {
		module := &MockModule{name: "test_queue"}
		attribute := "queue_arn"

		output := NewOutput(module, attribute)

		assert.Equal(t, module, output.module)
		assert.Equal(t, attribute, output.attribute)
	})

	t.Run("creates output with different attributes", func(t *testing.T) {
		module := &MockModule{name: "my_module"}
		attributes := []string{"id", "arn", "url", "name"}

		for _, attr := range attributes {
			output := NewOutput(module, attr)
			assert.Equal(t, attr, output.attribute)
		}
	})
}

func TestOutput_Ref(t *testing.T) {
	t.Run("returns terraform reference string", func(t *testing.T) {
		module := &MockModule{name: "my_queue"}
		output := NewOutput(module, "queue_arn")

		ref := output.Ref()

		assert.Equal(t, "module.my_queue.queue_arn", ref)
	})

	t.Run("returns correct reference for different modules", func(t *testing.T) {
		testCases := []struct {
			moduleName string
			attribute  string
			expected   string
		}{
			{"orders_queue", "queue_arn", "module.orders_queue.queue_arn"},
			{"users_table", "table_name", "module.users_table.table_name"},
			{"notifications", "topic_arn", "module.notifications.topic_arn"},
			{"data_bucket", "bucket_id", "module.data_bucket.bucket_id"},
		}

		for _, tc := range testCases {
			module := &MockModule{name: tc.moduleName}
			output := NewOutput(module, tc.attribute)

			assert.Equal(t, tc.expected, output.Ref())
		}
	})

	t.Run("handles empty module name", func(t *testing.T) {
		module := &MockModule{name: ""}
		output := NewOutput(module, "attr")

		assert.Equal(t, "module..attr", output.Ref())
	})

	t.Run("handles empty attribute", func(t *testing.T) {
		module := &MockModule{name: "test"}
		output := NewOutput(module, "")

		assert.Equal(t, "module.test.", output.Ref())
	})
}

func TestOutput_String(t *testing.T) {
	t.Run("returns same as Ref", func(t *testing.T) {
		module := &MockModule{name: "my_queue"}
		output := NewOutput(module, "queue_url")

		assert.Equal(t, output.Ref(), output.String())
	})
}

func TestModuleCall_ToHCL(t *testing.T) {
	t.Run("generates basic HCL", func(t *testing.T) {
		mc := ModuleCall{
			Name:    "my_module",
			Source:  "terraform-aws-modules/sqs/aws",
			Version: "~> 4.0",
		}

		hcl := mc.ToHCL()

		assert.Contains(t, hcl, `module "my_module"`)
		assert.Contains(t, hcl, `source  = "terraform-aws-modules/sqs/aws"`)
		assert.Contains(t, hcl, `version = "~> 4.0"`)
	})

	t.Run("generates HCL without version", func(t *testing.T) {
		mc := ModuleCall{
			Name:   "my_module",
			Source: "terraform-aws-modules/vpc/aws",
		}

		hcl := mc.ToHCL()

		assert.Contains(t, hcl, `module "my_module"`)
		assert.Contains(t, hcl, `source  = "terraform-aws-modules/vpc/aws"`)
		assert.NotContains(t, hcl, "version")
	})

	t.Run("generates HCL with arguments map", func(t *testing.T) {
		mc := ModuleCall{
			Name:    "test",
			Source:  "source",
			Version: "1.0",
			Args: map[string]interface{}{
				"name": "test",
				"size": 42,
			},
		}

		hcl := mc.ToHCL()

		// Args are not yet rendered (TODO in implementation)
		assert.Contains(t, hcl, `module "test"`)
		assert.Contains(t, hcl, "}")
	})
}

func TestWithValidation(t *testing.T) {
	t.Run("calls Validate when module implements Validator", func(t *testing.T) {
		expectedErr := assert.AnError
		validator := &MockValidator{
			MockModule:  MockModule{name: "test"},
			validateErr: expectedErr,
		}

		err := WithValidation(validator)

		assert.Equal(t, expectedErr, err)
	})

	t.Run("returns nil when module does not implement Validator", func(t *testing.T) {
		module := &MockModule{name: "test"}

		err := WithValidation(module)

		require.NoError(t, err)
	})

	t.Run("returns nil when validation succeeds", func(t *testing.T) {
		validator := &MockValidator{
			MockModule:  MockModule{name: "test"},
			validateErr: nil,
		}

		err := WithValidation(validator)

		require.NoError(t, err)
	})
}

func TestNewStack(t *testing.T) {
	t.Run("creates empty stack with name", func(t *testing.T) {
		name := "my-stack"
		stack := NewStack(name)

		assert.NotNil(t, stack)
		assert.Equal(t, name, stack.Name)
		assert.Empty(t, stack.Modules)
		assert.NotNil(t, stack.Dependencies)
		assert.Empty(t, stack.Dependencies)
	})

	t.Run("creates stack with empty name", func(t *testing.T) {
		stack := NewStack("")

		assert.NotNil(t, stack)
		assert.Empty(t, stack.Name)
	})
}

func TestStack_AddModule(t *testing.T) {
	t.Run("adds single module", func(t *testing.T) {
		stack := NewStack("test")
		module := &MockModule{name: "module1"}

		stack.AddModule(module)

		assert.Len(t, stack.Modules, 1)
		assert.Equal(t, module, stack.Modules[0])
	})

	t.Run("adds multiple modules", func(t *testing.T) {
		stack := NewStack("test")
		module1 := &MockModule{name: "module1"}
		module2 := &MockModule{name: "module2"}
		module3 := &MockModule{name: "module3"}

		stack.AddModule(module1)
		stack.AddModule(module2)
		stack.AddModule(module3)

		assert.Len(t, stack.Modules, 3)
		assert.Equal(t, module1, stack.Modules[0])
		assert.Equal(t, module2, stack.Modules[1])
		assert.Equal(t, module3, stack.Modules[2])
	})

	t.Run("maintains module order", func(t *testing.T) {
		stack := NewStack("test")
		modules := make([]*MockModule, 10)

		for i := 0; i < 10; i++ {
			modules[i] = &MockModule{name: string(rune('a' + i))}
			stack.AddModule(modules[i])
		}

		for i := 0; i < 10; i++ {
			assert.Equal(t, modules[i], stack.Modules[i])
		}
	})
}

func TestStack_AddDependency(t *testing.T) {
	t.Run("adds single dependency", func(t *testing.T) {
		stack := NewStack("test")

		stack.AddDependency("module2", "module1")

		assert.Len(t, stack.Dependencies["module2"], 1)
		assert.Contains(t, stack.Dependencies["module2"], "module1")
	})

	t.Run("adds multiple dependencies for same module", func(t *testing.T) {
		stack := NewStack("test")

		stack.AddDependency("module3", "module1")
		stack.AddDependency("module3", "module2")

		assert.Len(t, stack.Dependencies["module3"], 2)
		assert.Contains(t, stack.Dependencies["module3"], "module1")
		assert.Contains(t, stack.Dependencies["module3"], "module2")
	})

	t.Run("tracks dependencies for multiple modules", func(t *testing.T) {
		stack := NewStack("test")

		stack.AddDependency("module2", "module1")
		stack.AddDependency("module3", "module1")
		stack.AddDependency("module3", "module2")

		assert.Len(t, stack.Dependencies, 2)
		assert.Len(t, stack.Dependencies["module2"], 1)
		assert.Len(t, stack.Dependencies["module3"], 2)
	})

	t.Run("allows duplicate dependencies", func(t *testing.T) {
		stack := NewStack("test")

		stack.AddDependency("module2", "module1")
		stack.AddDependency("module2", "module1")

		// Implementation allows duplicates (doesn't deduplicate)
		assert.Len(t, stack.Dependencies["module2"], 2)
	})
}

func TestStack_ToHCL(t *testing.T) {
	t.Run("generates empty HCL for empty stack", func(t *testing.T) {
		stack := NewStack("empty")

		hcl, err := stack.ToHCL()

		require.NoError(t, err)
		assert.Empty(t, hcl)
	})

	t.Run("generates HCL for single module", func(t *testing.T) {
		stack := NewStack("test")
		module := &MockModule{
			name:   "test_module",
			config: "# test config",
		}
		stack.AddModule(module)

		hcl, err := stack.ToHCL()

		require.NoError(t, err)
		assert.Contains(t, hcl, "# test config")
	})

	t.Run("generates HCL for multiple modules", func(t *testing.T) {
		stack := NewStack("test")
		module1 := &MockModule{name: "module1", config: "# config 1"}
		module2 := &MockModule{name: "module2", config: "# config 2"}

		stack.AddModule(module1)
		stack.AddModule(module2)

		hcl, err := stack.ToHCL()

		require.NoError(t, err)
		assert.Contains(t, hcl, "# config 1")
		assert.Contains(t, hcl, "# config 2")
	})

	t.Run("returns error when module configuration fails", func(t *testing.T) {
		stack := NewStack("test")
		module := &MockModule{
			name: "bad_module",
			err:  assert.AnError,
		}
		stack.AddModule(module)

		hcl, err := stack.ToHCL()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate config for bad_module")
		assert.Empty(t, hcl)
	})

	t.Run("stops at first error", func(t *testing.T) {
		stack := NewStack("test")
		module1 := &MockModule{name: "module1", err: assert.AnError}
		module2 := &MockModule{name: "module2", config: "# config 2"}

		stack.AddModule(module1)
		stack.AddModule(module2)

		hcl, err := stack.ToHCL()

		assert.Error(t, err)
		assert.Empty(t, hcl)
	})
}

func TestStack_Validate(t *testing.T) {
	t.Run("validates empty stack", func(t *testing.T) {
		stack := NewStack("empty")

		err := stack.Validate()

		require.NoError(t, err)
	})

	t.Run("validates stack with non-validator modules", func(t *testing.T) {
		stack := NewStack("test")
		module := &MockModule{name: "module1"}

		stack.AddModule(module)

		err := stack.Validate()

		require.NoError(t, err)
	})

	t.Run("validates stack with valid validator modules", func(t *testing.T) {
		stack := NewStack("test")
		validator := &MockValidator{
			MockModule:  MockModule{name: "valid"},
			validateErr: nil,
		}

		stack.AddModule(validator)

		err := stack.Validate()

		require.NoError(t, err)
	})

	t.Run("returns error when validation fails", func(t *testing.T) {
		stack := NewStack("test")
		validator := &MockValidator{
			MockModule:  MockModule{name: "invalid"},
			validateErr: assert.AnError,
		}

		stack.AddModule(validator)

		err := stack.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed for invalid")
	})

	t.Run("stops at first validation error", func(t *testing.T) {
		stack := NewStack("test")
		validator1 := &MockValidator{
			MockModule:  MockModule{name: "module1"},
			validateErr: assert.AnError,
		}
		validator2 := &MockValidator{
			MockModule:  MockModule{name: "module2"},
			validateErr: nil,
		}

		stack.AddModule(validator1)
		stack.AddModule(validator2)

		err := stack.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "module1")
	})

	t.Run("validates mixed modules", func(t *testing.T) {
		stack := NewStack("test")
		module := &MockModule{name: "regular"}
		validator := &MockValidator{
			MockModule:  MockModule{name: "validator"},
			validateErr: nil,
		}

		stack.AddModule(module)
		stack.AddModule(validator)

		err := stack.Validate()

		require.NoError(t, err)
	})
}

func TestStack_CompleteWorkflow(t *testing.T) {
	t.Run("creates and validates complete stack", func(t *testing.T) {
		stack := NewStack("serverless-app")

		// Add modules
		queue := &MockModule{name: "orders_queue", config: "# SQS config"}
		table := &MockModule{name: "orders_table", config: "# DynamoDB config"}
		topic := &MockModule{name: "notifications", config: "# SNS config"}

		stack.AddModule(queue)
		stack.AddModule(table)
		stack.AddModule(topic)

		// Add dependencies
		stack.AddDependency("orders_queue", "orders_table")
		stack.AddDependency("notifications", "orders_queue")

		// Validate
		err := stack.Validate()
		require.NoError(t, err)

		// Generate HCL
		hcl, err := stack.ToHCL()
		require.NoError(t, err)
		assert.Contains(t, hcl, "# SQS config")
		assert.Contains(t, hcl, "# DynamoDB config")
		assert.Contains(t, hcl, "# SNS config")

		// Check structure
		assert.Len(t, stack.Modules, 3)
		assert.Len(t, stack.Dependencies, 2)
	})
}

func TestModule_Interface(t *testing.T) {
	t.Run("MockModule implements Module interface", func(t *testing.T) {
		var _ Module = (*MockModule)(nil)
	})

	t.Run("module interface defines required methods", func(t *testing.T) {
		module := &MockModule{name: "test", config: "config"}

		// LocalName method exists
		assert.Equal(t, "test", module.LocalName())

		// Configuration method exists
		config, err := module.Configuration()
		require.NoError(t, err)
		assert.Equal(t, "config", config)
	})
}

func TestValidator_Interface(t *testing.T) {
	t.Run("Validator interface can be implemented", func(t *testing.T) {
		var _ Validator = (*MockValidator)(nil)
	})

	t.Run("WithValidation works with interface", func(t *testing.T) {
		validator := &MockValidator{validateErr: nil}

		err := WithValidation(validator)

		require.NoError(t, err)
	})
}

// BenchmarkNewStack benchmarks stack creation
func BenchmarkNewStack(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewStack("bench-stack")
	}
}

// BenchmarkStackAddModule benchmarks adding modules to a stack
func BenchmarkStackAddModule(b *testing.B) {
	stack := NewStack("bench")
	module := &MockModule{name: "bench-module"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.AddModule(module)
	}
}

// BenchmarkStackAddDependency benchmarks adding dependencies
func BenchmarkStackAddDependency(b *testing.B) {
	stack := NewStack("bench")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stack.AddDependency("module2", "module1")
	}
}

// BenchmarkStackValidate benchmarks stack validation
func BenchmarkStackValidate(b *testing.B) {
	stack := NewStack("bench")
	for i := 0; i < 10; i++ {
		stack.AddModule(&MockModule{name: "module"})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = stack.Validate()
	}
}

// BenchmarkOutputRef benchmarks output reference generation
func BenchmarkOutputRef(b *testing.B) {
	module := &MockModule{name: "bench_module"}
	output := NewOutput(module, "attribute")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = output.Ref()
	}
}
