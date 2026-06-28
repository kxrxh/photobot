const DEFAULT_PREFERRED_BINS = 15
const DECIMAL_PLACES = 1
const ROUNDING_MULTIPLIER = 100

export function calculateAdaptiveBinSize(
	min: number,
	max: number,
	preferredBins: number = DEFAULT_PREFERRED_BINS
): number {
	if (min >= max || !Number.isFinite(min) || !Number.isFinite(max)) {
		return 1
	}

	const range = max - min
	const binSize = range / preferredBins

	const magnitude = 10 ** Math.floor(Math.log10(binSize))
	const normalized = binSize / magnitude

	let niceNormalized: number
	if (normalized <= 1) {
		niceNormalized = 1
	} else if (normalized <= 2) {
		niceNormalized = 2
	} else if (normalized <= 5) {
		niceNormalized = 5
	} else {
		niceNormalized = 10
	}

	return niceNormalized * magnitude
}

export function formatBinKey(start: number, end: number): string {
	return `${start.toFixed(DECIMAL_PLACES)}-${end.toFixed(DECIMAL_PLACES)}`
}

export function roundValue(value: number, decimals: number = DECIMAL_PLACES): number {
	const multiplier = ROUNDING_MULTIPLIER * 10 ** (decimals - DECIMAL_PLACES)
	return Math.round(value * multiplier) / multiplier
}

export function isValidRange(min: number, max: number): boolean {
	return (
		!Number.isNaN(min) &&
		!Number.isNaN(max) &&
		min < max &&
		min >= 0 &&
		Number.isFinite(min) &&
		Number.isFinite(max)
	)
}
