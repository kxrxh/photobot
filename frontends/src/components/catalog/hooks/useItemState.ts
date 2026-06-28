import { useCallback, useReducer } from "react"
import type { Analysis, KalibriObject } from "@/api/analysis/types"

export interface ItemState {
	name: string
	description: string
	harmfulness: string
	characteristics: Array<{ name: string; value: string }>
	analyses: Analysis[]
	photos: Array<File | string | null>
	allObjectsData: Record<string, KalibriObject[]>
	excludedObjects: Array<number>
}

const defaultState: ItemState = {
	name: "",
	description: "",
	harmfulness: "",
	characteristics: [],
	analyses: [],
	photos: [],
	allObjectsData: {},
	excludedObjects: [],
}

type ItemAction =
	| { type: "SET_NAME"; name: string }
	| { type: "SET_DESCRIPTION"; description: string }
	| { type: "SET_HARMFULNESS"; harmfulness: string }
	| { type: "SET_CHARACTERISTICS"; characteristics: Array<{ name: string; value: string }> }
	| { type: "ADD_ANALYSIS"; analysis: Analysis }
	| { type: "REMOVE_ANALYSIS"; analysisId: string }
	| { type: "SET_PHOTOS"; photos: Array<File | string | null> }
	| { type: "SET_ALL_OBJECTS_DATA"; data: Record<string, KalibriObject[]> }
	| { type: "TOGGLE_EXCLUDE_OBJECT"; objectId: number }
	| { type: "MERGE"; patch: Partial<ItemState> }
	| { type: "MERGE_FN"; fn: (prev: ItemState) => Partial<ItemState> }

function itemReducer(state: ItemState, action: ItemAction): ItemState {
	switch (action.type) {
		case "SET_NAME":
			return { ...state, name: action.name }
		case "SET_DESCRIPTION":
			return { ...state, description: action.description }
		case "SET_HARMFULNESS":
			return { ...state, harmfulness: action.harmfulness }
		case "SET_CHARACTERISTICS":
			return { ...state, characteristics: action.characteristics }
		case "ADD_ANALYSIS": {
			if (state.analyses.some((a) => a.id === action.analysis.id)) {
				return state
			}
			return { ...state, analyses: [...state.analyses, action.analysis] }
		}
		case "REMOVE_ANALYSIS": {
			const newAnalyses = state.analyses.filter((a) => a.id !== action.analysisId)
			const newAllObjectsData = { ...state.allObjectsData }
			const objectsInAnalysis = newAllObjectsData[action.analysisId] || []
			const objectIdsInAnalysis = new Set(objectsInAnalysis.map((o) => o.id))
			delete newAllObjectsData[action.analysisId]
			return {
				...state,
				analyses: newAnalyses,
				allObjectsData: newAllObjectsData,
				excludedObjects: state.excludedObjects.filter((id) => !objectIdsInAnalysis.has(Number(id))),
			}
		}
		case "SET_PHOTOS":
			return { ...state, photos: action.photos }
		case "SET_ALL_OBJECTS_DATA":
			return {
				...state,
				allObjectsData: { ...state.allObjectsData, ...action.data },
			}
		case "TOGGLE_EXCLUDE_OBJECT": {
			const excludedObjects = state.excludedObjects.includes(action.objectId)
				? state.excludedObjects.filter((id) => id !== action.objectId)
				: [...state.excludedObjects, action.objectId]
			return { ...state, excludedObjects }
		}
		case "MERGE":
			return { ...state, ...action.patch }
		case "MERGE_FN":
			return { ...state, ...action.fn(state) }
		default:
			return state
	}
}

const useItemState = (initialState?: Partial<ItemState>) => {
	const [itemState, dispatch] = useReducer(itemReducer, { ...defaultState, ...initialState })

	const setName = useCallback((name: string) => dispatch({ type: "SET_NAME", name }), [])
	const setDescription = useCallback(
		(description: string) => dispatch({ type: "SET_DESCRIPTION", description }),
		[]
	)
	const setHarmfulness = useCallback(
		(harmfulness: string) => dispatch({ type: "SET_HARMFULNESS", harmfulness }),
		[]
	)
	const setCharacteristics = useCallback(
		(characteristics: Array<{ name: string; value: string }>) =>
			dispatch({ type: "SET_CHARACTERISTICS", characteristics }),
		[]
	)
	const addAnalysis = useCallback(
		(analysis: Analysis) => dispatch({ type: "ADD_ANALYSIS", analysis }),
		[]
	)
	const removeAnalysis = useCallback(
		(analysisId: string) => dispatch({ type: "REMOVE_ANALYSIS", analysisId }),
		[]
	)
	const setPhotos = useCallback(
		(photos: Array<File | string | null>) => dispatch({ type: "SET_PHOTOS", photos }),
		[]
	)
	const setAllObjectsData = useCallback(
		(data: Record<string, KalibriObject[]>) => dispatch({ type: "SET_ALL_OBJECTS_DATA", data }),
		[]
	)
	const toggleExcludeObject = useCallback(
		(objectId: number) => dispatch({ type: "TOGGLE_EXCLUDE_OBJECT", objectId }),
		[]
	)
	const setItemState = useCallback(
		(patch: Partial<ItemState> | ((prev: ItemState) => Partial<ItemState>)) => {
			if (typeof patch === "function") {
				dispatch({ type: "MERGE_FN", fn: patch })
			} else {
				dispatch({ type: "MERGE", patch })
			}
		},
		[]
	)

	return {
		itemState,
		setName,
		setDescription,
		setHarmfulness,
		setCharacteristics,
		addAnalysis,
		removeAnalysis,
		setPhotos,
		setItemState,
		setAllObjectsData,
		toggleExcludeObject,
	}
}

export default useItemState
