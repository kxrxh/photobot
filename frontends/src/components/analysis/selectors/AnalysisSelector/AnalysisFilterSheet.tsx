import type React from "react"
import { useId } from "react"
import FilterBottomSheet from "@/components/common/dialogs/FilterBottomSheet"
import { ModalSelect } from "@/components/common/ui/ModalSelect"

interface AnalysisFilters {
	sort_by: string
	sort_order: string
	id_analysis: string
	show_only_added: boolean
}

interface AnalysisFilterSheetProps {
	isOpen: boolean
	onClose: () => void
	filters: AnalysisFilters
	onFilterChange: (name: string, value: string | boolean) => void
	onApplyFilters: () => void
	onClearFilters: () => void
	hasAddedAnalyses: boolean
}

const sortByOptions = [
	{ value: "date_time", label: "Дате создания" },
	{ value: "id", label: "Номеру анализа" },
]

const sortOrderOptions = [
	{ value: "desc", label: "По убыванию" },
	{ value: "asc", label: "По возрастанию" },
]

const AnalysisFilterSheet: React.FC<AnalysisFilterSheetProps> = ({
	isOpen,
	onClose,
	filters,
	onFilterChange,
	onApplyFilters,
	onClearFilters,
	hasAddedAnalyses,
}) => {
	const sortById = useId()
	const sortOrderId = useId()
	const idAnalysisId = useId()
	const showOnlyAddedId = useId()

	const handleLocalFilterChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
		const { name, value, type } = e.target
		const checked = (e.target as HTMLInputElement).checked

		if (name === "show_only_added" && type === "checkbox" && checked && !hasAddedAnalyses) {
			alert("У вас нет добавленных анализов. Сначала добавьте анализы.")
			return
		}

		onFilterChange(name, type === "checkbox" ? checked : value)
	}

	return (
		<FilterBottomSheet
			isOpen={isOpen}
			onClose={onClose}
			title="Фильтры"
			onClear={onClearFilters}
			onApply={onApplyFilters}
		>
			<div className="space-y-4">
				<div className="form-control">
					<label htmlFor={sortById} className="label">
						<span className="font-medium label-text text-base-content">Сортировать по</span>
					</label>
					<ModalSelect
						id={sortById}
						title="Сортировать по"
						placeholder="Сортировать по"
						options={sortByOptions}
						value={filters.sort_by}
						onChange={(v) => onFilterChange("sort_by", v)}
						clearable={false}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={sortOrderId} className="label">
						<span className="font-medium label-text text-base-content">Порядок сортировки</span>
					</label>
					<ModalSelect
						id={sortOrderId}
						title="Порядок сортировки"
						placeholder="Порядок сортировки"
						options={sortOrderOptions}
						value={filters.sort_order}
						onChange={(v) => onFilterChange("sort_order", v)}
						clearable={false}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={idAnalysisId} className="label">
						<span className="font-medium label-text text-base-content">ID анализа</span>
					</label>
					<input
						id={idAnalysisId}
						type="text"
						name="id_analysis"
						value={filters.id_analysis}
						onChange={handleLocalFilterChange}
						placeholder="Введите ID анализа"
						className="w-full input input-bordered"
					/>
				</div>

				{hasAddedAnalyses && (
					<div className="form-control">
						<label className="flex items-center gap-3 p-4 border rounded-2xl cursor-pointer bg-primary/5 border-primary/20 hover:bg-primary/10 transition-colors">
							<input
								id={showOnlyAddedId}
								type="checkbox"
								name="show_only_added"
								checked={filters.show_only_added}
								onChange={handleLocalFilterChange}
								className="checkbox checkbox-primary checkbox-lg"
							/>
							<span className="font-medium text-base-content">Только добавленные анализы</span>
						</label>
					</div>
				)}
			</div>
		</FilterBottomSheet>
	)
}

export default AnalysisFilterSheet
