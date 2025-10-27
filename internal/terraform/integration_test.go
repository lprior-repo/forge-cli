//go:build integration
// +build integration

package terraform

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationNewExecutor tests creating a real executor
func TestIntegrationNewExecutor(t *testing.T) {
	t.Run("creates executor with terraform binary", func(t *testing.T) {
		exec := NewExecutor("terraform")

		assert.NotNil(t, exec.Init)
		assert.NotNil(t, exec.Plan)
		assert.NotNil(t, exec.Apply)
		assert.NotNil(t, exec.Destroy)
		assert.NotNil(t, exec.Output)
		assert.NotNil(t, exec.Validate)
	})
}

// TestIntegrationTerraformInit tests real terraform init
func TestIntegrationTerraformInit(t *testing.T) {
	// Create temporary directory with terraform config
	tmpDir := t.TempDir()

	// Write minimal terraform config
	tfConfig := `
terraform {
  required_version = ">= 1.0"
}

resource "null_resource" "test" {
  triggers = {
    timestamp = timestamp()
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	t.Run("init succeeds with valid config", func(t *testing.T) {
		err := exec.Init(ctx, tmpDir)
		assert.NoError(t, err, "Terraform init should succeed")

		// Verify .terraform directory was created
		terraformDir := filepath.Join(tmpDir, ".terraform")
		_, err = os.Stat(terraformDir)
		assert.NoError(t, err, ".terraform directory should exist")
	})

	t.Run("init with upgrade flag", func(t *testing.T) {
		err := exec.Init(ctx, tmpDir, Upgrade(true))
		assert.NoError(t, err, "Init with upgrade should succeed")
	})

	t.Run("init with backend=false", func(t *testing.T) {
		err := exec.Init(ctx, tmpDir, Backend(false))
		assert.NoError(t, err, "Init with backend=false should succeed")
	})
}

// TestIntegrationTerraformValidate tests real terraform validate
func TestIntegrationTerraformValidate(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("validate succeeds with valid config", func(t *testing.T) {
		tfConfig := `
resource "null_resource" "test" {
  triggers = {
    value = "test"
  }
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
		require.NoError(t, err)

		exec := NewExecutor("terraform")
		ctx := context.Background()

		// Init first
		err = exec.Init(ctx, tmpDir)
		require.NoError(t, err)

		// Validate
		err = exec.Validate(ctx, tmpDir)
		assert.NoError(t, err, "Validate should succeed with valid config")
	})

}

// TestIntegrationTerraformPlan tests real terraform plan
func TestIntegrationTerraformPlan(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
resource "null_resource" "test" {
  triggers = {
    timestamp = "static"
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	// Init first
	err = exec.Init(ctx, tmpDir)
	require.NoError(t, err)

	t.Run("plan detects changes on fresh config", func(t *testing.T) {
		hasChanges, err := exec.Plan(ctx, tmpDir)
		assert.NoError(t, err, "Plan should succeed")
		assert.True(t, hasChanges, "Plan should detect changes on fresh config")
	})

	t.Run("plan with output file", func(t *testing.T) {
		planFile := filepath.Join(tmpDir, "plan.tfplan")
		hasChanges, err := exec.Plan(ctx, tmpDir, PlanOut(planFile))
		assert.NoError(t, err, "Plan with output should succeed")
		assert.True(t, hasChanges, "Should detect changes")

		// Verify plan file was created
		_, err = os.Stat(planFile)
		assert.NoError(t, err, "Plan file should exist")
	})
}

// TestIntegrationTerraformApply tests real terraform apply
func TestIntegrationTerraformApply(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
resource "null_resource" "test" {
  triggers = {
    timestamp = "static-value"
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	// Init first
	err = exec.Init(ctx, tmpDir)
	require.NoError(t, err)

	t.Run("apply creates resources", func(t *testing.T) {
		err := exec.Apply(ctx, tmpDir, AutoApprove(true))
		assert.NoError(t, err, "Apply should succeed")

		// Verify state file was created
		stateFile := filepath.Join(tmpDir, "terraform.tfstate")
		_, err = os.Stat(stateFile)
		assert.NoError(t, err, "State file should exist after apply")
	})

	t.Run("plan shows no changes after apply", func(t *testing.T) {
		hasChanges, err := exec.Plan(ctx, tmpDir)
		assert.NoError(t, err, "Plan should succeed")
		assert.False(t, hasChanges, "Plan should show no changes after successful apply")
	})
}

// TestIntegrationTerraformOutput tests real terraform output
func TestIntegrationTerraformOutput(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
resource "null_resource" "test" {
  triggers = {
    value = "test"
  }
}

output "test_output" {
  value = "hello-world"
}

output "number_output" {
  value = 42
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	// Init and apply first
	err = exec.Init(ctx, tmpDir)
	require.NoError(t, err)

	err = exec.Apply(ctx, tmpDir, AutoApprove(true))
	require.NoError(t, err)

	t.Run("output returns all outputs", func(t *testing.T) {
		outputs, err := exec.Output(ctx, tmpDir)
		assert.NoError(t, err, "Output should succeed")
		assert.Contains(t, outputs, "test_output")
		assert.Contains(t, outputs, "number_output")
		assert.Equal(t, "hello-world", outputs["test_output"])
		assert.Equal(t, float64(42), outputs["number_output"])
	})
}

// TestIntegrationTerraformDestroy tests real terraform destroy
func TestIntegrationTerraformDestroy(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
resource "null_resource" "test" {
  triggers = {
    value = "test"
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	// Init and apply first
	err = exec.Init(ctx, tmpDir)
	require.NoError(t, err)

	err = exec.Apply(ctx, tmpDir, AutoApprove(true))
	require.NoError(t, err)

	stateFile := filepath.Join(tmpDir, "terraform.tfstate")

	t.Run("destroy removes resources", func(t *testing.T) {
		err := exec.Destroy(ctx, tmpDir, DestroyAutoApprove(true))
		assert.NoError(t, err, "Destroy should succeed")

		// Verify state file shows no resources
		stateData, err := os.ReadFile(stateFile)
		require.NoError(t, err)
		assert.Contains(t, string(stateData), `"resources": []`, "State should show no resources after destroy")
	})
}

// TestIntegrationTerraformWorkflow tests complete workflow
func TestIntegrationTerraformWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
variable "instance_name" {
  type    = string
  default = "test-instance"
}

resource "null_resource" "instance" {
  triggers = {
    name = var.instance_name
  }
}

output "instance_name" {
  value = var.instance_name
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	t.Run("complete workflow: init -> plan -> apply -> output -> destroy", func(t *testing.T) {
		// 1. Init
		err := exec.Init(ctx, tmpDir)
		require.NoError(t, err, "Init should succeed")

		// 2. Validate
		err = exec.Validate(ctx, tmpDir)
		require.NoError(t, err, "Validate should succeed")

		// 3. Plan
		hasChanges, err := exec.Plan(ctx, tmpDir)
		require.NoError(t, err, "Plan should succeed")
		require.True(t, hasChanges, "Should have changes")

		// 4. Apply
		err = exec.Apply(ctx, tmpDir, AutoApprove(true))
		require.NoError(t, err, "Apply should succeed")

		// 5. Output
		outputs, err := exec.Output(ctx, tmpDir)
		require.NoError(t, err, "Output should succeed")
		assert.Equal(t, "test-instance", outputs["instance_name"])

		// 6. Plan again (should be no changes)
		hasChanges, err = exec.Plan(ctx, tmpDir)
		require.NoError(t, err, "Second plan should succeed")
		assert.False(t, hasChanges, "Should have no changes after apply")

		// 7. Destroy
		err = exec.Destroy(ctx, tmpDir, DestroyAutoApprove(true))
		require.NoError(t, err, "Destroy should succeed")
	})
}

// TestIntegrationTerraformWithVariables tests using variables
func TestIntegrationTerraformWithVariables(t *testing.T) {
	tmpDir := t.TempDir()

	tfConfig := `
variable "test_var" {
  type = string
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

	// Create tfvars file
	tfvarsContent := `test_var = "integration-test"`
	tfvarsFile := filepath.Join(tmpDir, "test.tfvars")
	err = os.WriteFile(tfvarsFile, []byte(tfvarsContent), 0644)
	require.NoError(t, err)

	exec := NewExecutor("terraform")
	ctx := context.Background()

	t.Run("workflow with var file", func(t *testing.T) {
		// Init
		err := exec.Init(ctx, tmpDir)
		require.NoError(t, err)

		// Plan with var file
		hasChanges, err := exec.Plan(ctx, tmpDir, PlanVarFile(tfvarsFile))
		require.NoError(t, err)
		assert.True(t, hasChanges)

		// Apply with var file
		err = exec.Apply(ctx, tmpDir, AutoApprove(true), ApplyVarFile(tfvarsFile))
		require.NoError(t, err)

		// Check output
		outputs, err := exec.Output(ctx, tmpDir)
		require.NoError(t, err)
		assert.Equal(t, "integration-test", outputs["test_output"])

		// Destroy with var file
		err = exec.Destroy(ctx, tmpDir, DestroyAutoApprove(true), DestroyVarFile(tfvarsFile))
		require.NoError(t, err)
	})
}

// BenchmarkTerraformInit benchmarks terraform init performance
func BenchmarkTerraformInit(b *testing.B) {
	tmpDir := b.TempDir()

	tfConfig := `
resource "null_resource" "test" {
  triggers = {
    value = "test"
  }
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(tfConfig), 0644)
	if err != nil {
		b.Fatal(err)
	}

	exec := NewExecutor("terraform")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		exec.Init(ctx, tmpDir)
	}
}
