import type { KalibriObject } from "@/api/analysis/types"
import {
	addUniqueObjectsToFraction,
	classifyObjects,
	classifyObjectsSelective,
	createNewFraction,
	DEFAULT_FRACTION,
	DEFAULT_FRACTION_ID,
	findObjectInFraction,
	removeObjectsFromFraction,
	toObjectsWithSource,
} from "@/hooks/fractions/fractionsLogic"
import type {
	ClassificationRules,
	FractionType,
	KalibriObjectWithSource,
} from "@/hooks/fractions/types"

export type PendingReclassification = {
	action: () => void
	modalData: {
		manualObjectsCount: number
		actionType: "adding_analysis" | "changing_rules" | "manual_reclassify"
		affectedFractions: string[]
		newObjectsCount?: number
		selectiveAction?: () => void
	}
}

export interface FractionsState {
	fractions: FractionType[]
	pendingReclassification: PendingReclassification | null
}

export const initialFractionsState: FractionsState = {
	fractions: [DEFAULT_FRACTION],
	pendingReclassification: null,
}

export type FractionsAction =
	| { type: "SET_FRACTIONS"; fractions: FractionType[] }
	| { type: "SET_FRACTIONS_FN"; fn: (prev: FractionType[]) => FractionType[] }
	| { type: "SET_PENDING"; pending: PendingReclassification | null }
	| { type: "ADD_FRACTION" }
	| { type: "REMOVE_FRACTION"; fractionId: string }
	| { type: "CLEAR_FRACTIONS" }
	| { type: "EDIT_FRACTION_NAME"; fractionId: string; newName: string }
	| {
			type: "MOVE_OBJECT"
			objectId: number
			sourceFractionId: string
			targetFractionId: string
	  }
	| {
			type: "MOVE_OBJECTS"
			objectsToMove: { objectId: number; sourceFractionId: string }[]
			targetFractionId: string
	  }
	| { type: "REMOVE_BY_ANALYSIS"; analysisId: string }
	| { type: "REMOVE_BY_CATALOG"; catalogItemId: number }
	| {
			type: "ADD_OBJECTS"
			fractionId: string
			objects: KalibriObjectWithSource[]
	  }
	| {
			type: "APPLY_RULES"
			fractionId: string
			rules?: ClassificationRules
			selective: boolean
	  }
	| { type: "CLASSIFY_ALL" }

export function fractionsReducer(state: FractionsState, action: FractionsAction): FractionsState {
	switch (action.type) {
		case "SET_FRACTIONS":
			return { ...state, fractions: action.fractions }
		case "SET_FRACTIONS_FN":
			return { ...state, fractions: action.fn(state.fractions) }
		case "SET_PENDING":
			return { ...state, pendingReclassification: action.pending }
		case "ADD_FRACTION":
			return { ...state, fractions: [...state.fractions, createNewFraction()] }
		case "REMOVE_FRACTION": {
			if (action.fractionId === DEFAULT_FRACTION_ID) return state
			const fractionToRemove = state.fractions.find((f) => f.id === action.fractionId)
			if (!fractionToRemove) return state
			const updated = state.fractions.map((fraction) =>
				fraction.id === DEFAULT_FRACTION_ID
					? addUniqueObjectsToFraction(fraction, fractionToRemove.objects)
					: fraction
			)
			return { ...state, fractions: updated.filter((f) => f.id !== action.fractionId) }
		}
		case "CLEAR_FRACTIONS":
			return { ...state, fractions: [DEFAULT_FRACTION] }
		case "EDIT_FRACTION_NAME": {
			if (action.fractionId === DEFAULT_FRACTION_ID) return state
			return {
				...state,
				fractions: state.fractions.map((f) =>
					f.id === action.fractionId ? { ...f, name: action.newName } : f
				),
			}
		}
		case "MOVE_OBJECT": {
			const { objectId, sourceFractionId, targetFractionId } = action
			if (sourceFractionId === targetFractionId) return state
			const sourceFraction = state.fractions.find((f) => f.id === sourceFractionId)
			if (!sourceFraction) return state
			const objectToMove = findObjectInFraction(sourceFraction, objectId)
			if (!objectToMove) return state
			return {
				...state,
				fractions: state.fractions.map((fraction) => {
					if (fraction.id === sourceFractionId) {
						return removeObjectsFromFraction(fraction, new Set([objectId]))
					}
					if (fraction.id === targetFractionId) {
						if (fraction.objects.some((obj) => obj.id === objectId)) return fraction
						return addUniqueObjectsToFraction(fraction, [
							{ ...objectToMove, classificationState: "manual" },
						])
					}
					return fraction
				}),
			}
		}
		case "MOVE_OBJECTS": {
			const validObjects: KalibriObjectWithSource[] = []
			const objectIdsToRemove = new Set<number>()
			for (const { objectId, sourceFractionId } of action.objectsToMove) {
				if (sourceFractionId === action.targetFractionId) continue
				const sourceFraction = state.fractions.find((f) => f.id === sourceFractionId)
				if (!sourceFraction) continue
				const objectToMove = findObjectInFraction(sourceFraction, objectId)
				if (objectToMove) {
					validObjects.push({ ...objectToMove, classificationState: "manual" })
					objectIdsToRemove.add(objectId)
				}
			}
			if (validObjects.length === 0) return state
			return {
				...state,
				fractions: state.fractions.map((fraction) => {
					const withoutMoved = removeObjectsFromFraction(fraction, objectIdsToRemove)
					return fraction.id === action.targetFractionId
						? addUniqueObjectsToFraction(withoutMoved, validObjects)
						: withoutMoved
				}),
			}
		}
		case "REMOVE_BY_ANALYSIS":
			return {
				...state,
				fractions: state.fractions.map((fraction) => ({
					...fraction,
					objects: fraction.objects.filter(
						(obj) =>
							!(obj.source.type === "analysis" && String(obj.source.sourceId) === action.analysisId)
					),
				})),
			}
		case "REMOVE_BY_CATALOG":
			return {
				...state,
				fractions: state.fractions.map((fraction) => ({
					...fraction,
					objects: fraction.objects.filter(
						(obj) =>
							!(obj.source.type === "catalog" && obj.source.sourceId === action.catalogItemId)
					),
				})),
			}
		case "ADD_OBJECTS": {
			const hasRules = state.fractions.some((f) => f.classificationRules)
			if (hasRules) {
				return {
					...state,
					fractions: classifyObjectsSelective(state.fractions, action.objects),
				}
			}
			return {
				...state,
				fractions: state.fractions.map((fraction) =>
					fraction.id === action.fractionId
						? addUniqueObjectsToFraction(fraction, action.objects)
						: fraction
				),
			}
		}
		case "APPLY_RULES": {
			const updated = state.fractions.map((f) =>
				f.id === action.fractionId ? { ...f, classificationRules: action.rules } : f
			)
			return {
				...state,
				fractions: action.selective ? classifyObjectsSelective(updated) : classifyObjects(updated),
			}
		}
		case "CLASSIFY_ALL":
			return { ...state, fractions: classifyObjects(state.fractions) }
		default:
			return state
	}
}

export function buildObjectsForAdd(
	newObjects: KalibriObject[],
	source: { type: "analysis" | "catalog"; sourceId: string | number }
): KalibriObjectWithSource[] {
	return toObjectsWithSource(newObjects, source)
}
