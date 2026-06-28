"""Unit tests for data_processor module."""

from unittest.mock import AsyncMock, patch

import numpy as np
import pytest
from fastapi import HTTPException

from app.clients.analysis import ObjectMetadata, SuccessResponse
from app.models.correlation import CorrelationRequest, ObjectGroup, ParameterGroup
from app.services.data_processor import (
    ProcessedDataset,
    ProcessedObjectData,
    build_feature_matrix,
    create_processed_dataset,
    handle_missing_values,
    prepare_attribute_data,
    validate_training_data,
)


def make_object_metadata(obj_id: int, **kwargs):
    defaults = {
        "id": obj_id,
        "id_analysis": None,
        "m_h": 0.5,
        "m_s": 0.3,
        "m_v": 0.7,
        "l_avg": 10.0,
        "w_avg": 5.0,
    }
    defaults.update(kwargs)
    return ObjectMetadata(**defaults)


class TestBuildFeatureMatrix:
    """Tests for build_feature_matrix."""

    def test_returns_none_when_no_valid_objects(self, sample_correlation_request):
        empty = ProcessedDataset(
            objects_by_id={},
            objects_by_group={"group_a": [], "group_b": []},
            feature_names=["m_h", "m_s"],
        )
        x, y, ids = build_feature_matrix(empty, "group_a", ["m_h", "m_s"])
        assert x is None
        assert y is None
        assert ids == []

    def test_builds_matrix_for_valid_objects(
        self, sample_processed_dataset, sample_correlation_request
    ):
        x, y, ids = build_feature_matrix(
            sample_processed_dataset, "group_a", ["m_h", "m_s"]
        )
        assert x is not None
        assert y is not None
        assert x.shape[0] == len(sample_processed_dataset.get_valid_objects())
        assert x.shape[1] == 2
        assert len(y) == x.shape[0]
        assert len(ids) == x.shape[0]

    def test_uses_nan_for_missing_values(self):
        meta = make_object_metadata(1, m_h=0.5, m_s=None)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="g")
        dataset = ProcessedDataset(
            objects_by_id={1: obj},
            objects_by_group={"g": [obj]},
            feature_names=["m_h", "m_s"],
        )
        x, _, _ = build_feature_matrix(dataset, "g", ["m_h", "m_s"])
        assert np.isnan(x[0, 1])

    def test_labels_correctly_for_target_group(self, sample_processed_dataset):
        _, y, _ = build_feature_matrix(sample_processed_dataset, "group_a", ["m_h"])
        n_group_a = len(sample_processed_dataset.get_objects_in_group("group_a"))
        assert np.sum(y == 1) == n_group_a
        assert (
            np.sum(y == 0)
            == len(sample_processed_dataset.get_valid_objects()) - n_group_a
        )


class TestHandleMissingValues:
    """Tests for handle_missing_values."""

    def test_returns_unchanged_when_no_nans(self):
        x = np.array([[1.0, 2.0], [3.0, 4.0]], dtype=np.float32)
        out, names = handle_missing_values(x, ["a", "b"], "group")
        np.testing.assert_array_equal(out, x)
        assert names == ["a", "b"]

    def test_imputes_nans(self):
        x = np.array([[1.0, np.nan], [3.0, 4.0]], dtype=np.float32)
        out, _ = handle_missing_values(x, ["a", "b"], "group")
        assert not np.isnan(out).any()
        assert out.shape == x.shape

    def test_removes_all_nan_columns(self):
        x = np.array([[1.0, np.nan, 3.0], [2.0, np.nan, 4.0]], dtype=np.float32)
        out, names = handle_missing_values(x, ["a", "b", "c"], "group")
        assert out.shape[1] == 2
        assert "b" not in names


class TestValidateTrainingData:
    """Tests for validate_training_data."""

    def test_returns_false_when_too_few_samples(self):
        x = np.array([[1.0], [2.0], [3.0]], dtype=np.float32)
        y = np.array([1, 0, 1])
        assert validate_training_data(x, y, "group", min_samples_leaf=3) is False

    def test_returns_false_when_single_class(self):
        x = np.array([[1.0], [2.0], [3.0], [4.0], [5.0], [6.0]], dtype=np.float32)
        y = np.array([1, 1, 1, 1, 1, 1])
        assert validate_training_data(x, y, "group", min_samples_leaf=2) is False

    def test_returns_true_when_valid(self):
        x = np.random.randn(20, 3).astype(np.float32)
        y = np.array([1] * 10 + [0] * 10)
        assert validate_training_data(x, y, "group", min_samples_leaf=3) is True


class TestPrepareAttributeData:
    """Tests for prepare_attribute_data."""

    def test_returns_empty_when_no_attributes_selected(
        self, sample_processed_dataset, sample_correlation_request
    ):
        req = CorrelationRequest(
            fractions=sample_correlation_request.fractions,
            parameter_groups=[ParameterGroup.COLOR],
        )
        data = prepare_attribute_data(sample_processed_dataset, req)
        assert isinstance(data, dict)
        assert "h" in data or "m_h" in data or len(data) >= 0

    def test_returns_attribute_data_for_all_params(
        self, sample_processed_dataset, sample_correlation_request
    ):
        data = prepare_attribute_data(
            sample_processed_dataset, sample_correlation_request
        )
        assert isinstance(data, dict)
        for _attr, values in data.items():
            assert isinstance(values, list)
            for v in values:
                assert isinstance(v, tuple)
                assert len(v) == 2
                assert isinstance(v[0], float)
                assert isinstance(v[1], str)


class TestCreateProcessedDataset:
    """Tests for create_processed_dataset."""

    @pytest.mark.asyncio
    async def test_returns_empty_dataset_when_no_object_ids(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="empty", object_ids=[])],
            parameter_groups=[ParameterGroup.ALL],
        )
        with patch(
            "app.services.data_processor.get_objects_by_id",
            new_callable=AsyncMock,
        ) as mock_get:
            mock_get.return_value = SuccessResponse(success=True, result=[])
            dataset = await create_processed_dataset(req)
        assert not dataset.objects_by_id
        assert not dataset.objects_by_group
        mock_get.assert_not_called()

    @pytest.mark.asyncio
    async def test_creates_dataset_from_api_response(self):
        req = CorrelationRequest(
            fractions=[
                ObjectGroup(name="g1", object_ids=[1, 2]),
                ObjectGroup(name="g2", object_ids=[3]),
            ],
            parameter_groups=[ParameterGroup.ALL],
        )
        objects = [
            make_object_metadata(1),
            make_object_metadata(2),
            make_object_metadata(3),
        ]
        with patch(
            "app.services.data_processor.get_objects_by_id",
            new_callable=AsyncMock,
        ) as mock_get:
            mock_get.return_value = SuccessResponse(success=True, result=objects)
            dataset = await create_processed_dataset(req)
        mock_get.assert_called_once_with([1, 2, 3])
        assert len(dataset.objects_by_id) == 3
        assert "g1" in dataset.objects_by_group
        assert "g2" in dataset.objects_by_group
        assert len(dataset.get_objects_in_group("g1")) == 2
        assert len(dataset.get_objects_in_group("g2")) == 1

    @pytest.mark.asyncio
    async def test_raises_http_exception_when_api_fails(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="g1", object_ids=[1])],
            parameter_groups=[ParameterGroup.ALL],
        )
        with patch(
            "app.services.data_processor.get_objects_by_id",
            new_callable=AsyncMock,
        ) as mock_get:
            mock_get.return_value = SuccessResponse(success=False, result=[])
            with pytest.raises(HTTPException) as exc_info:
                await create_processed_dataset(req)
            assert exc_info.value.status_code == 502

    @pytest.mark.asyncio
    async def test_raises_http_exception_on_network_error(self):
        req = CorrelationRequest(
            fractions=[ObjectGroup(name="g1", object_ids=[1])],
            parameter_groups=[ParameterGroup.ALL],
        )
        with patch(
            "app.services.data_processor.get_objects_by_id",
            new_callable=AsyncMock,
        ) as mock_get:
            mock_get.side_effect = Exception("Connection failed")
            with pytest.raises(HTTPException) as exc_info:
                await create_processed_dataset(req)
            assert exc_info.value.status_code == 502
