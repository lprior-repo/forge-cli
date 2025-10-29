output "lambda_function_name" {
  description = "Lambda function name"
  value       = aws_lambda_function.main.function_name
}

output "lambda_function_arn" {
  description = "Lambda function ARN"
  value       = aws_lambda_function.main.arn
}

output "lambda_invoke_arn" {
  description = "Lambda invoke ARN"
  value       = aws_lambda_function.main.invoke_arn
}

output "api_gateway_url" {
  description = "API Gateway URL"
  value       = aws_apigatewayv2_stage.default.invoke_url
}

output "api_endpoint" {
  description = "Full API endpoint URL"
  value       = "${aws_apigatewayv2_stage.default.invoke_url}${var.service_name}"
}

output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = aws_dynamodb_table.main.name
}

output "dynamodb_table_arn" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.main.arn
}
