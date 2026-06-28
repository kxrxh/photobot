import logging
import time

import aiohttp
import jwt
from fastapi import HTTPException
from opentelemetry import trace

from ..core.config import settings

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)


class IdentityClient:
    """Client for interacting with the Identity Service."""

    def __init__(self, base_url: str, service_id: str, service_secret: str):
        self.base_url = base_url
        self.service_id = service_id
        self.service_secret = service_secret

    async def login_as_service(self, audience: str) -> dict[str, str]:
        """Authenticate as a service and return access and refresh tokens."""
        with tracer.start_as_current_span("auth.service.login") as span:
            span.set_attribute("auth.audience", audience)
            url = f"{self.base_url}/auth/login"
            payload = {
                "service_id": self.service_id,
                "service_secret": self.service_secret,
                "audience": audience,
            }
            headers = {"X-Grant-Type": "client_credentials"}

            async with (
                aiohttp.ClientSession() as session,
                session.post(url, json=payload, headers=headers) as response,
            ):
                span.set_attribute("http.response.status_code", response.status)
                return await self._handle_response(response, "login")

    async def refresh_tokens(self, refresh_token: str) -> dict[str, str]:
        """Refresh tokens using a refresh token."""
        with tracer.start_as_current_span("auth.service.refresh_tokens"):
            url = f"{self.base_url}/auth/refresh"
            payload = {"refresh_token": refresh_token}

            async with (
                aiohttp.ClientSession() as session,
                session.post(url, json=payload) as response,
            ):
                return await self._handle_response(response, "refresh")

    async def _handle_response(
        self, response: aiohttp.ClientResponse, action: str
    ) -> dict:
        """Handle standardized API responses, unwrapping success/result if present."""
        text = await response.text()

        try:
            data = await response.json()
        except Exception as err:
            logger.error(f"Failed to parse {action} response as JSON: {text}")
            raise HTTPException(
                status_code=response.status,
                detail=f"Identity service {action} failed: {text}",
            ) from err

        if response.status != 200:
            # Handle standardized error response
            if isinstance(data, dict) and not data.get("success", True):
                error = data.get("error", {})
                message = error.get("message", "Unknown error")
                details = error.get("details", "")
                error_msg = f"{message}: {details}" if details else message
                logger.error(
                    f"Identity service {action} failed ({response.status}): {error_msg}"
                )
                raise HTTPException(
                    status_code=response.status,
                    detail=f"Identity service {action} failed: {error_msg}",
                )

            logger.error(
                f"Identity service {action} failed ({response.status}): {text}"
            )
            raise HTTPException(
                status_code=response.status,
                detail=f"Identity service {action} failed: {text}",
            )

        # Handle standardized success response
        if isinstance(data, dict) and data.get("success") is True and "result" in data:
            return data["result"]

        return data


class ServiceTokenManager:
    """Manages service-to-service authentication tokens with automatic rotation."""

    def __init__(self, identity_client: IdentityClient, audience: str):
        self.client = identity_client
        self.audience = audience
        self.access_token: str | None = None
        self.refresh_token: str | None = None
        self.expires_at: float = 0

    async def get_token(self) -> str:
        """Get a valid access token, performing login or refresh if necessary."""
        if self.access_token and time.time() < self.expires_at - 60:
            return self.access_token

        try:
            if self.refresh_token:
                try:
                    logger.info("Refreshing service token...")
                    tokens = await self.client.refresh_tokens(self.refresh_token)
                    self._update_tokens(tokens)
                    if self.access_token is None:
                        raise Exception(
                            "access_token is unexpectedly None after refresh"
                        )
                    return self.access_token
                except Exception as e:
                    logger.warning(
                        f"Failed to refresh token, falling back to login: {e!s}"
                    )

            logger.info(f"Logging in as service for audience: {self.audience}")
            tokens = await self.client.login_as_service(self.audience)
            self._update_tokens(tokens)
            if self.access_token is None:
                raise Exception("access_token is unexpectedly None after login")
            return self.access_token
        except Exception as e:
            logger.error(f"Failed to get service token: {e!s}")
            raise HTTPException(
                status_code=500, detail="Service-to-service authentication failed"
            ) from e

    def _update_tokens(self, tokens: dict[str, str]):
        self.access_token = tokens.get("access_token")
        self.refresh_token = tokens.get("refresh_token")

        if not self.access_token:
            logger.error("No access_token found in identity service response")
            return

        # Decode token to get expiration time
        try:
            payload = jwt.decode(self.access_token, options={"verify_signature": False})
            self.expires_at = payload.get("exp", 0)
        except Exception as e:
            logger.error(f"Failed to decode service token: {e!s}")
            # Fallback: assume 1 hour expiration if decoding fails
            self.expires_at = time.time() + 3600


# Initialize identity client
identity_client = IdentityClient(
    base_url=settings.auth_service_url,
    service_id=settings.service_id,
    service_secret=settings.service_secret,
)

# Initialize token manager for Analysis Service
analysis_token_manager = ServiceTokenManager(
    identity_client, audience=settings.analysis_service_audience
)
