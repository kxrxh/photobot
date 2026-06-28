from pathlib import Path
from typing import Annotated

from dotenv import load_dotenv
from pydantic import AliasChoices, Field, field_validator
from pydantic_settings import BaseSettings, NoDecode, SettingsConfigDict

for env_file in (
    ".env",
    ".env.local",
    "compose.env",
    ".env.development",
    ".env.production",
):
    path = Path.cwd() / env_file
    if path.exists():
        load_dotenv(path)


def _aliases(name: str) -> AliasChoices:
    return AliasChoices(name, f"CORRELATION_{name}")


class AppConfig(BaseSettings):
    model_config = SettingsConfigDict(
        extra="ignore",
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
    )

    title: str = Field(default="Correlation Calculator Service")
    description: str = Field(
        default="A standalone service to calculate object correlations based on attributes."
    )
    version: str = Field(default="0.1.0")

    host: str = Field(default="0.0.0.0")  # noqa: S104
    port: int = Field(default=6000, validation_alias=_aliases("PORT"))
    log_level: str = Field(default="INFO", validation_alias=_aliases("LOG_LEVEL"))
    debug: bool = Field(default=False, validation_alias=_aliases("DEBUG"))
    reload: bool = Field(default=False, validation_alias=_aliases("RELOAD"))

    cors_origins: Annotated[list[str], NoDecode] = Field(
        default=["*"], validation_alias=_aliases("CORS_ORIGINS")
    )
    cors_methods: Annotated[list[str], NoDecode] = Field(
        default=["GET", "POST", "OPTIONS"], validation_alias=_aliases("CORS_METHODS")
    )
    cors_headers: Annotated[list[str], NoDecode] = Field(
        default=["*"], validation_alias=_aliases("CORS_HEADERS")
    )

    auth_service_url: str = Field(
        default="http://localhost:8080/api/v1",
        validation_alias=_aliases("AUTH_SERVICE_URL"),
    )
    jwt_audience: str = Field(
        default="correlation-service", validation_alias=_aliases("JWT_AUDIENCE")
    )

    analysis_service_url: str = Field(
        default="http://localhost:6040/api/v1/",
        validation_alias=_aliases("ANALYSIS_SERVICE_URL"),
    )
    analysis_service_audience: str = Field(
        default="analysis-service",
        validation_alias=_aliases("ANALYSIS_SERVICE_AUDIENCE"),
    )
    service_id: str = Field(
        default="correlation-service", validation_alias=_aliases("SERVICE_ID")
    )
    service_secret: str = Field(default="", validation_alias=_aliases("SERVICE_SECRET"))
    otel_service_name: str = Field(
        default="correlation-service", validation_alias=_aliases("OTEL_SERVICE_NAME")
    )
    otel_exporter_otlp_endpoint: str = Field(
        default="", validation_alias=_aliases("OTEL_EXPORTER_OTLP_ENDPOINT")
    )
    otel_exporter_otlp_protocol: str = Field(
        default="http/protobuf",
        validation_alias=_aliases("OTEL_EXPORTER_OTLP_PROTOCOL"),
    )
    otel_resource_attributes: str = Field(
        default="service.namespace=photobot",
        validation_alias=_aliases("OTEL_RESOURCE_ATTRIBUTES"),
    )
    otel_sdk_disabled: bool = Field(
        default=False, validation_alias=_aliases("OTEL_SDK_DISABLED")
    )

    @field_validator("log_level")
    @classmethod
    def validate_log_level(cls, v: str) -> str:
        valid_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
        if v.upper() not in valid_levels:
            raise ValueError(f"Log level must be one of: {', '.join(valid_levels)}")
        return v.upper()

    @field_validator("port")
    @classmethod
    def validate_port(cls, v: int) -> int:
        if not (1 <= v <= 65535):
            raise ValueError("Port must be between 1 and 65535")
        return v

    @field_validator("otel_exporter_otlp_protocol")
    @classmethod
    def validate_otel_protocol(cls, v: str) -> str:
        value = v.strip().lower()
        if value != "http/protobuf":
            raise ValueError("OTEL_EXPORTER_OTLP_PROTOCOL must be http/protobuf")
        return value

    @field_validator("cors_origins", "cors_methods", "cors_headers", mode="before")
    @classmethod
    def validate_cors_lists(cls, v: list[str] | str) -> list[str]:
        if isinstance(v, str):
            v = [item.strip() for item in v.split(",") if item.strip()]
        if not v:
            raise ValueError("CORS settings cannot be empty")
        return v

    @property
    def jwks_url(self) -> str:
        return f"{self.auth_service_url}/auth/.well-known/jwks.json"

    def get_uvicorn_config(self) -> dict:
        from .logging import get_logging_config

        return {
            "host": self.host,
            "port": self.port,
            "log_level": self.log_level.lower(),
            "reload": self.reload or self.debug,
            "log_config": get_logging_config(),
        }


settings = AppConfig()
