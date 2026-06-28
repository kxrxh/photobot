FROM golang:1.26-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum* ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-w -s" -o /out/reports ./cmd/api

FROM debian:bookworm-slim

ENV DEBIAN_FRONTEND=noninteractive
ENV CHROME_BIN=/usr/bin/chromium
ENV PORT=6050

RUN apt-get update && apt-get install -y --no-install-recommends \
    chromium \
    ca-certificates \
    tini \
    fonts-liberation \
    fonts-noto-color-emoji \
    fonts-recommended \
    && rm -rf /var/lib/apt/lists/*

RUN useradd -m -u 10001 appuser
USER appuser
WORKDIR /app

COPY --from=builder --chown=appuser:appuser /out/reports /app/reports
COPY --chown=appuser:appuser templates /app/templates

EXPOSE 6050

ENTRYPOINT ["tini", "--"]
CMD ["/app/reports"]