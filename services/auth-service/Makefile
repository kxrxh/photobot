COMPOSE := docker compose -f compose.yml --env-file compose.env
COMPOSE_DEV := docker compose -f docker-compose-dev.yml

.PHONY: generate test-integration start run reload stop logs logs-service \
	dev dev-setup dev-deps dev-up dev-down dev-logs dev-ps dev-reload dev-keys dev-run dev-seed

generate:
	go generate ./...
	sqlc generate

test-integration:
	go test -tags=integration -count=1 -parallel 1 -v ./integration/...

start:
	$(COMPOSE) up -d

run:
	$(COMPOSE) up

reload:
	$(COMPOSE) down
	$(COMPOSE) up -d --build

stop:
	$(COMPOSE) down

logs:
	$(COMPOSE) logs -f

logs-service:
	$(COMPOSE) logs -f $(SERVICE)

dev: dev-setup
	DEV_MODE=true go run ./cmd/app

dev-setup: dev-keys
	@test -f compose.env || cp compose.dev.env.example compose.env
	@mkdir -p persistence/postgres
	$(COMPOSE_DEV) up -d --wait

dev-deps dev-up:
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

dev-keys:
	@mkdir -p keys
	@test -f keys/private.pem || openssl genrsa -out keys/private.pem 2048
	@test -f keys/public.pem  || openssl rsa -in keys/private.pem -pubout -out keys/public.pem

dev-run:
	DEV_MODE=true go run ./cmd/app

dev-seed:
	DEV_MODE=true go run ./cmd/seed-dev
