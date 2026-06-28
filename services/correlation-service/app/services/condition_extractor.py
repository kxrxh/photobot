import logging
from collections.abc import Callable
from typing import Any

import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.tree import DecisionTreeClassifier

from ..models.correlation import Condition, CorrelationRequest
from .condition_tester import compute_condition_metrics
from .data_processor import ProcessedDataset, build_feature_matrix

logger = logging.getLogger(__name__)

# scikit-learn tree constants
TREE_UNDEFINED = -2
TREE_LEAF = -1

# Re-define constants for consistency
MIN_F1_LOOSE = 0.10
MIN_PRECISION = 0.10
MIN_RECALL_LOOSE = 0.05
F_BETA = 2.0


def _fbeta(precision: float, recall: float, beta: float = F_BETA) -> float:
    if precision <= 0.0 or recall <= 0.0:
        return 0.0
    b2 = beta * beta
    denom = (b2 * precision) + recall
    return ((1.0 + b2) * precision * recall / denom) if denom > 0.0 else 0.0


def extract_tree_conditions(
    model: Any,
    feature_names: list[str],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
    algorithm_name: str = "decision_tree",
    compute_metrics_fn: Callable | None = None,
) -> list[Condition]:
    """
    Extract conditions from tree models (decision tree or random forest).
    """
    metrics_fn = compute_metrics_fn or compute_condition_metrics
    try:
        if algorithm_name == "random_forest":
            return extract_random_forest_conditions(
                model, feature_names, group_name, dataset, data, metrics_fn
            )
        return extract_single_tree_conditions(
            model, feature_names, group_name, dataset, data, metrics_fn
        )
    except Exception as e:
        logger.error(f"Error extracting tree conditions: {e}")
        return []


def extract_random_forest_conditions(
    rf: RandomForestClassifier,
    feature_names: list[str],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
    compute_metrics_fn: Callable | None = None,
) -> list[Condition]:
    """
    Extract conditions from Random Forest.
    """
    try:
        best_conditions = []
        best_score = (-1.0, -1.0, -1.0, float("inf"))

        for tree in rf.estimators_[:10]:
            try:
                conditions = extract_single_tree_conditions(
                    tree, feature_names, group_name, dataset, data, compute_metrics_fn
                )
                if not conditions:
                    continue

                test_result = compute_metrics_fn(conditions, group_name, dataset, data)
                path_score = (
                    _fbeta(test_result.precision, test_result.recall),
                    test_result.precision,
                    test_result.recall,
                    -len(conditions),
                )

                if path_score > best_score:
                    best_score = path_score
                    best_conditions = conditions
            except Exception as e:
                logger.debug(f"Error evaluating tree estimator: {e}")
                continue

        if best_conditions:
            score, prec, rec, _ = best_score
            if score < MIN_F1_LOOSE or prec < MIN_PRECISION or rec < MIN_RECALL_LOOSE:
                return []

        return best_conditions
    except Exception as e:
        logger.error(f"Error extracting RF conditions: {e}")
        return []


def extract_single_tree_conditions(
    dt: DecisionTreeClassifier,
    feature_names: list[str],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
    compute_metrics_fn: Callable | None = None,
) -> list[Condition]:
    """
    Extract conditions from a single decision tree.
    """
    metrics_fn = compute_metrics_fn or compute_condition_metrics
    tree_ = dt.tree_
    if tree_.node_count <= 1:
        return []

    try:
        target_class_index = list(dt.classes_).index(1)
    except ValueError:
        return []

    best_conditions = []
    best_score = (-1.0, -1.0, -1.0, float("inf"), float("inf"))

    def find_paths(node_id: int, current_conditions: list[Condition], depth: int = 0):
        nonlocal best_conditions, best_score
        if depth > 20:
            return

        feature_index = tree_.feature[node_id]
        if feature_index == TREE_UNDEFINED:
            leaf_values = tree_.value[node_id][0]
            predicted_class_index = np.argmax(leaf_values)

            if (
                predicted_class_index == target_class_index
                and leaf_values[target_class_index] > 0
            ):
                test_res = metrics_fn(current_conditions, group_name, dataset, data)
                path_score = (
                    _fbeta(test_res.precision, test_res.recall),
                    test_res.precision,
                    test_res.recall,
                    -len(current_conditions),
                    -depth,
                )

                if path_score > best_score:
                    best_score = path_score
                    best_conditions = current_conditions[:]
            return

        threshold = tree_.threshold[node_id]
        attr = feature_names[feature_index]
        find_paths(
            tree_.children_left[node_id],
            [
                *current_conditions,
                Condition(attribute=attr, operator="<=", value=float(threshold)),
            ],
            depth + 1,
        )
        find_paths(
            tree_.children_right[node_id],
            [
                *current_conditions,
                Condition(attribute=attr, operator=">", value=float(threshold)),
            ],
            depth + 1,
        )

    find_paths(0, [])
    if best_conditions:
        score, prec, rec, _, _ = best_score
        if score < MIN_F1_LOOSE or prec < MIN_PRECISION or rec < MIN_RECALL_LOOSE:
            return []

    return best_conditions


def extract_linear_conditions(
    model: Any,
    feature_names: list[str],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
) -> list[Condition]:
    """
    Extract conditions from linear models.
    """
    try:
        if not hasattr(model, "coef_"):
            return []

        coef = model.coef_[0] if len(model.coef_.shape) > 1 else model.coef_
        importance = np.abs(coef)
        top_indices = np.argsort(importance)[-5:][::-1]

        x_matrix, y_vector = build_feature_matrix(dataset, group_name, feature_names)[
            :2
        ]
        if x_matrix is None:
            return []

        # Real-world: greedy selection optimizing recall-weighted F2 (not just separation).
        # This intentionally allows some false positives if recall improves meaningfully.
        candidate_conditions: list[Condition] = []

        percentiles = [5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95]
        for idx in top_indices:
            if importance[idx] < 0.05:
                continue

            f_name = feature_names[idx]
            f_values = x_matrix[:, idx]
            f_coef = float(coef[idx])
            target_vals = f_values[y_vector == 1]
            other_vals = f_values[y_vector == 0]
            if len(target_vals) == 0 or len(other_vals) == 0:
                continue

            op = ">=" if f_coef > 0 else "<="
            for p in percentiles:
                try:
                    t = float(np.percentile(target_vals, p))
                except Exception as e:
                    logger.debug(f"Error calculating percentile {p} for {f_name}: {e}")
                    continue
                candidate_conditions.append(
                    Condition(attribute=f_name, operator=op, value=t)
                )

        if not candidate_conditions:
            return []

        selected: list[Condition] = []
        best_score = 0.0

        # Allow up to 2 conditions; more tends to kill recall in practice.
        for _ in range(2):
            best_c = None
            best_c_score = best_score
            for c in candidate_conditions:
                if c in selected:
                    continue
                trial = [*selected, c]
                res = compute_condition_metrics(trial, group_name, dataset, data)
                score = _fbeta(res.precision, res.recall)
                if score > best_c_score:
                    best_c_score = score
                    best_c = c
            if best_c is None:
                break
            selected.append(best_c)
            best_score = best_c_score

        return selected
    except Exception as e:
        logger.error(f"Error extracting linear conditions: {e}")
        return []
