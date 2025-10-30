# Lambda deployment package
data "archive_file" "lambda" {
  type        = "zip"
  source_dir  = "${path.module}/../.build/lambda"
  output_path = "${path.module}/../.build/lambda.zip"
}

# Lambda function
resource "aws_lambda_function" "main" {
  filename         = data.archive_file.lambda.output_path
  function_name    = "${var.service_name}-${var.environment}"
  role            = aws_iam_role.lambda.arn
  handler         = "service.handlers.handle_request.lambda_handler"
  source_code_hash = data.archive_file.lambda.output_base64sha256
  runtime         = var.lambda_runtime
  timeout         = var.lambda_timeout
  memory_size     = var.lambda_memory_size

  environment {
    variables = {
      POWERTOOLS_SERVICE_NAME      = var.service_name
      LOG_LEVEL                    = "INFO"
      ENVIRONMENT                  = var.environment
      TABLE_NAME                   = aws_dynamodb_table.main.name
      IDEMPOTENCY_TABLE_NAME       = aws_dynamodb_table.main.name
    }
  }

  tracing_config {
    mode = "Active"
  }

  depends_on = [
    aws_iam_role_policy_attachment.lambda_logs,
    aws_cloudwatch_log_group.lambda,
  ]
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.service_name}-${var.environment}"
  retention_in_days = 7
}

# Lambda permission for API Gateway
resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.main.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.main.execution_arn}/*/*"
}
