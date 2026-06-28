from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse

from .api.v1.router import api_router
from .core.config import settings
from .core.logging import logger
from .core.observability import configure_observability


def create_app() -> FastAPI:
    app = FastAPI(
        title=settings.title,
        description=settings.description,
        version=settings.version,
        docs_url="/docs",
        redoc_url="/redoc",
    )

    from fastapi.middleware.cors import CORSMiddleware

    app.add_middleware(
        CORSMiddleware,  # type: ignore[arg-type]
        allow_origins=settings.cors_origins,
        allow_credentials=True,
        allow_methods=settings.cors_methods,
        allow_headers=settings.cors_headers,
    )

    app.include_router(api_router, prefix="/api/v1")
    # For backward compatibility with the old path /correlation-api/
    app.include_router(api_router)

    configure_observability(app)

    return app


app = create_app()


@app.exception_handler(HTTPException)
async def http_exception_handler(request: Request, exc: HTTPException) -> JSONResponse:
    logger.error(
        f"HTTP {exc.status_code} error on {request.method} {request.url}: {exc.detail}"
    )
    return JSONResponse(
        status_code=exc.status_code,
        content={"error": exc.detail, "status_code": exc.status_code},
    )


@app.exception_handler(Exception)
async def general_exception_handler(request: Request, exc: Exception) -> JSONResponse:
    logger.error(
        f"Unexpected error on {request.method} {request.url}: {exc!s}", exc_info=True
    )
    return JSONResponse(
        status_code=500,
        content={
            "error": "An unexpected error occurred",
            "status_code": 500,
            "detail": str(exc)
            if settings.log_level == "DEBUG"
            else "Internal server error",
        },
    )


@app.get("/health", summary="Health Check", tags=["Health"])
async def health_check() -> dict:
    return {"status": "healthy", "service": settings.title, "version": settings.version}


@app.get("/info", summary="Service Information", tags=["Info"])
async def service_info() -> dict:
    return {
        "service": settings.title,
        "version": settings.version,
        "description": settings.description,
        "log_level": settings.log_level,
    }


def main() -> None:
    import uvicorn

    uvicorn_config = settings.get_uvicorn_config()
    uvicorn.run("app.main:app", **uvicorn_config)


if __name__ == "__main__":
    main()
