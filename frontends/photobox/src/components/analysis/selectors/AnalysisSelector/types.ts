interface StatData {
	min: number
	max: number
	avg: number
	median: number
}

interface AnalysisStatistics {
	// Geometric characteristics
	length?: StatData
	width?: StatData
	sq?: StatData
	lwRatio?: StatData
	// RGB color characteristics
	r?: StatData
	g?: StatData
	b?: StatData
	// HSV color characteristics
	h?: StatData
	s?: StatData
	v?: StatData
	// Index signature for dynamic access
	[key: string]: StatData | undefined
}

interface AnalysisFilters {
	sort_by: string
	sort_order: string
	id_analysis: string
	product: string
	show_only_added: boolean
}
export type { AnalysisFilters, AnalysisStatistics, StatData }
