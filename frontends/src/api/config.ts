export const API_ENDPOINTS = {
	auth: import.meta.env.VITE_AUTH_API_URL ?? "",
	catalog: import.meta.env.VITE_WEED_API_URL ?? "",
	classification: import.meta.env.VITE_CLASSIFICATION_API_URL ?? "",
	analysis: import.meta.env.VITE_ANALYSIS_API_URL ?? "",
	reports: import.meta.env.VITE_REPORTS_API_URL ?? "",
} as const
