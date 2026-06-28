FROM python:3.13-slim AS builder

COPY --from=ghcr.io/astral-sh/uv:0.6.5 /uv /uvx /bin/

ENV UV_SYSTEM_PYTHON=1
ENV UV_LINK_MODE=copy
ENV UV_COMPILE_BYTECODE=1

WORKDIR /app
COPY pyproject.toml uv.lock ./

RUN uv sync --frozen --no-dev --no-install-project

FROM python:3.13-slim

COPY --from=ghcr.io/astral-sh/uv:0.6.5 /uv /uvx /bin/

ENV UV_SYSTEM_PYTHON=1
ENV UV_LINK_MODE=copy
ENV UV_COMPILE_BYTECODE=1

WORKDIR /app
COPY --from=builder /app/.venv /app/.venv
COPY pyproject.toml uv.lock ./
COPY app ./app

RUN uv sync --frozen --no-dev --no-editable

ENV PATH="/app/.venv/bin:$PATH"
ARG APP_UID=10001
ARG APP_GID=10001
RUN groupadd -g "${APP_GID}" app \
    && useradd -u "${APP_UID}" -g app -m -s /usr/sbin/nologin app \
    && chown -R app:app /app
USER app
ENV PORT=6021
EXPOSE 6021

CMD ["correlation-service"]
