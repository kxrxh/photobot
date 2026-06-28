import { useInfiniteQuery, useQueryClient } from "@tanstack/react-query"
import { AnimatePresence, motion } from "framer-motion"
import type React from "react"
import { useEffect, useMemo, useRef, useState } from "react"
import { createPortal } from "react-dom"
import {
	FaExclamationTriangle,
	FaFilter,
	FaList,
	FaPlus,
	FaRegFileAlt,
	FaTrash,
} from "react-icons/fa"
import { IoClose } from "react-icons/io5"
import { useInView } from "react-intersection-observer"
import { fetchAnalyses, fetchAnalysisById } from "@/api/analysis"
import type { Analysis, AnalysisWithObjects } from "@/api/analysis/types"
import { CACHE } from "@/api/queryConfig"
import { queryKeys } from "@/api/queryKeys"
import type { PaginatedResponse } from "@/api/types"
import AnalysisViewSheet from "@/components/analysis/dialogs/AnalysisViewSheet"
import Loading from "@/components/common/ui/Loading"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { useMessenger } from "@/hooks/useMessenger"
import { downloadAnalysisReport } from "@/utils/analysis"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import { resolveChatPlatform } from "@/utils/messengerPlatform"
import { AnalysisCard } from "./AnalysisCard"
import AnalysisFilterSheet from "./AnalysisFilterSheet"
import type { AnalysisFilters } from "./types"
import { useAnalysisSourceHydration } from "./useAnalysisSourceHydration"

const AnalysisSkeleton = () => (
	<div className="relative flex flex-col gap-4 rounded-3xl border border-base-200 bg-base-100 p-4 shadow-sm backdrop-blur-md animate-pulse sm:p-5">
		<div className="flex flex-col gap-4 sm:flex-row sm:items-center">
			<div className="relative w-full shrink-0 overflow-hidden rounded-2xl border border-base-200 bg-base-200 aspect-4/3 sm:w-28 sm:aspect-square">
				<div className="absolute inset-0 bg-linear-to-br from-base-300 via-base-200 to-base-300" />
			</div>
			<div className="flex flex-1 flex-col gap-3">
				<div className="h-4 w-40 rounded-full bg-base-300" />
				<div className="h-3 w-32 rounded-full bg-base-300" />
			</div>
		</div>
		<div className="flex flex-wrap items-center gap-2">
			<div className="h-9 w-28 rounded-full bg-base-300" />
			<div className="h-9 w-28 rounded-full bg-base-300" />
			<div className="h-9 w-32 rounded-full bg-base-300" />
		</div>
	</div>
)

interface AnalysisSelectorDialogProps {
	isOpen: boolean
	onClose: () => void
	selectedAnalysisIds: string[]
	onAddAnalysis: (analysis: Analysis) => void
	onRemoveAnalysis: (analysisId: string) => void
	onRemoveAllAnalyses: () => void
	hasAddedAnalyses: boolean
	selectionMode?: "single" | "multiple"
	mode?: "selection" | "view-only"
	onOpenCreateDialog?: () => void
	/**
	 * When true, sheet stacks below the layout footer so "Назад в меню" stays visible.
	 * Use for /analysis/list and /analysis/create only — embedded dialogs must keep the header close.
	 */
	usesGlobalMenuBackButton?: boolean
}

const PAGE_SIZE = 10

const AnalysisSelectorDialog: React.FC<AnalysisSelectorDialogProps> = ({
	isOpen,
	onClose,
	selectedAnalysisIds,
	onAddAnalysis,
	onRemoveAnalysis,
	onRemoveAllAnalyses,
	hasAddedAnalyses,
	selectionMode = "multiple",
	mode = "selection",
	onOpenCreateDialog,
	usesGlobalMenuBackButton = false,
}) => {
	const [filters, setFilters] = useState<AnalysisFilters>({
		sort_by: "date_time",
		sort_order: "desc",
		id_analysis: "",
		product: "",
		show_only_added: false,
	})

	const [pendingFilters, setPendingFilters] = useState<AnalysisFilters>({
		sort_by: "date_time",
		sort_order: "desc",
		id_analysis: "",
		product: "",
		show_only_added: false,
	})

	const [isFilterDialogOpen, setIsFilterDialogOpen] = useState(false)
	const [isViewDialogOpen, setIsViewDialogOpen] = useState(false)
	const [selectedAnalysisForView, setSelectedAnalysisForView] =
		useState<AnalysisWithObjects | null>(null)
	const [isLoadingViewAnalysis, setIsLoadingViewAnalysis] = useState(false)
	const analysesListRef = useRef<HTMLDivElement>(null)
	const queryClient = useQueryClient()
	const [showSelected, setShowSelected] = useState(false)
	const [downloadingAnalysisIds, setDownloadingAnalysisIds] = useState<Set<string>>(new Set())

	const [isPulling, setIsPulling] = useState(false)
	const touchStart = useRef(0)

	useEffect(() => {
		if (selectedAnalysisIds.length === 0) {
			setShowSelected(false)
		}
	}, [selectedAnalysisIds])

	const { userId } = useAuth()
	const { showError, showSuccess } = useAlert()
	const { requestFileDownload, hapticFeedback, platform } = useMessenger()

	const {
		data: analysesData,
		isLoading: loading,
		isFetching,
		error,
		fetchNextPage,
		hasNextPage,
		isFetchingNextPage,
		refetch,
	} = useInfiniteQuery<PaginatedResponse<Analysis>, Error>({
		queryKey: queryKeys.analyses.list({
			userId: userId ?? 0,
			sort_by: filters.sort_by,
			sort_order: filters.sort_order,
			product: filters.product,
			id_analysis: filters.id_analysis,
		}),
		queryFn: ({ pageParam = 0 }) => {
			const params = {
				limit: PAGE_SIZE,
				offset: pageParam as number,
				sort_by: filters.sort_by as "date_time" | "id" | "product",
				sort_order: filters.sort_order as "asc" | "desc",
				...(filters.product && { product: filters.product }),
				...(filters.id_analysis && { id: filters.id_analysis }),
			}
			if (userId) {
				return fetchAnalyses(params)
			}
			return Promise.reject(new Error("User not found"))
		},
		enabled: isOpen && !!userId,
		initialPageParam: 0,
		getNextPageParam: (lastPage) => {
			const nextOffset = lastPage.offset + lastPage.limit
			if (nextOffset < lastPage.total) {
				return nextOffset
			}
			return undefined
		},
		staleTime: CACHE.staleTime,
		gcTime: CACHE.gcTime,
		refetchOnWindowFocus: false,
		refetchOnMount: false,
		refetchOnReconnect: true,
	})

	const allAnalyses = useMemo(
		() => analysesData?.pages.flatMap((page) => page.data) ?? [],
		[analysesData?.pages]
	)

	const isLoadingData =
		loading || (isFetching && !isFetchingNextPage && (!analysesData || allAnalyses.length === 0))

	const { ref: loadMoreRef, inView } = useInView({
		threshold: 0,
		rootMargin: "200px",
		triggerOnce: false,
		skip: !hasNextPage || isFetchingNextPage,
	})

	const selectedAnalyses = allAnalyses.filter((analysis) =>
		selectedAnalysisIds.includes(analysis.id)
	)

	const filteredAndSortedAnalyses = useMemo(() => {
		let analyses = [...allAnalyses]

		if (filters.show_only_added) {
			analyses = analyses.filter((analysis) => selectedAnalysisIds.includes(analysis.id))
		}

		return analyses
	}, [allAnalyses, filters.show_only_added, selectedAnalysisIds])

	const hydratedSourceUrlsByAnalysisId = useAnalysisSourceHydration(
		filteredAndSortedAnalyses,
		isOpen
	)

	useEffect(() => {
		if (isOpen && analysesListRef.current) {
			analysesListRef.current.scrollTop = 0
		}
	}, [isOpen])

	useEffect(() => {
		if (inView && hasNextPage && !isFetchingNextPage) {
			fetchNextPage()
		}
	}, [inView, hasNextPage, isFetchingNextPage, fetchNextPage])

	const handleTouchStart = (e: React.TouchEvent<HTMLDivElement>) => {
		touchStart.current = e.touches[0].clientY
	}

	const handleTouchMove = (e: React.TouchEvent<HTMLDivElement>) => {
		if (!analysesListRef.current) return

		const currentTouch = e.touches[0].clientY
		const scrollTop = analysesListRef.current.scrollTop

		if (scrollTop === 0 && currentTouch > touchStart.current + 50) {
			setIsPulling(true)
		}
	}

	const handleTouchEnd = () => {
		if (isPulling) {
			refetch()
			setIsPulling(false)
		}
	}

	const handleFilterChange = (name: string, value: string | boolean) => {
		setPendingFilters((prev) => ({
			...prev,
			[name]: value,
		}))
	}

	const handleClearFilters = () => {
		const clearedFilters = {
			sort_by: "date_time",
			sort_order: "desc",
			id_analysis: "",
			product: "",
			show_only_added: false,
		}
		setPendingFilters(clearedFilters)
		setFilters(clearedFilters)
		setIsFilterDialogOpen(false)
	}

	const handleApplyFilters = () => {
		setFilters(pendingFilters)
		setIsFilterDialogOpen(false)
	}

	const handleOpenFilterDialog = () => {
		setPendingFilters(filters)
		setIsFilterDialogOpen(true)
	}

	const handleCloseFilterDialog = () => {
		setIsFilterDialogOpen(false)
		setPendingFilters(filters)
	}

	const handleAnalysisClick = (analysis: Analysis) => {
		hapticFeedback.impactOccurred("light")

		const isSelected = selectedAnalysisIds.includes(analysis.id)

		if (selectionMode === "multiple") {
			if (isSelected) {
				onRemoveAnalysis(analysis.id)
			} else {
				onAddAnalysis(analysis)
			}
		} else {
			if (isSelected) {
				onRemoveAnalysis(analysis.id)
			} else {
				if (selectedAnalysisIds.length > 0) {
					onRemoveAnalysis(selectedAnalysisIds[0])
				}
				onAddAnalysis(analysis)
			}
		}
	}

	const handleViewAnalysis = async (analysis: Analysis) => {
		setIsLoadingViewAnalysis(true)
		try {
			const fullAnalysis = await queryClient.fetchQuery({
				queryKey: queryKeys.analyses.detail(analysis.id),
				queryFn: () => fetchAnalysisById(analysis.id),
				staleTime: CACHE.staleTime,
			})
			setSelectedAnalysisForView(fullAnalysis)
			setIsViewDialogOpen(true)
		} catch (error) {
			log.devError("Failed to fetch analysis:", error)
			showError(getUserFacingErrorMessage(error))
		} finally {
			setIsLoadingViewAnalysis(false)
		}
	}

	const handleDownloadReport = async (analysis: Analysis) => {
		const analysisId = analysis.id

		setDownloadingAnalysisIds((prev) => new Set(prev).add(analysisId))

		try {
			const chatPlatform = resolveChatPlatform(platform)
			const { sendToChatOk } = await downloadAnalysisReport(
				analysis.id,
				requestFileDownload,
				chatPlatform ? { platform: chatPlatform } : undefined
			)
			showSuccess(sendToChatOk ? "Отчёт скачан. Отправлен в чат." : "Отчёт скачан.")
		} catch (error) {
			log.devError("Error downloading report:", error)
			showError(getUserFacingErrorMessage(error))
		} finally {
			setDownloadingAnalysisIds((prev) => {
				const newSet = new Set(prev)
				newSet.delete(analysisId)
				return newSet
			})
		}
	}

	const handleCloseViewDialog = () => {
		setIsViewDialogOpen(false)
		setSelectedAnalysisForView(null)
	}

	return createPortal(
		<AnimatePresence>
			{isOpen && (
				<>
					<motion.div
						initial={{ opacity: 0 }}
						animate={{ opacity: 1 }}
						exit={{ opacity: 0 }}
						transition={{ duration: 0.2 }}
						className="fixed inset-0 z-40 bg-black/50"
						onClick={onClose}
					/>

					<motion.div
						initial={{ opacity: 0 }}
						animate={{ opacity: 1 }}
						exit={{ opacity: 0 }}
						transition={{ duration: 0.2 }}
						className={`fixed inset-0 flex flex-col bg-base-100 ${usesGlobalMenuBackButton ? "z-40" : "z-50"}`}
					>
						<div className="sticky top-0 z-10 bg-base-100 border-b border-base-200">
							<div className="flex justify-between items-center p-2">
								<div className="flex flex-col gap-1">
									<h1 className="text-xl font-bold text-base-content leading-tight">Анализы</h1>
									{mode === "selection" && selectedAnalysisIds.length > 0 && (
										<p className="text-sm text-base-content/70">
											Выбрано: {selectedAnalysisIds.length}
										</p>
									)}
								</div>
								<div className="flex gap-2 items-center">
									{onOpenCreateDialog && (
										<button
											type="button"
											onClick={onOpenCreateDialog}
											className="btn btn-soft btn-circle btn-sm btn-primary"
											title="Создать анализ"
											aria-label="Открыть диалог создания анализа"
										>
											<FaPlus size={18} />
										</button>
									)}
									{mode === "selection" &&
										selectionMode === "multiple" &&
										selectedAnalysisIds.length > 0 && (
											<button
												type="button"
												onClick={() => setShowSelected(!showSelected)}
												className={`btn btn-soft btn-circle btn-sm relative ${
													showSelected ? "btn-primary" : "btn-ghost"
												}`}
												title="Показать выбранные"
												aria-label="Показать выбранные анализы"
											>
												<FaList size={18} />
												{!showSelected && (
													<span className="absolute -top-1 -right-1 flex items-center justify-center w-5 h-5 rounded-full bg-primary text-primary-content text-xs font-bold">
														{selectedAnalysisIds.length}
													</span>
												)}
											</button>
										)}
									<button
										type="button"
										onClick={handleOpenFilterDialog}
										className="btn btn-soft btn-circle btn-sm"
										title="Фильтры"
										aria-label="Открыть фильтры"
									>
										<FaFilter size={18} />
									</button>
									{!usesGlobalMenuBackButton && (
										<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть диалог" />
									)}
								</div>
							</div>
						</div>

						{(filters.product || filters.id_analysis) && (
							<div className="flex gap-2 overflow-x-auto px-4 py-2 border-b border-base-200 no-scrollbar">
								{filters.product && (
									<div className="badge badge-primary gap-2 shrink-0">
										Продукт: {filters.product}
										<button
											type="button"
											onClick={() => handleFilterChange("product", "")}
											className="btn btn-ghost btn-xs btn-circle"
											aria-label="Удалить фильтр по продукту"
										>
											<IoClose size={14} />
										</button>
									</div>
								)}
								{filters.id_analysis && (
									<div className="badge badge-primary gap-2 shrink-0">
										ID: {filters.id_analysis}
										<button
											type="button"
											onClick={() => handleFilterChange("id_analysis", "")}
											className="btn btn-ghost btn-xs btn-circle"
											aria-label="Удалить фильтр по ID"
										>
											<IoClose size={14} />
										</button>
									</div>
								)}
								<button
									type="button"
									onClick={handleClearFilters}
									className="badge badge-ghost gap-2 shrink-0"
								>
									Очистить все
								</button>
							</div>
						)}

						<AnimatePresence>
							{mode === "selection" && showSelected && selectedAnalyses.length > 0 && (
								<motion.div
									initial={{ height: 0, opacity: 0 }}
									animate={{ height: "auto", opacity: 1 }}
									exit={{ height: 0, opacity: 0 }}
									transition={{ duration: 0.2, ease: "easeInOut" }}
									className="overflow-hidden border-b border-base-200"
								>
									<div className="p-2 space-y-2 bg-base-100">
										<div className="flex justify-between items-center">
											<h3 className="font-semibold">Выбрано: {selectedAnalyses.length}</h3>
											<button
												type="button"
												className="btn btn-ghost btn-xs"
												onClick={onRemoveAllAnalyses}
											>
												<FaTrash className="mr-1" />
												Очистить все
											</button>
										</div>
										<div className="py-1 pr-1 pl-2 space-y-2 max-h-36 overflow-y-auto">
											{selectedAnalyses.map((analysis: Analysis) => (
												<div
													key={`selected-${analysis.id}`}
													className="flex justify-between items-center p-2 rounded-lg bg-base-200"
												>
													<span className="text-sm">Анализ #{analysis.id}</span>
													<button
														type="button"
														className="btn btn-ghost btn-xs btn-circle"
														onClick={() => onRemoveAnalysis(analysis.id)}
													>
														<IoClose size={16} />
													</button>
												</div>
											))}
										</div>
									</div>
								</motion.div>
							)}
						</AnimatePresence>

						<AnalysisFilterSheet
							isOpen={isFilterDialogOpen}
							onClose={handleCloseFilterDialog}
							filters={pendingFilters}
							onFilterChange={handleFilterChange}
							onApplyFilters={handleApplyFilters}
							onClearFilters={handleClearFilters}
							hasAddedAnalyses={hasAddedAnalyses}
						/>

						{isLoadingViewAnalysis && (
							<div className="fixed inset-0 z-60 flex items-center justify-center bg-black/30 backdrop-blur-sm">
								<div className="flex flex-col items-center gap-3 rounded-2xl bg-base-100 p-6 shadow-xl">
									<span className="loading loading-spinner loading-lg text-primary" />
									<span className="text-sm font-medium text-base-content">Загрузка анализа...</span>
								</div>
							</div>
						)}

						<AnalysisViewSheet
							isOpen={isViewDialogOpen}
							onClose={handleCloseViewDialog}
							analysis={selectedAnalysisForView}
						/>

						<div
							className={`overflow-y-auto flex-1 p-2 space-y-4 ${usesGlobalMenuBackButton ? "pb-16" : ""}`}
							ref={analysesListRef}
							onTouchStart={handleTouchStart}
							onTouchMove={handleTouchMove}
							onTouchEnd={handleTouchEnd}
						>
							{error && !isFetching ? (
								<div className="flex flex-col justify-center items-center h-full">
									<div className="max-w-md p-6 text-center">
										<div className="mb-4 text-primary">
											<FaExclamationTriangle size={48} className="mx-auto" />
										</div>
										<h3 className="mb-2 text-lg font-semibold">Ошибка загрузки</h3>
										<p className="mb-4 text-sm text-base-content/70">
											{getUserFacingErrorMessage(error)}
										</p>
										<button
											type="button"
											onClick={() => refetch()}
											className="btn btn-primary btn-sm"
										>
											Повторить попытку
										</button>
									</div>
								</div>
							) : isLoadingData ? (
								<div className="space-y-4 p-2">
									{["first", "second", "third"].map((key) => (
										<AnalysisSkeleton key={`skeleton-${key}`} />
									))}
								</div>
							) : filteredAndSortedAnalyses.length === 0 ? (
								<div className="flex flex-col justify-center items-center h-full">
									<div className="max-w-md p-6 text-center">
										<div className="mb-4 text-base-content/50">
											<FaRegFileAlt size={48} className="mx-auto" />
										</div>
										<h3 className="mb-2 text-lg font-semibold">Анализы не найдены</h3>
										<p className="text-sm text-base-content/70">
											{filters.show_only_added
												? "Нет добавленных анализов"
												: "Попробуйте изменить параметры фильтрации"}
										</p>
									</div>
								</div>
							) : (
								<>
									{filteredAndSortedAnalyses.map((analysis) => (
										<AnalysisCard
											key={analysis.id}
											analysis={analysis}
											imageUrls={hydratedSourceUrlsByAnalysisId[analysis.id]}
											isSelected={selectedAnalysisIds.includes(analysis.id)}
											isDownloading={downloadingAnalysisIds.has(analysis.id)}
											mode={mode}
											selectionMode={selectionMode}
											onSelect={handleAnalysisClick}
											onView={handleViewAnalysis}
											onDownload={handleDownloadReport}
										/>
									))}
									{hasNextPage && (
										<div ref={loadMoreRef} className="flex justify-center py-4">
											{isFetchingNextPage && <Loading />}
										</div>
									)}
								</>
							)}
						</div>
					</motion.div>
				</>
			)}
		</AnimatePresence>,
		document.body
	)
}

export default AnalysisSelectorDialog
