# Terraform Infrastructure - NO CloudFormation!

This Python Lambda project uses **100% Terraform** for infrastructure deployment.

## Why Terraform (Not CloudFormation/CDK)?

✅ **Multi-cloud capable** - Not locked into AWS
✅ **Better state management** - Remote state, locking, versioning
✅ **Mature ecosystem** - 1000+ providers
✅ **Type-safe generation** - Generated from Go code
✅ **No vendor lock-in** - Can migrate to other clouds

## Generated Terraform Files

All Terraform code is **programmatically generated** using type-safe Go:

```
terraform/
├── main.tf         # Provider configuration
├── variables.tf    # Input variables
├── outputs.tf      # Output values
├── lambda.tf       # Lambda function + CloudWatch logs
├── iam.tf          # IAM roles and policies
├── dynamodb.tf     # DynamoDB table
└── apigateway.tf   # API Gateway v2 (HTTP API)
```

## Infrastructure Components

### 1. Lambda Function (`lambda.tf`)
- **Runtime**: Python 3.13
- **Handler**: `service.handlers.handle_request.lambda_handler`
- **Timeout**: 30 seconds (configurable)
- **Memory**: 512 MB (configurable)
- **X-Ray Tracing**: Enabled
- **Environment Variables**:
  - `POWERTOOLS_SERVICE_NAME`
  - `LOG_LEVEL`
  - `TABLE_NAME` (DynamoDB)

### 2. API Gateway v2 (`apigateway.tf`)
- **Type**: HTTP API (cheaper & faster than REST API)
- **CORS**: Configured for web apps
- **Logging**: CloudWatch access logs
- **Route**: `POST /api/orders`
- **Integration**: Lambda proxy integration

### 3. DynamoDB Table (`dynamodb.tf`)
- **Billing**: Pay-per-request (on-demand)
- **Encryption**: Server-side encryption enabled
- **PITR**: Point-in-time recovery enabled
- **Hash Key**: `id` (String)

### 4. IAM Roles (`iam.tf`)
- **Lambda Execution Role**: AssumeRole for Lambda service
- **CloudWatch Logs**: Basic execution role attached
- **X-Ray**: Daemon write access
- **DynamoDB**: GetItem, PutItem, UpdateItem, DeleteItem, Query, Scan

### 5. CloudWatch Logs
- **Lambda logs**: `/aws/lambda/orders-service-dev`
- **API Gateway logs**: `/aws/apigateway/orders-service-dev`
- **Retention**: 7 days

## Deployment

### Prerequisites

```bash
# Install Terraform
brew install terraform  # macOS
# or
sudo apt-get install terraform  # Ubuntu

# Configure AWS credentials
aws configure
```

### Deploy

```bash
cd terraform

# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Deploy infrastructure
terraform apply

# Get outputs
terraform output
```

### Outputs

After deployment, you'll get:

```hcl
Outputs:

api_endpoint = "https://abc123.execute-api.us-east-1.amazonaws.com/api/orders"
api_gateway_url = "https://abc123.execute-api.us-east-1.amazonaws.com"
dynamodb_table_arn = "arn:aws:dynamodb:us-east-1:123456789:table/orders-table"
dynamodb_table_name = "orders-table"
lambda_function_arn = "arn:aws:lambda:us-east-1:123456789:function:orders-service-dev"
lambda_function_name = "orders-service-dev"
lambda_invoke_arn = "arn:aws:lambda:us-east-1:123456789:function:orders-service-dev"
```

## Testing the API

```bash
# Get the API endpoint
API_URL=$(terraform output -raw api_endpoint)

# Test the Lambda
curl -X POST $API_URL \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-order",
    "count": 5
  }'
```

## Customization

Edit `variables.tf` to customize:

```hcl
variable "lambda_timeout" {
  default = 60  # Increase timeout
}

variable "lambda_memory_size" {
  default = 1024  # Increase memory
}

variable "dynamodb_billing_mode" {
  default = "PROVISIONED"  # Switch to provisioned capacity
}
```

## State Management

For production, use remote state:

```hcl
# Add to main.tf
terraform {
  backend "s3" {
    bucket         = "my-terraform-state"
    key            = "orders-service/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

## Comparison: Terraform vs CloudFormation

| Feature | Terraform | CloudFormation |
|---------|-----------|----------------|
| Multi-cloud | ✅ Yes | ❌ AWS only |
| State management | ✅ Built-in | ⚠️ Manual |
| Module ecosystem | ✅ 1000+ providers | ⚠️ Limited |
| Type safety | ✅ HCL validation | ⚠️ JSON/YAML |
| Community | ✅ Large | ⚠️ AWS-focused |
| Learning curve | ⚠️ Medium | ⚠️ Medium |
| AWS integration | ✅ Excellent | ✅ Native |

## Generated vs Hand-Written

This Terraform code is **100% programmatically generated** using type-safe Go:

```go
// From internal/generators/python/terraform.go
func (g *Generator) generateTerraformLambda() string {
    return fmt.Sprintf(`
resource "aws_lambda_function" "main" {
  function_name = "${var.service_name}-${var.environment}"
  runtime      = var.lambda_runtime
  handler      = "service.handlers.handle_request.lambda_handler"
  ...
}
`, ...)
}
```

### Benefits:
- ✅ **Type-safe**: Compiler catches errors
- ✅ **Consistent**: Same structure every time
- ✅ **Maintainable**: Update generator, regenerate all
- ✅ **Testable**: Unit tests verify generation
- ✅ **No templates**: Pure Go code

## Cleanup

```bash
# Destroy all infrastructure
terraform destroy

# Confirm with 'yes'
```

---

**Generated by Forge** - Type-safe serverless infrastructure toolkit
**No CloudFormation. No CDK. Just Terraform.**
