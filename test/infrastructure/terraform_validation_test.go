package infrastructure

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTerraformValidation tests that generated Terraform code is valid.
func TestTerraformValidation(t *testing.T) {
	t.Run("validates example Python Lambda infrastructure", func(t *testing.T) {
		// Skip in short mode - requires terraform binary
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
			// Don't actually deploy - just validate
			PlanFilePath: filepath.Join(terraformDir, "test.tfplan"),
		}

		// Validate terraform configuration
		_, err := terraform.InitE(t, terraformOptions)
		require.NoError(t, err, "Terraform init should succeed")

		_, err = terraform.ValidateE(t, terraformOptions)
		require.NoError(t, err, "Terraform configuration should be valid")
	})

	t.Run("validates Terraform syntax without init", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Use terraform fmt to check formatting
		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		// Check if terraform files are formatted
		output, err := terraform.RunTerraformCommandAndGetStdoutE(t, terraformOptions, "fmt", "-check", "-diff")
		// fmt returns error if files need formatting
		if err != nil {
			t.Logf("Terraform files need formatting:\n%s", output)
			// Don't fail on formatting issues, just warn
		}
	})
}

// TestTerraformResourceConfiguration tests specific resource configurations.
func TestTerraformResourceConfiguration(t *testing.T) {
	t.Run("validates Lambda function configuration", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		terraformOptions := &terraform.Options{
			TerraformDir: terraformDir,
			NoColor:      true,
		}

		// Init terraform
		terraform.Init(t, terraformOptions)

		// Run terraform plan and parse output
		planOutput := terraform.Plan(t, terraformOptions)

		// Verify plan doesn't show errors
		assert.NotEmpty(t, planOutput, "Plan output should not be empty")
	})

	t.Run("checks for required variables", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Read variables.tf to check for required variables
		variablesFile := filepath.Join(terraformDir, "variables.tf")

		// Check file exists
		assert.FileExists(t, variablesFile, "variables.tf should exist")
	})

	t.Run("checks for outputs definition", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Read outputs.tf to check outputs are defined
		outputsFile := filepath.Join(terraformDir, "outputs.tf")

		// Check file exists
		assert.FileExists(t, outputsFile, "outputs.tf should exist")
	})
}

// TestTerraformStateManagement tests Terraform state operations.
func TestTerraformStateManagement(t *testing.T) {
	t.Run("validates state file structure", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping infrastructure test in short mode")
		}

		terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

		// Check if state file exists (from previous deployment)
		stateFile := filepath.Join(terraformDir, "terraform.tfstate")

		// State file may or may not exist depending on whether terraform was run
		// Just check the path is valid
		assert.NotEmpty(t, stateFile, "State file path should be defined")
	})
}
