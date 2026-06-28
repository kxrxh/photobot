"""Unit tests for CorrelationCalculatorService."""

from unittest.mock import AsyncMock, patch

import pytest
from fastapi import HTTPException

from app.models.correlation import (
    Condition,
    CorrelationRequest,
    CorrelationWithTest,
    ObjectGroup,
    ParameterGroup,
)
from app.services.correlation import CorrelationCalculatorService
from app.services.data_processor import ProcessedDataset, ProcessedObjectData


def make_object_metadata(obj_id: int, **kwargs):
    from .conftest import make_object_metadata as _make

    return _make(obj_id, **kwargs)


@pytest.fixture
def service_dataset():
    objs = []
    for i in range(20):
        meta = make_object_metadata(i, m_h=0.5 + (i % 2) * 0.3, m_s=0.4, m_v=0.6)
        group = "group_a" if i < 10 else "group_b"
        objs.append(ProcessedObjectData(object_id=i, metadata=meta, group_name=group))
    return ProcessedDataset(
        objects_by_id={o.object_id: o for o in objs},
        objects_by_group={"group_a": objs[:10], "group_b": objs[10:]},
        feature_names=["m_h", "m_s", "m_v"],
    )


@pytest.fixture
def service_request():
    return CorrelationRequest(
        fractions=[
            ObjectGroup(name="group_a", object_ids=list(range(10))),
            ObjectGroup(name="group_b", object_ids=list(range(10, 20))),
        ],
        parameter_groups=[ParameterGroup.ALL],
    )


class TestFbeta:
    """Tests for _fbeta static method."""

    def test_returns_zero_when_precision_zero(self):
        assert CorrelationCalculatorService._fbeta(0.0, 0.5, 2.0) == 0.0

    def test_returns_zero_when_recall_zero(self):
        assert CorrelationCalculatorService._fbeta(0.5, 0.0, 2.0) == 0.0

    def test_returns_value_when_both_positive(self):
        result = CorrelationCalculatorService._fbeta(0.8, 0.6, 2.0)
        assert 0 < result < 1

    def test_beta_affects_result(self):
        r1 = CorrelationCalculatorService._fbeta(0.8, 0.6, 1.0)
        r2 = CorrelationCalculatorService._fbeta(0.8, 0.6, 2.0)
        assert r1 != r2


class TestPostProcessConditions:
    """Tests for post_process_conditions."""

    def test_returns_empty_when_no_conditions(self, service_dataset, service_request):
        result = CorrelationCalculatorService.post_process_conditions(
            [], service_dataset, "group_a", service_request
        )
        assert result == []

    def test_returns_simplified_conditions_when_possible(
        self, service_dataset, service_request
    ):
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.4),
            Condition(attribute="m_s", operator="<=", value=0.8),
        ]
        result = CorrelationCalculatorService.post_process_conditions(
            conditions, service_dataset, "group_a", service_request
        )
        assert isinstance(result, list)
        assert len(result) >= 1


class TestFindGroupConditions:
    """Tests for find_group_conditions."""

    def test_returns_empty_when_no_feature_names(
        self, service_dataset, service_request
    ):
        dataset = ProcessedDataset(
            objects_by_id=service_dataset.objects_by_id,
            objects_by_group=service_dataset.objects_by_group,
            feature_names=[],
        )
        result = CorrelationCalculatorService.find_group_conditions(
            "group_a", dataset, service_request
        )
        assert result == []

    def test_returns_empty_when_empty_feature_matrix(self, service_request):
        empty_dataset = ProcessedDataset(
            objects_by_id={},
            objects_by_group={"group_a": [], "group_b": []},
            feature_names=["m_h"],
        )
        result = CorrelationCalculatorService.find_group_conditions(
            "group_a", empty_dataset, service_request
        )
        assert result == []

    def test_returns_conditions_when_valid_data(self, service_dataset, service_request):
        result = CorrelationCalculatorService.find_group_conditions(
            "group_a", service_dataset, service_request
        )
        assert isinstance(result, list)
        for c in result:
            assert isinstance(c, Condition)


class TestCalculatePairwiseCorrelation:
    """Tests for calculate_pairwise_correlation_optimized."""

    @pytest.mark.asyncio
    async def test_returns_empty_results_when_no_objects(self):
        req = CorrelationRequest(
            fractions=[
                ObjectGroup(name="g1", object_ids=[]),
                ObjectGroup(name="g2", object_ids=[]),
            ],
            parameter_groups=[ParameterGroup.ALL],
        )
        with patch(
            "app.services.correlation.create_processed_dataset",
            new_callable=AsyncMock,
        ) as mock_create:
            mock_create.return_value = ProcessedDataset(
                objects_by_id={},
                objects_by_group={"g1": [], "g2": []},
                feature_names=[],
            )
            result = await CorrelationCalculatorService.calculate_pairwise_correlation(
                req
            )
        assert len(result) == 2
        assert all(r.conditions == [] for r in result)
        assert all(r.test_results is None for r in result)

    @pytest.mark.asyncio
    async def test_returns_results_for_valid_request(
        self, service_dataset, service_request
    ):
        with (
            patch(
                "app.services.correlation.create_processed_dataset",
                new_callable=AsyncMock,
            ) as mock_create,
            patch(
                "app.services.correlation.prepare_attribute_data",
            ) as mock_prep,
        ):
            mock_create.return_value = service_dataset
            mock_prep.return_value = {
                "m_h": [(0.5, "group_a"), (0.2, "group_b")],
                "m_s": [(0.4, "group_a"), (0.7, "group_b")],
                "m_v": [(0.6, "group_a"), (0.3, "group_b")],
            }
            result = await CorrelationCalculatorService.calculate_pairwise_correlation(
                service_request
            )
        assert isinstance(result, list)
        assert len(result) == 2
        for r in result:
            assert isinstance(r, CorrelationWithTest)
            assert r.name in ["group_a", "group_b"]

    @pytest.mark.asyncio
    async def test_raises_http_exception_on_internal_error(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="g1", object_ids=[1])],
            parameter_groups=[ParameterGroup.ALL],
        )
        with patch(
            "app.services.correlation.create_processed_dataset",
            new_callable=AsyncMock,
        ) as mock_create:
            mock_create.side_effect = RuntimeError("Unexpected")
            with pytest.raises(HTTPException) as exc_info:
                await CorrelationCalculatorService.calculate_pairwise_correlation(req)
            assert exc_info.value.status_code == 500
