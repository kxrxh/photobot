from fastapi import FastAPI
from opentelemetry import trace
from opentelemetry.exporter.otlp.proto.http.trace_exporter import OTLPSpanExporter
from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
from opentelemetry.instrumentation.fastapi import FastAPIInstrumentor
from opentelemetry.sdk.resources import Resource
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor

from .config import settings


def _parse_resource_attributes(raw: str) -> dict[str, str]:
    attrs: dict[str, str] = {}
    for part in raw.split(","):
        pair = part.strip()
        if not pair:
            continue
        key, sep, value = pair.partition("=")
        if not sep:
            continue
        key = key.strip()
        value = value.strip()
        if key:
            attrs[key] = value
    return attrs


def configure_observability(app: FastAPI) -> None:
    if settings.otel_sdk_disabled:
        return
    if not settings.otel_exporter_otlp_endpoint:
        return

    resource_attributes = _parse_resource_attributes(settings.otel_resource_attributes)
    resource_attributes["service.name"] = settings.otel_service_name
    resource = Resource.create(resource_attributes)

    provider = TracerProvider(resource=resource)
    exporter = OTLPSpanExporter(endpoint=settings.otel_exporter_otlp_endpoint)
    provider.add_span_processor(BatchSpanProcessor(exporter))
    trace.set_tracer_provider(provider)

    FastAPIInstrumentor.instrument_app(app)
    AioHttpClientInstrumentor().instrument()

    @app.on_event("shutdown")
    async def _shutdown_tracing() -> None:
        provider.shutdown()
