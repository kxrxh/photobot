import type { AnalysisRequest } from "@/api/analysis/types"

export const REQUESTS_TAB_FIXTURE_TIME = "2026-03-16T12:58:59.000Z"

export function analysisRequestFixture(overrides: Partial<AnalysisRequest> = {}): AnalysisRequest {
	return {
		id: "request-1",
		user_id: "42",
		product: "Tomato",
		status: "created",
		created_at: REQUESTS_TAB_FIXTURE_TIME,
		updated_at: REQUESTS_TAB_FIXTURE_TIME,
		...overrides,
	}
}
