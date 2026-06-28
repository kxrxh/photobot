import type { KalibriObject } from "@/api/analysis/types"

export const DEFAULT_FRACTION_ID = "0"
export const GRID_CLASSES =
	"grid grid-cols-4 gap-2 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 xl:grid-cols-12"

export const createObjectLookup = (
	analysesObjectsMap: Record<string, KalibriObject[]>,
	analysisIds: (string | number)[]
): Record<number, KalibriObject & { analysisId: string }> => {
	const lookup: Record<number, KalibriObject & { analysisId: string }> = {}

	for (const analysisId of analysisIds) {
		const idStr = typeof analysisId === "string" ? analysisId : analysisId.toString()
		const objects = analysesObjectsMap[idStr] || []
		for (const obj of objects) {
			lookup[obj.id] = { ...obj, analysisId: idStr }
		}
	}

	return lookup
}
