from aws_lambda_powertools.event_handler import APIGatewayRestResolver

# API configuration
API_PATH = '/api/orders'

# Initialize API resolver
app = APIGatewayRestResolver()
