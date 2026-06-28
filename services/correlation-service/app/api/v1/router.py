from fastapi import APIRouter

from .endpoints import correlation

api_router = APIRouter()
api_router.include_router(
    correlation.router, prefix="/correlation", tags=["Correlation"]
)
