# End-to-End Infrastructure Tests

This directory contains E2E infrastructure tests for each runtime generator in Forge.

## Overview

E2E tests verify that:
1. Generated project code is valid and builds correctly
2. Terraform infrastructure deploys successfully to AWS
3. All AWS resources are configured correctly (Lambda, API Gateway, DynamoDB, etc.)
4. The deployed API functions as expected
5. Infrastructure can be torn down cleanly

## Test Organization

```
test/e2e/
├── python/          # Python Lambda E2E tests
│   ├── lambda_infra_test.go
│   └── helpers.go
├── go/              # Go Lambda E2E tests (TODO)
└── nodejs/          # Node.js Lambda E2E tests (TODO)
```

Each runtime has its own package with dedicated infrastructure tests.

## Running E2E Tests

### Prerequisites

1. **AWS Credentials**: Configure AWS credentials
   ```bash
   export AWS_ACCESS_KEY_ID="your-key"
   export AWS_SECRET_ACCESS_KEY="your-secret"
   export AWS_REGION="us-east-1"
   ```

2. **Terraform**: Install Terraform
   ```bash
   brew install terraform  # macOS
   ```

3. **Dependencies**: Install test dependencies
   ```bash
   cd test/e2e
   go mod download
   ```

### Run All E2E Tests

```bash
# From project root
make test-e2e

# Or manually
go test -v -tags=e2e -timeout=30m ./test/e2e/...
```

### Run Specific Runtime Tests

```bash
# Python Lambda tests only
go test -v -tags=e2e -timeout=30m ./test/e2e/python/

# With verbose output
go test -v -tags=e2e -timeout=30m ./test/e2e/python/ -run TestPythonLambdaInfrastructure
```

### Keep Infrastructure Running

By default, tests tear down infrastructure after completion. To keep it running:

```bash
SKIP_TEARDOWN=true go test -v -tags=e2e ./test/e2e/python/
```

## What Gets Tested

### Python Lambda E2E Tests

#### Lambda Function
- ✅ Runtime configuration (Python 3.13)
- ✅ Handler path
- ✅ Timeout and memory settings
- ✅ Environment variables
- ✅ X-Ray tracing enabled
- ✅ IAM role attached

#### DynamoDB Table
- ✅ Table exists and is active
- ✅ Pay-per-request billing mode
- ✅ Correct key schema (hash key: id)
- ✅ Server-side encryption enabled
- ✅ Point-in-time recovery enabled

#### CloudWatch Logs
- ✅ Lambda log group exists
- ✅ API Gateway log group exists
- ✅ Retention period set to 7 days

#### API Gateway
- ✅ API endpoint is accessible
- ✅ Correct route configuration

## Adding New Runtime Tests

To add E2E tests for a new runtime (e.g., Node.js):

1. Create directory: `test/e2e/nodejs/`
2. Add `helpers.go` with common functions
3. Create `lambda_infra_test.go` with Terratest assertions
4. Follow the Python example structure
5. Update this README

## Test Dependencies

E2E tests use:
- **Terratest** - Terraform testing framework
- **AWS SDK Go** - AWS resource verification
- **testify** - Assertions and test utilities

## Cost Considerations

E2E tests deploy real AWS resources which incur costs:
- Lambda: ~$0.20 per 1M requests (Free tier: 1M requests/month)
- API Gateway: ~$1.00 per 1M requests (Free tier: 1M requests/month)
- DynamoDB: Pay-per-request (~$1.25 per 1M writes)
- CloudWatch Logs: ~$0.50 per GB

**Estimated cost per test run**: < $0.01 (within free tier limits)

## CI/CD Integration

Example GitHub Actions workflow:

```yaml
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.22'

      - name: Install Terraform
        uses: hashicorp/setup-terraform@v2

      - name: Configure AWS Credentials
        uses: aws-actions/configure-aws-credentials@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Run E2E Tests
        run: make test-e2e
        env:
          AWS_REGION: us-east-1
```

## Troubleshooting

### Test Timeout

If tests timeout:
```bash
# Increase timeout (default: 30m)
go test -v -tags=e2e -timeout=60m ./test/e2e/python/
```

### AWS Permission Errors

Ensure your AWS credentials have permissions for:
- Lambda (create/read/delete functions)
- API Gateway v2 (create/read/delete APIs)
- DynamoDB (create/read/delete tables)
- CloudWatch Logs (create/read log groups)
- IAM (create/attach/delete roles and policies)

### Terraform State Issues

Tests use local Terraform state. If state gets corrupted:
```bash
# Clean up state
rm -rf examples/generated-python-lambda/terraform/.terraform
rm -f examples/generated-python-lambda/terraform/terraform.tfstate*
```

## Future Enhancements

- [ ] Add HTTP request testing to verify API functionality
- [ ] Test Lambda invocation with sample payloads
- [ ] Verify DynamoDB data persistence
- [ ] Add performance benchmarks
- [ ] Parallel test execution for multiple runtimes
- [ ] Smoke tests for critical paths only (faster than full E2E)
