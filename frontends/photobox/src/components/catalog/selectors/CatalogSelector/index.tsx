import { useInfiniteQuery } from "@tanstack/react-query"
import { AnimatePresence, motion } from "framer-motion"
import type React from "react"
import { useEffect, useMemo, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { FaList, FaSearch, FaTrash } from "react-icons/fa"
import { IoClose } from "react-icons/io5"
import { useInView } from "react-intersection-observer"
import { fetchWeedList } from "@/api/catalog"
import type { WeedListItem } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import CatalogItemDetailSheet from "@/components/catalog/dialogs/CatalogItemDetailSheet"
import CatalogSelectorCard from "@/components/catalog/selectors/CatalogSelector/CatalogSelectorCard"
import ErrorPage from "@/components/common/layout/ErrorPage"
import Loading from "@/components/common/ui/Loading"

const PAGE_SIZE = 12

const CATALOG_GRID_SKELETON_KEYS = Array.from({ length: 6 }, () => crypto.randomUUID())

interface CatalogSelectorSheetProps {
	isOpen: boolean
	onClose: () => void
	initialSelectedItems: WeedListItem[]
	onConfirm: (selectedItems: WeedListItem[]) => void
}

const CatalogSelectorSheet: React.FC<CatalogSelectorSheetProps> = ({
	isOpen,
	onClose,
	initialSelectedItems,
	onConfirm,
}) => {
	const dialogRef = useRef<HTMLDialogElement>(null)
	const [selectedItems, setSelectedItems] = useState<WeedListItem[]>(initialSelectedItems)
	const allItemsMap = useRef(new Map<number, WeedListItem>())
	const [isPreviewOpen, setIsPreviewOpen] = useState(false)
	const [itemForPreview, setItemForPreview] = useState<WeedListItem | null>(null)
	const [showSelected, setShowSelected] = useState(false)
	const [searchName, setSearchName] = useState("")

	const { ref, inView } = useInView()

	const { data, error, fetchNextPage, hasNextPage, isFetchingNextPage, status } = useInfiniteQuery({
		queryKey: queryKeys.catalog.itemSelector(searchName),
		queryFn: ({ pageParam }) =>
			fetchWeedList({
				limit: PAGE_SIZE,
				offset: pageParam,
				sort_order: "desc",
				name: searchName || undefined,
				main_group: undefined,
				main_subgroup: undefined,
				subgroup: undefined,
				is_quarantine: undefined,
			}),
		initialPageParam: 0,
		getNextPageParam: (lastPage) => {
			const nextOffset = lastPage.offset + lastPage.limit
			if (nextOffset < lastPage.total) {
				return nextOffset
			}
			return undefined
		},
		enabled: isOpen,
		staleTime: 5 * 60 * 1000,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
	})

	const catalogItems = useMemo(() => data?.pages.flatMap((page) => page.data) ?? [], [data])

	useEffect(() => {
		if (isOpen) {
			setSelectedItems(initialSelectedItems)
			initialSelectedItems.forEach((item) => {
				if (!allItemsMap.current.has(item.id)) {
					allItemsMap.current.set(item.id, item)
				}
			})
		}
	}, [isOpen, initialSelectedItems])

	useEffect(() => {
		catalogItems.forEach((item) => {
			if (!allItemsMap.current.has(item.id)) {
				allItemsMap.current.set(item.id, item)
			}
		})
	}, [catalogItems])

	const selectedItemsWithData = useMemo(() => {
		return selectedItems
			.map((selected) => {
				const fullItem = allItemsMap.current.get(selected.id)
				return fullItem || selected
			})
			.filter(Boolean)
	}, [selectedItems])

	useEffect(() => {
		if (selectedItems.length === 0) {
			setShowSelected(false)
		}
	}, [selectedItems])

	useEffect(() => {
		if (dialogRef.current) {
			if (isOpen) {
				dialogRef.current.showModal()
			} else {
				dialogRef.current.close()
			}
		}
	}, [isOpen])

	useEffect(() => {
		if (inView && hasNextPage && !isFetchingNextPage) {
			fetchNextPage()
		}
	}, [inView, hasNextPage, isFetchingNextPage, fetchNextPage])

	const handlePreviewItem = (item: WeedListItem) => {
		setItemForPreview(item)
		setIsPreviewOpen(true)
	}

	const handleClosePreview = () => {
		setIsPreviewOpen(false)
		setItemForPreview(null)
	}

	const handleItemClick = (item: WeedListItem) => {
		setSelectedItems((prev) => {
			const isSelected = prev.some((i) => i.id === item.id)
			if (isSelected) {
				return prev.filter((i) => i.id !== item.id)
			}
			return [...prev, item]
		})
	}

	const handleRemoveAllItems = () => {
		setSelectedItems([])
	}

	const handleConfirm = () => {
		onConfirm(selectedItems)
		onClose()
	}

	return createPortal(
		<>
			<dialog ref={dialogRef} className="modal">
				<div className="flex fixed inset-0 z-50 justify-center items-center">
					<div className="flex relative flex-col w-full h-full bg-base-100">
						<div className="flex sticky top-0 z-10 justify-between items-center p-2 border-b bg-base-100 border-base-200">
							<h2 className="text-xl font-bold">Выберите из каталога</h2>
							<div className="flex gap-3 items-center">
								<div className="indicator">
									<button
										type="button"
										onClick={() => setShowSelected(!showSelected)}
										className={`btn btn-ghost btn-circle btn-sm ${
											showSelected ? "btn-active" : ""
										}`}
										disabled={selectedItems.length === 0}
									>
										<FaList size={18} />
									</button>
									{selectedItems.length > 0 && (
										<span className="indicator-item indicator-bottom indicator-end badge badge-primary badge-xs">
											{selectedItems.length}
										</span>
									)}
								</div>
								<button type="button" onClick={onClose} className="btn btn-sm btn-ghost btn-circle">
									<IoClose size={24} />
								</button>
							</div>
						</div>

						<div className="p-2 border-b border-base-200">
							<div className="relative">
								<FaSearch className="absolute left-3 top-1/2 z-10 transform -translate-y-1/2 pointer-events-none text-base-content/50" />
								<input
									type="text"
									placeholder="Поиск по названию..."
									value={searchName}
									onChange={(e) => setSearchName(e.target.value)}
									className="pr-4 pl-10 w-full input input-bordered"
								/>
							</div>
						</div>

						<AnimatePresence>
							{showSelected && selectedItemsWithData.length > 0 && (
								<motion.div
									initial={{ height: 0, opacity: 0 }}
									animate={{ height: "auto", opacity: 1 }}
									exit={{ height: 0, opacity: 0 }}
									transition={{ duration: 0.2, ease: "easeInOut" }}
									className="overflow-hidden border-b border-base-200"
								>
									<div className="p-2 space-y-2 bg-base-100">
										<div className="flex justify-between items-center">
											<h3 className="font-semibold">Выбрано: {selectedItemsWithData.length}</h3>
											<button
												type="button"
												className="btn btn-ghost btn-xs"
												onClick={handleRemoveAllItems}
											>
												<FaTrash className="mr-1" />
												Очистить все
											</button>
										</div>
										<div className="py-1 pr-1 pl-2 space-y-2 max-h-36 overflow-y-auto">
											{selectedItemsWithData.map((item: WeedListItem) => (
												<div
													key={item.id}
													className="flex justify-between items-center p-2 rounded-lg bg-base-200"
												>
													<span className="text-sm">{item.name}</span>
													<button
														type="button"
														className="btn btn-ghost btn-xs btn-circle"
														onClick={() => handleItemClick(item)}
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

						<div className="flex-1 overflow-y-auto p-2">
							{status === "pending" ? (
								<div className="grid grid-cols-2 gap-4">
									{Array.from({ length: 6 }).map((_, index) => (
										<div
											key={CATALOG_GRID_SKELETON_KEYS[index]}
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
								<ErrorPage error={error} />
							) : (
								<>
									<div className="grid grid-cols-2 gap-4">
										{catalogItems.length === 0 ? (
											<div className="col-span-2 text-center text-base-content/60 py-4">
												Ничего не найдено.
											</div>
										) : (
											catalogItems.map((item) => (
												<CatalogSelectorCard
													key={item.id}
													item={item}
													isSelected={selectedItems.some((i) => i.id === item.id)}
													onSelect={() => handleItemClick(item)}
													onViewDetails={() => handlePreviewItem(item)}
												/>
											))
										)}
									</div>
									<div ref={ref}>
										{hasNextPage && (
											<div className="flex justify-center py-4">
												{isFetchingNextPage && <Loading />}
											</div>
										)}
									</div>
								</>
							)}
						</div>

						<div className="p-2 border-t border-base-200">
							<button type="button" className="btn btn-primary w-full" onClick={handleConfirm}>
								Подтвердить ({selectedItems.length})
							</button>
						</div>
					</div>
				</div>
			</dialog>
			<CatalogItemDetailSheet
				isOpen={isPreviewOpen}
				onClose={handleClosePreview}
				catalogItemId={itemForPreview?.id ?? null}
			/>
		</>,
		document.body
	)
}

export default CatalogSelectorSheet
