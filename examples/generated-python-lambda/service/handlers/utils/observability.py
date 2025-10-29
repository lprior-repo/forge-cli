from aws_lambda_powertools import Logger, Metrics, Tracer

# Initialize Powertools
logger = Logger(service='orders-service')
metrics = Metrics(namespace='orders-service', service='orders-service')
tracer = Tracer(service='orders-service')
