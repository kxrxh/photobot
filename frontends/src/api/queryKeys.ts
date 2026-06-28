export const queryKeys = {
	analyses: {
		all: ["analyses"] as const,
		list: (params?: Record<string, unknown>) => ["analyses", params] as const,
		detail: (id: string) => ["analyses", id] as const,
		objects: (analysisId: string) => ["analysisObjects", analysisId] as const,
		withObjects: (scopeId: string, analysisIds?: string) =>
			["analyses", scopeId, analysisIds ?? ""] as const,
		objectsBatch: (scopeId: string) => ["analysesObjects", scopeId] as const,
	},
	catalog: {
		items: ["catalogItems"] as const,
		itemsList: (filters: unknown, searchName: string) =>
			["catalogItems", filters, searchName] as const,
		itemSelector: (search?: string) => ["catalogItemsSelector", search ?? ""] as const,
		itemDetails: (id: string) => ["itemDetails", id] as const,
		weedDetails: (id: string) => ["weedDetails", id] as const,
		proposals: ["catalog-proposals"] as const,
		proposalsModeration: (userId: number, requestByParam: string, sortOrder: string) =>
			["catalog-proposals", "moderation", userId, requestByParam, sortOrder] as const,
		proposalsUser: (userId: number, statusFilter: string) =>
			["catalog-proposals", "user", userId, statusFilter] as const,
		proposal: (id: string) => ["catalog-proposal", id] as const,
		objects: (catalogAnalysisId: string) => ["catalogObjects", catalogAnalysisId] as const,
	},
	classifications: {
		all: ["classifications"] as const,
		search: (term?: string, product?: string) =>
			["classifications", term ?? "", product ?? ""] as const,
	},
	markups: (params?: { name?: string }) => ["markups", params ?? {}] as const,
	notes: (catalogItemId: string) => ["notes", catalogItemId] as const,
	products: ["products"] as const,
	params: ["params"] as const,
	requests: {
		all: ["requests"] as const,
		list: (statusFilter?: string) => ["requests", statusFilter ?? ""] as const,
	},
	requestObjects: (requestId: string) => ["requestObjects", requestId] as const,
	userActiveClassification: ["user-active-classification"] as const,
	classificationParams: ["classification-params"] as const,
	completeClassification: (id: string) => ["complete-classification", id] as const,
} as const
