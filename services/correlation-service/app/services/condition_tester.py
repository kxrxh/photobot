import logging

import numpy as np
from sklearn.metrics import roc_auc_score

from ..clients.analysis import ObjectMetadata
from ..models.correlation import Condition, ConditionTestResult, CorrelationRequest
from .data_processor import ProcessedDataset, ProcessedObjectData
from .ml_engine import EnhancedConditionTestResult

logger = logging.getLogger(__name__)


def compute_condition_metrics(
    conditions: list[Condition],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
) -> ConditionTestResult:
    """
    Compute precision, recall, F1, and accuracy for conditions against dataset.
    """
    try:
        if not conditions:
            return ConditionTestResult(
                true_positives=0,
                false_positives=0,
                true_negatives=0,
                false_negatives=0,
                precision=0.0,
                recall=0.0,
                accuracy=0.0,
                f1_score=0.0,
            )

        target_objects = dataset.get_objects_in_group(group_name)
        other_objects = []
        for group in data.fractions:
            if group.name != group_name:
                other_objects.extend(dataset.get_objects_in_group(group.name))

        if not target_objects and not other_objects:
            return ConditionTestResult(
                true_positives=0,
                false_positives=0,
                true_negatives=0,
                false_negatives=0,
                precision=0.0,
                recall=0.0,
                accuracy=0.0,
                f1_score=0.0,
            )

        tp = sum(1 for obj in target_objects if meets_all_conditions(obj, conditions))
        fn = len(target_objects) - tp
        fp = sum(1 for obj in other_objects if meets_all_conditions(obj, conditions))
        tn = len(other_objects) - fp

        precision = tp / (tp + fp) if (tp + fp) > 0 else 0.0
        recall = tp / (tp + fn) if (tp + fn) > 0 else 0.0
        f1_score = (
            2 * precision * recall / (precision + recall)
            if (precision + recall) > 0
            else 0.0
        )
        total = tp + tn + fp + fn
        accuracy = (tp + tn) / total if total > 0 else 0.0

        return ConditionTestResult(
            true_positives=tp,
            false_positives=fp,
            true_negatives=tn,
            false_negatives=fn,
            precision=precision,
            recall=recall,
            accuracy=accuracy,
            f1_score=f1_score,
        )
    except Exception as e:
        logger.error(f"Error testing conditions: {e}")
        return ConditionTestResult(
            true_positives=0,
            false_positives=0,
            true_negatives=0,
            false_negatives=0,
            precision=0.0,
            recall=0.0,
            accuracy=0.0,
            f1_score=0.0,
        )


def evaluate_conditions(
    conditions: list[Condition],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
    cv_scores: list[float] | None = None,
) -> EnhancedConditionTestResult:
    """
    Evaluate conditions with extended metrics (AUC, balanced accuracy, confidence).
    """
    try:
        basic_result = compute_condition_metrics(conditions, group_name, dataset, data)
        target_objects = dataset.get_objects_in_group(group_name)
        other_objects = []
        for group in data.fractions:
            if group.name != group_name:
                other_objects.extend(dataset.get_objects_in_group(group.name))

        if not target_objects or not other_objects:
            return EnhancedConditionTestResult(
                true_positives=basic_result.true_positives,
                false_positives=basic_result.false_positives,
                true_negatives=basic_result.true_negatives,
                false_negatives=basic_result.false_negatives,
                precision=basic_result.precision,
                recall=basic_result.recall,
                accuracy=basic_result.accuracy,
                f1_score=basic_result.f1_score,
            )

        sensitivity = basic_result.recall
        denominator = basic_result.true_negatives + basic_result.false_positives
        specificity = (
            basic_result.true_negatives / denominator if denominator > 0 else 0.0
        )
        balanced_accuracy = (sensitivity + specificity) / 2

        y_true = np.array([1] * len(target_objects) + [0] * len(other_objects))
        y_pred_scores = np.array(
            [
                calculate_condition_match_score(obj, conditions)
                for obj in target_objects + other_objects
            ]
        )

        auc_score = balanced_accuracy
        try:
            if (
                len(y_pred_scores) >= 3
                and len(np.unique(y_pred_scores)) > 1
                and not (
                    np.any(np.isnan(y_pred_scores))
                    or np.any(np.isinf(y_pred_scores))
                    or len(np.unique(y_true)) < 2
                )
            ):
                auc_score = roc_auc_score(y_true, y_pred_scores)
        except Exception as e:
            logger.debug(f"Error calculating ROC AUC: {e}")

        confidence_interval = (basic_result.f1_score, basic_result.f1_score)
        cross_val_list = []
        if cv_scores is not None and len(cv_scores) > 1:
            try:
                arr = np.array(cv_scores)
                if (
                    arr.size > 1
                    and not np.isnan(arr).all()
                    and not np.any(np.isinf(arr))
                ):
                    mean, std = np.nanmean(arr), np.nanstd(arr)
                    if not (np.isnan(mean) or np.isnan(std) or std == 0):
                        confidence_interval = (
                            max(0, mean - 1.96 * std),
                            min(1, mean + 1.96 * std),
                        )
                cross_val_list = arr.tolist()
            except Exception as e:
                logger.debug(f"Error processing CV scores: {e}")

        return EnhancedConditionTestResult(
            true_positives=basic_result.true_positives,
            false_positives=basic_result.false_positives,
            true_negatives=basic_result.true_negatives,
            false_negatives=basic_result.false_negatives,
            precision=basic_result.precision,
            recall=basic_result.recall,
            accuracy=basic_result.accuracy,
            f1_score=basic_result.f1_score,
            balanced_accuracy=balanced_accuracy,
            auc_score=auc_score,
            confidence_interval_95=confidence_interval,
            cross_val_scores=cross_val_list,
            feature_importance={},
        )
    except Exception as e:
        logger.error(f"Error in enhanced testing: {e}")
        return EnhancedConditionTestResult(
            true_positives=0,
            false_positives=0,
            true_negatives=0,
            false_negatives=0,
            precision=0.0,
            recall=0.0,
            accuracy=0.0,
            f1_score=0.0,
        )


def calculate_condition_match_score(
    processed_obj: ProcessedObjectData, conditions: list[Condition]
) -> float:
    if not conditions:
        return 1.0
    met = sum(1 for c in conditions if meets_single_condition(processed_obj, c))
    return met / len(conditions)


def meets_single_condition(
    processed_obj: ProcessedObjectData, condition: Condition
) -> bool:
    if processed_obj.metadata is None:
        return False
    val = getattr(processed_obj.metadata, condition.attribute, None)
    if val is None or not isinstance(val, (int, float)):
        return False

    if condition.operator == ">":
        return val > condition.value
    if condition.operator == ">=":
        return val >= condition.value
    if condition.operator == "<":
        return val < condition.value
    if condition.operator == "<=":
        return val <= condition.value
    if condition.operator == "==":
        return abs(val - condition.value) < 1e-6
    if condition.operator == "!=":
        return abs(val - condition.value) >= 1e-6
    return False


def meets_all_conditions(
    processed_obj: ProcessedObjectData, conditions: list[Condition]
) -> bool:
    if not conditions:
        return False
    obj = processed_obj.metadata
    return all(meets_condition(obj, c) for c in conditions)


def meets_condition(obj: ObjectMetadata, condition: Condition) -> bool:
    try:
        val = getattr(obj, condition.attribute, None)
        if val is None or not isinstance(val, (int, float)):
            return False
        v, t = float(val), float(condition.value)
        if condition.operator == ">":
            return v > t
        if condition.operator == ">=":
            return v >= t
        if condition.operator == "<":
            return v < t
        if condition.operator == "<=":
            return v <= t
        if condition.operator == "==":
            return abs(v - t) < 1e-6
        if condition.operator == "!=":
            return abs(v - t) >= 1e-6
        return False
    except Exception:
        return False
