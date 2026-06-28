COMPOSE := docker compose -f compose.yml --env-file compose.env
COMPOSE_DEV := docker compose -f docker-compose-dev.yml

.PHONY: start build update reload reload-analysis stop logs \
	dev dev-up dev-down dev-logs dev-ps dev-reload \
	test test.integration

test:
	go test ./...

# Requires Docker. If DOCKER_HOST is unset, uses the active `docker context` endpoint so
# testcontainers matches `docker` CLI (OrbStack, Colima, remote daemons). Postgres tests skip
# if the daemon is still unreachable (e.g. no Docker in CI).
test.integration:
	DOCKER_HOST="$${DOCKER_HOST:-$$(docker context inspect -f '{{.Endpoints.docker.Host}}' 2>/dev/null)}" \
		go test -tags=integration -timeout=15m ./integration/...

start:
	$(COMPOSE) up -d

build:
	GIT_COMMIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) $(COMPOSE) build

update:
	GIT_COMMIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) $(COMPOSE) up -d --build

reload:
	$(COMPOSE) down
	GIT_COMMIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo unknown) $(COMPOSE) up -d --build

reload-analysis:
	$(COMPOSE) up -d --build analysis_service

stop:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

# Local dependencies only (no analysis_service image)
dev dev-up:
	$(COMPOSE_DEV) up -d

dev-down:
	$(COMPOSE_DEV) down

dev-logs:
	$(COMPOSE_DEV) logs -f

dev-ps:
	$(COMPOSE_DEV) ps

dev-reload:
	$(COMPOSE_DEV) down
	$(COMPOSE_DEV) up -d
