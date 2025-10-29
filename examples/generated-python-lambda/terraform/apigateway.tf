# API Gateway HTTP API (v2)
resource "aws_apigatewayv2_api" "main" {
  name          = "${var.service_name}-${var.environment}"
  protocol_type = "HTTP"
  description   = "Orders service API - Create order endpoint"

  cors_configuration {
    allow_origins = ["*"]
    allow_methods = ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allow_headers = ["content-type", "x-amz-date", "authorization", "x-api-key"]
    max_age       = 300
  }
}

# API Gateway stage
resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.main.id
  name        = "$default"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gateway.arn
    format = jsonencode({
      requestId      = "$context.requestId"
      ip             = "$context.identity.sourceIp"
      requestTime    = "$context.requestTime"
      httpMethod     = "$context.httpMethod"
      routeKey       = "$context.routeKey"
      status         = "$context.status"
      protocol       = "$context.protocol"
      responseLength = "$context.responseLength"
    })
  }
}

# CloudWatch Log Group for API Gateway
resource "aws_cloudwatch_log_group" "api_gateway" {
  name              = "/aws/apigateway/${var.service_name}-${var.environment}"
  retention_in_days = 7
}

# API Gateway integration with Lambda
resource "aws_apigatewayv2_integration" "lambda" {
  api_id             = aws_apigatewayv2_api.main.id
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
  integration_uri    = aws_lambda_function.main.invoke_arn
  payload_format_version = "2.0"
}

# API Gateway route
resource "aws_apigatewayv2_route" "main" {
  api_id    = aws_apigatewayv2_api.main.id
  route_key = "POST /api/orders"
  target    = "integrations/${aws_apigatewayv2_integration.lambda.id}"
}
