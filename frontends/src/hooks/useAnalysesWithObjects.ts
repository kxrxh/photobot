import { useQuery } from "@tanstack/react-query"
import { fetchAnalysisById, fetchAnalysisObjects } from "@/api/analysis"
import type { AnalysisWithObjects, KalibriObject } from "@/api/analysis/types"
import { CACHE } from "@/api/queryConfig"
import { queryKeys } from "@/api/queryKeys"
import { log } from "@/utils/log"

export interface UseAnalysesWithObjectsResult {
	analyses: AnalysisWithObjects[]
	objectsData: Record<string, KalibriObject[]>
	isFetching: boolean
}

export function useAnalysesWithObjects(
	scopeId: string,
	analysisIds: (string | number)[],
	options?: { enabled?: boolean }
): UseAnalysesWithObjectsResult {
	const ids = analysisIds.map(String)
	const enabled = (options?.enabled ?? true) && ids.length > 0

	const { data: analysesData, isFetching: isAnalysesFetching } = useQuery({
		queryKey: queryKeys.analyses.withObjects(scopeId, ids.join(",")),
		queryFn: async () => {
			const results = await Promise.allSettled(ids.map((id) => fetchAnalysisById(id)))
			const analyses: AnalysisWithObjects[] = []

			for (let i = 0; i < results.length; i += 1) {
				const result = results[i]
				if (result?.status === "fulfilled") {
					analyses.push(result.value)
					continue
				}

				const missingId = ids[i]
				log.devError(
					`Analysis ${missingId} referenced in ${scopeId} is unavailable and will be skipped.`,
					result?.reason
				)
			}

			return analyses
		},
		enabled,
		staleTime: CACHE.staleTime,
	})

	const { data: objectsData, isFetching: isObjectsFetching } = useQuery({
		queryKey: queryKeys.analyses.objectsBatch(scopeId),
		queryFn: async () => {
			if (!analysesData || analysesData.length === 0) return {}
			const results = await Promise.allSettled(analysesData.map((a) => fetchAnalysisObjects(a.id)))
			const objectsByAnalysis: Record<string, KalibriObject[]> = {}

			for (let i = 0; i < analysesData.length; i += 1) {
				const analysis = analysesData[i]
				if (!analysis) {
					continue
				}
				const result = results[i]

				if (result?.status === "fulfilled") {
					objectsByAnalysis[analysis.id] = result.value
					continue
				}

				log.devError(
					`Failed to fetch objects for analysis ${analysis.id}; using empty list.`,
					result?.reason
				)
				objectsByAnalysis[analysis.id] = []
			}

			return objectsByAnalysis
		},
		enabled: enabled && !!analysesData && analysesData.length > 0,
		staleTime: CACHE.staleTime,
	})

	return {
		analyses: analysesData ?? [],
		objectsData: objectsData ?? {},
		isFetching: isAnalysesFetching || isObjectsFetching,
	}
}
