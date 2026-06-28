import dataclasses
import logging
from collections.abc import Callable
from typing import Any

import numpy as np
from sklearn.ensemble import RandomForestClassifier
from sklearn.linear_model import LogisticRegression
from sklearn.model_selection import cross_val_score
from sklearn.svm import SVC
from sklearn.tree import DecisionTreeClassifier

from ..models.correlation import Condition, CorrelationRequest
from .data_processor import ProcessedDataset

logger = logging.getLogger(__name__)


@dataclasses.dataclass
class EnhancedConditionTestResult:
    """Enhanced test result with additional metrics and confidence intervals."""

    true_positives: int
    false_positives: int
    true_negatives: int
    false_negatives: int
    precision: float
    recall: float
    accuracy: float
    f1_score: float
    balanced_accuracy: float = 0.0
    auc_score: float = 0.0
    confidence_interval_95: tuple[float, float] = (0.0, 0.0)
    cross_val_scores: list[float] = dataclasses.field(default_factory=list)
    feature_importance: dict[str, float] = dataclasses.field(default_factory=dict)


@dataclasses.dataclass
class AlgorithmResult:
    """Result from a single algorithm with its performance metrics."""

    algorithm_name: str
    conditions: list[Condition]
    test_result: EnhancedConditionTestResult
    feature_importance: dict[str, float]
    model_params: dict


def run_multiple_algorithms(
    x_matrix: np.ndarray,
    y_vector: np.ndarray,
    feature_names: list[str],
    group_name: str,
    dataset: ProcessedDataset,
    data: CorrelationRequest,
    cv_folds: int = 5,
    extract_tree_fn: Callable | None = None,
    extract_linear_fn: Callable | None = None,
    test_enhanced_fn: Callable | None = None,
) -> list[AlgorithmResult]:
    """
    Run multiple algorithms and return results.
    """
    if extract_tree_fn is None or extract_linear_fn is None or test_enhanced_fn is None:
        logger.error(
            "Required extraction or testing functions not provided to run_multiple_algorithms"
        )
        return []

    algorithms: dict[str, dict[str, Any]] = {
        "random_forest": {
            "model": RandomForestClassifier(
                n_estimators=100,
                max_depth=10,
                min_samples_leaf=3,
                class_weight="balanced",
                random_state=42,
                n_jobs=-1,
            ),
            "params": {"n_estimators": 100, "max_depth": 10, "min_samples_leaf": 3},
        },
        "decision_tree": {
            "model": DecisionTreeClassifier(
                max_depth=5,
                min_samples_leaf=3,
                class_weight="balanced",
                random_state=42,
            ),
            "params": {"max_depth": 5, "min_samples_leaf": 3},
        },
        "logistic_regression": {
            "model": LogisticRegression(
                C=1.0,
                class_weight="balanced",
                random_state=42,
                max_iter=2000,
                solver="liblinear",
            ),
            "params": {"C": 1.0, "max_iter": 2000, "solver": "liblinear"},
        },
        "svm": {
            "model": SVC(
                C=1.0,
                kernel="rbf",
                class_weight="balanced",
                probability=True,
                random_state=42,
            ),
            "params": {"C": 1.0, "kernel": "rbf"},
        },
    }

    results = []
    for alg_name, alg_config in algorithms.items():
        try:
            logger.debug(f"Testing {alg_name} for group {group_name}")
            model: Any = alg_config["model"]

            try:
                cv_folds_actual = min(cv_folds, len(np.unique(y_vector)))
                if cv_folds_actual < 2:
                    cv_scores = np.array([0.0])
                else:
                    cv_scores = cross_val_score(
                        model, x_matrix, y_vector, cv=cv_folds_actual, scoring="f1"
                    )
            except Exception as e:
                logger.debug(f"Cross-validation failed for {alg_name}: {e}")
                cv_scores = np.array([0.0])

            model.fit(x_matrix, y_vector)

            if alg_name in ["random_forest", "decision_tree"]:
                conditions = extract_tree_fn(
                    model, feature_names, group_name, dataset, data, alg_name
                )
            else:
                conditions = extract_linear_fn(
                    model, feature_names, group_name, dataset, data
                )

            if not conditions:
                continue

            test_result = test_enhanced_fn(
                conditions, group_name, dataset, data, cv_scores
            )

            if hasattr(model, "feature_importances_"):
                importance = dict(
                    zip(feature_names, model.feature_importances_, strict=False)
                )
            elif hasattr(model, "coef_"):
                coef = (
                    np.abs(model.coef_[0])
                    if len(model.coef_.shape) > 1
                    else np.abs(model.coef_)
                )
                importance = dict(zip(feature_names, coef, strict=False))
            else:
                importance = {}

            results.append(
                AlgorithmResult(
                    algorithm_name=alg_name,
                    conditions=conditions,
                    test_result=test_result,
                    feature_importance=importance,
                    model_params=alg_config["params"],
                )
            )

        except Exception as e:
            logger.warning(f"Algorithm {alg_name} failed: {e}")
            continue

    return results
