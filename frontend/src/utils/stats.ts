import type { KalibriObject } from "@/api/analysis/types"
import type { WeedStatistics } from "@/api/catalog/types"
import type { AnalysisStatistics } from "@/components/analysis/selectors/AnalysisSelector/types"
import type { CatalogFilters } from "@/hooks/useCatalogFilters"

export interface StatData {
	min: number
	max: number
	avg: number
	median: number
}

export const calculateMetricStats = (values: number[]): StatData => {
	if (values.length === 0) {
		return { min: 0, max: 0, avg: 0, median: 0 }
	}

	let min = values[0]
	let max = values[0]
	let sum = values[0]

	for (let i = 1; i < values.length; i++) {
		const val = values[i]
		min = Math.min(min, val)
		max = Math.max(max, val)
		sum += val
	}

	const avg = sum / values.length

	const sorted = [...values].sort((a, b) => a - b)
	const median =
		sorted.length % 2 === 0
			? (sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2
			: sorted[Math.floor(sorted.length / 2)]

	return {
		min,
		max,
		avg: Number(avg.toFixed(2)),
		median: Number(median.toFixed(2)),
	}
}

const filterIncludedObjects = (
	allObjectsData: Record<string, KalibriObject[]>,
	excludedIds: Set<number>
): KalibriObject[] => {
	const excludedSet = new Set(excludedIds)
	return Object.values(allObjectsData)
		.flat()
		.filter((obj) => !excludedSet.has(obj.id))
}

const extractNumericValues = (objects: KalibriObject[], key: keyof KalibriObject): number[] => {
	return objects.map((obj) => obj[key]).filter((val): val is number => typeof val === "number")
}

const calculateLwRatios = (objects: KalibriObject[]): number[] => {
	return objects
		.map((obj) => {
			if (typeof obj.l === "number" && typeof obj.w === "number" && obj.w !== 0) {
				return obj.l / obj.w
			}
			return Number.NaN
		})
		.filter((val) => !Number.isNaN(val))
}

export const calculateStatsFromObjects = (
	allObjectsData: Record<string, KalibriObject[]>,
	excludedIds: number[]
): AnalysisStatistics | null => {
	const includedObjects = filterIncludedObjects(allObjectsData, new Set(excludedIds))

	if (includedObjects.length === 0) {
		return null
	}

	const stats: AnalysisStatistics = {}
	const metricMapping: {
		[key in keyof AnalysisStatistics]?: keyof KalibriObject
	} = {
		length: "l",
		width: "w",
		sq: "sq",
		r: "r",
		g: "g",
		b: "b",
		h: "h",
		s: "s",
		v: "v",
	}

	for (const key of Object.keys(metricMapping) as (keyof typeof metricMapping)[]) {
		const objKey = metricMapping[key]
		if (!objKey) continue

		const values = extractNumericValues(includedObjects, objKey)
		if (values.length > 0) {
			stats[key as keyof AnalysisStatistics] = calculateMetricStats(values)
		}
	}

	const lwRatioValues = calculateLwRatios(includedObjects)
	if (lwRatioValues.length > 0) {
		stats.lwRatio = calculateMetricStats(lwRatioValues)
	}

	return stats
}

const calculateBackendStats = (
	values: number[]
): { avg: number; median: number; min: number; max: number } => {
	if (values.length === 0) {
		return { avg: 0, median: 0, min: 0, max: 0 }
	}

	let min = values[0]
	let max = values[0]
	let sum = values[0]

	for (let i = 1; i < values.length; i++) {
		const val = values[i]
		min = Math.min(min, val)
		max = Math.max(max, val)
		sum += val
	}

	const avg = sum / values.length
	const sorted = [...values].sort((a, b) => a - b)
	const median =
		sorted.length % 2 === 0
			? (sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2
			: sorted[Math.floor(sorted.length / 2)]

	return {
		avg: Number(avg.toFixed(2)),
		median: Number(median.toFixed(2)),
		min,
		max,
	}
}

const METRIC_MAPPING = [
	{ key: "w", objKey: "w" as keyof KalibriObject },
	{ key: "l", objKey: "l" as keyof KalibriObject },
	{ key: "sq", objKey: "sq" as keyof KalibriObject },
	{ key: "r", objKey: "r" as keyof KalibriObject },
	{ key: "g", objKey: "g" as keyof KalibriObject },
	{ key: "b", objKey: "b" as keyof KalibriObject },
	{ key: "h", objKey: "h" as keyof KalibriObject },
	{ key: "s", objKey: "s" as keyof KalibriObject },
	{ key: "v", objKey: "v" as keyof KalibriObject },
	{ key: "brt", objKey: "brt" as keyof KalibriObject },
	{ key: "solid", objKey: "solid" as keyof KalibriObject },
	{ key: "sq_sqcrl", objKey: "sq_sqcrl" as keyof KalibriObject },
] as const

/** Maps computed analysis stats into the catalog API `WeedStatistics` shape. */
export const convertToBackendStatistics = (
	allObjectsData: Record<string, KalibriObject[]>,
	excludedIds: number[]
): WeedStatistics | null => {
	const includedObjects = filterIncludedObjects(allObjectsData, new Set(excludedIds))

	if (includedObjects.length === 0) {
		return null
	}

	const result: WeedStatistics = {} as WeedStatistics

	for (const { key, objKey } of METRIC_MAPPING) {
		const values = extractNumericValues(includedObjects, objKey)
		const stats = calculateBackendStats(values)

		;(result as WeedStatistics)[`${key}_avg`] = stats.avg
		;(result as WeedStatistics)[`${key}_median`] = stats.median
		;(result as WeedStatistics)[`${key}_min`] = stats.min
		;(result as WeedStatistics)[`${key}_max`] = stats.max
	}

	const lwValues = calculateLwRatios(includedObjects)
	const lwStats = calculateBackendStats(lwValues)
	result.lw_avg = lwStats.avg
	result.lw_median = lwStats.median
	result.lw_min = lwStats.min
	result.lw_max = lwStats.max

	return result
}

const calculateSmartRange = (
	values: number[],
	bands: number
): { min: number; max: number } | undefined => {
	if (values.length === 0) return undefined

	let min = values[0]
	let max = values[0]

	for (let i = 1; i < values.length; i++) {
		const val = values[i]
		min = Math.min(min, val)
		max = Math.max(max, val)
	}

	const range = max - min

	if (range === 0) {
		return {
			min: Math.max(0, min - 1),
			max: max + 1,
		}
	}

	const bandSize = range / bands
	const newMin = Math.max(0, min - bandSize * 0.5)
	const newMax = max + bandSize * 0.5

	return {
		min: Math.round(newMin * 100) / 100,
		max: Math.round(newMax * 100) / 100,
	}
}

const RANGE_FIELD_MAPPING = [
	{ key: "l", objKey: "l" as keyof KalibriObject },
	{ key: "w", objKey: "w" as keyof KalibriObject },
	{ key: "h", objKey: "h" as keyof KalibriObject },
	{ key: "s", objKey: "s" as keyof KalibriObject },
	{ key: "v", objKey: "v" as keyof KalibriObject },
	{ key: "r", objKey: "r" as keyof KalibriObject },
	{ key: "g", objKey: "g" as keyof KalibriObject },
	{ key: "b", objKey: "b" as keyof KalibriObject },
	{ key: "brt", objKey: "brt" as keyof KalibriObject },
	{ key: "sq_sqcrl", objKey: "sq_sqcrl" as keyof KalibriObject },
] as const

/** Derives default catalog numeric filter ranges from a set of objects. */
export const buildSmartRangesFromObjects = (
	objects: KalibriObject[],
	bands: number = 2
): Partial<CatalogFilters> => {
	if (objects.length === 0) {
		return {}
	}

	const result: Partial<CatalogFilters> = {}

	for (const { key, objKey } of RANGE_FIELD_MAPPING) {
		const values = extractNumericValues(objects, objKey)
		const range = calculateSmartRange(values, bands)

		if (range) {
			const resultWithFilters = result as Record<string, number | undefined>
			resultWithFilters[`${key}_min`] = range.min
			resultWithFilters[`${key}_max`] = range.max
		}
	}

	const lwValues = calculateLwRatios(objects)
	const lwRange = calculateSmartRange(lwValues, bands)
	if (lwRange) {
		result.lw_min = lwRange.min
		result.lw_max = lwRange.max
	}

	return result
}
