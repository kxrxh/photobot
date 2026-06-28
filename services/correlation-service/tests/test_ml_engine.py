"""Unit tests for ml_engine module."""

import numpy as np
import pytest

from app.models.correlation import (
    Condition,
    CorrelationRequest,
    ObjectGroup,
    ParameterGroup,
)
from app.services.data_processor import ProcessedDataset, ProcessedObjectData
from app.services.ml_engine import (
    AlgorithmResult,
    EnhancedConditionTestResult,
    run_multiple_algorithms,
)


def make_object_metadata(obj_id: int, **kwargs):
    from .conftest import make_object_metadata as _make

    return _make(obj_id, **kwargs)


@pytest.fixture
def ml_dataset():
    objs = []
    for i in range(30):
        meta = make_object_metadata(
            i,
            m_h=0.4 + (i % 2) * 0.3,
            m_s=0.3 + (i % 3) * 0.15,
            m_v=0.6,
        )
        group = "group_a" if i < 15 else "group_b"
        objs.append(ProcessedObjectData(object_id=i, metadata=meta, group_name=group))
    return ProcessedDataset(
        objects_by_id={o.object_id: o for o in objs},
        objects_by_group={"group_a": objs[:15], "group_b": objs[15:]},
        feature_names=["m_h", "m_s", "m_v"],
    )


@pytest.fixture
def ml_request():
    return CorrelationRequest(
        fractions=[
            ObjectGroup(name="group_a", object_ids=list(range(15))),
            ObjectGroup(name="group_b", object_ids=list(range(15, 30))),
        ],
        parameter_groups=[ParameterGroup.ALL],
    )


def _mock_extract_tree(model, features, group, dataset, data, algo_name=None):
    return [Condition(attribute="m_h", operator=">=", value=0.5)]


def _mock_extract_linear(model, features, group, dataset, data):
    return [Condition(attribute="m_s", operator="<=", value=0.6)]


def _mock_test_enhanced(conditions, group, dataset, data, cv_scores=None):
    return EnhancedConditionTestResult(
        true_positives=5,
        false_positives=2,
        true_negatives=8,
        false_negatives=5,
        precision=0.71,
        recall=0.5,
        accuracy=0.65,
        f1_score=0.59,
    )


class TestRunMultipleAlgorithms:
    """Tests for run_multiple_algorithms."""

    def test_returns_empty_when_extract_fns_none(self, ml_dataset, ml_request):
        x = np.array([[0.5, 0.3, 0.6]] * 15 + [[0.2, 0.8, 0.3]] * 15, dtype=np.float32)
        y = np.array([1] * 15 + [0] * 15)
        result = run_multiple_algorithms(
            x,
            y,
            ["m_h", "m_s", "m_v"],
            "group_a",
            ml_dataset,
            ml_request,
            extract_tree_fn=None,
            extract_linear_fn=None,
            test_enhanced_fn=None,
        )
        assert result == []

    def test_returns_results_when_fns_provided(self, ml_dataset, ml_request):
        x = np.array(
            [[0.6, 0.3, 0.6]] * 10 + [[0.3, 0.7, 0.3]] * 10 + [[0.5, 0.5, 0.5]] * 10,
            dtype=np.float32,
        )
        y = np.array([1] * 10 + [0] * 10 + [1] * 10)
        result = run_multiple_algorithms(
            x,
            y,
            ["m_h", "m_s", "m_v"],
            "group_a",
            ml_dataset,
            ml_request,
            extract_tree_fn=_mock_extract_tree,
            extract_linear_fn=_mock_extract_linear,
            test_enhanced_fn=_mock_test_enhanced,
        )
        assert isinstance(result, list)
        for r in result:
            assert isinstance(r, AlgorithmResult)
            assert r.algorithm_name in [
                "random_forest",
                "decision_tree",
                "logistic_regression",
                "svm",
            ]
            assert isinstance(r.conditions, list)
            assert isinstance(r.test_result, EnhancedConditionTestResult)
