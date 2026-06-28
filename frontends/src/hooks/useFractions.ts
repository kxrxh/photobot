import { type Dispatch, type SetStateAction, useCallback, useReducer } from "react"
import type { KalibriObject } from "@/api/analysis/types"
import { countManualObjects, hasManualObjects } from "@/hooks/fractions/fractionsLogic"
import {
	buildObjectsForAdd,
	fractionsReducer,
	initialFractionsState,
	type PendingReclassification,
} from "@/hooks/fractions/fractionsReducer"
import type { ClassificationRules, FractionType } from "@/hooks/fractions/types"

export type {
	ClassificationRule,
	ClassificationRules,
	FractionType,
	KalibriObjectWithSource,
} from "@/hooks/fractions/types"

export const useFractions = () => {
	const [state, dispatch] = useReducer(fractionsReducer, initialFractionsState)
	const { fractions, pendingReclassification } = state

	const setFractions = useCallback((value: SetStateAction<FractionType[]>) => {
		if (typeof value === "function") {
			dispatch({ type: "SET_FRACTIONS_FN", fn: value })
		} else {
			dispatch({ type: "SET_FRACTIONS", fractions: value })
		}
	}, []) as Dispatch<SetStateAction<FractionType[]>>

	const runOrConfirm = useCallback(
		(action: () => void, modalData: PendingReclassification["modalData"]) => {
			if (hasManualObjects(fractions)) {
				dispatch({ type: "SET_PENDING", pending: { action, modalData } })
			} else {
				action()
			}
		},
		[fractions]
	)

	const addNewFraction = useCallback(() => dispatch({ type: "ADD_FRACTION" }), [])
	const removeFraction = useCallback(
		(fractionId: string) => dispatch({ type: "REMOVE_FRACTION", fractionId }),
		[]
	)
	const clearFractions = useCallback(() => dispatch({ type: "CLEAR_FRACTIONS" }), [])
	const editFractionName = useCallback(
		(fractionId: string, newName: string) =>
			dispatch({ type: "EDIT_FRACTION_NAME", fractionId, newName }),
		[]
	)
	const moveObjectToFraction = useCallback(
		(objectId: number, sourceFractionId: string, targetFractionId: string) =>
			dispatch({ type: "MOVE_OBJECT", objectId, sourceFractionId, targetFractionId }),
		[]
	)
	const moveMultipleObjectsToFraction = useCallback(
		(objectsToMove: { objectId: number; sourceFractionId: string }[], targetFractionId: string) =>
			dispatch({ type: "MOVE_OBJECTS", objectsToMove, targetFractionId }),
		[]
	)
	const removeObjectsByAnalysisId = useCallback(
		(analysisId: string) => dispatch({ type: "REMOVE_BY_ANALYSIS", analysisId }),
		[]
	)
	const removeObjectsByCatalogItemId = useCallback(
		(catalogItemId: number) => dispatch({ type: "REMOVE_BY_CATALOG", catalogItemId }),
		[]
	)
	const classifyObjectsAutomatically = useCallback(() => dispatch({ type: "CLASSIFY_ALL" }), [])

	const addObjectsToFraction = useCallback(
		(
			fractionId: string,
			newObjects: KalibriObject[],
			source: { type: "analysis" | "catalog"; sourceId: string | number }
		) => {
			const objectsWithSource = buildObjectsForAdd(newObjects, source)
			const action = () =>
				dispatch({
					type: "ADD_OBJECTS",
					fractionId,
					objects: objectsWithSource,
				})

			runOrConfirm(action, {
				manualObjectsCount: countManualObjects(fractions),
				actionType: "adding_analysis",
				affectedFractions: fractions.filter((f) => f.classificationRules).map((f) => f.name),
				newObjectsCount: newObjects.length,
			})
		},
		[fractions, runOrConfirm]
	)

	const setClassificationRules = useCallback(
		(fractionId: string, rules?: ClassificationRules) => {
			const action = () => dispatch({ type: "APPLY_RULES", fractionId, rules, selective: false })
			const selectiveAction = () =>
				dispatch({ type: "APPLY_RULES", fractionId, rules, selective: true })

			const affectedFractions = fractions
				.filter((f) => f.id === fractionId || f.classificationRules)
				.map((f) => f.name)

			runOrConfirm(action, {
				manualObjectsCount: countManualObjects(fractions),
				actionType: "changing_rules",
				affectedFractions: affectedFractions.length > 0 ? affectedFractions : [],
				selectiveAction,
			})
		},
		[fractions, runOrConfirm]
	)

	const handleReclassificationChoice = useCallback(
		(choice: "preserve_manual" | "reclassify_all" | "cancel") => {
			if (!pendingReclassification) return
			if (choice === "preserve_manual" && pendingReclassification.modalData.selectiveAction) {
				pendingReclassification.modalData.selectiveAction()
			} else if (choice === "preserve_manual" || choice === "reclassify_all") {
				pendingReclassification.action()
			}
			dispatch({ type: "SET_PENDING", pending: null })
		},
		[pendingReclassification]
	)

	const requestReclassification = useCallback(
		(action: () => void, modalData: PendingReclassification["modalData"]) => {
			runOrConfirm(action, modalData)
		},
		[runOrConfirm]
	)

	return {
		fractions,
		setFractions,
		addNewFraction,
		removeFraction,
		clearFractions,
		editFractionName,
		moveObjectToFraction,
		moveMultipleObjectsToFraction,
		addObjectsToFraction,
		removeObjectsByAnalysisId,
		removeObjectsByCatalogItemId,
		setClassificationRules,
		classifyObjectsAutomatically,
		hasManualObjects: () => hasManualObjects(fractions),
		requestReclassification,
		handleReclassificationChoice,
		pendingReclassification,
	}
}
