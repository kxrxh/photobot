"""Unit tests for condition_extractor module."""

import numpy as np
import pytest
from sklearn.ensemble import RandomForestClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.tree import DecisionTreeClassifier

from app.models.correlation import (
    Condition,
    CorrelationRequest,
    ObjectGroup,
    ParameterGroup,
)
from app.services.condition_extractor import (
    extract_linear_conditions,
    extract_tree_conditions,
)
from app.services.condition_tester import compute_condition_metrics
from app.services.data_processor import ProcessedDataset, ProcessedObjectData


def make_object_metadata(obj_id: int, **kwargs):
    from .conftest import make_object_metadata as _make

    return _make(obj_id, **kwargs)


@pytest.fixture
def extractor_dataset():
    objs = []
    for i in range(30):
        meta = make_object_metadata(
            i,
            m_h=0.5 + (i % 2) * 0.3,
            m_s=0.4,
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
def extractor_request():
    return CorrelationRequest(
        fractions=[
            ObjectGroup(name="group_a", object_ids=list(range(15))),
            ObjectGroup(name="group_b", object_ids=list(range(15, 30))),
        ],
        parameter_groups=[ParameterGroup.ALL],
    )


class TestExtractTreeConditions:
    """Tests for extract_tree_conditions."""

    def test_returns_empty_for_trivial_tree(self, extractor_dataset, extractor_request):
        dt = DecisionTreeClassifier(max_depth=1, min_samples_leaf=10, random_state=42)
        x = np.array([[0.6, 0.4, 0.6]] * 15 + [[0.2, 0.4, 0.3]] * 15, dtype=np.float32)
        y = np.array([1] * 15 + [0] * 15)
        dt.fit(x, y)
        result = extract_tree_conditions(
            dt,
            ["m_h", "m_s", "m_v"],
            "group_a",
            extractor_dataset,
            extractor_request,
            algorithm_name="decision_tree",
            compute_metrics_fn=compute_condition_metrics,
        )
        assert isinstance(result, list)

    def test_returns_conditions_for_fitted_tree(
        self, extractor_dataset, extractor_request
    ):
        dt = DecisionTreeClassifier(max_depth=3, min_samples_leaf=2, random_state=42)
        x = np.array(
            [[0.7, 0.3, 0.6]] * 10 + [[0.2, 0.7, 0.3]] * 10 + [[0.5, 0.5, 0.5]] * 10,
            dtype=np.float32,
        )
        y = np.array([1] * 10 + [0] * 10 + [1] * 5 + [0] * 5)
        dt.fit(x, y)
        result = extract_tree_conditions(
            dt,
            ["m_h", "m_s", "m_v"],
            "group_a",
            extractor_dataset,
            extractor_request,
            algorithm_name="decision_tree",
            compute_metrics_fn=compute_condition_metrics,
        )
        assert isinstance(result, list)
        for c in result:
            assert isinstance(c, Condition)
            assert c.attribute in ["m_h", "m_s", "m_v"]
            assert c.operator in [">", "<="]

    def test_random_forest_returns_conditions(
        self, extractor_dataset, extractor_request
    ):
        rf = RandomForestClassifier(
            n_estimators=5, max_depth=3, min_samples_leaf=2, random_state=42
        )
        x = np.array(
            [[0.7, 0.3, 0.6]] * 15 + [[0.2, 0.7, 0.3]] * 15,
            dtype=np.float32,
        )
        y = np.array([1] * 15 + [0] * 15)
        rf.fit(x, y)
        result = extract_tree_conditions(
            rf,
            ["m_h", "m_s", "m_v"],
            "group_a",
            extractor_dataset,
            extractor_request,
            algorithm_name="random_forest",
            compute_metrics_fn=compute_condition_metrics,
        )
        assert isinstance(result, list)


class TestExtractLinearConditions:
    """Tests for extract_linear_conditions."""

    def test_returns_empty_when_no_coef(self):
        model = object()
        dataset = ProcessedDataset(
            objects_by_id={},
            objects_by_group={"a": [], "b": []},
            feature_names=[],
        )
        req = CorrelationRequest(
            fractions=[
                ObjectGroup(name="a", object_ids=[]),
                ObjectGroup(name="b", object_ids=[]),
            ],
            parameter_groups=[ParameterGroup.ALL],
        )
        result = extract_linear_conditions(model, [], "a", dataset, req)
        assert result == []

    def test_returns_conditions_for_logistic_regression(
        self, extractor_dataset, extractor_request
    ):
        lr = LogisticRegression(random_state=42, max_iter=500)
        x = np.array(
            [[0.7, 0.3, 0.6]] * 15 + [[0.2, 0.7, 0.3]] * 15,
            dtype=np.float32,
        )
        y = np.array([1] * 15 + [0] * 15)
        lr.fit(x, y)
        result = extract_linear_conditions(
            lr, ["m_h", "m_s", "m_v"], "group_a", extractor_dataset, extractor_request
        )
        assert isinstance(result, list)
        for c in result:
            assert isinstance(c, Condition)
