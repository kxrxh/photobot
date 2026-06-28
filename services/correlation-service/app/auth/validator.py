import logging
from typing import Annotated, Any

import jwt
from fastapi import HTTPException, Security
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from jwt import PyJWKClient

from ..core.config import settings

logger = logging.getLogger(__name__)


class JWTValidator:
    def __init__(self, jwks_url: str):
        self.jwks_url = jwks_url
        self.jwks_client = PyJWKClient(jwks_url)

    async def validate_token(self, token: str) -> dict[str, Any]:
        try:
            # Fetch the signing key from JWKS
            signing_key = self.jwks_client.get_signing_key_from_jwt(token)

            # Decode and validate the token
            return jwt.decode(
                token,
                signing_key.key,
                algorithms=["RS256"],
                audience=settings.jwt_audience,
            )
        except jwt.ExpiredSignatureError:
            logger.warning("Token has expired")
            raise HTTPException(status_code=401, detail="Token has expired") from None
        except jwt.InvalidTokenError as e:
            logger.warning(f"Invalid token: {e!s}")
            raise HTTPException(status_code=401, detail=f"Invalid token: {e!s}") from e
        except Exception as e:
            logger.error(f"Error validating token: {e!s}")
            raise HTTPException(
                status_code=401, detail="Could not validate credentials"
            ) from e


# Initialize validator
jwt_validator = JWTValidator(settings.jwks_url)

security = HTTPBearer()


async def get_current_user(
    credentials: Annotated[HTTPAuthorizationCredentials, Security(security)],
) -> dict[str, Any]:
    return await jwt_validator.validate_token(credentials.credentials)
