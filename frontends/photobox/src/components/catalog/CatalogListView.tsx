import { useEffect, useMemo, useRef, useState } from "react"
import { FaFilter, FaList, FaPlus, FaSearch, FaSyncAlt } from "react-icons/fa"
import type { WeedListItem } from "@/api/catalog/types"
import CatalogCard from "@/components/catalog/components/CatalogCard"
import CatalogFilterSheet from "@/components/catalog/dialogs/CatalogFilterSheet"
import ErrorPage from "@/components/common/layout/ErrorPage"
import Loading from "@/components/common/ui/Loading"
import type { CatalogFilters } from "@/hooks/useCatalogFilters"

export type CatalogListStatus = "pending" | "error" | "success"

export interface CatalogListViewProps {
	items: WeedListItem[]
	filters: CatalogFilters
	searchName: string
	onFilterChange: (name: string, value: string | boolean | number | undefined) => void
	onApplyFilters: () => void
	onClearFilters: () => void
	onSearchNameChange: (value: string) => void
	onViewDetails: (item: WeedListItem) => void
	onEditItem?: (item: WeedListItem) => void
	showEditButton?: boolean
	onAddItem: () => void
	onOpenProposals: () => void
	isModerator: boolean
	status: CatalogListStatus
	error?: Error
	hasNextPage: boolean
	isFetchingNextPage: boolean
	isRefreshing: boolean
	onRefresh: () => void
	loadMoreRef: React.Ref<HTMLDivElement | null>
	isFilterDialogOpen: boolean
	onOpenFilterDialog: () => void
	onCloseFilterDialog: () => void
}

export default function CatalogListView({
	items,
	filters,
	searchName,
	onFilterChange,
	onApplyFilters,
	onClearFilters,
	onSearchNameChange,
	onViewDetails,
	onEditItem,
	showEditButton,
	onAddItem,
	onOpenProposals,
	isModerator,
	status,
	error,
	hasNextPage,
	isFetchingNextPage,
	isRefreshing,
	onRefresh,
	loadMoreRef,
	isFilterDialogOpen,
	onOpenFilterDialog,
	onCloseFilterDialog,
}: CatalogListViewProps) {
	const [isPulling, setIsPulling] = useState(false)
	const [pullDistance, setPullDistance] = useState(0)
	const touchStart = useRef(0)
	const PULL_THRESHOLD = 80

	useEffect(() => {
		if (!isRefreshing) {
			setPullDistance(0)
		}
	}, [isRefreshing])

	const getScrollTop = () =>
		window.scrollY || document.documentElement.scrollTop || document.body.scrollTop || 0

	const handleTouchStart = (e: React.TouchEvent<HTMLDivElement>) => {
		touchStart.current = e.touches[0].clientY
		setIsPulling(false)
		setPullDistance(0)
	}

	const handleTouchMove = (e: React.TouchEvent<HTMLDivElement>) => {
		if (isRefreshing) return
		const currentTouch = e.touches[0].clientY
		const scrollTop = getScrollTop()
		const delta = Math.max(0, currentTouch - touchStart.current)

		if (scrollTop === 0 && delta > 0) {
			setIsPulling(true)
			setPullDistance(Math.min(delta, 140))
		}
	}

	const handleTouchEnd = () => {
		if (!isPulling) return
		const shouldRefresh = pullDistance >= PULL_THRESHOLD
		setIsPulling(false)
		setPullDistance(0)
		if (!isRefreshing && shouldRefresh) {
			onRefresh()
		}
	}

	const pullProgress = useMemo(() => Math.min(1, pullDistance / PULL_THRESHOLD), [pullDistance])
	const showPullIndicator = isPulling || pullDistance > 0 || isRefreshing

	return (
		<div onTouchStart={handleTouchStart} onTouchMove={handleTouchMove} onTouchEnd={handleTouchEnd}>
			<header className="flex sticky top-0 z-50 flex-col gap-2 px-2 py-2 w-full bg-base-100">
				<div className="flex justify-between items-center align-center">
					<h1 className="text-2xl font-bold">Каталог</h1>
					<div className="flex gap-2">
						<button type="button" className="btn btn-sm btn-circle" onClick={onOpenFilterDialog}>
							<FaFilter size={18} />
						</button>
						{showEditButton && (
							<>
								<button
									type="button"
									className="btn btn-primary btn-sm max-sm:btn-circle"
									onClick={onAddItem}
								>
									<FaPlus className="w-3 h-3" />
									<span className="hidden sm:inline">Новая заявка</span>
								</button>
								<button type="button" className="btn btn-primary btn-sm" onClick={onOpenProposals}>
									<FaList className="w-3 h-3" />
									{isModerator ? "Предложения" : "Мои предложения"}
								</button>
							</>
						)}
					</div>
				</div>

				<div className="relative">
					<FaSearch className="absolute left-3 top-1/2 z-10 transform -translate-y-1/2 pointer-events-none text-base-content/50" />
					<input
						type="text"
						placeholder="Поиск по названию..."
						value={searchName}
						onChange={(e) => onSearchNameChange(e.target.value)}
						className="pr-4 pl-10 w-full input input-bordered"
					/>
				</div>
			</header>

			<div className="sticky top-16 z-40 flex justify-center pointer-events-none h-0">
				<div
					className={`transition-opacity duration-200 ${
						showPullIndicator ? "opacity-100" : "opacity-0"
					}`}
				>
					<div className="flex items-center gap-2 rounded-full bg-base-200/90 px-3 py-1.5 shadow-lg text-xs font-medium text-base-content">
						<div
							className={`flex h-6 w-6 items-center justify-center rounded-full bg-base-100 shadow-inner ${
								isRefreshing ? "animate-spin" : ""
							}`}
							style={{
								transform: isRefreshing ? "none" : `rotate(${pullProgress * 180}deg)`,
								transition: "transform 150ms ease",
							}}
						>
							<FaSyncAlt className="h-3.5 w-3.5 text-primary" />
						</div>
						<div className="flex flex-col leading-tight text-left">
							<span>{isRefreshing ? "Обновляем каталог…" : "Потяните вниз для обновления"}</span>
							{!isRefreshing && (
								<span className="text-[10px] text-base-content/60">
									{pullProgress >= 1 ? "Отпустите, чтобы обновить" : "Продолжайте тянуть"}
								</span>
							)}
						</div>
						<div className="relative h-1.5 w-16 rounded-full bg-base-300/70 overflow-hidden">
							<div
								className="absolute inset-y-0 left-0 rounded-full bg-primary transition-all duration-150"
								style={{ width: `${isRefreshing ? 100 : pullProgress * 100}%` }}
							/>
						</div>
					</div>
				</div>
			</div>

			{status === "pending" ? (
				<div className="grid grid-cols-2 gap-4 px-2">
					{Array.from({ length: 6 }, (_, i) => `skeleton-${i}`).map((skeletonId) => (
						<div
							key={skeletonId}
							className="flex flex-col p-4 rounded-lg border shadow-sm border-base-300 bg-base-100"
						>
							<div className="w-full h-32 skeleton rounded-lg mb-3" />
							<div className="w-3/4 h-4 skeleton rounded mb-2" />
							<div className="w-1/2 h-3 skeleton rounded mb-2" />
							<div className="w-full h-8 skeleton rounded" />
						</div>
					))}
				</div>
			) : status === "error" ? (
				<div className="min-h-[50vh]">
					<ErrorPage error={error as Error} />
				</div>
			) : (
				<>
					<div className="grid grid-cols-2 gap-4 px-2">
						{items.length === 0 ? (
							<div className="col-span-2 text-center text-base-content/60 py-4">
								Ничего не найдено.
							</div>
						) : (
							items.map((item) => (
								<CatalogCard
									key={item.id}
									item={item}
									onViewDetails={onViewDetails}
									onEditItem={onEditItem}
									showEditButton={showEditButton}
								/>
							))
						)}
					</div>
					<div ref={loadMoreRef}>
						{hasNextPage && (
							<div className="flex justify-center py-4">{isFetchingNextPage && <Loading />}</div>
						)}
					</div>
				</>
			)}

			<CatalogFilterSheet
				isOpen={isFilterDialogOpen}
				onClose={onCloseFilterDialog}
				filters={filters}
				onFilterChange={onFilterChange}
				onApplyFilters={onApplyFilters}
				onClearFilters={onClearFilters}
			/>
		</div>
	)
}
