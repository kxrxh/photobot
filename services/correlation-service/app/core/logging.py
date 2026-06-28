import logging

from opentelemetry import trace

from .config import settings


class TraceContextFilter(logging.Filter):
    def filter(self, record: logging.LogRecord) -> bool:
        span = trace.get_current_span()
        context = span.get_span_context()
        if context.is_valid:
            record.trace_id = f"{context.trace_id:032x}"
            record.span_id = f"{context.span_id:016x}"
        else:
            record.trace_id = "-"
            record.span_id = "-"
        return True


def get_logging_config() -> dict:
    """Get comprehensive logging configuration dictionary."""
    log_level = getattr(logging, settings.log_level, logging.INFO)

    return {
        "version": 1,
        "disable_existing_loggers": False,
        "formatters": {
            "standard": {
                "format": "%(asctime)s - %(name)s - %(levelname)s - trace_id=%(trace_id)s span_id=%(span_id)s - %(message)s",
                "datefmt": "%Y-%m-%d %H:%M:%S",
            },
        },
        "filters": {
            "trace_context": {"()": TraceContextFilter},
        },
        "handlers": {
            "console": {
                "class": "logging.StreamHandler",
                "formatter": "standard",
                "stream": "ext://sys.stdout",
                "level": log_level,
                "filters": ["trace_context"],
            },
        },
        "root": {
            "handlers": ["console"],
            "level": log_level,
        },
        "loggers": {
            "app": {
                "handlers": ["console"],
                "level": log_level,
                "propagate": False,
            },
            "uvicorn": {
                "handlers": ["console"],
                "level": log_level,
                "propagate": False,
            },
            "uvicorn.access": {
                "handlers": ["console"],
                "level": log_level,
                "propagate": False,
            },
            "uvicorn.error": {
                "handlers": ["console"],
                "level": log_level,
                "propagate": False,
            },
        },
    }


def setup_logging() -> logging.Logger:
    """Setup logging with consistent format across all components."""
    import logging.config

    # Apply the logging configuration
    logging.config.dictConfig(get_logging_config())

    # Return main app logger
    return logging.getLogger("app")


logger = setup_logging()
