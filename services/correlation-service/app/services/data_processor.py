import dataclasses
import logging

import numpy as np
from fastapi import HTTPException, status

from ..clients.analysis import ObjectMetadata, get_objects_by_id
from ..models.correlation import CorrelationRequest, ParameterGroup

logger = logging.getLogger(__name__)


@dataclasses.dataclass
class ProcessedObjectData:
    """Efficient data structure for processed object information."""

    object_id: int
    metadata: ObjectMetadata
    feature_vector: np.ndarray | None = None
    group_name: str | None = None
    is_valid: bool = True


@dataclasses.dataclass
class ProcessedDataset:
    """Container for all processed data with efficient lookups."""

    objects_by_id: dict[int, ProcessedObjectData]
    objects_by_group: dict[str, list[ProcessedObjectData]]
    feature_names: list[str]
    feature_matrix: np.ndarray | None = None
    labels: np.ndarray | None = None
    object_ids_order: list[int] | None = None

    def get_objects_in_group(self, group_name: str) -> list[ProcessedObjectData]:
        """Get all objects in a specific group."""
        return self.objects_by_group.get(group_name, [])

    def get_valid_objects(self) -> list[ProcessedObjectData]:
        """Get all valid objects."""
        return [obj for obj in self.objects_by_id.values() if obj.is_valid]


async def create_processed_dataset(data: CorrelationRequest) -> ProcessedDataset:
    """
    Create a processed dataset in a single pass to reduce redundant processing.
    """
    all_object_ids: set[int] = set()
    group_membership: dict[int, str] = {}

    for group in data.fractions:
        all_object_ids.update(group.object_ids)
        for obj_id in group.object_ids:
            group_membership[obj_id] = group.name

    if not all_object_ids:
        return ProcessedDataset(objects_by_id={}, objects_by_group={}, feature_names=[])

    try:
        response = await get_objects_by_id(list(all_object_ids))
        if not response.success:
            raise Exception("Analysis API did not return success.")

        objects_dict: dict[int, ObjectMetadata] = {
            obj.id: obj for obj in response.result
        }

        objects_by_id: dict[int, ProcessedObjectData] = {}
        objects_by_group: dict[str, list[ProcessedObjectData]] = {}
        missing_ids = all_object_ids - set(objects_dict.keys())

        if missing_ids:
            logger.warning(
                f"Could not find or process data for object IDs: {missing_ids}"
            )

        for group in data.fractions:
            objects_by_group[group.name] = []

        for obj_id in all_object_ids:
            if obj_id in objects_dict:
                metadata = objects_dict[obj_id]
                group_name = group_membership.get(obj_id)

                processed_obj = ProcessedObjectData(
                    object_id=obj_id,
                    metadata=metadata,
                    group_name=group_name,
                    is_valid=True,
                )

                objects_by_id[obj_id] = processed_obj

                if group_name and group_name in objects_by_group:
                    objects_by_group[group_name].append(processed_obj)

        logger.info(
            f"Successfully processed dataset with {len(objects_by_id)} objects "
            f"across {len(objects_by_group)} groups."
        )

        return ProcessedDataset(
            objects_by_id=objects_by_id,
            objects_by_group=objects_by_group,
            feature_names=[],
        )

    except Exception as e:
        logger.error(f"Error creating processed dataset: {e!s}", exc_info=True)
        raise HTTPException(
            status_code=status.HTTP_502_BAD_GATEWAY,
            detail=f"Failed to create processed dataset: {e!s}",
        ) from e


def build_feature_matrix(
    dataset: ProcessedDataset, target_group: str, feature_names: list[str]
) -> tuple[np.ndarray | None, np.ndarray | None, list[int]]:
    """
    Efficiently build feature matrix and labels from ProcessedDataset.
    """
    try:
        valid_objects = dataset.get_valid_objects()
        if not valid_objects:
            return None, None, []

        n_objects = len(valid_objects)
        n_features = len(feature_names)

        x_matrix = np.empty((n_objects, n_features), dtype=np.float32)
        y_vector = np.empty(n_objects, dtype=np.int8)
        object_ids_in_order = np.empty(n_objects, dtype=np.int64)

        attr_getters = {
            attr: lambda obj, a=attr: getattr(obj, a, None) for attr in feature_names
        }

        for i, processed_obj in enumerate(valid_objects):
            obj = processed_obj.metadata
            object_ids_in_order[i] = processed_obj.object_id

            for j, attr in enumerate(feature_names):
                value = attr_getters[attr](obj)
                x_matrix[i, j] = (
                    float(value)
                    if value is not None and isinstance(value, (int, float))
                    else np.nan
                )

            y_vector[i] = processed_obj.group_name == target_group

        return x_matrix, y_vector, object_ids_in_order.tolist()

    except Exception as e:
        logger.error(f"Error building feature matrix: {e!s}")
        return None, None, []


def handle_missing_values(
    x_matrix: np.ndarray, feature_names: list[str], group_name: str
) -> tuple[np.ndarray, list[str]]:
    """
    Handle missing values in feature matrix.
    """
    if not np.isnan(x_matrix).any():
        return x_matrix, feature_names

    logger.debug(f"NaNs detected in feature matrix for group {group_name}. Imputing.")

    try:
        from sklearn.impute import SimpleImputer

        if x_matrix.shape[0] >= 2:
            nan_all_cols = np.all(np.isnan(x_matrix), axis=0)
            if np.any(nan_all_cols):
                logger.warning(
                    f"Removing {np.sum(nan_all_cols)} all-NaN columns for group {group_name}"
                )
                x_matrix = np.delete(x_matrix, np.where(nan_all_cols)[0], axis=1)
                feature_names = [
                    name for i, name in enumerate(feature_names) if not nan_all_cols[i]
                ]

                if x_matrix.shape[1] == 0:
                    return np.array([]).reshape(0, 0), []

            if np.isnan(x_matrix).any() and x_matrix.shape[1] >= 1:
                imputer = SimpleImputer(strategy="mean")
                x_matrix = imputer.fit_transform(x_matrix)
            else:
                x_matrix = np.nan_to_num(x_matrix, nan=0.0)
        else:
            x_matrix = np.nan_to_num(x_matrix, nan=0.0)

    except ImportError:
        logger.error("Scikit-learn SimpleImputer not available.")
        x_matrix = np.nan_to_num(x_matrix, nan=0.0)
    except Exception as e:
        logger.error(f"Error during imputation: {e}")
        x_matrix = np.nan_to_num(x_matrix, nan=0.0)

    return x_matrix, feature_names


def validate_training_data(
    x_matrix: np.ndarray, y_vector: np.ndarray, group_name: str, min_samples_leaf: int
) -> bool:
    """
    Validate that training data meets minimum requirements.
    """
    min_samples = max(min_samples_leaf * 2, 5)

    if x_matrix.shape[0] < min_samples:
        logger.warning(
            f"Not enough samples ({x_matrix.shape[0]}) for group {group_name}. Min required: {min_samples}"
        )
        return False

    if len(np.unique(y_vector)) < 2:
        logger.warning(f"Only one class present for group {group_name}. Cannot train.")
        return False

    return True


def prepare_attribute_data(
    dataset: ProcessedDataset, data: CorrelationRequest
) -> dict[str, list[tuple[float, str]]]:
    """
    Prepare attribute data mapping attribute names to (value, group_name) tuples.
    """
    parameter_groups_data = {
        "color": frozenset(
            [
                "h",
                "s",
                "v",
                "r",
                "g",
                "b",
                "m_h",
                "m_s",
                "m_v",
                "m_r",
                "m_g",
                "m_b",
                "min_h",
                "min_s",
                "min_v",
                "max_h",
                "max_s",
                "max_v",
            ]
        ),
        "geometry": frozenset(
            [
                "geometry",
                "l",
                "w",
                "l_w",
                "sq",
                "sq_sqcrl",
                "pr",
                "solid",
                "corners",
                "hu1",
                "hu2",
                "hu3",
                "hu4",
                "hu5",
                "hu6",
            ]
        ),
        "median": frozenset(
            [
                "h_m",
                "s_m",
                "v_m",
                "r_m",
                "g_m",
                "b_m",
                "h_avg",
                "s_avg",
                "v_avg",
                "r_avg",
                "g_avg",
                "b_avg",
                "brt_m",
                "brt_avg",
                "l_avg",
                "l_m",
                "w_avg",
                "w_m",
            ]
        ),
    }

    ignore_fields_data = frozenset(
        ["id", "geometry", "id_analysis", "id_image", "file", "color_rhs", "class"]
    )

    all_attributes = []
    for field, field_info in ObjectMetadata.model_fields.items():
        if field in ignore_fields_data:
            continue
        # Check if the annotation is float, int or Optional versions thereof
        annotation = field_info.annotation
        if annotation in (float, int, float | None, int | None):
            all_attributes.append(field)

    if ParameterGroup.ALL not in data.parameter_groups:
        selected_attrs = set()
        for group in data.parameter_groups:
            if group_attrs := parameter_groups_data.get(group.value):
                selected_attrs.update(group_attrs)
        attributes = [attr for attr in all_attributes if attr in selected_attrs]
    else:
        attributes = all_attributes

    if not attributes:
        logger.warning("No attributes selected or available.")
        return {}

    valid_objects = dataset.get_valid_objects()
    object_count = len(valid_objects)
    attribute_data = {attr: [None] * object_count for attr in attributes}
    attr_lengths = dict.fromkeys(attributes, 0)

    for processed_obj in valid_objects:
        if not processed_obj.group_name:
            continue

        obj = processed_obj.metadata
        group_name = processed_obj.group_name

        for attr in attributes:
            value = getattr(obj, attr, None)
            if value is not None and isinstance(value, (int, float)):
                idx = attr_lengths[attr]
                attribute_data[attr][idx] = (float(value), group_name)
                attr_lengths[attr] += 1

    result = {}
    for attr in attributes:
        values = attribute_data[attr][: attr_lengths[attr]]
        if values:
            values.sort(key=lambda x: x[0])
            result[attr] = values

    return result
