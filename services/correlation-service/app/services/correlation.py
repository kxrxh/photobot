import logging

from fastapi import HTTPException, status
from opentelemetry import trace

from ..models.correlation import Condition, CorrelationRequest, CorrelationWithTest
from .condition_extractor import extract_linear_conditions, extract_tree_conditions
from .condition_tester import compute_condition_metrics, evaluate_conditions
from .data_processor import (
    ProcessedDataset,
    build_feature_matrix,
    create_processed_dataset,
    handle_missing_values,
    prepare_attribute_data,
    validate_training_data,
)
from .feature_selector import select_important_features
from .ml_engine import run_multiple_algorithms

logger = logging.getLogger(__name__)
tracer = trace.get_tracer(__name__)


class CorrelationCalculatorService:
    PERFORMANCE_DEGRADATION_TOLERANCE = 0.95
    # Real-world: allow simpler 1-condition rules to increase recall
    MIN_CONDITIONS_TO_KEEP = 1
    MIN_F1_STRICT = 0.18
    MIN_F1_LOOSE = 0.15
    MIN_PRECISION = 0.20
    MIN_RECALL_STRICT = 0.15
    MIN_RECALL_LOOSE = 0.10
    # Prefer recall in real-world usage (F2)
    F_BETA = 2.0

    @staticmethod
    def _fbeta(precision: float, recall: float, beta: float) -> float:
        if precision <= 0.0 or recall <= 0.0:
            return 0.0
        b2 = beta * beta
        denom = (b2 * precision) + recall
        return ((1.0 + b2) * precision * recall / denom) if denom > 0.0 else 0.0

    @staticmethod
    async def calculate_pairwise_correlation(
        data: CorrelationRequest,
    ) -> list[CorrelationWithTest]:
        with tracer.start_as_current_span("correlation.service.calculate_pairwise"):
            try:
                logger.info("Creating optimized processed dataset...")
                with tracer.start_as_current_span("correlation.dataset.create"):
                    dataset = await create_processed_dataset(data)
                if not dataset.objects_by_id:
                    return [
                        CorrelationWithTest(
                            name=g.name, conditions=[], test_results=None
                        )
                        for g in data.fractions
                    ]

                logger.info("Preparing attribute data...")
                with tracer.start_as_current_span("correlation.attributes.prepare"):
                    attr_data = prepare_attribute_data(dataset, data)
                if not attr_data:
                    return [
                        CorrelationWithTest(
                            name=g.name, conditions=[], test_results=None
                        )
                        for g in data.fractions
                    ]

                dataset.feature_names = list(attr_data.keys())
                results: list[CorrelationWithTest] = []

                for target_group in data.fractions:
                    with tracer.start_as_current_span(
                        "correlation.group.process"
                    ) as group_span:
                        group_span.set_attribute(
                            "correlation.group.name", target_group.name
                        )
                        logger.info(f"--- Processing group: {target_group.name} ---")
                        conditions = CorrelationCalculatorService.find_group_conditions(
                            target_group.name, dataset, data
                        )

                        if not conditions:
                            results.append(
                                CorrelationWithTest(
                                    name=target_group.name,
                                    conditions=[],
                                    test_results=None,
                                )
                            )
                            continue

                        test_res = compute_condition_metrics(
                            conditions, target_group.name, dataset, data
                        )
                        if (
                            test_res.f1_score
                            < CorrelationCalculatorService.MIN_F1_STRICT
                            or test_res.precision
                            < CorrelationCalculatorService.MIN_PRECISION
                            or test_res.recall
                            < CorrelationCalculatorService.MIN_RECALL_STRICT
                        ):
                            results.append(
                                CorrelationWithTest(
                                    name=target_group.name,
                                    conditions=[],
                                    test_results=None,
                                )
                            )
                        else:
                            results.append(
                                CorrelationWithTest(
                                    name=target_group.name,
                                    conditions=conditions,
                                    test_results=test_res,
                                )
                            )

                # Retry with looser constraints if needed
                successful = sum(1 for r in results if r.conditions)
                if len(data.fractions) > 1 and successful < max(
                    2, (len(data.fractions) + 1) // 2
                ):
                    for i, res in enumerate(results):
                        if not res.conditions:
                            target = data.fractions[i]
                            retry_conds = (
                                CorrelationCalculatorService.find_group_conditions(
                                    target.name,
                                    dataset,
                                    data,
                                    cv_folds=3,
                                )
                            )
                            if retry_conds:
                                retry_test = compute_condition_metrics(
                                    retry_conds, target.name, dataset, data
                                )
                                retry_score = CorrelationCalculatorService._fbeta(
                                    retry_test.precision,
                                    retry_test.recall,
                                    CorrelationCalculatorService.F_BETA,
                                )
                                if (
                                    retry_score
                                    >= CorrelationCalculatorService.MIN_F1_LOOSE
                                    and retry_test.precision
                                    >= CorrelationCalculatorService.MIN_PRECISION
                                    and retry_test.recall
                                    >= CorrelationCalculatorService.MIN_RECALL_LOOSE
                                ):
                                    results[i] = CorrelationWithTest(
                                        name=target.name,
                                        conditions=retry_conds,
                                        test_results=retry_test,
                                    )

                return sorted(
                    results,
                    key=lambda r: r.test_results.accuracy if r.test_results else 0.0,
                    reverse=True,
                )

            except HTTPException as e:
                raise e
            except Exception as e:
                logger.error(f"Unexpected error: {e}", exc_info=True)
                raise HTTPException(
                    status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
                    detail="Internal error during calculation.",
                ) from e

    @staticmethod
    def find_group_conditions(
        group_name: str,
        dataset: ProcessedDataset,
        data: CorrelationRequest,
        cv_folds: int = 5,
    ) -> list[Condition]:
        try:
            f_names = dataset.feature_names
            if not f_names:
                return []

            if len(f_names) > 20:
                f_names = select_important_features(
                    dataset, group_name, max_features=20, min_importance=0.005
                )

            x_matrix, y_vector = build_feature_matrix(dataset, group_name, f_names)[:2]
            if x_matrix is None or y_vector is None or x_matrix.size == 0:
                return []

            # Explicitly assert for type checkers
            assert x_matrix is not None
            assert y_vector is not None

            x_matrix, f_names = handle_missing_values(x_matrix, f_names, group_name)
            if x_matrix.shape[1] == 0 or not validate_training_data(
                x_matrix, y_vector, group_name, min_samples_leaf=3
            ):
                return []

            alg_results = run_multiple_algorithms(
                x_matrix,
                y_vector,
                f_names,
                group_name,
                dataset,
                data,
                cv_folds,
                extract_tree_fn=extract_tree_conditions,
                extract_linear_fn=extract_linear_conditions,
                test_enhanced_fn=evaluate_conditions,
            )
            if alg_results:
                # Prefer recall-heavy quality instead of pure F1
                best = max(
                    alg_results,
                    key=lambda x: CorrelationCalculatorService._fbeta(
                        x.test_result.precision,
                        x.test_result.recall,
                        CorrelationCalculatorService.F_BETA,
                    ),
                )
                return CorrelationCalculatorService.post_process_conditions(
                    best.conditions, dataset, group_name, data
                )

            return []
        except Exception as e:
            logger.error(f"Error in find_group_conditions: {e}", exc_info=True)
            return []

    @staticmethod
    def post_process_conditions(
        conditions: list[Condition],
        dataset: ProcessedDataset,
        group_name: str,
        data: CorrelationRequest,
    ) -> list[Condition]:
        if not conditions:
            return []
        try:
            orig_res = compute_condition_metrics(conditions, group_name, dataset, data)
            orig_score = CorrelationCalculatorService._fbeta(
                orig_res.precision, orig_res.recall, CorrelationCalculatorService.F_BETA
            )
            simplified = conditions.copy()
            while len(simplified) > CorrelationCalculatorService.MIN_CONDITIONS_TO_KEEP:
                best_idx, best_f1 = None, None
                for i in range(len(simplified)):
                    test_c = simplified[:i] + simplified[i + 1 :]
                    res = compute_condition_metrics(test_c, group_name, dataset, data)
                    score = CorrelationCalculatorService._fbeta(
                        res.precision, res.recall, CorrelationCalculatorService.F_BETA
                    )
                    if (
                        score
                        >= orig_score
                        * CorrelationCalculatorService.PERFORMANCE_DEGRADATION_TOLERANCE
                        and res.precision >= CorrelationCalculatorService.MIN_PRECISION
                        and res.recall >= CorrelationCalculatorService.MIN_RECALL_LOOSE
                        and (best_f1 is None or score > best_f1)
                    ):
                        best_f1, best_idx = score, i
                if best_idx is not None:
                    simplified.pop(best_idx)
                else:
                    break
            return simplified
        except Exception:
            return conditions
