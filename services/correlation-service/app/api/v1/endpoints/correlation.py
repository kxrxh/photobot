import logging
from typing import Any

from fastapi import APIRouter, Depends, HTTPException
from opentelemetry import trace

from ....auth.validator import get_current_user
from ....models.correlation import CorrelationRequest, CorrelationWithTest
from ....services.correlation import CorrelationCalculatorService

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)
router = APIRouter()


@router.post(
    "/",
    response_model=list[CorrelationWithTest],
    summary="Calculate Pairwise Correlation Conditions",
)
async def calculate_correlation(
    request_data: CorrelationRequest,
    user: dict[str, Any] = Depends(get_current_user),  # noqa: B008
) -> list[CorrelationWithTest]:
    with tracer.start_as_current_span("correlation.request.calculate") as span:
        group_names = [group.name for group in request_data.fractions]
        user_id = str(user.get("sub", "unknown"))
        span.set_attribute("correlation.groups.count", len(group_names))
        span.set_attribute("enduser.id", user_id)
        logger.info(
            f"Starting correlation calculation for groups: {group_names} (User: {user_id})"
        )

        try:
            results = await CorrelationCalculatorService.calculate_pairwise_correlation(
                data=request_data,
            )
            span.set_attribute("correlation.groups.with_conditions", len(results))
            logger.info(
                f"Correlation calculation completed. Processed {len(group_names)} groups."
            )
            return results
        except HTTPException:
            raise
        except ValueError as e:
            logger.warning(f"Validation error: {e!s}")
            raise HTTPException(
                status_code=400, detail=f"Invalid request data: {e!s}"
            ) from e
        except Exception as e:
            logger.error(f"Unexpected error: {e!s}", exc_info=True)
            raise HTTPException(
                status_code=500,
                detail="Failed to calculate correlations due to an internal error",
            ) from e
