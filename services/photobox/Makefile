start:
	docker compose -f compose.yml --env-file compose.env up -d

run:
	docker compose -f compose.yml --env-file compose.env up

reload:
	docker compose -f compose.yml --env-file compose.env down
	docker compose -f compose.yml --env-file compose.env up -d --build

stop:
	docker compose -f compose.yml --env-file compose.env down

logs:
	docker compose -f compose.yml --env-file compose.env logs -f

logs-service:
	docker compose -f compose.yml --env-file compose.env logs -f $(SERVICE)

generate:
	go generate ./internal/weed/... ./internal/proposal/...

test:
	go test ./...

test-integration:
	go test -tags=integration -count=1 -timeout=15m ./integration/...

build:
	go build ./cmd/app/...
