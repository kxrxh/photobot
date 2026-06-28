"""Unit tests for correlation API endpoint."""

from unittest.mock import AsyncMock, patch

from fastapi.testclient import TestClient

from app.auth.validator import get_current_user
from app.main import create_app
from app.models.correlation import CorrelationWithTest


class TestCalculateCorrelationEndpoint:
    """Tests for POST / correlation endpoint."""

    def test_requires_authentication(self):
        app = create_app()
        client = TestClient(app)
        response = client.post(
            "/api/v1/correlation/",
            json={
                "fractions": [{"name": "g1", "object_ids": []}],
                "parameter_groups": ["all"],
            },
        )
        assert response.status_code == 401

    def test_returns_results_when_authenticated(self):
        async def fake_user():
            return {"sub": "test-user"}

        async def mock_calc(data):
            return [
                CorrelationWithTest(
                    name="g1",
                    conditions=[],
                    test_results=None,
                ),
            ]

        app = create_app()
        app.dependency_overrides[get_current_user] = fake_user

        with patch(
            "app.api.v1.endpoints.correlation.CorrelationCalculatorService.calculate_pairwise_correlation",
            new_callable=AsyncMock,
            side_effect=mock_calc,
        ):
            client = TestClient(app)
            response = client.post(
                "/api/v1/correlation/",
                json={
                    "fractions": [{"name": "g1", "object_ids": [1, 2]}],
                    "parameter_groups": ["all"],
                },
                headers={"Authorization": "Bearer fake-token"},
            )

        assert response.status_code == 200
        data = response.json()
        assert isinstance(data, list)
        assert len(data) >= 1
        assert data[0]["name"] == "g1"
