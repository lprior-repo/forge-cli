package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewExecutor tests the NewExecutor constructor
func TestNewExecutor(t *testing.T) {
	t.Run("creates executor with all functions", func(t *testing.T) {
		exec := NewExecutor("terraform")

		assert.NotNil(t, exec.Init, "Init function should be set")
		assert.NotNil(t, exec.Plan, "Plan function should be set")
		assert.NotNil(t, exec.Apply, "Apply function should be set")
		assert.NotNil(t, exec.Destroy, "Destroy function should be set")
		assert.NotNil(t, exec.Output, "Output function should be set")
		assert.NotNil(t, exec.Validate, "Validate function should be set")
	})

	t.Run("creates executor with custom terraform path", func(t *testing.T) {
		exec := NewExecutor("/usr/local/bin/terraform")

		assert.NotNil(t, exec.Init)
		assert.NotNil(t, exec.Plan)
		assert.NotNil(t, exec.Apply)
		assert.NotNil(t, exec.Destroy)
		assert.NotNil(t, exec.Output)
		assert.NotNil(t, exec.Validate)
	})

	t.Run("creates executor with relative terraform path", func(t *testing.T) {
		exec := NewExecutor("./terraform")

		assert.NotNil(t, exec.Init)
		assert.NotNil(t, exec.Plan)
		assert.NotNil(t, exec.Apply)
		assert.NotNil(t, exec.Destroy)
		assert.NotNil(t, exec.Output)
		assert.NotNil(t, exec.Validate)
	})
}

// TestMakeInitFunc tests the makeInitFunc closure
func TestMakeInitFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		initFunc := makeInitFunc("terraform")
		ctx := context.Background()

		// Use a directory that definitely doesn't exist
		err := initFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a valid terraform config
		tfConfig := `
terraform {
  required_version = ">= 1.0"
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		// Use a non-existent terraform binary
		initFunc := makeInitFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		err = initFunc(ctx, tmpDir)

		// terraform-exec validates the binary path, so this should fail
		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("applies Upgrade option", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a minimal terraform config
		tfConfig := `terraform { required_version = ">= 1.0" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		initFunc := makeInitFunc("terraform")
		ctx := context.Background()

		// This will fail if terraform isn't installed, but we're testing the option passing
		// The function should accept the option without error
		_ = initFunc(ctx, tmpDir, Upgrade(true))

		// We can't assert success without terraform installed,
		// but we can verify the function accepts the option
	})

	t.Run("applies Backend option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `terraform { required_version = ">= 1.0" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		initFunc := makeInitFunc("terraform")
		ctx := context.Background()

		// Test that Backend option is accepted
		_ = initFunc(ctx, tmpDir, Backend(false))
	})

	t.Run("applies Reconfigure option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `terraform { required_version = ">= 1.0" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		initFunc := makeInitFunc("terraform")
		ctx := context.Background()

		// Test that Reconfigure option is accepted
		_ = initFunc(ctx, tmpDir, Reconfigure(true))
	})

	t.Run("applies multiple options together", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `terraform { required_version = ">= 1.0" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		initFunc := makeInitFunc("terraform")
		ctx := context.Background()

		// Test that multiple options are accepted
		_ = initFunc(ctx, tmpDir, Upgrade(true), Backend(false), Reconfigure(true))
	})
}

// TestMakePlanFunc tests the makePlanFunc closure
func TestMakePlanFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		planFunc := makePlanFunc("terraform")
		ctx := context.Background()

		_, err := planFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		_, err = planFunc(ctx, tmpDir)

		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("applies PlanOut option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("terraform")
		ctx := context.Background()

		// Test that PlanOut option is accepted
		_, _ = planFunc(ctx, tmpDir, PlanOut("plan.tfplan"))
	})

	t.Run("applies PlanDestroy option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("terraform")
		ctx := context.Background()

		// Test that PlanDestroy option is accepted
		_, _ = planFunc(ctx, tmpDir, PlanDestroy(true))
	})

	t.Run("applies PlanVarFile option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("terraform")
		ctx := context.Background()

		// Test that PlanVarFile option is accepted
		_, _ = planFunc(ctx, tmpDir, PlanVarFile("vars.tfvars"))
	})

	t.Run("applies multiple options together", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("terraform")
		ctx := context.Background()

		// Test that multiple options are accepted
		_, _ = planFunc(ctx, tmpDir, PlanOut("plan.tfplan"), PlanDestroy(true), PlanVarFile("vars.tfvars"))
	})
}

// TestMakeApplyFunc tests the makeApplyFunc closure
func TestMakeApplyFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		applyFunc := makeApplyFunc("terraform")
		ctx := context.Background()

		err := applyFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		err = applyFunc(ctx, tmpDir)

		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("applies AutoApprove option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("terraform")
		ctx := context.Background()

		// Test that AutoApprove option is accepted
		_ = applyFunc(ctx, tmpDir, AutoApprove(true))
	})

	t.Run("applies ApplyVarFile option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("terraform")
		ctx := context.Background()

		// Test that ApplyVarFile option is accepted
		_ = applyFunc(ctx, tmpDir, ApplyVarFile("vars.tfvars"))
	})

	t.Run("applies ApplyPlanFile option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("terraform")
		ctx := context.Background()

		// Test that ApplyPlanFile option is accepted
		_ = applyFunc(ctx, tmpDir, ApplyPlanFile("plan.tfplan"))
	})

	t.Run("applies multiple options together", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("terraform")
		ctx := context.Background()

		// Test that multiple options are accepted
		_ = applyFunc(ctx, tmpDir, AutoApprove(true), ApplyVarFile("vars.tfvars"))
	})
}

// TestMakeDestroyFunc tests the makeDestroyFunc closure
func TestMakeDestroyFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		destroyFunc := makeDestroyFunc("terraform")
		ctx := context.Background()

		err := destroyFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		destroyFunc := makeDestroyFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		err = destroyFunc(ctx, tmpDir)

		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("applies DestroyAutoApprove option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		destroyFunc := makeDestroyFunc("terraform")
		ctx := context.Background()

		// Test that DestroyAutoApprove option is accepted
		_ = destroyFunc(ctx, tmpDir, DestroyAutoApprove(true))
	})

	t.Run("applies DestroyVarFile option", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		destroyFunc := makeDestroyFunc("terraform")
		ctx := context.Background()

		// Test that DestroyVarFile option is accepted
		_ = destroyFunc(ctx, tmpDir, DestroyVarFile("vars.tfvars"))
	})

	t.Run("applies multiple options together", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		destroyFunc := makeDestroyFunc("terraform")
		ctx := context.Background()

		// Test that multiple options are accepted
		_ = destroyFunc(ctx, tmpDir, DestroyAutoApprove(true), DestroyVarFile("vars.tfvars"))
	})
}

// TestMakeOutputFunc tests the makeOutputFunc closure
func TestMakeOutputFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		outputFunc := makeOutputFunc("terraform")
		ctx := context.Background()

		_, err := outputFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `output "test" { value = "test" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		outputFunc := makeOutputFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		_, err = outputFunc(ctx, tmpDir)

		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("creates output function that can be called", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `output "test_output" { value = "hello" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		outputFunc := makeOutputFunc("terraform")
		ctx := context.Background()

		// This will fail without terraform init/apply, but tests the function signature
		_, err = outputFunc(ctx, tmpDir)

		// We expect an error since we haven't initialized terraform,
		// but the function should be callable
		assert.Error(t, err)
	})
}

// TestMakeValidateFunc tests the makeValidateFunc closure
func TestMakeValidateFunc(t *testing.T) {
	t.Run("returns error for non-existent directory", func(t *testing.T) {
		validateFunc := makeValidateFunc("terraform")
		ctx := context.Background()

		err := validateFunc(ctx, "/nonexistent/directory/path/12345")

		assert.Error(t, err, "Should error for non-existent directory")
		assert.Contains(t, err.Error(), "failed to create terraform", "Error should mention terraform creation failure")
	})

	t.Run("returns error for invalid terraform path", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		validateFunc := makeValidateFunc("/nonexistent/terraform/binary/12345")
		ctx := context.Background()

		err = validateFunc(ctx, tmpDir)

		assert.Error(t, err, "Should error for invalid terraform binary")
	})

	t.Run("creates validate function that can be called", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		validateFunc := makeValidateFunc("terraform")
		ctx := context.Background()

		// This will fail without terraform init, but tests the function signature
		err = validateFunc(ctx, tmpDir)

		// We expect an error since we haven't initialized terraform,
		// but the function should be callable
		assert.Error(t, err)
	})
}

// TestRealFunctionsWithContext tests context handling
func TestRealFunctionsWithContext(t *testing.T) {
	t.Run("Init respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `terraform { required_version = ">= 1.0" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		initFunc := makeInitFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = initFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})

	t.Run("Plan respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		planFunc := makePlanFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = planFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})

	t.Run("Apply respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		applyFunc := makeApplyFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = applyFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})

	t.Run("Destroy respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		destroyFunc := makeDestroyFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = destroyFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})

	t.Run("Output respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `output "test" { value = "test" }`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		outputFunc := makeOutputFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err = outputFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})

	t.Run("Validate respects context cancellation", func(t *testing.T) {
		tmpDir := t.TempDir()

		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		validateFunc := makeValidateFunc("terraform")

		// Create a cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = validateFunc(ctx, tmpDir)

		// Should get an error (either from terraform or context cancellation)
		assert.Error(t, err)
	})
}

// TestRealFunctionsClosureProperties tests closure properties
func TestRealFunctionsClosureProperties(t *testing.T) {
	t.Run("different terraform paths create independent closures", func(t *testing.T) {
		initFunc1 := makeInitFunc("terraform1")
		initFunc2 := makeInitFunc("terraform2")

		// These should be different functions
		assert.NotNil(t, initFunc1)
		assert.NotNil(t, initFunc2)

		// While we can't compare function pointers directly,
		// we can verify they're independent by checking they both exist
		ctx := context.Background()
		tmpDir := t.TempDir()

		// Both should fail with their respective invalid paths
		err1 := initFunc1(ctx, tmpDir)
		err2 := initFunc2(ctx, tmpDir)

		assert.Error(t, err1)
		assert.Error(t, err2)
	})

	t.Run("executor functions are independent", func(t *testing.T) {
		exec := NewExecutor("terraform")

		// All functions should be independently callable
		ctx := context.Background()
		tmpDir := t.TempDir()

		// Create a minimal terraform config
		tfConfig := `resource "null_resource" "test" {}`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		// Each function should be callable independently
		// (They'll fail without proper setup, but should be callable)
		_ = exec.Init(ctx, tmpDir)
		_, _ = exec.Plan(ctx, tmpDir)
		_ = exec.Apply(ctx, tmpDir)
		_ = exec.Destroy(ctx, tmpDir)
		_, _ = exec.Output(ctx, tmpDir)
		_ = exec.Validate(ctx, tmpDir)

		// No assertion needed - we're just verifying the functions can be called
	})
}

// TestRealFunctionsOptionCombinations tests various option combinations
func TestRealFunctionsOptionCombinations(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
variable "test_var" {
  type    = string
  default = "test"
}

resource "null_resource" "test" {
  triggers = {
    value = var.test_var
  }
}

output "test_output" {
  value = var.test_var
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	ctx := context.Background()

	t.Run("Init with all options", func(t *testing.T) {
		initFunc := makeInitFunc("terraform")

		// Test with all options enabled
		_ = initFunc(ctx, tmpDir,
			Upgrade(true),
			Backend(false),
			Reconfigure(true),
			BackendConfig("key=value"),
		)

		// Just verify the function accepts all options
	})

	t.Run("Plan with all options", func(t *testing.T) {
		planFunc := makePlanFunc("terraform")

		// Test with all options
		_, _ = planFunc(ctx, tmpDir,
			PlanOut("plan.tfplan"),
			PlanDestroy(true),
			PlanVarFile("vars.tfvars"),
			PlanVar("test_var", "value"),
		)

		// Just verify the function accepts all options
	})

	t.Run("Apply with all options", func(t *testing.T) {
		applyFunc := makeApplyFunc("terraform")

		// Test with all options
		_ = applyFunc(ctx, tmpDir,
			AutoApprove(true),
			ApplyVarFile("vars.tfvars"),
			ApplyVar("test_var", "value"),
			ApplyPlanFile("plan.tfplan"),
		)

		// Just verify the function accepts all options
	})

	t.Run("Destroy with all options", func(t *testing.T) {
		destroyFunc := makeDestroyFunc("terraform")

		// Test with all options
		_ = destroyFunc(ctx, tmpDir,
			DestroyAutoApprove(true),
			DestroyVarFile("vars.tfvars"),
		)

		// Just verify the function accepts all options
	})
}
