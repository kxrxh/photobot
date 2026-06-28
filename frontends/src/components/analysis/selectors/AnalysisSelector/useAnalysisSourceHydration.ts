import { useQueries } from "@tanstack/react-query"
import { useMemo } from "react"
import { fetchAnalysisById } from "@/api/analysis"
import type { Analysis } from "@/api/analysis/types"
import { CACHE } from "@/api/queryConfig"
import { queryKeys } from "@/api/queryKeys"
import { getAnalysisSourceUrls } from "@/utils/image"

/**
 * Fetches analysis details for list cards missing preview URLs (cached via React Query).
 */
export function useAnalysisSourceHydration(analyses: Analysis[], isOpen: boolean) {
	const idsToHydrate = useMemo(
		() =>
			analyses
				.filter((analysis) => getAnalysisSourceUrls(analysis).length === 0)
				.map((analysis) => analysis.id),
		[analyses]
	)

	const detailQueries = useQueries({
		queries: idsToHydrate.map((id) => ({
			queryKey: queryKeys.analyses.detail(id),
			queryFn: () => fetchAnalysisById(id),
			enabled: isOpen && Boolean(id),
			staleTime: CACHE.staleTime,
			gcTime: CACHE.gcTime,
		})),
	})

	return useMemo(() => {
		const map: Record<string, string[]> = {}
		for (let i = 0; i < idsToHydrate.length; i++) {
			const data = detailQueries[i]?.data
			if (!data) continue
			const urls = getAnalysisSourceUrls(data)
			if (urls.length > 0) {
				map[idsToHydrate[i]] = urls
			}
		}
		return map
	}, [detailQueries, idsToHydrate])
}
