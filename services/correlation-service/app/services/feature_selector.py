import logging

from sklearn.ensemble import RandomForestClassifier
from sklearn.feature_selection import SelectKBest, f_classif

from .data_processor import (
    ProcessedDataset,
    build_feature_matrix,
    handle_missing_values,
)

logger = logging.getLogger(__name__)


def select_important_features(
    dataset: ProcessedDataset,
    group_name: str,
    max_features: int = 20,
    min_importance: float = 0.01,
    use_fast_method: bool = True,
) -> list[str]:
    """
    Select the most important features.
    """
    try:
        if not dataset.feature_names:
            return []

        x_matrix, y_vector = build_feature_matrix(
            dataset, group_name, dataset.feature_names
        )[:2]
        if x_matrix is None or x_matrix.shape[0] < 10:
            return dataset.feature_names[:max_features]

        x_imputed, f_names = handle_missing_values(
            x_matrix, dataset.feature_names, group_name
        )
        if x_imputed.shape[1] == 0:
            return []

        if use_fast_method:
            k = min(max_features, x_imputed.shape[1])
            selector = SelectKBest(score_func=f_classif, k=k)
            selector.fit(x_imputed, y_vector)
            selected_mask = selector.get_support()
            feature_scores = selector.scores_
            f_importance = {
                f_names[i]: feature_scores[i] if selected_mask[i] else 0.0
                for i in range(len(f_names))
            }
            sorted_f = sorted(f_importance.items(), key=lambda x: x[1], reverse=True)
            return [item[0] for item in sorted_f[:max_features]]
        rf = RandomForestClassifier(
            n_estimators=50,
            max_depth=8,
            min_samples_leaf=3,
            class_weight="balanced",
            random_state=42,
            n_jobs=-1,
        )
        rf.fit(x_imputed, y_vector)
        f_importance = dict(zip(f_names, rf.feature_importances_, strict=False))
        sorted_f = sorted(f_importance.items(), key=lambda x: x[1], reverse=True)
        selected = [f for f, imp in sorted_f if imp >= min_importance]
        if not selected:
            selected = [item[0] for item in sorted_f[:max_features]]
        return selected[:max_features]

    except Exception as e:
        logger.warning(f"Feature selection failed for {group_name}: {e}")
        return dataset.feature_names[:max_features]
