import type { KalibriObject } from "@/api/analysis/types"

export interface FieldStatistics {
	min: number
	max: number
	avg: number
	med: number
	stddev: number
	skew: number
}

export type FieldKey = "l" | "w" | "sq" | "l_w" | "r" | "g" | "b" | "h" | "s" | "v"

export type FieldStatisticsResult = Partial<Record<FieldKey, FieldStatistics>>

export interface ClassStatistics extends Partial<Record<FieldKey, FieldStatistics>> {
	count: number
}

export type ClassStatisticsResult = Record<string, ClassStatistics>

function calculateMedian(values: number[]): number {
	if (values.length === 0) {
		return 0
	}

	const sorted = [...values].sort((a, b) => a - b)
	const mid = Math.floor(sorted.length / 2)

	if (sorted.length % 2 === 0) {
		const val1 = sorted[mid - 1]
		const val2 = sorted[mid]
		if (val1 !== undefined && val2 !== undefined) {
			return (val1 + val2) / 2
		}
		return 0
	}

	const val = sorted[mid]
	return val !== undefined ? val : 0
}

function calculateStandardDeviation(values: number[], mean?: number): number {
	if (values.length < 2) {
		return 0
	}

	const avg = mean !== undefined ? mean : values.reduce((a, b) => a + b, 0) / values.length
	const variance = values.reduce((sum, val) => sum + (val - avg) ** 2, 0) / (values.length - 1)
	return Math.sqrt(variance)
}

function calculateSkewness(values: number[]): number {
	if (values.length < 3) {
		return 0
	}

	const n = values.length

	const mean = values.reduce((a, b) => a + b, 0) / n
	const variance = values.reduce((sum, val) => sum + (val - mean) ** 2, 0) / (n - 1)
	const std = Math.sqrt(variance)

	if (std === 0) {
		return 0
	}

	const thirdMoment = values.reduce((sum, val) => {
		const standardized = (val - mean) / std
		return sum + standardized ** 3
	}, 0)

	const biasCorrection = n / ((n - 1) * (n - 2))

	return biasCorrection * thirdMoment
}

const FIELDS_TO_STATS: Array<{ key: string; name: FieldKey }> = [
	{ key: "l", name: "l" },
	{ key: "w", name: "w" },
	{ key: "sq", name: "sq" },
	{ key: "l_w", name: "l_w" },
	{ key: "r", name: "r" },
	{ key: "g", name: "g" },
	{ key: "b", name: "b" },
	{ key: "h", name: "h" },
	{ key: "s", name: "s" },
	{ key: "v", name: "v" },
]

function getNumericFieldValue(obj: KalibriObject, field: string): number | null {
	const record = obj as unknown as Record<string, unknown>
	const directValue = record[field]
	if (typeof directValue === "number" && Number.isFinite(directValue)) {
		return directValue
	}

	if (field === "l_w") {
		const length = obj.l
		const width = obj.w
		if (
			typeof length === "number" &&
			Number.isFinite(length) &&
			typeof width === "number" &&
			Number.isFinite(width) &&
			width !== 0
		) {
			return length / width
		}
	}

	return null
}

function statisticsFromValues(values: number[]): FieldStatistics | null {
	if (values.length === 0) return null
	const sorted = [...values].sort((a, b) => a - b)
	const sum = values.reduce((acc, val) => acc + val, 0)
	const avg = sum / values.length
	const min = sorted[0]
	const max = sorted[sorted.length - 1]
	if (min === undefined || max === undefined) return null
	return {
		min,
		max,
		avg,
		med: calculateMedian(values),
		stddev: calculateStandardDeviation(values, avg),
		skew: calculateSkewness(values),
	}
}

function calculateFieldStatistics(objects: KalibriObject[], field: string): FieldStatistics | null {
	const values: number[] = []
	for (const obj of objects) {
		const value = getNumericFieldValue(obj, field)
		if (value !== null) {
			values.push(value)
		}
	}
	return statisticsFromValues(values)
}

export function calculateFieldStatisticsResult(objects: KalibriObject[]): FieldStatisticsResult {
	const result: FieldStatisticsResult = {}
	if (objects.length === 0) return result

	const valuesByField: Partial<Record<FieldKey, number[]>> = {}
	for (const { name } of FIELDS_TO_STATS) {
		valuesByField[name] = []
	}
	for (const obj of objects) {
		for (const { key, name } of FIELDS_TO_STATS) {
			const value = getNumericFieldValue(obj, key)
			if (value !== null) {
				valuesByField[name]?.push(value)
			}
		}
	}

	for (const { name } of FIELDS_TO_STATS) {
		const values = valuesByField[name]
		if (values && values.length > 0) {
			const stats = statisticsFromValues(values)
			if (stats) result[name] = stats
		}
	}

	return result
}

export function calculateClassStatisticsResult(
	objects: KalibriObject[]
): ClassStatisticsResult | null {
	if (objects.length === 0) {
		return null
	}

	const classGroups: Record<string, KalibriObject[]> = {}
	for (const obj of objects) {
		const className = obj.class
		if (typeof className === "string" && className.trim().length > 0) {
			if (!classGroups[className]) {
				classGroups[className] = []
			}
			classGroups[className].push(obj)
		}
	}

	if (Object.keys(classGroups).length === 0) {
		return null
	}

	const result: ClassStatisticsResult = {}

	const fields: Array<{ key: string; name: FieldKey }> = [
		{ key: "l", name: "l" },
		{ key: "w", name: "w" },
		{ key: "sq", name: "sq" },
		{ key: "l_w", name: "l_w" },
		{ key: "r", name: "r" },
		{ key: "g", name: "g" },
		{ key: "b", name: "b" },
		{ key: "h", name: "h" },
		{ key: "s", name: "s" },
		{ key: "v", name: "v" },
	]

	for (const [className, classObjects] of Object.entries(classGroups)) {
		const classStats: ClassStatistics = {
			count: classObjects.length,
		}

		for (const { key, name } of fields) {
			const fieldStats = calculateFieldStatistics(classObjects, key)
			if (fieldStats) {
				classStats[name] = fieldStats
			}
		}

		result[className] = classStats
	}

	return result
}

export function calculateKernelBrokenPercentage(
	objects: KalibriObject[],
	product: string
): number | null {
	const productLower = product.toLowerCase()
	const isKernels = productLower.includes("kernel")

	if (!isKernels || objects.length === 0) {
		return null
	}

	let totalMass = 0
	let brokenMass = 0

	for (const obj of objects) {
		const length = obj.l
		const width = obj.w

		if (
			typeof length !== "number" ||
			typeof width !== "number" ||
			!Number.isFinite(length) ||
			!Number.isFinite(width) ||
			length <= 0 ||
			width <= 0
		) {
			continue
		}

		const semiLength = length / 2
		const semiWidth = width / 2
		const volume = (4 / 3) * Math.PI * semiLength * (semiWidth * semiWidth)
		const mass = 0.55 * volume
		totalMass += mass

		if (typeof obj.class === "string" && ["broken", "ломанные"].includes(obj.class.toLowerCase())) {
			brokenMass += mass
		}
	}

	if (totalMass === 0) {
		return null
	}

	const percentage = (brokenMass / totalMass) * 100
	return Math.round(percentage * 10) / 10
}

export function calculateSampleStdFromValues(values: number[]): number | null {
	if (values.length < 2) {
		return null
	}

	const mean = values.reduce((sum, value) => sum + value, 0) / values.length
	const variance = values.reduce((sum, value) => sum + (value - mean) ** 2, 0) / (values.length - 1)
	return Math.sqrt(variance)
}
