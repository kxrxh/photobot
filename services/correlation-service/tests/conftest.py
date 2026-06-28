"""Shared pytest fixtures and configuration for correlation service tests."""

from unittest.mock import MagicMock

import numpy as np
import pytest

from app.clients.analysis import ObjectMetadata, SuccessResponse
from app.models.correlation import (
    Condition,
    CorrelationRequest,
    ObjectGroup,
    ParameterGroup,
)
from app.services.data_processor import ProcessedDataset, ProcessedObjectData


def make_object_metadata(
    obj_id: int,
    **kwargs: float | int | str | None,
) -> ObjectMetadata:
    """Create ObjectMetadata with defaults for tests."""
    defaults = {
        "id": obj_id,
        "id_analysis": None,
        "class_": None,
        "geometry": None,
        "m_h": 0.5,
        "m_s": 0.3,
        "m_v": 0.7,
        "l_avg": 10.0,
        "w_avg": 5.0,
        "h": 180.0,
        "s": 0.5,
        "v": 0.8,
    }
    defaults.update(kwargs)
    return ObjectMetadata(**defaults)


@pytest.fixture
def sample_object_metadata() -> ObjectMetadata:
    """Single object metadata for tests."""
    return make_object_metadata(1, m_h=0.5, m_s=0.3, m_v=0.7)


@pytest.fixture
def sample_condition() -> Condition:
    """Sample condition for tests."""
    return Condition(attribute="m_h", operator=">=", value=0.4)


@pytest.fixture
def sample_conditions() -> list[Condition]:
    """List of sample conditions."""
    return [
        Condition(attribute="m_h", operator=">=", value=0.3),
        Condition(attribute="m_s", operator="<=", value=0.8),
    ]


@pytest.fixture
def sample_correlation_request() -> CorrelationRequest:
    """Sample correlation request with two groups."""
    return CorrelationRequest(
        fractions=[
            ObjectGroup(name="group_a", object_ids=[1, 2, 3]),
            ObjectGroup(name="group_b", object_ids=[4, 5, 6]),
        ],
        parameter_groups=[ParameterGroup.ALL],
    )


@pytest.fixture
def sample_processed_dataset(
    sample_correlation_request: CorrelationRequest,
) -> ProcessedDataset:
    """Create a ProcessedDataset with mock objects for testing."""
    group_a_objs = [
        ProcessedObjectData(
            object_id=i,
            metadata=make_object_metadata(i, m_h=0.5 + i * 0.1, m_s=0.3),
            group_name="group_a",
        )
        for i in [1, 2, 3]
    ]
    group_b_objs = [
        ProcessedObjectData(
            object_id=i,
            metadata=make_object_metadata(i, m_h=0.2 + i * 0.05, m_s=0.8),
            group_name="group_b",
        )
        for i in [4, 5, 6]
    ]
    objects_by_id = {o.object_id: o for o in group_a_objs + group_b_objs}
    objects_by_group = {
        "group_a": group_a_objs,
        "group_b": group_b_objs,
    }
    return ProcessedDataset(
        objects_by_id=objects_by_id,
        objects_by_group=objects_by_group,
        feature_names=["m_h", "m_s", "m_v", "l_avg", "w_avg"],
    )


@pytest.fixture
def sample_feature_matrix() -> tuple[np.ndarray, np.ndarray]:
    """Sample X matrix and y vector for ML tests."""
    x = np.array(
        [
            [0.5, 0.3, 0.7],
            [0.6, 0.4, 0.6],
            [0.7, 0.2, 0.8],
            [0.2, 0.8, 0.3],
            [0.25, 0.75, 0.35],
            [0.3, 0.85, 0.25],
        ],
        dtype=np.float32,
    )
    y = np.array([1, 1, 1, 0, 0, 0], dtype=np.int8)
    return x, y


@pytest.fixture
def mock_get_objects_by_id():
    """Factory fixture to create mock for get_objects_by_id."""

    def _create_mock(objects: list[ObjectMetadata] | None = None):
        async def mock_fn(object_ids: list[int]):
            if objects is None:
                return SuccessResponse(
                    success=True,
                    result=[make_object_metadata(oid) for oid in object_ids],
                )
            return SuccessResponse(success=True, result=objects)

        return mock_fn

    return _create_mock


@pytest.fixture
def mock_analysis_token_manager():
    """Mock analysis_token_manager.get_token to return a test token."""

    def _create_mock():
        manager = MagicMock()
        manager.get_token = MagicMock(return_value="test-token-123")
        return manager

    return _create_mock
