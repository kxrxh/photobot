import type { Analysis, AnalysisWithObjects } from "@/api/analysis/types"

/** Shared clock anchor for AnalysisSelector tests. */
export const ANALYSIS_FIXTURE_DATE_TIME = "2026-03-16T12:58:59.000Z"

const defaultUserId = 1

/**
 * Minimal list row for infinite-query pages; defaults match common test fixtures.
 */
export function analysisListRow(
	overrides: Partial<Analysis> & Required<Pick<Analysis, "id" | "files_source" | "files_output">>
): Analysis {
	return {
		date_time: ANALYSIS_FIXTURE_DATE_TIME,
		user_id: defaultUserId,
		...overrides,
	}
}

export function analysisListPage(
	rows: Analysis[],
	options?: { total?: number; limit?: number; offset?: number }
) {
	const limit = options?.limit ?? 10
	const offset = options?.offset ?? 0
	const total = options?.total ?? rows.length
	return { data: rows, total, limit, offset }
}

/** Detail payload returned by `fetchAnalysisById` in tests. */
export function analysisWithObjects(
	overrides: Partial<AnalysisWithObjects> &
		Required<Pick<Analysis, "id" | "files_source" | "files_output">>
): AnalysisWithObjects {
	return {
		date_time: ANALYSIS_FIXTURE_DATE_TIME,
		user_id: defaultUserId,
		objects: [],
		...overrides,
	}
}
