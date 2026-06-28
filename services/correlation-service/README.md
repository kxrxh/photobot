# Correlation Service

Standalone service to calculate object correlations based on attributes.

## Setup

```bash
cp env.example .env          # for make run
cp env.example compose.env   # for make start
make install
```

## Run

```bash
make run
```

## Test

```bash
make test
```

## Dev testing

Run AuthService with `JWT_ACCESS_EXPIRY_MINUTES=1440` (24h) in its env. Set `SERVICE_SECRET`, use `make token`. Tokens last all day for Postman.

## Docker

```bash
make start   # up -d
make stop    # down
make logs    # logs -f
make reload  # rebuild and restart
```
