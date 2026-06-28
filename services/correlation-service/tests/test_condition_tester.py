"""Unit tests for condition_tester module."""

from app.clients.analysis import ObjectMetadata
from app.models.correlation import (
    Condition,
)
from app.services.condition_tester import (
    calculate_condition_match_score,
    compute_condition_metrics,
    evaluate_conditions,
    meets_all_conditions,
    meets_condition,
    meets_single_condition,
)
from app.services.data_processor import ProcessedDataset, ProcessedObjectData


class TestMeetsSingleCondition:
    """Tests for meets_single_condition."""

    def test_returns_false_when_metadata_is_none(self, sample_condition):
        obj = ProcessedObjectData(object_id=1, metadata=None, group_name="test")
        assert meets_single_condition(obj, sample_condition) is False

    def test_returns_false_when_attribute_value_is_none(self, sample_condition):
        meta = ObjectMetadata(id=1, m_h=None)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, sample_condition) is False

    def test_returns_false_when_attribute_not_numeric(self, sample_condition):
        class FakeMeta:
            id = 1
            m_h = "not_a_number"

        obj = ProcessedObjectData(
            object_id=1,
            metadata=FakeMeta(),
            group_name="test",
        )
        assert meets_single_condition(obj, sample_condition) is False

    def test_operator_greater_than_true(self):
        cond = Condition(attribute="m_h", operator=">", value=0.4)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_greater_than_false(self):
        cond = Condition(attribute="m_h", operator=">", value=0.6)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is False

    def test_operator_greater_equal_true(self):
        cond = Condition(attribute="m_h", operator=">=", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_less_than_true(self):
        cond = Condition(attribute="m_h", operator="<", value=0.8)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_less_equal_true(self):
        cond = Condition(attribute="m_h", operator="<=", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_equals_true(self):
        cond = Condition(attribute="m_h", operator="==", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_equals_with_float_tolerance(self):
        cond = Condition(attribute="m_h", operator="==", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5000001)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_not_equals_true(self):
        cond = Condition(attribute="m_h", operator="!=", value=0.3)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is True

    def test_operator_not_equals_false(self):
        cond = Condition(attribute="m_h", operator="!=", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is False

    def test_unknown_operator_returns_false(self):
        cond = Condition(attribute="m_h", operator="~", value=0.5)
        meta = ObjectMetadata(id=1, m_h=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        assert meets_single_condition(obj, cond) is False


class TestMeetsCondition:
    """Tests for meets_condition (uses ObjectMetadata directly)."""

    def test_returns_false_when_value_none(self):
        cond = Condition(attribute="m_h", operator=">=", value=0.4)
        meta = ObjectMetadata(id=1, m_h=None)
        assert meets_condition(meta, cond) is False

    def test_returns_false_when_value_not_numeric(self):
        class FakeMeta:
            m_h = "string"

        meta = FakeMeta()
        cond = Condition(attribute="m_h", operator=">=", value=0.4)
        assert meets_condition(meta, cond) is False

    def test_handles_exception_returns_false(self):
        class BadObj:
            @property
            def m_h(self):
                raise RuntimeError("bad")

        meta = BadObj()
        cond = Condition(attribute="m_h", operator=">=", value=0.4)
        assert meets_condition(meta, cond) is False


class TestMeetsAllConditions:
    """Tests for meets_all_conditions."""

    def test_returns_false_when_empty_conditions(self, sample_object_metadata):
        obj = ProcessedObjectData(
            object_id=1, metadata=sample_object_metadata, group_name="test"
        )
        assert meets_all_conditions(obj, []) is False

    def test_returns_true_when_all_met(self):
        meta = ObjectMetadata(id=1, m_h=0.6, m_s=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.5),
            Condition(attribute="m_s", operator="<=", value=0.8),
        ]
        assert meets_all_conditions(obj, conditions) is True

    def test_returns_false_when_one_fails(self):
        meta = ObjectMetadata(id=1, m_h=0.3, m_s=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.5),
            Condition(attribute="m_s", operator="<=", value=0.8),
        ]
        assert meets_all_conditions(obj, conditions) is False


class TestCalculateConditionMatchScore:
    """Tests for calculate_condition_match_score."""

    def test_returns_one_when_empty_conditions(self, sample_object_metadata):
        obj = ProcessedObjectData(
            object_id=1, metadata=sample_object_metadata, group_name="test"
        )
        assert calculate_condition_match_score(obj, []) == 1.0

    def test_returns_partial_when_some_met(self):
        meta = ObjectMetadata(id=1, m_h=0.6, m_s=0.9)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.5),
            Condition(attribute="m_s", operator="<=", value=0.8),
        ]
        assert calculate_condition_match_score(obj, conditions) == 0.5

    def test_returns_one_when_all_met(self):
        meta = ObjectMetadata(id=1, m_h=0.6, m_s=0.5)
        obj = ProcessedObjectData(object_id=1, metadata=meta, group_name="test")
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.5),
            Condition(attribute="m_s", operator="<=", value=0.8),
        ]
        assert calculate_condition_match_score(obj, conditions) == 1.0


class TestComputeConditionMetrics:
    """Tests for compute_condition_metrics."""

    def test_returns_zeros_when_empty_conditions(
        self, sample_processed_dataset, sample_correlation_request
    ):
        result = compute_condition_metrics(
            [], "group_a", sample_processed_dataset, sample_correlation_request
        )
        assert result.true_positives == 0
        assert result.false_positives == 0
        assert result.precision == 0.0
        assert result.recall == 0.0
        assert result.f1_score == 0.0

    def test_returns_zeros_when_no_objects(self, sample_correlation_request):
        empty_dataset = ProcessedDataset(
            objects_by_id={},
            objects_by_group={"group_a": [], "group_b": []},
            feature_names=[],
        )
        result = compute_condition_metrics(
            [Condition(attribute="m_h", operator=">=", value=0.0)],
            "group_a",
            empty_dataset,
            sample_correlation_request,
        )
        assert result.true_positives == 0
        assert result.precision == 0.0

    def test_computes_metrics_for_valid_conditions(
        self, sample_processed_dataset, sample_correlation_request
    ):
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.4),
        ]
        result = compute_condition_metrics(
            conditions, "group_a", sample_processed_dataset, sample_correlation_request
        )
        assert result.true_positives >= 0
        assert result.false_positives >= 0
        assert 0 <= result.precision <= 1
        assert 0 <= result.recall <= 1
        assert 0 <= result.accuracy <= 1
        assert 0 <= result.f1_score <= 1


class TestEvaluateConditions:
    """Tests for evaluate_conditions."""

    def test_returns_enhanced_result_with_auc(
        self, sample_processed_dataset, sample_correlation_request
    ):
        conditions = [
            Condition(attribute="m_h", operator=">=", value=0.3),
        ]
        result = evaluate_conditions(
            conditions, "group_a", sample_processed_dataset, sample_correlation_request
        )
        assert hasattr(result, "auc_score")
        assert hasattr(result, "balanced_accuracy")
        assert hasattr(result, "confidence_interval_95")
        assert (
            0 <= result.auc_score <= 1 or result.auc_score == result.balanced_accuracy
        )

    def test_handles_empty_target_or_other_groups(self, sample_correlation_request):
        dataset = ProcessedDataset(
            objects_by_id={},
            objects_by_group={"group_a": [], "group_b": []},
            feature_names=[],
        )
        result = evaluate_conditions([], "group_a", dataset, sample_correlation_request)
        assert result.true_positives == 0
        assert result.f1_score == 0.0
