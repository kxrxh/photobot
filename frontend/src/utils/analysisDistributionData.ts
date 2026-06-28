import type { KalibriObject } from "@/api/analysis/types"
import {
	calculateAdaptiveBinSize,
	formatBinKey,
	isValidRange,
	roundValue,
} from "@/utils/analysisBinning"

const MASS_1000_RED = "#e74c3c"

export interface HistogramBin {
	label: string
	count: number
	color?: string
}

function getMass1000Value(obj: KalibriObject): number | null {
	const value = obj.mass_1000
	if (typeof value === "number" && Number.isFinite(value)) return value
	return null
}

/** Same bucket logic as ReportsService `createMass1000DistributionData`. */
export function buildMass1000Histogram(objects: KalibriObject[]): HistogramBin[] {
	if (!Array.isArray(objects) || objects.length === 0) return []

	const chartData: Record<string, { count: number; color?: string }> = {}

	for (const obj of objects) {
		const mass1000 = getMass1000Value(obj)
		if (mass1000 === null) continue

		if (mass1000 < 28) {
			const currentCount = chartData["<28"]?.count ?? 0
			chartData["<28"] = { count: currentCount + 1, color: MASS_1000_RED }
			continue
		}

		const binKey =
			mass1000 < 30
				? formatBinKey(28, 30)
				: formatBinKey(
						30 + Math.floor((mass1000 - 30) / 5) * 5,
						30 + (Math.floor((mass1000 - 30) / 5) + 1) * 5
					)
		const currentCount = chartData[binKey]?.count ?? 0
		chartData[binKey] = { count: currentCount + 1 }
	}

	return Object.entries(chartData).map(([label, v]) => ({
		label,
		count: v.count,
		color: v.color,
	}))
}

function extractNumericValue(obj: KalibriObject, field: string): number | null {
	const record = obj as unknown as Record<string, unknown>
	const value = record[field]
	if (typeof value === "number" && Number.isFinite(value)) {
		return value
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

/** Same binning as ReportsService `createDistributionData`. */
export function buildFieldHistogram(
	objects: KalibriObject[],
	field: string,
	min: number,
	max: number
): HistogramBin[] {
	if (!Array.isArray(objects) || objects.length === 0 || !isValidRange(min, max)) {
		return []
	}

	const binSize = calculateAdaptiveBinSize(min, max)
	const bins = new Map<string, number>()

	for (let binStart = min; binStart < max; binStart += binSize) {
		const binEnd = Math.min(binStart + binSize, max)
		const binKey = formatBinKey(binStart, binEnd)
		bins.set(binKey, 0)
	}

	for (const obj of objects) {
		const value = extractNumericValue(obj, field)
		if (value === null) continue

		const roundedValue = roundValue(value)
		const binIndex = Math.floor((roundedValue - min) / binSize)
		const binStart = min + binIndex * binSize
		const binEnd = Math.min(binStart + binSize, max)
		const binKey = formatBinKey(binStart, binEnd)

		if (bins.has(binKey)) {
			bins.set(binKey, (bins.get(binKey) || 0) + 1)
		} else {
			const lastBinKey = Array.from(bins.keys()).pop()
			if (lastBinKey) {
				bins.set(lastBinKey, (bins.get(lastBinKey) || 0) + 1)
			}
		}
	}

	const out: HistogramBin[] = []
	bins.forEach((count, label) => {
		if (count > 0) {
			out.push({ label, count })
		}
	})
	return out
}
