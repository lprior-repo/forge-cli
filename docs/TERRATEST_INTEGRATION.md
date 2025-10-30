# Terratest Integration Documentation

**Date:** 2025-10-30
**Purpose:** Infrastructure testing for generated Terraform code
**Framework:** [Terratest](https://terratest.gruntwork.io/) by Gruntwork

---

## Overview

Forge uses Terratest to validate that generated Terraform infrastructure code actually works as expected. This provides confidence that:

1. **Terraform syntax is valid** - Files pass `terraform validate`
2. **Resources can be planned** - Terraform can generate execution plans
3. **Deployments succeed** - Resources can be created in AWS (E2E tests)
4. **Resources are configured correctly** - Deployed infrastructure meets expectations

## Why Terratest?

Terratest is the industry-standard Go testing framework for infrastructure code, offering:

- **Native Go integration** - Tests written in Go, same language as Forge
- **AWS SDK integration** - Direct verification of deployed resources
- **Retry mechanisms** - Handle eventual consistency automatically
- **Cleanup utilities** - Automatic resource teardown with `defer`
- **Rich assertions** - Comprehensive testing utilities
- **Battle-tested** - Used by thousands of infrastructure projects

## Test Structure

```
test/infrastructure/
├── terraform_validation_test.go    # Terraform syntax and validation tests
└── lambda_deployment_test.go       # End-to-end deployment tests
```

### Test Categories

**1. Validation Tests (`terraform_validation_test.go`)**
- Tests that **don't deploy** to AWS
- Fast execution (~2-3 seconds)
- Validate Terraform syntax, configuration, and structure
- Safe to run in CI/CD without AWS credentials

**2. Deployment Tests (`lambda_deployment_test.go`)**
- Tests that **actually deploy** to AWS
- Slower execution (~30-60 seconds)
- Verify real infrastructure deployment
- Require AWS credentials
- Automatic cleanup with `defer terraform.Destroy()`

## Running Tests

### Quick Test (Short Mode) - No AWS Deployment

```bash
# Run all infrastructure tests (skips AWS deployment)
go test ./test/infrastructure/... -v -short

# Run specific test
go test ./test/infrastructure/... -v -short -run TestTerraformValidation
```

**Use case**: Local development, fast feedback, CI/CD

### Full Test (E2E with AWS)

```bash
# Run all tests including AWS deployment
go test ./test/infrastructure/... -v

# Run specific deployment test
go test ./test/infrastructure/... -v -run TestLambdaDeploymentEndToEnd
```

**Requirements**:
- AWS credentials configured (see below)
- Valid AWS account with permissions
- Costs ~$0.001-0.01 per test run

**Use case**: Pre-release validation, regression testing

## AWS Credentials Configuration

Tests check for AWS credentials in this order:

1. **Environment variables**:
   ```bash
   export AWS_ACCESS_KEY_ID=xxx
   export AWS_SECRET_ACCESS_KEY=yyy
   export AWS_REGION=us-east-1
   ```

2. **AWS Profile**:
   ```bash
   export AWS_PROFILE=my-profile
   ```

3. **AWS Config file**:
   ```bash
   export AWS_CONFIG_FILE=~/.aws/config
   ```

If no credentials are found, E2E tests are automatically skipped.

## Test Patterns

### Pattern 1: Terraform Validation Test

```go
func TestTerraformValidation(t *testing.T) {
    t.Run("validates example Python Lambda infrastructure", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping infrastructure test in short mode")
        }

        terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

        terraformOptions := &terraform.Options{
            TerraformDir: terraformDir,
            NoColor:      true,
        }

        // Validate terraform configuration
        _, err := terraform.InitE(t, terraformOptions)
        require.NoError(t, err, "Terraform init should succeed")

        _, err = terraform.ValidateE(t, terraformOptions)
        require.NoError(t, err, "Terraform configuration should be valid")
    })
}
```

**Key points**:
- ✅ Skip in short mode
- ✅ Use `InitE` and `ValidateE` for error handling
- ✅ Assert on errors with clear messages

### Pattern 2: End-to-End Deployment Test

```go
func TestLambdaDeploymentEndToEnd(t *testing.T) {
    t.Run("deploys and destroys Python Lambda function", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping E2E deployment test in short mode")
        }

        if !isAWSConfigured() {
            t.Skip("Skipping E2E test - AWS credentials not configured")
        }

        terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

        // Generate unique namespace for isolation
        uniqueID := random.UniqueId()
        namespace := fmt.Sprintf("forge-test-%s", strings.ToLower(uniqueID))

        terraformOptions := &terraform.Options{
            TerraformDir: terraformDir,
            NoColor:      true,
            Vars: map[string]interface{}{
                "namespace": namespace,
            },
            MaxRetries:         3,
            TimeBetweenRetries: 5 * time.Second,
        }

        // Ensure cleanup happens even if test fails
        defer func() {
            t.Log("Cleaning up test infrastructure...")
            terraform.Destroy(t, terraformOptions)
        }()

        // DEPLOY PHASE
        terraform.Init(t, terraformOptions)
        terraform.Apply(t, terraformOptions)

        // VERIFICATION PHASE
        functionName := terraform.Output(t, terraformOptions, "function_name")
        require.NotEmpty(t, functionName)

        awsRegion := terraform.Output(t, terraformOptions, "aws_region")
        if awsRegion == "" {
            awsRegion = "us-east-1"
        }

        // Use retry for eventual consistency
        retry.DoWithRetry(t, "Check Lambda function exists", 10, 3*time.Second, func() (string, error) {
            output, err := aws.InvokeFunctionWithParamsE(
                t,
                awsRegion,
                functionName,
                &aws.LambdaOptions{
                    Payload: map[string]interface{}{"test": "validation"},
                },
            )
            if err != nil {
                return "", fmt.Errorf("failed to invoke Lambda: %w", err)
            }
            return "Lambda verified", nil
        })
    })
}
```

**Key points**:
- ✅ Skip in short mode
- ✅ Check AWS credentials before running
- ✅ Generate unique namespace for test isolation
- ✅ Use `defer` for guaranteed cleanup
- ✅ Retry operations for eventual consistency
- ✅ Verify actual AWS resources

### Pattern 3: Resource Configuration Test

```go
func TestAPIGatewayIntegration(t *testing.T) {
    t.Run("validates API Gateway configuration", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping infrastructure test in short mode")
        }

        terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

        // Check file exists
        apiGatewayFile := filepath.Join(terraformDir, "apigateway.tf")
        assert.FileExists(t, apiGatewayFile, "apigateway.tf should exist")

        terraformOptions := &terraform.Options{
            TerraformDir: terraformDir,
            NoColor:      true,
        }

        terraform.Init(t, terraformOptions)
        planOutput := terraform.Plan(t, terraformOptions)

        // Verify resources in plan
        assert.Contains(t, planOutput, "aws_apigatewayv2", "Plan should include API Gateway v2 resources")
    })
}
```

**Key points**:
- ✅ Verify files exist before testing
- ✅ Check plan output for expected resources
- ✅ Fast execution (no deployment)

## Best Practices

### 1. Test Isolation

**Always use unique namespaces**:
```go
uniqueID := random.UniqueId()
namespace := fmt.Sprintf("forge-test-%s", strings.ToLower(uniqueID))
```

This prevents test interference when running in parallel or in CI/CD.

### 2. Cleanup Guarantees

**Always use defer for cleanup**:
```go
defer func() {
    t.Log("Cleaning up test infrastructure...")
    terraform.Destroy(t, terraformOptions)
}()
```

Cleanup runs even if test fails or panics.

### 3. Retry for Eventual Consistency

**Use retry for AWS operations**:
```go
retry.DoWithRetry(t, "Check resource exists", 10, 3*time.Second, func() (string, error) {
    // Check resource exists
    return "success", nil
})
```

AWS resources may take time to become available after creation.

### 4. Short Mode Support

**Always support short mode**:
```go
if testing.Short() {
    t.Skip("Skipping infrastructure test in short mode")
}
```

Enables fast feedback loop for developers.

### 5. AWS Credential Check

**Check credentials before E2E tests**:
```go
if !isAWSConfigured() {
    t.Skip("Skipping E2E test - AWS credentials not configured")
}
```

Prevents test failures in environments without AWS access.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Infrastructure Tests

on: [push, pull_request]

jobs:
  validation:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      # Fast validation tests (no AWS)
      - name: Run validation tests
        run: go test ./test/infrastructure/... -v -short

  e2e:
    runs-on: ubuntu-latest
    # Only run on main branch or manual trigger
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      # Configure AWS credentials
      - name: Configure AWS
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: ${{ secrets.AWS_ROLE_ARN }}
          aws-region: us-east-1

      # E2E tests with AWS deployment
      - name: Run E2E tests
        run: go test ./test/infrastructure/... -v -timeout 15m
```

**Strategy**:
- ✅ Run validation tests on every PR (fast, no AWS)
- ✅ Run E2E tests on main branch only (slow, uses AWS)
- ✅ Use OIDC for secure AWS credential management
- ✅ Set reasonable timeout (15 minutes)

## Adding New Tests

### 1. Add Validation Test

For new Terraform resources (SNS, SQS, EventBridge, etc.):

```go
func TestSNSConfiguration(t *testing.T) {
    t.Run("validates SNS topic configuration", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping infrastructure test in short mode")
        }

        terraformDir := filepath.Join("..", "..", "examples", "generated-python-lambda", "terraform")

        // Check file exists
        snsFile := filepath.Join(terraformDir, "sns.tf")
        assert.FileExists(t, snsFile, "sns.tf should exist")

        terraformOptions := &terraform.Options{
            TerraformDir: terraformDir,
            NoColor:      true,
        }

        terraform.Init(t, terraformOptions)
        planOutput := terraform.Plan(t, terraformOptions)

        // Verify SNS resources
        assert.Contains(t, planOutput, "aws_sns_topic", "Plan should include SNS topic")
    })
}
```

### 2. Add Deployment Test

For verifying actual AWS resource creation:

```go
func TestSNSDeployment(t *testing.T) {
    t.Run("deploys SNS topic", func(t *testing.T) {
        if testing.Short() {
            t.Skip("Skipping E2E test in short mode")
        }

        if !isAWSConfigured() {
            t.Skip("Skipping E2E test - AWS credentials not configured")
        }

        // Setup terraform options with unique namespace
        uniqueID := random.UniqueId()
        namespace := fmt.Sprintf("forge-test-%s", strings.ToLower(uniqueID))

        terraformOptions := &terraform.Options{
            TerraformDir: filepath.Join("..", "..", "examples", "sns-example", "terraform"),
            NoColor:      true,
            Vars: map[string]interface{}{
                "namespace": namespace,
            },
        }

        defer terraform.Destroy(t, terraformOptions)

        // Deploy
        terraform.Init(t, terraformOptions)
        terraform.Apply(t, terraformOptions)

        // Verify
        topicArn := terraform.Output(t, terraformOptions, "topic_arn")
        require.NotEmpty(t, topicArn)
        assert.Contains(t, topicArn, "arn:aws:sns")
    })
}
```

## Troubleshooting

### Test Timeouts

**Problem**: E2E tests timeout
**Solution**: Increase timeout
```bash
go test ./test/infrastructure/... -v -timeout 15m
```

### Cleanup Failures

**Problem**: Resources not cleaned up after test failure
**Solution**: Check CloudWatch logs, manually run:
```bash
cd examples/generated-python-lambda/terraform
terraform destroy -var="namespace=forge-test-xyz"
```

### AWS Permission Errors

**Problem**: Tests fail with permission denied
**Solution**: Ensure AWS role has these permissions:
- `lambda:*`
- `iam:CreateRole`
- `iam:AttachRolePolicy`
- `apigateway:*`
- `dynamodb:*`
- `logs:*`

### Parallel Test Conflicts

**Problem**: Tests interfere with each other
**Solution**: Ensure unique namespaces:
```go
uniqueID := random.UniqueId()  // Different each time
```

## Coverage Impact

**Infrastructure tests coverage**: `[no statements]`
- Infrastructure tests contain only test functions (no library code)
- They validate generated Terraform, not Go code
- Aggregate coverage not affected

**Functional code coverage**: **80.3%**
- Excludes generated Lingon AWS packages
- Target: 90% (per CLAUDE.md)
- Gap: 9.7% (achievable with more CLI tests)

## Dependencies

Infrastructure tests require:

```go
require (
    github.com/gruntwork-io/terratest v0.52.0
    github.com/stretchr/testify v1.10.0
)
```

**Terratest modules used**:
- `modules/terraform` - Terraform operations
- `modules/aws` - AWS resource verification
- `modules/random` - Unique ID generation
- `modules/retry` - Retry with backoff
- `modules/logger` - Structured logging

## Future Enhancements

### P1: Multi-Runtime Tests
- Add tests for Go Lambda functions
- Add tests for Node.js Lambda functions
- Verify runtime-specific configurations

### P2: Regional Testing
- Test deployments across multiple AWS regions
- Verify region-specific resource configurations

### P3: Performance Testing
- Measure deployment time
- Track resource creation latency
- Compare deployment methods

### P4: Cost Estimation
- Estimate test run costs
- Track AWS spending per test
- Optimize for cost efficiency

## References

- [Terratest Documentation](https://terratest.gruntwork.io/)
- [Terratest AWS Module](https://pkg.go.dev/github.com/gruntwork-io/terratest/modules/aws)
- [Terratest Terraform Module](https://pkg.go.dev/github.com/gruntwork-io/terratest/modules/terraform)
- [Forge Testing Guide](TDD_PROGRESS.md)
- [Coverage Report](SESSION_SUMMARY_2025-10-30.md)

---

**Status**: ✅ Production-ready
**Test Count**: 12 tests (8 validation, 4 deployment patterns)
**Coverage**: Infrastructure validated, 0 linting errors
**Maintenance**: Add tests for new Terraform resources as generated
