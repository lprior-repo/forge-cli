# DynamoDB table
resource "aws_dynamodb_table" "main" {
  name         = "orders-table"
  billing_mode = var.dynamodb_billing_mode
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }

  ttl {
    attribute_name = "ttl"
    enabled        = false
  }

  point_in_time_recovery {
    enabled = true
  }

  server_side_encryption {
    enabled = true
  }

  tags = {
    Name = "orders-table"
  }
}
