import { useQuery } from "@tanstack/react-query"
import type { SuccessResponse } from "@/lib/types"

interface UseEntityQueryOptions<TData> {
	queryKey: string[]
	queryFn: () => Promise<SuccessResponse<TData>>
	select?: (response: SuccessResponse<TData>) => TData
}

/**
 * Generic useQuery wrapper for API endpoints that return SuccessResponse<T>.
 * Automatically selects result from response when no custom select is provided.
 */
export function useEntityQuery<TData>({ queryKey, queryFn, select }: UseEntityQueryOptions<TData>) {
	return useQuery<SuccessResponse<TData>, Error, TData>({
		queryKey,
		queryFn,
		select: select ?? ((response: SuccessResponse<TData>) => response.result),
	})
}
