from aws_lambda_powertools.event_handler import APIGatewayHttpResolver

# API configuration
API_PATH = '/api/orders'

# Initialize API resolver (HTTP API v2)
app = APIGatewayHttpResolver()
