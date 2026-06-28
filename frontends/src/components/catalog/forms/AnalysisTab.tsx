import { useNavigate } from "@tanstack/react-router"
import { type FC, useCallback, useEffect, useId, useRef, useState } from "react"
import { FaPlus } from "react-icons/fa"
import { MdOutlineAnalytics } from "react-icons/md"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import AnalysisSelectorDialog from "@/components/analysis/selectors/AnalysisSelector"
import { useAlert } from "@/hooks/useAlert"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import AnalysisItem from "./AnalysisItem"

/** Matches `GeneralInfoTab` / `CharacteristicsTab` section titles */
const SECTION_TITLE_CLASS =
	"mb-1.5 block text-xs font-semibold uppercase tracking-wide text-base-content/45"

interface AnalysisTabProps {
	selectedAnalyses?: Analysis[]
	onAddAnalysis: (analysis: Analysis) => void
	onRemoveAnalysis: (analysisId: string) => void
	allObjectsData: Record<string, KalibriObject[]>
	excludedObjects: number[]
	onToggleExclude: (objectId: number) => void
	onObjectsDataChange: (data: Record<string, KalibriObject[]>) => void
	isParentLoading?: boolean
	readOnly?: boolean
}

const AnalysisTab: FC<AnalysisTabProps> = ({
	selectedAnalyses = [],
	onAddAnalysis,
	onRemoveAnalysis,
	allObjectsData,
	excludedObjects,
	onToggleExclude,
	onObjectsDataChange,
	isParentLoading = false,
	readOnly = false,
}) => {
	const sectionTitleId = useId()
	const navigate = useNavigate()
	const { showError } = useAlert()
	const [isFetchingAnalysisObjects, setIsFetchingAnalysisObjects] = useState(false)
	const [loadingObjectImages, setLoadingObjectImages] = useState<Record<string, boolean>>({})
	const [expandedStates, setExpandedStates] = useState<Record<string, boolean>>({})
	const [isSelectorDialogOpen, setIsSelectorDialogOpen] = useState(false)

	const hasAddedAnalyses = selectedAnalyses.length > 0

	const prevAnalysisIdsRef = useRef<Set<string>>(new Set())

	useEffect(() => {
		const currentAnalysisIds = new Set(selectedAnalyses.map((a) => a.id))
		const previousAnalysisIds = prevAnalysisIdsRef.current

		const removedAnalysisIds = Array.from(previousAnalysisIds).filter(
			(id) => !currentAnalysisIds.has(id)
		)

		const addedAnalysisIds = selectedAnalyses.filter(
			(analysis) => !previousAnalysisIds.has(analysis.id)
		)

		if (removedAnalysisIds.length > 0) {
			setLoadingObjectImages((prev) => {
				const newLoadingState = { ...prev }
				for (const id of removedAnalysisIds) {
					delete newLoadingState[id]
				}
				return newLoadingState
			})
			setExpandedStates((prev) => {
				const newExpandedStates = { ...prev }
				for (const id of removedAnalysisIds) {
					delete newExpandedStates[id]
				}
				return newExpandedStates
			})
		}

		const shouldFetchData =
			addedAnalysisIds.length > 0 || (selectedAnalyses.length > 0 && previousAnalysisIds.size === 0)

		if (shouldFetchData) {
			const fetchObjectData = async () => {
				const analysesToFetch = selectedAnalyses.filter((analysis) => !allObjectsData[analysis.id])

				if (analysesToFetch.length === 0) {
					setIsFetchingAnalysisObjects(false)
					return
				}

				setIsFetchingAnalysisObjects(true)
				const results = await Promise.allSettled(
					analysesToFetch.map((analysis) => fetchAnalysisObjects(analysis.id))
				)

				const newObjectData: Record<string, KalibriObject[]> = {}
				const failedIds: string[] = []
				let firstReason: unknown

				analysesToFetch.forEach((analysis, i) => {
					const settled = results[i]
					if (settled.status === "fulfilled") {
						newObjectData[analysis.id] = settled.value
					} else {
						newObjectData[analysis.id] = []
						failedIds.push(analysis.id)
						if (firstReason === undefined) {
							firstReason = settled.reason
						}
						log.devError(`Failed to fetch objects for analysis ${analysis.id}:`, settled.reason)
					}
				})

				onObjectsDataChange(newObjectData)

				if (failedIds.length > 0) {
					if (failedIds.length === analysesToFetch.length) {
						showError(getUserFacingErrorMessage(firstReason))
					} else {
						showError(`Не удалось загрузить объекты для анализов: ${failedIds.join(", ")}.`)
					}
				}

				setIsFetchingAnalysisObjects(false)
			}
			void fetchObjectData()
		} else if (selectedAnalyses.length === 0) {
			setIsFetchingAnalysisObjects(false)
		}

		prevAnalysisIdsRef.current = currentAnalysisIds
	}, [selectedAnalyses, allObjectsData, onObjectsDataChange, showError])

	const fetchAnalysisObjectsIfNeeded = useCallback(
		async (analysisId: string) => {
			if (allObjectsData[analysisId] || loadingObjectImages[analysisId]) {
				return
			}

			setLoadingObjectImages((prev) => ({ ...prev, [analysisId]: true }))
			try {
				const objects = await fetchAnalysisObjects(analysisId)

				onObjectsDataChange({ [analysisId]: objects })
			} catch (err) {
				log.devError(`Failed to fetch objects for analysis ${analysisId}:`, err)
				showError(getUserFacingErrorMessage(err))
				onObjectsDataChange({ [analysisId]: [] })
			} finally {
				setLoadingObjectImages((prev) => ({ ...prev, [analysisId]: false }))
			}
		},
		[allObjectsData, loadingObjectImages, onObjectsDataChange, showError]
	)

	const toggleExpand = useCallback(
		(analysisId: string) => {
			setExpandedStates((prev) => {
				const newState = { ...prev, [analysisId]: !prev[analysisId] }
				if (newState[analysisId]) {
					fetchAnalysisObjectsIfNeeded(analysisId)
				}
				return newState
			})
		},
		[fetchAnalysisObjectsIfNeeded]
	)

	const handleRemoveAllAnalyses = useCallback(() => {
		for (const analysis of selectedAnalyses) {
			onRemoveAnalysis(analysis.id)
		}
		setLoadingObjectImages({})
		setExpandedStates({})
	}, [selectedAnalyses, onRemoveAnalysis])

	const isLoading = isFetchingAnalysisObjects || isParentLoading
	const analysisCount = selectedAnalyses.length

	return (
		<div className="animate-fadeIn space-y-6 px-4 pb-8">
			<section aria-labelledby={sectionTitleId} className="flex flex-col gap-4">
				<div className="pt-1">
					<h2 id={sectionTitleId} className={SECTION_TITLE_CLASS}>
						Анализы
					</h2>
					<p className="text-xs leading-relaxed text-base-content/55">
						Привяжите анализы к записи: по их объектам считаются характеристики на соседней вкладке.
						Можно исключать отдельные объекты из расчёта.
					</p>
				</div>

				{isLoading ? (
					<div className="flex flex-col items-center gap-3 rounded-2xl border border-base-200 bg-base-200/25 px-4 py-10">
						<span className="loading loading-spinner loading-lg text-primary" />
						<p className="text-center text-sm text-base-content/65">
							Загрузка анализов и объектов…
						</p>
					</div>
				) : analysisCount > 0 ? (
					<>
						{!readOnly ? (
							<div className="space-y-3 rounded-2xl border border-base-200 bg-base-200/25 p-4">
								<div className="flex flex-wrap items-center gap-2">
									<span className="badge badge-sm border border-base-300/60 bg-base-100/80 font-medium text-base-content/85">
										Анализов в списке: {analysisCount}
									</span>
								</div>
								<p className="text-xs leading-relaxed text-base-content/55">
									Нажмите на карточку, чтобы развернуть объекты и отметить исключения.
								</p>
								<button
									type="button"
									onClick={() => setIsSelectorDialogOpen(true)}
									className="btn btn-primary btn-md w-full gap-2 rounded-xl shadow-sm"
								>
									<FaPlus className="h-4 w-4 shrink-0 opacity-90" aria-hidden />
									Добавить анализ
								</button>
							</div>
						) : (
							<div className="rounded-2xl border border-base-200 bg-base-200/20 px-4 py-3">
								<p className="text-xs text-base-content/65">
									Привязано анализов: {analysisCount}. Редактирование недоступно в режиме просмотра
									заявки.
								</p>
							</div>
						)}

						<div className="space-y-3">
							{selectedAnalyses.map((analysis) => {
								const analysisId = analysis.id
								return (
									<AnalysisItem
										key={analysisId}
										analysis={analysis}
										isExpanded={expandedStates[analysisId]}
										onToggleExpand={toggleExpand}
										onRemove={readOnly ? undefined : onRemoveAnalysis}
										excludedObjects={excludedObjects}
										onToggleExclude={readOnly ? undefined : onToggleExclude}
									/>
								)
							})}
						</div>
					</>
				) : readOnly ? (
					<div className="rounded-2xl border border-base-200 bg-base-200/20 px-4 py-5 text-center">
						<p className="text-sm leading-relaxed text-base-content/70">
							В этой заявке пока нет привязанных анализов. После добавления анализов модератором они
							появятся здесь.
						</p>
					</div>
				) : (
					<div className="overflow-hidden rounded-2xl border border-primary/20 bg-linear-to-br from-primary/[0.07] via-base-200/30 to-base-200/50 shadow-sm">
						<div className="flex flex-col items-center gap-4 px-4 py-8 text-center sm:px-6">
							<div
								className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/15 text-primary shadow-inner"
								aria-hidden
							>
								<MdOutlineAnalytics className="h-8 w-8" />
							</div>
							<div className="max-w-sm space-y-2">
								<h3 className="text-base font-semibold text-base-content">Анализы не выбраны</h3>
								<p className="text-sm leading-relaxed text-base-content/65">
									Добавьте один или несколько анализов — они будут связаны с этой записью каталога,
									а объекты появятся в характеристиках.
								</p>
							</div>
							<button
								type="button"
								onClick={() => setIsSelectorDialogOpen(true)}
								className="btn btn-primary btn-md mt-1 w-full max-w-xs gap-2 rounded-xl shadow-sm"
							>
								<FaPlus className="h-4 w-4 shrink-0 opacity-90" aria-hidden />
								Выбрать анализы
							</button>
						</div>
					</div>
				)}
			</section>

			{!readOnly && (
				<AnalysisSelectorDialog
					isOpen={isSelectorDialogOpen}
					onClose={() => setIsSelectorDialogOpen(false)}
					selectedAnalysisIds={selectedAnalyses.map((analysis) => analysis.id)}
					onAddAnalysis={onAddAnalysis}
					onRemoveAnalysis={onRemoveAnalysis}
					onRemoveAllAnalyses={handleRemoveAllAnalyses}
					hasAddedAnalyses={hasAddedAnalyses}
					onOpenCreateDialog={() => {
						setIsSelectorDialogOpen(false)
						void navigate({ to: "/analysis/create", search: { openRequest: undefined } })
					}}
				/>
			)}
		</div>
	)
}

export default AnalysisTab
