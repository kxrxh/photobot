.PHONY: dev dev-setup install

dev: dev-setup
	bun run dev

dev-setup:
	@test -f .env.development.local || cp .env.development.local.example .env.development.local
	@test -d node_modules || bun install

install:
	bun install
