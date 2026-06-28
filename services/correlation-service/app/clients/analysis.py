import logging
from typing import Any

import aiohttp
from opentelemetry import trace
from pydantic import BaseModel, ConfigDict

from ..auth.manager import analysis_token_manager
from ..core.config import settings

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)

ANALYSIS_SERVICE_URL = settings.analysis_service_url
if not ANALYSIS_SERVICE_URL.endswith("/"):
    ANALYSIS_SERVICE_URL += "/"


class ObjectMetadata(BaseModel):
    id: int
    id_analysis: int | None = None
    class_: str | None = None
    geometry: str | None = None
    m_h: float | None = None
    m_s: float | None = None
    m_v: float | None = None
    m_r: float | None = None
    m_g: float | None = None
    m_b: float | None = None
    l_avg: float | None = None
    w_avg: float | None = None
    brt_avg: float | None = None
    r_avg: float | None = None
    g_avg: float | None = None
    b_avg: float | None = None
    h_avg: float | None = None
    s_avg: float | None = None
    v_avg: float | None = None
    h: float | None = None
    s: float | None = None
    v: float | None = None
    h_m: float | None = None
    s_m: float | None = None
    v_m: float | None = None
    r_m: float | None = None
    g_m: float | None = None
    b_m: float | None = None
    brt_m: float | None = None
    w_m: float | None = None
    l_m: float | None = None
    l: float | None = None  # noqa: E741
    w: float | None = None
    l_w: float | None = None
    pr: float | None = None
    sq: float | None = None
    brt: float | None = None
    r: float | None = None
    g: float | None = None
    b: float | None = None
    solid: float | None = None
    min_h: float | None = None
    min_s: float | None = None
    min_v: float | None = None
    max_h: float | None = None
    max_s: float | None = None
    max_v: float | None = None
    entropy: float | None = None
    color_rhs: str | None = None
    sq_sqcrl: float | None = None
    hu1: float | None = None
    hu2: float | None = None
    hu3: float | None = None
    hu4: float | None = None
    hu5: float | None = None
    hu6: float | None = None

    model_config = ConfigDict(
        populate_by_name=True,
        extra="allow",
    )


class SuccessResponse(BaseModel):
    success: bool
    result: list[ObjectMetadata]


class ErrorResponse(BaseModel):
    success: bool
    error: Any


async def get_objects_by_id(object_ids: list[int]) -> SuccessResponse:
    with tracer.start_as_current_span("analysis.get_objects_by_id") as span:
        span.set_attribute("analysis.objects.count", len(object_ids))
        token = await analysis_token_manager.get_token()
        headers = {"Authorization": f"Bearer {token}"}

        async with (
            aiohttp.ClientSession() as session,
            session.post(
                f"{ANALYSIS_SERVICE_URL}objects/search",
                json={"objects": object_ids},
                headers=headers,
            ) as response,
        ):
            span.set_attribute("http.response.status_code", response.status)
            if response.status != 200:
                error_text = await response.text()
                try:
                    error_response = ErrorResponse.model_validate_json(error_text)
                    error_msg = error_response.error
                except Exception:
                    error_msg = error_text
                raise Exception(f"Failed to get objects: {error_msg}")

            return SuccessResponse.model_validate_json(await response.text())
