# Deployment Guide

Complete guide to deploy this Python Lambda service to AWS using Terraform and Task.

## Prerequisites

### 1. Install Tools

```bash
# Install Task (task runner)
brew install go-task/tap/go-task  # macOS
# or
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Install Terraform
brew install terraform  # macOS
# or
wget https://releases.hashicorp.com/terraform/1.6.0/terraform_1.6.0_linux_amd64.zip
unzip terraform_1.6.0_linux_amd64.zip
sudo mv terraform /usr/local/bin/

# Install Python 3.13+ and pip
# Arch Linux
sudo pacman -S python python-pip

# macOS
brew install python@3.13

# Ubuntu/Debian
sudo apt-get install python3.13 python3-pip

# Install Poetry (Python package manager)
curl -sSL https://install.python-poetry.org | python3 -

# Install AWS CLI
# Via pip
pip install --user awscli

# macOS
brew install awscli

# Arch Linux
sudo pacman -S aws-cli

# Verify installations
task --version
terraform --version
python3 --version
poetry --version
aws --version
```

### 2. Configure AWS Credentials

```bash
# Configure AWS CLI
aws configure

# Or export credentials
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# Verify credentials
aws sts get-caller-identity
```

## Quick Start

```bash
# Show available tasks
task --list

# Deploy everything
task deploy

# Test the API
task test-api

# View logs
task logs

# Destroy infrastructure
task destroy
```

## Step-by-Step Deployment

### Step 1: Install Dependencies

```bash
task install
```

This installs Python dependencies using Poetry.

### Step 2: Run Tests (Optional)

```bash
# Run full test suite
task full-test

# Or individual steps
task lint      # Lint code
task test      # Run tests
task format    # Format code
```

### Step 3: Build Lambda Package

```bash
task build
```

This:
- Creates `.build/lambda/` directory
- Copies service code
- Installs production dependencies
- Creates deployment package

### Step 4: Deploy with Terraform

```bash
# Option A: Deploy in one command
task deploy

# Option B: Step-by-step
task tf-init      # Initialize Terraform
task tf-plan      # Preview changes
task tf-apply     # Deploy infrastructure
```

### Step 5: Test the Deployment

```bash
# Get deployment info
task outputs

# Test API endpoint
task test-api

# Or manually
curl -X POST "$(cd terraform && terraform output -raw api_endpoint)" \
  -H "Content-Type: application/json" \
  -d '{"name": "test-order", "count": 5}'
```

### Step 6: Monitor

```bash
# View Lambda logs in real-time
task logs

# Check deployment status
task status

# Invoke Lambda directly
task invoke
```

## Available Tasks

| Task | Description |
|------|-------------|
| `task deploy` | Build and deploy everything |
| `task build` | Build Lambda deployment package |
| `task test` | Run Python tests |
| `task lint` | Lint Python code |
| `task format` | Format Python code |
| `task tf-init` | Initialize Terraform |
| `task tf-plan` | Preview Terraform changes |
| `task tf-apply` | Apply Terraform changes |
| `task destroy` | Destroy all infrastructure |
| `task outputs` | Show Terraform outputs |
| `task test-api` | Test deployed API |
| `task logs` | Tail Lambda logs |
| `task invoke` | Invoke Lambda directly |
| `task status` | Show deployment status |
| `task clean` | Clean build artifacts |

## Infrastructure Created

When you run `task deploy`, Terraform creates:

1. **Lambda Function**
   - Runtime: Python 3.13
   - Memory: 512 MB
   - Timeout: 30 seconds
   - X-Ray tracing enabled

2. **API Gateway v2 (HTTP API)**
   - Route: `POST /api/orders`
   - CORS enabled
   - Access logging to CloudWatch

3. **DynamoDB Table**
   - Name: `orders-table`
   - Billing: Pay-per-request
   - Encryption: Enabled
   - Point-in-time recovery: Enabled

4. **IAM Roles**
   - Lambda execution role
   - CloudWatch Logs permissions
   - DynamoDB access permissions
   - X-Ray write permissions

5. **CloudWatch Log Groups**
   - Lambda logs: `/aws/lambda/orders-service-dev`
   - API Gateway logs: `/aws/apigateway/orders-service-dev`
   - Retention: 7 days

## Outputs

After deployment:

```bash
task outputs
```

Returns:

```json
{
  "api_endpoint": "https://abc123.execute-api.us-east-1.amazonaws.com/api/orders",
  "api_gateway_url": "https://abc123.execute-api.us-east-1.amazonaws.com",
  "dynamodb_table_name": "orders-table",
  "lambda_function_name": "orders-service-dev"
}
```

## Testing

### Test via API Gateway

```bash
# Use task
task test-api

# Or manually
API_URL=$(cd terraform && terraform output -raw api_endpoint)
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "count": 10
  }' | jq '.'
```

Expected response:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "John Doe",
  "count": 10,
  "status": "created"
}
```

### Test via Direct Invoke

```bash
task invoke
```

### View Logs

```bash
# Real-time logs
task logs

# Or manually
aws logs tail "/aws/lambda/orders-service-dev" --follow
```

## Troubleshooting

### Build Issues

```bash
# Clean and rebuild
task clean
task build

# Check Python version
python3 --version  # Should be 3.13

# Check Poetry
poetry --version
```

### Deployment Issues

```bash
# Check AWS credentials
aws sts get-caller-identity

# Validate Terraform
task tf-validate

# Check Terraform state
cd terraform && terraform show
```

### API Issues

```bash
# Check Lambda logs
task logs

# Invoke directly to bypass API Gateway
task invoke

# Check API Gateway configuration
aws apigatewayv2 get-apis --region us-east-1
```

## Cleanup

### Destroy Everything

```bash
task destroy
```

This removes all AWS resources created by Terraform.

### Clean Local Files

```bash
task clean
```

This removes:
- `.build/` directory
- Terraform state files
- Terraform cache

## Cost Estimation

With the default configuration:

- **Lambda**: ~$0.20 per 1M requests
- **API Gateway**: ~$1.00 per 1M requests
- **DynamoDB**: Pay-per-request ($1.25 per 1M writes)
- **CloudWatch Logs**: ~$0.50 per GB ingested

**Free tier**: 1M Lambda requests/month, 1M API Gateway requests/month

**Estimated cost for testing**: < $0.01

## Next Steps

1. **Add Tests**: Add unit tests in `tests/`
2. **Customize**: Edit `terraform/variables.tf` for your needs
3. **CI/CD**: Integrate with GitHub Actions
4. **Monitoring**: Add CloudWatch alarms
5. **Production**: Create separate environments (dev/staging/prod)

## Resources

- [AWS Lambda Documentation](https://docs.aws.amazon.com/lambda/)
- [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Task Documentation](https://taskfile.dev/)
- [AWS Lambda Powertools](https://docs.powertools.aws.dev/lambda/python/)

---

**Ready to deploy?**

```bash
task deploy
```
