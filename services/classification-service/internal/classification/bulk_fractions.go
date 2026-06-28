package classification

import (
	"context"
	"fmt"
	"math"

	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *ClassificationService) bulkCreateFractionTree(
	ctx context.Context,
	qtx *database.Queries,
	classificationID uuid.UUID,
	fractions []Fraction,
) ([]Fraction, error) {
	if len(fractions) == 0 {
		return []Fraction{}, nil
	}

	names := make([]string, 0, len(fractions))
	orderIndexes := make([]int32, 0, len(fractions))
	fracPlans := make([]fractionInsertPlan, 0, len(fractions))

	for i, fraction := range fractions {
		orderIdx, err := orderIndexFromRange(i)
		if err != nil {
			return nil, httperr.Wrap(err, fiber.StatusBadRequest, "Order index overflow")
		}
		names = append(names, fraction.Name)
		orderIndexes = append(orderIndexes, orderIdx)
		fracPlans = append(fracPlans, fractionInsertPlan{sourceIndex: i, orderIndex: orderIdx})
	}

	createdFractions, err := qtx.BulkCreateFractions(ctx, database.BulkCreateFractionsParams{
		Names:            names,
		ClassificationID: classificationID,
		OrderIndexes:     orderIndexes,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("classification_id", classificationID.String()).
			Msg("bulk create fractions failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to create fractions",
		)
	}

	if len(createdFractions) != len(fracPlans) {
		return nil, httperr.New(
			fiber.StatusInternalServerError,
			"Fraction bulk insert count mismatch",
		)
	}

	fractionByOrder := make(map[int32]database.Fraction, len(createdFractions))
	for _, created := range createdFractions {
		fractionByOrder[created.OrderIndex] = created
	}

	fractionIDsBySource := make([]uuid.UUID, len(fractions))
	for _, plan := range fracPlans {
		created, ok := fractionByOrder[plan.orderIndex]
		if !ok {
			return nil, httperr.New(
				fiber.StatusInternalServerError,
				"Fraction bulk insert mapping failed",
			)
		}
		fractionIDsBySource[plan.sourceIndex] = created.ID
	}

	var (
		condFractionIDs []uuid.UUID
		condNames       []string
		condOperators   []database.LogicOperator
		condConnections []database.LogicOperator
		condOrderIdxs   []int32
		condPlans       []conditionInsertPlan
	)

	for fi, fraction := range fractions {
		fractionID := fractionIDsBySource[fi]
		for ci, condition := range fraction.Conditions {
			condFractionIDs = append(condFractionIDs, fractionID)
			condNames = append(condNames, condition.Name)
			condOperators = append(condOperators, database.LogicOperator(condition.Operator))
			condConnections = append(condConnections, database.LogicOperator(condition.Connection))
			condOrderIdxs = append(condOrderIdxs, condition.OrderIndex)
			condPlans = append(condPlans, conditionInsertPlan{
				fractionIndex:  fi,
				conditionIndex: ci,
				orderIndex:     condition.OrderIndex,
			})
		}
	}

	conditionIDsBySource := make([][]uuid.UUID, len(fractions))
	for i := range conditionIDsBySource {
		conditionIDsBySource[i] = make([]uuid.UUID, len(fractions[i].Conditions))
	}

	if len(condPlans) > 0 {
		createdConditions, err := qtx.BulkCreateConditions(ctx, database.BulkCreateConditionsParams{
			FractionIds:  condFractionIDs,
			Names:        condNames,
			Operators:    condOperators,
			Connections:  condConnections,
			OrderIndexes: condOrderIdxs,
		})
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("classification_id", classificationID.String()).
				Msg("bulk create conditions failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to create conditions",
			)
		}

		if len(createdConditions) != len(condPlans) {
			return nil, httperr.New(
				fiber.StatusInternalServerError,
				"Condition bulk insert count mismatch",
			)
		}

		conditionByKey := make(map[conditionKey]database.Condition, len(createdConditions))
		for _, created := range createdConditions {
			conditionByKey[conditionKey{
				fractionID: created.FractionID,
				orderIndex: created.OrderIndex,
			}] = created
		}

		for _, plan := range condPlans {
			fractionID := fractionIDsBySource[plan.fractionIndex]
			created, ok := conditionByKey[conditionKey{
				fractionID: fractionID,
				orderIndex: plan.orderIndex,
			}]
			if !ok {
				return nil, httperr.New(
					fiber.StatusInternalServerError,
					"Condition bulk insert mapping failed",
				)
			}
			conditionIDsBySource[plan.fractionIndex][plan.conditionIndex] = created.ID
		}
	}

	var (
		paramNames        []string
		paramOperators    []database.ParamOperator
		paramValues       []pgtype.Numeric
		paramConditionIDs []uuid.UUID
		paramPlans        []paramInsertPlan
	)

	for fi, fraction := range fractions {
		for ci, condition := range fraction.Conditions {
			conditionID := conditionIDsBySource[fi][ci]
			for pi, param := range condition.Params {
				numericValue := pgtype.Numeric{}
				if err := numericValue.Scan(fmt.Sprintf("%.2f", param.Value)); err != nil {
					s.logger.Error().
						Err(err).
						Str("condition_id", conditionID.String()).
						Str("param_name", param.Name).
						Float32("param_value", param.Value).
						Msg("parse param value failed")
					return nil, httperr.Wrap(
						err,
						fiber.StatusInternalServerError,
						"Failed to parse param value",
					)
				}
				paramNames = append(paramNames, param.Name)
				paramOperators = append(paramOperators, database.ParamOperator(param.Operator))
				paramValues = append(paramValues, numericValue)
				paramConditionIDs = append(paramConditionIDs, conditionID)
				paramPlans = append(paramPlans, paramInsertPlan{
					fractionIndex:  fi,
					conditionIndex: ci,
					paramIndex:     pi,
					sourceValue:    param.Value,
				})
			}
		}
	}

	paramsBySource := make([][][]Param, len(fractions))
	for fi, fraction := range fractions {
		paramsBySource[fi] = make([][]Param, len(fraction.Conditions))
		for ci := range fraction.Conditions {
			paramsBySource[fi][ci] = make([]Param, len(fraction.Conditions[ci].Params))
		}
	}

	if len(paramPlans) > 0 {
		createdParams, err := qtx.BulkCreateParams(ctx, database.BulkCreateParamsParams{
			Names:        paramNames,
			Operators:    paramOperators,
			Values:       paramValues,
			ConditionIds: paramConditionIDs,
		})
		if err != nil {
			s.logger.Error().
				Err(err).
				Str("classification_id", classificationID.String()).
				Msg("bulk create params failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to create params",
			)
		}
		if len(createdParams) != len(paramPlans) {
			return nil, httperr.New(
				fiber.StatusInternalServerError,
				"Param bulk insert count mismatch",
			)
		}

		paramByKey := make(map[paramKey]database.Param, len(createdParams))
		for _, created := range createdParams {
			paramByKey[paramKey{conditionID: created.ConditionID, name: created.Name}] = created
		}

		for i, plan := range paramPlans {
			conditionID := conditionIDsBySource[plan.fractionIndex][plan.conditionIndex]
			created, ok := paramByKey[paramKey{conditionID: conditionID, name: paramNames[i]}]
			if !ok {
				return nil, httperr.New(
					fiber.StatusInternalServerError,
					"Param bulk insert mapping failed",
				)
			}
			paramsBySource[plan.fractionIndex][plan.conditionIndex][plan.paramIndex] = Param{
				ID:       created.ID,
				Name:     created.Name,
				Operator: string(created.Operator),
				Value:    plan.sourceValue,
			}
		}
	}

	result := make([]Fraction, 0, len(fractions))
	for fi, fraction := range fractions {
		var orderIndex int32
		for _, plan := range fracPlans {
			if plan.sourceIndex == fi {
				orderIndex = plan.orderIndex
				break
			}
		}

		conditions := make([]Condition, 0, len(fraction.Conditions))
		for ci, condition := range fraction.Conditions {
			conditions = append(conditions, Condition{
				ID:         conditionIDsBySource[fi][ci],
				Name:       condition.Name,
				Operator:   condition.Operator,
				Connection: condition.Connection,
				OrderIndex: condition.OrderIndex,
				Params:     paramsBySource[fi][ci],
			})
		}

		result = append(result, Fraction{
			ID:         fractionIDsBySource[fi],
			Name:       fraction.Name,
			Conditions: conditions,
			OrderIndex: orderIndex,
		})
	}

	return result, nil
}

func orderIndexFromRange(i int) (int32, error) {
	if i < math.MinInt32 || i > math.MaxInt32 {
		return 0, fmt.Errorf("order index %d overflows int32", i)
	}
	return int32(i), nil
}
