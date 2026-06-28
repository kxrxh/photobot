import type { Analysis, ChannelStats, KalibriObject } from "@/api/analysis/types"
import {
	calculateSampleStdFromValues,
	type FieldStatisticsResult,
} from "@/utils/analysisStatistics"

function channel(analysis: Analysis, key: "l" | "w" | "t"): ChannelStats | undefined {
	return analysis.analysis_params?.[key] ?? analysis[key]
}

export interface PdfLikeKpiCard {
	id: string
	title: string
	value: string
	detail?: string
}

export interface PdfLikeKpis {
	cards: PdfLikeKpiCard[]
	brokenPercentLabel: string
	objectCount: number
	colorRhs?: string
	rgbSwatch?: { r: number; g: number; b: number }
}

function formatMm(n: number): string {
	return `${n.toFixed(2)} мм`
}

function formatGrams(n: number): string {
	return `${n} гр`
}

export function buildPdfLikeKpis(
	analysis: Analysis,
	objects: KalibriObject[] | undefined,
	fieldStats: FieldStatisticsResult
): PdfLikeKpis {
	const ap = analysis.analysis_params
	const product = analysis.product ?? ""
	const productLower = product.toLowerCase()

	const lenCh = channel(analysis, "l")
	const widthCh = channel(analysis, "w")
	const thickCh = channel(analysis, "t")

	const lenAvg = lenCh?.avg
	let lenStd = lenCh?.stddev
	const widthAvg = widthCh?.avg
	let widthStd = widthCh?.stddev

	if (objects && objects.length > 0) {
		if (fieldStats.l?.stddev !== undefined) {
			lenStd = fieldStats.l.stddev
		}
		if (fieldStats.w?.stddev !== undefined) {
			widthStd = fieldStats.w.stddev
		}
	}

	const thicknessValues =
		objects
			?.map((o) => o.t)
			.filter((v): v is number => typeof v === "number" && Number.isFinite(v)) ?? []
	const thicknessStd = calculateSampleStdFromValues(thicknessValues)

	const thickAvg = thickCh?.avg
	let thickStd = thickCh?.stddev
	if (thicknessStd !== null) {
		thickStd = thicknessStd
	}

	let brokenPers: number | undefined
	if (ap && typeof ap.broken_percent === "number" && !Number.isNaN(ap.broken_percent)) {
		brokenPers = ap.broken_percent
	} else if (ap && typeof ap.mass_percent === "number" && !Number.isNaN(ap.mass_percent)) {
		brokenPers = ap.mass_percent
	}

	let brokenPercentLabel = "% негодных"
	if (productLower.includes("kernal") || productLower.includes("kernel")) {
		brokenPercentLabel = "Ломанные %"
	}

	const objectCount = objects?.length ?? 0

	const mass1000 = ap?.mass_1000
	const sampleMass = ap?.mass

	let entities50: string | undefined
	if (typeof sampleMass === "number" && Number.isFinite(sampleMass) && sampleMass > 0) {
		entities50 = (50 / sampleMass).toFixed(2)
	}

	let count50: string | undefined
	const count50Val = ap?.count_50
	if (
		productLower === "seeds_striped" &&
		typeof count50Val === "number" &&
		Number.isFinite(count50Val) &&
		count50Val > 0
	) {
		count50 = count50Val.toString()
	}

	const colorRhs = ap?.color_rhs
	let rgbSwatch: { r: number; g: number; b: number } | undefined
	if (fieldStats.r && fieldStats.g && fieldStats.b) {
		rgbSwatch = {
			r: Math.round(fieldStats.r.avg),
			g: Math.round(fieldStats.g.avg),
			b: Math.round(fieldStats.b.avg),
		}
	}

	const cards: PdfLikeKpiCard[] = []

	if (objects !== undefined) {
		cards.push({
			id: "objectCount",
			title: "Объектов",
			value: String(objectCount),
		})
	}

	if (lenAvg !== undefined && Number.isFinite(lenAvg)) {
		cards.push({
			id: "len",
			title: "Длина",
			value: formatMm(lenAvg),
			detail:
				lenStd !== undefined && Number.isFinite(lenStd) ? `σ: ${formatMm(lenStd)}` : undefined,
		})
	}

	if (widthAvg !== undefined && Number.isFinite(widthAvg)) {
		cards.push({
			id: "width",
			title: "Ширина",
			value: formatMm(widthAvg),
			detail:
				widthStd !== undefined && Number.isFinite(widthStd)
					? `σ: ${formatMm(widthStd)}`
					: undefined,
		})
	}

	if (thickAvg !== undefined && Number.isFinite(thickAvg) && thickAvg !== 0) {
		cards.push({
			id: "thick",
			title: "Толщина",
			value: formatMm(thickAvg),
			detail:
				thickStd !== undefined && Number.isFinite(thickStd)
					? `σ: ${formatMm(thickStd)}`
					: undefined,
		})
	}

	if (mass1000 !== undefined && Number.isFinite(mass1000) && mass1000 !== 0) {
		cards.push({
			id: "mass1000",
			title: "Масса 1000 зерен",
			value: formatGrams(mass1000),
		})
	}

	if (sampleMass !== undefined && Number.isFinite(sampleMass) && sampleMass !== 0) {
		cards.push({
			id: "sampleMass",
			title: "Масса образца",
			value: formatGrams(sampleMass),
		})
	}

	if (productLower === "seeds" && entities50 && entities50 !== "0.00") {
		cards.push({
			id: "entities50",
			title: "Образцов на 50 грамм",
			value: entities50,
		})
	}

	if (count50) {
		cards.push({
			id: "count50",
			title: "Количество в 50 грамм",
			value: count50,
		})
	}

	if (brokenPers !== undefined && Number.isFinite(brokenPers)) {
		cards.push({
			id: "broken",
			title: brokenPercentLabel,
			value: `${brokenPers} %`,
		})
	}

	if (colorRhs && colorRhs !== "-") {
		cards.push({
			id: "rhs",
			title: "Цвет по RHS",
			value: colorRhs,
		})
	}

	return {
		cards,
		brokenPercentLabel,
		objectCount,
		colorRhs: colorRhs && colorRhs !== "-" ? colorRhs : undefined,
		rgbSwatch,
	}
}
