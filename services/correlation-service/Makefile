COMPOSE = docker compose -f compose.yml --env-file compose.env

install:
	uv sync --all-extras

start:
	$(COMPOSE) up -d

stop:
	$(COMPOSE) down

reload: stop
	$(COMPOSE) up -d --build

logs:
	$(COMPOSE) logs -f

test:
	uv run pytest

format:
	uv run ruff format .

run:
	uv run correlation-service

token:
	@sh scripts/get_token.sh
