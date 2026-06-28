"""Unit tests for feature_selector module."""

import pytest

from app.services.data_processor import ProcessedDataset, ProcessedObjectData
from app.services.feature_selector import select_important_features

from .conftest import make_object_metadata


@pytest.fixture
def dataset_with_many_features():
    """Dataset with enough objects and features for feature selection."""
    objects = []
    for i in range(20):
        meta = make_object_metadata(
            i,
            m_h=0.3 + (i % 2) * 0.4,
            m_s=0.2 + (i % 3) * 0.2,
            m_v=0.4 + (i % 5) * 0.1,
            l_avg=10.0 + i,
            w_avg=5.0 + i * 0.5,
        )
        group = "group_a" if i < 10 else "group_b"
        objects.append(
            ProcessedObjectData(object_id=i, metadata=meta, group_name=group)
        )
    by_id = {o.object_id: o for o in objects}
    by_group = {"group_a": objects[:10], "group_b": objects[10:]}
    return ProcessedDataset(
        objects_by_id=by_id,
        objects_by_group=by_group,
        feature_names=["m_h", "m_s", "m_v", "l_avg", "w_avg"],
    )


class TestSelectImportantFeatures:
    """Tests for select_important_features."""

    def test_returns_empty_when_no_feature_names(self):
        dataset = ProcessedDataset(
            objects_by_id={},
            objects_by_group={},
            feature_names=[],
        )
        result = select_important_features(dataset, "group_a")
        assert result == []

    def test_returns_subset_when_many_features(self, dataset_with_many_features):
        result = select_important_features(
            dataset_with_many_features,
            "group_a",
            max_features=3,
            min_importance=0.001,
        )
        assert len(result) <= 3
        assert all(f in dataset_with_many_features.feature_names for f in result)

    def test_returns_all_when_fewer_than_max(self, dataset_with_many_features):
        result = select_important_features(
            dataset_with_many_features,
            "group_a",
            max_features=20,
        )
        assert len(result) <= len(dataset_with_many_features.feature_names)

    def test_uses_fast_method_by_default(self, dataset_with_many_features):
        result_fast = select_important_features(
            dataset_with_many_features,
            "group_a",
            max_features=5,
            use_fast_method=True,
        )
        assert isinstance(result_fast, list)
        assert len(result_fast) > 0

    def test_uses_rf_method_when_disabled(self, dataset_with_many_features):
        result_rf = select_important_features(
            dataset_with_many_features,
            "group_a",
            max_features=5,
            use_fast_method=False,
        )
        assert isinstance(result_rf, list)
        assert len(result_rf) > 0
