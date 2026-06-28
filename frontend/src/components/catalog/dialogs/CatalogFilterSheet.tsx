import { useMutation } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import type React from "react"
import { useEffect, useId, useState } from "react"
import { fetchAnalysisObjects } from "@/api/analysis"
import AnalysisSelectorSheet from "@/components/analysis/selectors/AnalysisSelector"
import FilterBottomSheet from "@/components/common/dialogs/FilterBottomSheet"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { type CatalogFilters, defaultFilters } from "@/hooks/useCatalogFilters"
import { useClassifications } from "@/hooks/useClassifications"
import { buildSmartRangesFromObjects } from "@/utils/stats"

interface CatalogFilterSheetProps {
	isOpen: boolean
	onClose: () => void
	filters: CatalogFilters
	onFilterChange: (name: string, value: string | boolean | number | undefined) => void
	onApplyFilters: () => void
	onClearFilters: () => void
}

const SORT_ORDER_OPTIONS = [
	{ value: "asc", label: "По возрастанию" },
	{ value: "desc", label: "По убыванию" },
]

const CatalogFilterSheet: React.FC<CatalogFilterSheetProps> = ({
	isOpen,
	onClose,
	filters,
	onFilterChange,
	onApplyFilters,
	onClearFilters,
}) => {
	const navigate = useNavigate()
	const [localFilters, setLocalFilters] = useState<CatalogFilters>(filters)
	const [isAnalysisDialogOpen, setIsAnalysisDialogOpen] = useState(false)
	const [selectedAnalysisId, setSelectedAnalysisId] = useState<string | null>(null)
	const [shouldPrefillOnClose, setShouldPrefillOnClose] = useState(false)

	const sortOrderId = useId()
	const mainGroupId = useId()
	const mainSubgroupId = useId()
	const subgroupId = useId()
	const lMinId = useId()
	const lMaxId = useId()

	const prefillFromAnalysis = useMutation({
		mutationFn: (analysisId: string) => fetchAnalysisObjects(analysisId),
		onSuccess: (objects) => {
			const bands = buildSmartRangesFromObjects(objects, 2)
			setLocalFilters((prev) => ({ ...prev, ...bands }))
		},
	})

	const isPrefilling = prefillFromAnalysis.isPending

	useEffect(() => {
		if (isOpen) {
			setLocalFilters(filters)
		}
	}, [isOpen, filters])

	useEffect(() => {
		if (!isAnalysisDialogOpen && shouldPrefillOnClose && selectedAnalysisId) {
			void prefillFromAnalysis.mutateAsync(selectedAnalysisId)
			setShouldPrefillOnClose(false)
		}
	}, [isAnalysisDialogOpen, shouldPrefillOnClose, selectedAnalysisId, prefillFromAnalysis])

	const { data: classifications } = useClassifications({ enabled: isOpen })

	const handleNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const { name, value } = e.target
		setLocalFilters((prev) => ({
			...prev,
			[name]: value === "" ? undefined : Number(value),
		}))
	}

	const handleApplyFilters = () => {
		Object.entries(localFilters).forEach(([key, value]) => {
			onFilterChange(key, value)
		})
		onApplyFilters()
		onClose()
	}

	const handleClearFilters = () => {
		setLocalFilters(defaultFilters)
		onClearFilters()
		onClose()
	}

	const formContent = (
		<>
			<div className="space-y-4">
				<div className="form-control">
					<label htmlFor={sortOrderId} className="label">
						<span className="font-medium label-text text-base-content">Порядок сортировки</span>
					</label>
					<ModalSelect
						id={sortOrderId}
						title="Порядок сортировки"
						placeholder="Порядок сортировки"
						options={SORT_ORDER_OPTIONS}
						value={localFilters.sort_order}
						onChange={(v) =>
							setLocalFilters((prev) => ({
								...prev,
								sort_order: v === "asc" || v === "desc" ? v : prev.sort_order,
							}))
						}
						clearable={false}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={mainGroupId} className="label">
						<span className="font-medium label-text text-base-content">Основная классификация</span>
					</label>
					<ModalSelect
						id={mainGroupId}
						title="Основная классификация"
						placeholder="Все"
						options={Object.entries(classifications?.main_groups || {}).map(([id, name]) => ({
							value: id,
							label: name,
						}))}
						value={localFilters.main_group ?? ""}
						onChange={(v) =>
							setLocalFilters((prev) => ({ ...prev, main_group: v === "" ? undefined : v }))
						}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={mainSubgroupId} className="label">
						<span className="font-medium label-text text-base-content">
							Вторичная классификация
						</span>
					</label>
					<ModalSelect
						id={mainSubgroupId}
						title="Вторичная классификация"
						placeholder="Все"
						options={Object.entries(classifications?.main_subgroups || {}).map(([id, name]) => ({
							value: id,
							label: name,
						}))}
						value={localFilters.main_subgroup ?? ""}
						onChange={(v) =>
							setLocalFilters((prev) => ({ ...prev, main_subgroup: v === "" ? undefined : v }))
						}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={subgroupId} className="label">
						<span className="font-medium label-text text-base-content">Подгруппа</span>
					</label>
					<ModalSelect
						id={subgroupId}
						title="Подгруппа"
						placeholder="Все"
						options={Object.entries(classifications?.subgroups || {}).map(([id, name]) => ({
							value: id,
							label: name,
						}))}
						value={localFilters.subgroup ?? ""}
						onChange={(v) =>
							setLocalFilters((prev) => ({ ...prev, subgroup: v === "" ? undefined : v }))
						}
					/>
				</div>
			</div>

			<div className="relative space-y-4" aria-busy={isPrefilling} aria-live="polite">
				<div className="flex items-center justify-between">
					<h4 className="text-lg font-semibold text-base-content">Статистики</h4>
					<button
						type="button"
						className="btn btn-sm btn-primary btn-soft"
						onClick={() => setIsAnalysisDialogOpen(true)}
					>
						Заполнить из анализа
					</button>
				</div>

				<div className="grid grid-cols-2 gap-3">
					<div className="relative">
						<label className="label" htmlFor={lMinId}>
							<span className="label-text font-medium text-base-content">Длина (L)</span>
						</label>
						<div className="flex gap-2">
							<input
								id={lMinId}
								name="l_min"
								type="number"
								value={localFilters.l_min ?? ""}
								onChange={handleNumberChange}
								className="input input-bordered w-full"
								placeholder="мин"
							/>
							<input
								id={lMaxId}
								name="l_max"
								type="number"
								value={localFilters.l_max ?? ""}
								onChange={handleNumberChange}
								className="input input-bordered w-full"
								placeholder="макс"
							/>
						</div>
						{isPrefilling && (
							<div className="absolute inset-0 z-30 flex items-center justify-center rounded-lg bg-base-100/70 backdrop-blur-sm">
								<span className="loading loading-spinner loading-lg" />
								<span className="sr-only">Загрузка</span>
							</div>
						)}
					</div>

					{(
						[
							["w_min", "w_max", "Ширина (W)"],
							["lw_min", "lw_max", "L/W"],
							["h_min", "h_max", "H"],
							["s_min", "s_max", "S"],
							["v_min", "v_max", "V"],
							["r_min", "r_max", "R"],
							["g_min", "g_max", "G"],
							["b_min", "b_max", "B"],
							["brt_min", "brt_max", "Brt"],
							["sq_sqcrl_min", "sq_sqcrl_max", "Sq/SqCrl"],
						] as const
					).map(([minKey, maxKey, label]) => {
						const minVal = localFilters[minKey]
						const maxVal = localFilters[maxKey]
						return (
							<div key={minKey}>
								<label className="label" htmlFor={minKey}>
									<span className="label-text font-medium text-base-content">{label}</span>
								</label>
								<div className="flex gap-2">
									<input
										id={minKey}
										name={minKey}
										type="number"
										value={typeof minVal === "number" ? minVal : ""}
										onChange={handleNumberChange}
										className="input input-bordered w-full"
										placeholder="мин"
									/>
									<input
										id={maxKey}
										name={maxKey}
										type="number"
										value={typeof maxVal === "number" ? maxVal : ""}
										onChange={handleNumberChange}
										className="input input-bordered w-full"
										placeholder="макс"
									/>
								</div>
							</div>
						)
					})}
				</div>
			</div>
		</>
	)

	return (
		<>
			<FilterBottomSheet
				isOpen={isOpen}
				onClose={onClose}
				title="Фильтры каталога"
				onClear={handleClearFilters}
				onApply={handleApplyFilters}
			>
				{formContent}
			</FilterBottomSheet>
			<AnalysisSelectorSheet
				isOpen={isAnalysisDialogOpen}
				onClose={() => setIsAnalysisDialogOpen(false)}
				selectedAnalysisIds={selectedAnalysisId ? [selectedAnalysisId] : []}
				onAddAnalysis={(analysis) => {
					setSelectedAnalysisId(analysis.id)
					setShouldPrefillOnClose(true)
				}}
				onRemoveAnalysis={() => setSelectedAnalysisId(null)}
				onRemoveAllAnalyses={() => setSelectedAnalysisId(null)}
				hasAddedAnalyses={false}
				selectionMode="single"
				onOpenCreateDialog={() => {
					setIsAnalysisDialogOpen(false)
					void navigate({ to: "/analysis/create", search: { openRequest: undefined } })
				}}
			/>
		</>
	)
}

export default CatalogFilterSheet
