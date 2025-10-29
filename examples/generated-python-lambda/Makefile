.PHONY: install test format lint deploy clean

install:
	poetry install

test:
	poetry run pytest tests/ -v

format:
	poetry run ruff format .

lint:
	poetry run ruff check .
	poetry run mypy service/

deploy:
	cdk deploy

clean:
	find . -type d -name __pycache__ -exec rm -rf {} +
	find . -type f -name "*.pyc" -delete
	rm -rf .pytest_cache .coverage htmlcov/ dist/ build/
