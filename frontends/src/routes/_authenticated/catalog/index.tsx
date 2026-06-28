import { useInfiniteQuery } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect, useMemo, useRef, useState } from "react"
import { useInView } from "react-intersection-observer"
import { fetchWeedList } from "@/api/catalog"
import type { WeedListItem } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import CatalogListView from "@/components/catalog/CatalogListView"
import {
	canEditCatalogItem,
	isModeratorRole,
} from "@/components/catalog/components/CatalogProposalsShared"
import { useAuth } from "@/contexts/AuthContext"
import { useCatalogFilters } from "@/hooks/useCatalogFilters"

const PAGE_SIZE = 12

export const Route = createFileRoute("/_authenticated/catalog/")({
	component: () => {
		const navigate = useNavigate()
		const { roles } = useAuth()
		const [isFilterDialogOpen, setIsFilterDialogOpen] = useState(false)
		const { filters, searchName, updateFilter, clearFilters, updateSearchName } =
			useCatalogFilters()

		const isModerator = isModeratorRole(roles)
		const showEditButton = canEditCatalogItem(roles)

		const STORAGE_KEY = "catalog:listState"
		const hasRestoredRef = useRef(false)

		const saveListState = () => {
			try {
				sessionStorage.setItem(STORAGE_KEY, JSON.stringify({ filters, searchName }))
			} catch {}
		}

		useEffect(() => {
			if (hasRestoredRef.current) return
			hasRestoredRef.current = true
			try {
				const raw = sessionStorage.getItem(STORAGE_KEY)
				if (!raw) return
				const parsed = JSON.parse(raw) as {
					filters?: typeof filters
					searchName?: string
				}
				if (parsed?.filters) {
					Object.entries(parsed.filters).forEach(([k, v]) => {
						updateFilter(k as string, v as unknown as string | boolean | number | undefined)
					})
				}
				if (typeof parsed?.searchName === "string") {
					updateSearchName(parsed.searchName)
				}
			} catch {
			} finally {
				sessionStorage.removeItem(STORAGE_KEY)
			}
		}, [updateFilter, updateSearchName])

		const {
			data,
			error,
			fetchNextPage,
			hasNextPage,
			isFetchingNextPage,
			isFetching,
			refetch,
			status,
		} = useInfiniteQuery({
			queryKey: queryKeys.catalog.itemsList(filters, searchName),
			queryFn: ({ pageParam }) =>
				fetchWeedList({
					name: searchName || undefined,
					main_group: filters.main_group,
					main_subgroup: filters.main_subgroup,
					subgroup: filters.subgroup,
					is_quarantine: filters.is_quarantine,
					limit: PAGE_SIZE,
					offset: pageParam,
					sort_order: filters.sort_order,
					l_min: filters.l_min,
					l_max: filters.l_max,
					w_min: filters.w_min,
					w_max: filters.w_max,
					lw_min: filters.lw_min,
					lw_max: filters.lw_max,
					h_min: filters.h_min,
					h_max: filters.h_max,
					s_min: filters.s_min,
					s_max: filters.s_max,
					v_min: filters.v_min,
					v_max: filters.v_max,
					r_min: filters.r_min,
					r_max: filters.r_max,
					g_min: filters.g_min,
					g_max: filters.g_max,
					b_min: filters.b_min,
					b_max: filters.b_max,
					brt_min: filters.brt_min,
					brt_max: filters.brt_max,
					sq_sqcrl_min: filters.sq_sqcrl_min,
					sq_sqcrl_max: filters.sq_sqcrl_max,
				}),
			initialPageParam: 0,
			getNextPageParam: (lastPage) => {
				const nextOffset = lastPage.offset + lastPage.limit
				if (nextOffset < lastPage.total) {
					return nextOffset
				}
				return undefined
			},
			staleTime: 2 * 60 * 1000,
			refetchOnWindowFocus: false,
			refetchOnReconnect: true,
		})

		const { ref, inView } = useInView()

		useEffect(() => {
			if (inView && hasNextPage && !isFetchingNextPage) {
				fetchNextPage()
			}
		}, [inView, hasNextPage, isFetchingNextPage, fetchNextPage])

		const catalogItems = useMemo(() => data?.pages.flatMap((page) => page.data) ?? [], [data])

		const handleViewDetails = (item: WeedListItem) => {
			saveListState()
			navigate({
				to: "/catalog/$catalogItemId",
				params: { catalogItemId: String(item.id) },
			})
		}

		const handleAddItem = () => {
			saveListState()
			navigate({ to: "/catalog/add" })
		}

		const handleEditItem = (item: WeedListItem) => {
			saveListState()
			navigate({
				to: "/catalog/$catalogItemId/edit",
				params: { catalogItemId: String(item.id) },
			})
		}

		const handleFilterChange = (name: string, value: string | boolean | number | undefined) => {
			updateFilter(name, value)
		}

		const handleApplyFilters = () => {
			setIsFilterDialogOpen(false)
		}

		const listStatus: "pending" | "error" | "success" =
			status === "pending" ? "pending" : status === "error" ? "error" : "success"
		const isRefreshing = isFetching && !isFetchingNextPage

		return (
			<CatalogListView
				items={catalogItems}
				filters={filters}
				searchName={searchName}
				onFilterChange={handleFilterChange}
				onApplyFilters={handleApplyFilters}
				onClearFilters={clearFilters}
				onSearchNameChange={updateSearchName}
				onViewDetails={handleViewDetails}
				onEditItem={handleEditItem}
				showEditButton={showEditButton}
				onAddItem={handleAddItem}
				onOpenProposals={() => navigate({ to: "/catalog/proposals" })}
				isModerator={isModerator}
				status={listStatus}
				error={error as Error | undefined}
				hasNextPage={!!hasNextPage}
				isFetchingNextPage={isFetchingNextPage}
				isRefreshing={isRefreshing}
				onRefresh={() => refetch()}
				loadMoreRef={ref}
				isFilterDialogOpen={isFilterDialogOpen}
				onOpenFilterDialog={() => setIsFilterDialogOpen(true)}
				onCloseFilterDialog={() => setIsFilterDialogOpen(false)}
			/>
		)
	},
})
