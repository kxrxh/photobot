import { useQuery, useQueryClient } from "@tanstack/react-query"
import type React from "react"
import { useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { FaCheck, FaEye, FaSearch } from "react-icons/fa"
import { listMarkups } from "@/api/markup"
import type { Markup, MarkupFilters } from "@/api/markup/types"
import { queryKeys } from "@/api/queryKeys"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import MarkupViewSheet from "./MarkupViewSheet"

const MARKUP_LIST_SKELETON_KEYS = Array.from({ length: 5 }, () => crypto.randomUUID())

interface MarkupSelectorSheetProps {
	isOpen: boolean
	onClose: () => void
	onSelectMarkup: (markup: Markup) => void
	selectedMarkupId?: number
}

const formatDate = (date: string) => {
	return new Intl.DateTimeFormat("ru-RU", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	}).format(new Date(date))
}

const MarkupItem: React.FC<{
	markup: Markup
	isSelected: boolean
	onView: (markup: Markup) => void
	onSelect: (markup: Markup) => void
}> = ({ markup, isSelected, onView, onSelect }) => {
	return (
		<div className={`card card-border p-4 ${isSelected ? "ring-2 ring-primary" : ""}`}>
			<div className="flex flex-col gap-3 p-0 card-body">
				<div className="flex gap-4 justify-between items-start">
					<div className="flex overflow-hidden grow gap-2 items-center">
						<h3 className="text-base font-semibold truncate card-title">{markup.name}</h3>
					</div>
					<button
						type="button"
						onClick={(e) => {
							e.stopPropagation()
							onView(markup)
						}}
						className="btn btn-ghost btn-sm btn-square text-primary"
						title="Просмотр разметки"
					>
						<FaEye size={18} />
					</button>
				</div>
				<div className="flex flex-wrap gap-y-1 gap-x-4 items-center text-sm text-base-content/70">
					<div className="flex gap-1 items-center">
						Обновлено:
						<span className="badge badge-outline badge-sm">{formatDate(markup.updated_at)}</span>
					</div>
				</div>
				<div className="justify-end card-actions">
					<button
						type="button"
						onClick={(e) => {
							e.stopPropagation()
							onSelect(markup)
						}}
						className="w-full btn btn-primary btn-sm sm:w-auto"
						title="Выбрать разметку"
					>
						<FaCheck className="mr-1 text-lg" />
						Выбрать
					</button>
				</div>
			</div>
		</div>
	)
}

const MarkupSelectorSheet: React.FC<MarkupSelectorSheetProps> = ({
	isOpen,
	onClose,
	onSelectMarkup,
	selectedMarkupId,
}) => {
	const [filters, setFilters] = useState<MarkupFilters>({
		name: "",
	})

	const [debouncedFilterName, setDebouncedFilterName] = useState(filters.name)

	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedFilterName(filters.name)
		}, 500)

		return () => {
			clearTimeout(handler)
		}
	}, [filters.name])

	const {
		data: markups = [],
		isLoading: isMarkupsLoading,
		isFetching: isMarkupsFetching,
	} = useQuery({
		queryKey: queryKeys.markups({ name: debouncedFilterName }),
		queryFn: () => listMarkups(filters),
		staleTime: 5 * 60 * 1000,
		gcTime: 5 * 60 * 1000,
		enabled: isOpen,
	})

	const loading = isMarkupsLoading || isMarkupsFetching
	const [selectedMarkupForView, setSelectedMarkupForView] = useState<Markup | null>(null)
	const [isViewDialogOpen, setIsViewDialogOpen] = useState(false)
	const dialogRef = useRef<HTMLDialogElement>(null)
	const markupsListRef = useRef<HTMLDivElement>(null)

	const queryClient = useQueryClient()

	useEffect(() => {
		if (dialogRef.current) {
			if (isOpen) {
				dialogRef.current.showModal()
				if (markupsListRef.current) {
					markupsListRef.current.scrollTop = 0
				}
				queryClient.invalidateQueries({
					queryKey: queryKeys.markups(),
				})
			} else {
				dialogRef.current.close()
			}
		}
	}, [isOpen, queryClient])

	// Ensure state resets if dialog is closed by ESC or backdrop
	useEffect(() => {
		const dialog = dialogRef.current
		if (!dialog) return

		const handleDialogClose = () => {
			onClose()
		}
		dialog.addEventListener("close", handleDialogClose)

		return () => {
			dialog.removeEventListener("close", handleDialogClose)
		}
	}, [onClose])

	const handleFilterChange = (event: React.ChangeEvent<HTMLInputElement>) => {
		const { value } = event.target
		setFilters({
			name: value,
		})
	}

	const handleSelectMarkup = (markup: Markup) => {
		onSelectMarkup(markup)
		onClose()
	}

	const handleViewMarkup = (markup: Markup) => {
		setSelectedMarkupForView(markup)
		setIsViewDialogOpen(true)
	}

	return createPortal(
		<dialog ref={dialogRef} className="modal">
			<div className="flex fixed inset-0 z-50 justify-center items-center">
				<div className="flex relative flex-col w-full h-full bg-base-100">
					<div className="flex sticky top-0 z-10 justify-between items-center p-2 border-b border-base-200 bg-base-100">
						<h2 className="text-xl font-bold">Выберите разметку</h2>
						<div className="flex gap-4 items-center">
							<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
						</div>
					</div>

					<div className="p-2 border-b border-base-200">
						<label className="input input-bordered flex items-center gap-2 w-full">
							<input
								type="text"
								className="grow input-sm"
								placeholder="Поиск по названию"
								value={filters.name}
								onChange={handleFilterChange}
								title="Поиск разметок по названию"
							/>
							<FaSearch size={16} className="opacity-70" />
						</label>
					</div>

					<MarkupViewSheet
						isOpen={isViewDialogOpen}
						onClose={() => setIsViewDialogOpen(false)}
						markup={selectedMarkupForView}
					/>

					<div className="overflow-y-auto flex-1 p-2 space-y-2" ref={markupsListRef}>
						{loading ? (
							Array.from({ length: 5 }).map((_, index) => (
								<div key={MARKUP_LIST_SKELETON_KEYS[index]} className="p-4 card card-border">
									<div className="flex flex-col gap-4">
										<div className="w-2/3 h-5 skeleton" />
										<div className="w-1/2 h-4 skeleton" />
										<div className="self-end w-full h-8 skeleton sm:w-32" />
									</div>
								</div>
							))
						) : markups?.length === 0 ? (
							<div className="py-12 text-center text-base-content/70">
								<p>Список разметок пуст</p>
								<p className="text-sm">
									Добавьте первую разметку, используя соответствующую кнопку
								</p>
							</div>
						) : (
							markups?.map((markup) => (
								<MarkupItem
									key={markup.id}
									markup={markup}
									isSelected={String(selectedMarkupId) === String(markup.id)}
									onView={handleViewMarkup}
									onSelect={handleSelectMarkup}
								/>
							))
						)}
					</div>
				</div>
			</div>
		</dialog>,
		document.body
	)
}

export default MarkupSelectorSheet
