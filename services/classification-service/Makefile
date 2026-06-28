generate:
	go generate ./internal/classification/...
	go generate ./internal/markup/...
	go generate ./internal/merge/...
	go generate ./internal/params/...
	go generate ./internal/product/...
	go generate ./internal/user/...

test-integration:
	go test -tags=integration -count=1 -parallel 1 -v ./integration/...

start:
	docker compose -f compose.yml --env-file compose.env up -d

reload:
	docker compose -f compose.yml --env-file compose.env down
	docker compose -f compose.yml --env-file compose.env up -d --build

stop:
	docker compose -f compose.yml --env-file compose.env down

logs:
	docker compose -f compose.yml --env-file compose.env logs -f

dev:
	docker compose -f docker-compose-dev.yml up -d --build

dev-reload:
	docker compose -f docker-compose-dev.yml down
	docker compose -f docker-compose-dev.yml up -d --build