import { useQuery } from "@tanstack/react-query"
import type { ClassificationsResponse } from "@/api/catalog/classifications"
import { fetchClassifications } from "@/api/catalog/classifications"
import { CACHE } from "@/api/queryConfig"
import { queryKeys } from "@/api/queryKeys"

export function useClassifications(options?: { enabled?: boolean }) {
	return useQuery<ClassificationsResponse>({
		queryKey: queryKeys.classifications.all,
		queryFn: fetchClassifications,
		staleTime: CACHE.staleTime,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
		enabled: options?.enabled ?? true,
	})
}
