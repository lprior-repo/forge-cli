package state

import (
	"strings"
	"testing"
)

// TestGenerateStateBucketName tests bucket name generation (PURE function)
func TestGenerateStateBucketName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		want        string
	}{
		{
			name:        "simple project name",
			projectName: "my-app",
			want:        "forge-state-my-app",
		},
		{
			name:        "project name with underscores",
			projectName: "my_app",
			want:        "forge-state-my-app",
		},
		{
			name:        "uppercase project name",
			projectName: "MyApp",
			want:        "forge-state-myapp",
		},
		{
			name:        "mixed case with special chars",
			projectName: "My_Cool_App",
			want:        "forge-state-my-cool-app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateStateBucketName(tt.projectName)
			if got != tt.want {
				t.Errorf("GenerateStateBucketName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateLockTableName tests DynamoDB table name generation (PURE function)
func TestGenerateLockTableName(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		want        string
	}{
		{
			name:        "simple project name",
			projectName: "my-app",
			want:        "forge_locks_my_app",
		},
		{
			name:        "project name with dashes",
			projectName: "my-cool-app",
			want:        "forge_locks_my_cool_app",
		},
		{
			name:        "uppercase project name",
			projectName: "MyApp",
			want:        "forge_locks_myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateLockTableName(tt.projectName)
			if got != tt.want {
				t.Errorf("GenerateLockTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateStateKey tests state key generation with namespace support (PURE function)
func TestGenerateStateKey(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      string
	}{
		{
			name:      "no namespace",
			namespace: "",
			want:      "terraform.tfstate",
		},
		{
			name:      "with namespace",
			namespace: "pr-123",
			want:      "pr-123/terraform.tfstate",
		},
		{
			name:      "production namespace",
			namespace: "production",
			want:      "production/terraform.tfstate",
		},
		{
			name:      "staging namespace",
			namespace: "staging",
			want:      "staging/terraform.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateStateKey(tt.namespace)
			if got != tt.want {
				t.Errorf("GenerateStateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGenerateS3BucketSpec tests S3 bucket specification generation (PURE function)
func TestGenerateS3BucketSpec(t *testing.T) {
	projectName := "test-app"
	region := "us-west-2"

	spec := GenerateS3BucketSpec(projectName, region)

	// Verify immutable data structure
	if spec.Name != "forge-state-test-app" {
		t.Errorf("Expected bucket name forge-state-test-app, got %s", spec.Name)
	}

	if spec.Region != region {
		t.Errorf("Expected region %s, got %s", region, spec.Region)
	}

	if spec.EnableLogging {
		t.Error("Expected EnableLogging to be false by default")
	}

	// Verify tags
	if spec.Tags["Project"] != projectName {
		t.Errorf("Expected Project tag %s, got %s", projectName, spec.Tags["Project"])
	}

	if spec.Tags["ManagedBy"] != "forge" {
		t.Errorf("Expected ManagedBy tag forge, got %s", spec.Tags["ManagedBy"])
	}

	if spec.Tags["Purpose"] != "terraform-state" {
		t.Errorf("Expected Purpose tag terraform-state, got %s", spec.Tags["Purpose"])
	}
}

// TestGenerateDynamoDBTableSpec tests DynamoDB table specification generation (PURE function)
func TestGenerateDynamoDBTableSpec(t *testing.T) {
	projectName := "test-app"
	region := "us-east-1"

	spec := GenerateDynamoDBTableSpec(projectName, region)

	if spec.Name != "forge_locks_test_app" {
		t.Errorf("Expected table name forge_locks_test_app, got %s", spec.Name)
	}

	if spec.Region != region {
		t.Errorf("Expected region %s, got %s", region, spec.Region)
	}

	if spec.BillingMode != "PAY_PER_REQUEST" {
		t.Errorf("Expected PAY_PER_REQUEST billing mode, got %s", spec.BillingMode)
	}

	if spec.HashKey != "LockID" {
		t.Errorf("Expected HashKey LockID, got %s", spec.HashKey)
	}

	// Verify tags
	if spec.Tags["Project"] != projectName {
		t.Errorf("Expected Project tag %s, got %s", projectName, spec.Tags["Project"])
	}

	if spec.Tags["ManagedBy"] != "forge" {
		t.Errorf("Expected ManagedBy tag forge, got %s", spec.Tags["ManagedBy"])
	}
}

// TestGenerateBackendConfig tests backend configuration generation (PURE function)
func TestGenerateBackendConfig(t *testing.T) {
	tests := []struct {
		name        string
		projectName string
		region      string
		namespace   string
		wantKey     string
	}{
		{
			name:        "default environment",
			projectName: "my-app",
			region:      "us-east-1",
			namespace:   "",
			wantKey:     "terraform.tfstate",
		},
		{
			name:        "PR environment",
			projectName: "my-app",
			region:      "us-east-1",
			namespace:   "pr-456",
			wantKey:     "pr-456/terraform.tfstate",
		},
		{
			name:        "staging environment",
			projectName: "my-app",
			region:      "us-west-2",
			namespace:   "staging",
			wantKey:     "staging/terraform.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GenerateBackendConfig(tt.projectName, tt.region, tt.namespace)

			if config.Bucket != "forge-state-my-app" {
				t.Errorf("Expected bucket forge-state-my-app, got %s", config.Bucket)
			}

			if config.Key != tt.wantKey {
				t.Errorf("Expected key %s, got %s", tt.wantKey, config.Key)
			}

			if config.Region != tt.region {
				t.Errorf("Expected region %s, got %s", tt.region, config.Region)
			}

			if config.DynamoDBTable != "forge_locks_my_app" {
				t.Errorf("Expected table forge_locks_my_app, got %s", config.DynamoDBTable)
			}

			if !config.Encrypt {
				t.Error("Expected Encrypt to be true")
			}

			if !config.EnableLocking {
				t.Error("Expected EnableLocking to be true")
			}
		})
	}
}

// TestGenerateStateResources tests complete state resources generation (PURE function)
func TestGenerateStateResources(t *testing.T) {
	projectName := "production-app"
	region := "eu-west-1"
	namespace := "production"

	resources := GenerateStateResources(projectName, region, namespace)

	// Verify S3 bucket
	if resources.S3Bucket.Name != "forge-state-production-app" {
		t.Errorf("Unexpected S3 bucket name: %s", resources.S3Bucket.Name)
	}

	if resources.S3Bucket.Region != region {
		t.Errorf("Expected region %s, got %s", region, resources.S3Bucket.Region)
	}

	// Verify DynamoDB table
	if resources.DynamoDBTable.Name != "forge_locks_production_app" {
		t.Errorf("Unexpected DynamoDB table name: %s", resources.DynamoDBTable.Name)
	}

	// Verify backend config
	if resources.BackendConfig.Key != "production/terraform.tfstate" {
		t.Errorf("Expected key production/terraform.tfstate, got %s", resources.BackendConfig.Key)
	}
}

// TestRenderBackendTF tests backend.tf rendering (PURE function)
func TestRenderBackendTF(t *testing.T) {
	config := BackendConfig{
		Bucket:        "my-state-bucket",
		Key:           "terraform.tfstate",
		Region:        "us-east-1",
		DynamoDBTable: "my-lock-table",
		Encrypt:       true,
		EnableLocking: true,
	}

	tf := RenderBackendTF(config)

	// Verify essential elements
	if !strings.Contains(tf, "terraform {") {
		t.Error("Expected terraform block")
	}

	if !strings.Contains(tf, "backend \"s3\" {") {
		t.Error("Expected s3 backend block")
	}

	if !strings.Contains(tf, "bucket         = \"my-state-bucket\"") {
		t.Error("Expected bucket configuration")
	}

	if !strings.Contains(tf, "key            = \"terraform.tfstate\"") {
		t.Error("Expected key configuration")
	}

	if !strings.Contains(tf, "region         = \"us-east-1\"") {
		t.Error("Expected region configuration")
	}

	if !strings.Contains(tf, "encrypt        = true") {
		t.Error("Expected encrypt configuration")
	}

	if !strings.Contains(tf, "dynamodb_table = \"my-lock-table\"") {
		t.Error("Expected dynamodb_table configuration")
	}
}

// TestRenderBackendTF_WithoutLocking tests backend.tf without locking
func TestRenderBackendTF_WithoutLocking(t *testing.T) {
	config := BackendConfig{
		Bucket:        "my-state-bucket",
		Key:           "terraform.tfstate",
		Region:        "us-east-1",
		Encrypt:       true,
		EnableLocking: false, // No locking
	}

	tf := RenderBackendTF(config)

	if strings.Contains(tf, "dynamodb_table") {
		t.Error("Should not include dynamodb_table when EnableLocking is false")
	}
}

// TestRenderS3BucketTF tests S3 bucket Terraform generation (PURE function)
func TestRenderS3BucketTF(t *testing.T) {
	spec := S3BucketSpec{
		Name:   "test-bucket",
		Region: "us-west-2",
		Tags: map[string]string{
			"Environment": "production",
			"ManagedBy":   "forge",
		},
	}

	tf := RenderS3BucketTF(spec)

	// Verify essential resources
	if !strings.Contains(tf, "resource \"aws_s3_bucket\" \"terraform_state\"") {
		t.Error("Expected S3 bucket resource")
	}

	if !strings.Contains(tf, "bucket = \"test-bucket\"") {
		t.Error("Expected bucket name")
	}

	if !strings.Contains(tf, "Environment = \"production\"") {
		t.Error("Expected Environment tag")
	}

	if !strings.Contains(tf, "ManagedBy   = \"forge\"") {
		t.Error("Expected ManagedBy tag")
	}

	// Verify versioning
	if !strings.Contains(tf, "resource \"aws_s3_bucket_versioning\" \"terraform_state\"") {
		t.Error("Expected versioning resource")
	}

	if !strings.Contains(tf, "status = \"Enabled\"") {
		t.Error("Expected versioning enabled")
	}

	// Verify encryption
	if !strings.Contains(tf, "resource \"aws_s3_bucket_server_side_encryption_configuration\"") {
		t.Error("Expected encryption resource")
	}

	if !strings.Contains(tf, "sse_algorithm = \"AES256\"") {
		t.Error("Expected AES256 encryption")
	}

	// Verify public access block
	if !strings.Contains(tf, "resource \"aws_s3_bucket_public_access_block\"") {
		t.Error("Expected public access block resource")
	}

	if !strings.Contains(tf, "block_public_acls       = true") {
		t.Error("Expected block_public_acls = true")
	}
}

// TestRenderDynamoDBTableTF tests DynamoDB table Terraform generation (PURE function)
func TestRenderDynamoDBTableTF(t *testing.T) {
	spec := DynamoDBTableSpec{
		Name:        "test-locks",
		Region:      "us-east-1",
		BillingMode: "PAY_PER_REQUEST",
		HashKey:     "LockID",
		Tags: map[string]string{
			"Project": "test-app",
		},
	}

	tf := RenderDynamoDBTableTF(spec)

	// Verify essential elements
	if !strings.Contains(tf, "resource \"aws_dynamodb_table\" \"terraform_locks\"") {
		t.Error("Expected DynamoDB table resource")
	}

	if !strings.Contains(tf, "name         = \"test-locks\"") {
		t.Error("Expected table name")
	}

	if !strings.Contains(tf, "billing_mode = \"PAY_PER_REQUEST\"") {
		t.Error("Expected PAY_PER_REQUEST billing mode")
	}

	if !strings.Contains(tf, "hash_key     = \"LockID\"") {
		t.Error("Expected LockID hash key")
	}

	if !strings.Contains(tf, "attribute {") {
		t.Error("Expected attribute block")
	}

	if !strings.Contains(tf, "name = \"LockID\"") {
		t.Error("Expected LockID attribute")
	}

	if !strings.Contains(tf, "type = \"S\"") {
		t.Error("Expected string type attribute")
	}

	if !strings.Contains(tf, "Project = \"test-app\"") {
		t.Error("Expected Project tag")
	}
}

// TestRenderStateBootstrapTF tests complete bootstrap Terraform generation (PURE function)
func TestRenderStateBootstrapTF(t *testing.T) {
	resources := GenerateStateResources("test-app", "us-east-1", "")

	tf := RenderStateBootstrapTF(resources)

	// Verify provider
	if !strings.Contains(tf, "provider \"aws\" {") {
		t.Error("Expected AWS provider")
	}

	if !strings.Contains(tf, "region = \"us-east-1\"") {
		t.Error("Expected region configuration")
	}

	// Verify S3 bucket
	if !strings.Contains(tf, "resource \"aws_s3_bucket\" \"terraform_state\"") {
		t.Error("Expected S3 bucket resource")
	}

	// Verify DynamoDB table
	if !strings.Contains(tf, "resource \"aws_dynamodb_table\" \"terraform_locks\"") {
		t.Error("Expected DynamoDB table resource")
	}

	// Verify comments
	if !strings.Contains(tf, "Auto-generated by Forge") {
		t.Error("Expected auto-generated comment")
	}
}

// TestPureFunctionIdempotency tests that pure functions are idempotent
func TestPureFunctionIdempotency(t *testing.T) {
	projectName := "idempotency-test"
	region := "eu-central-1"
	namespace := "test"

	// Call function multiple times
	result1 := GenerateStateResources(projectName, region, namespace)
	result2 := GenerateStateResources(projectName, region, namespace)
	result3 := GenerateStateResources(projectName, region, namespace)

	// Verify all results are identical (same inputs → same outputs)
	if result1.S3Bucket.Name != result2.S3Bucket.Name || result2.S3Bucket.Name != result3.S3Bucket.Name {
		t.Error("Pure function should return identical results for same inputs")
	}

	if result1.BackendConfig.Key != result2.BackendConfig.Key || result2.BackendConfig.Key != result3.BackendConfig.Key {
		t.Error("Pure function should return identical results for same inputs")
	}

	if result1.DynamoDBTable.Name != result2.DynamoDBTable.Name || result2.DynamoDBTable.Name != result3.DynamoDBTable.Name {
		t.Error("Pure function should return identical results for same inputs")
	}
}

// TestRenderBackendTF_NamespaceAware tests namespace-aware state keys
func TestRenderBackendTF_NamespaceAware(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantKey   string
	}{
		{
			name:      "default environment",
			namespace: "",
			wantKey:   "terraform.tfstate",
		},
		{
			name:      "PR environment",
			namespace: "pr-789",
			wantKey:   "pr-789/terraform.tfstate",
		},
		{
			name:      "feature branch",
			namespace: "feature-auth",
			wantKey:   "feature-auth/terraform.tfstate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GenerateBackendConfig("test-app", "us-east-1", tt.namespace)
			tf := RenderBackendTF(config)

			expectedKey := fmt.Sprintf("key            = \"%s\"", tt.wantKey)
			if !strings.Contains(tf, expectedKey) {
				t.Errorf("Expected key %s in rendered TF, got:\n%s", expectedKey, tf)
			}
		})
	}
}
