package state

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	E "github.com/IBM/fp-go/either"
	"github.com/lewis/forge/internal/terraform"
)

// ProvisionResult represents the result of state backend provisioning (immutable data)
type ProvisionResult struct {
	BucketName     string
	TableName      string
	BackendTFPath  string
	BootstrapApplied bool
}

// WriteStateBootstrap writes bootstrap Terraform files to disk (ACTION - I/O)
// Creates a temporary .forge/bootstrap/ directory with Terraform code
func WriteStateBootstrap(projectDir string, resources StateResources) (string, error) {
	bootstrapDir := filepath.Join(projectDir, ".forge", "bootstrap")

	// Create bootstrap directory
	if err := os.MkdirAll(bootstrapDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bootstrap directory: %w", err)
	}

	// Write bootstrap.tf with S3 + DynamoDB resources
	bootstrapTF := RenderStateBootstrapTF(resources)
	bootstrapPath := filepath.Join(bootstrapDir, "bootstrap.tf")

	if err := os.WriteFile(bootstrapPath, []byte(bootstrapTF), 0644); err != nil {
		return "", fmt.Errorf("failed to write bootstrap.tf: %w", err)
	}

	return bootstrapDir, nil
}

// WriteBackendConfig writes backend.tf to infra/ directory (ACTION - I/O)
func WriteBackendConfig(projectDir string, config BackendConfig) (string, error) {
	infraDir := filepath.Join(projectDir, "infra")

	// Ensure infra directory exists
	if err := os.MkdirAll(infraDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create infra directory: %w", err)
	}

	// Write backend.tf
	backendTF := RenderBackendTF(config)
	backendPath := filepath.Join(infraDir, "backend.tf")

	if err := os.WriteFile(backendPath, []byte(backendTF), 0644); err != nil {
		return "", fmt.Errorf("failed to write backend.tf: %w", err)
	}

	return backendPath, nil
}

// ApplyBootstrap applies the bootstrap Terraform to provision S3 + DynamoDB (ACTION - I/O)
// This uses Terraform to create the state backend resources
func ApplyBootstrap(ctx context.Context, bootstrapDir string, exec terraform.Executor) error {
	// Initialize Terraform
	if err := exec.Init(ctx, bootstrapDir, terraform.Upgrade(false)); err != nil {
		return fmt.Errorf("terraform init failed: %w", err)
	}

	// Plan
	hasChanges, err := exec.Plan(ctx, bootstrapDir, terraform.PlanOut(filepath.Join(bootstrapDir, "tfplan")))
	if err != nil {
		return fmt.Errorf("terraform plan failed: %w", err)
	}

	if !hasChanges {
		// Resources already exist
		return nil
	}

	// Apply
	if err := exec.Apply(ctx, bootstrapDir,
		terraform.ApplyPlanFile(filepath.Join(bootstrapDir, "tfplan")),
		terraform.AutoApprove(true),
	); err != nil {
		return fmt.Errorf("terraform apply failed: %w", err)
	}

	return nil
}

// CleanupBootstrap removes the temporary bootstrap directory (ACTION - I/O)
func CleanupBootstrap(bootstrapDir string) error {
	return os.RemoveAll(bootstrapDir)
}

// ProvisionStateBackend provisions complete state backend (COMPOSITION - pure + impure)
// This is the main entry point that orchestrates state backend provisioning
// Uses Railway-Oriented Programming with Either monad
func ProvisionStateBackend(
	ctx context.Context,
	projectDir string,
	projectName string,
	region string,
	exec terraform.Executor,
) E.Either[error, ProvisionResult] {
	// PURE: Generate state resources specification
	resources := GenerateStateResources(projectName, region, "")

	// IMPURE: Write bootstrap Terraform files
	bootstrapDir, err := WriteStateBootstrap(projectDir, resources)
	if err != nil {
		return E.Left[ProvisionResult](err)
	}

	// IMPURE: Apply bootstrap to provision S3 + DynamoDB
	if err := ApplyBootstrap(ctx, bootstrapDir, exec); err != nil {
		return E.Left[ProvisionResult](fmt.Errorf("failed to provision state backend: %w", err))
	}

	// IMPURE: Write backend.tf to infra/
	backendPath, err := WriteBackendConfig(projectDir, resources.BackendConfig)
	if err != nil {
		return E.Left[ProvisionResult](err)
	}

	// IMPURE: Cleanup bootstrap directory (optional - can keep for debugging)
	// Commenting out for now so users can inspect if needed
	// _ = CleanupBootstrap(bootstrapDir)

	// SUCCESS: Return immutable result
	result := ProvisionResult{
		BucketName:       resources.S3Bucket.Name,
		TableName:        resources.DynamoDBTable.Name,
		BackendTFPath:    backendPath,
		BootstrapApplied: true,
	}

	return E.Right[error](result)
}

// ProvisionStateBackendSync is a synchronous wrapper for ProvisionStateBackend
// Returns (result, error) for easier use in non-functional contexts
func ProvisionStateBackendSync(
	ctx context.Context,
	projectDir string,
	projectName string,
	region string,
	exec terraform.Executor,
) (ProvisionResult, error) {
	result := ProvisionStateBackend(ctx, projectDir, projectName, region, exec)

	return E.Fold(
		func(err error) (ProvisionResult, error) {
			return ProvisionResult{}, err
		},
		func(res ProvisionResult) (ProvisionResult, error) {
			return res, nil
		},
	)(result)
}
