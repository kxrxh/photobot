import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useCallback, useEffect, useState } from "react"
import { FaCodeBranch, FaPlus } from "react-icons/fa"
import { RiResetLeftLine } from "react-icons/ri"
import { fetchAnalysisObjects, mergeAnalyses } from "@/api/analysis"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import AnalysisSelectorDialog from "@/components/analysis/selectors/AnalysisSelector"
import AnalysisItem from "@/components/catalog/forms/AnalysisItem"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"

export const Route = createFileRoute("/_authenticated/merge")({
	component: RouteComponent,
})

function RouteComponent() {
	const navigate = useNavigate()
	const { userId } = useAuth()
	const { showSuccess, showError } = useAlert()
	const [selectedAnalyses, setSelectedAnalyses] = useState<Analysis[]>([])
	const [allObjectsData, setAllObjectsData] = useState<Record<string, KalibriObject[]>>({})
	const [loadingAnalyses, setLoadingAnalyses] = useState(false)
	const [loadingObjectImages, setLoadingObjectImages] = useState<Record<string, boolean>>({})
	const [excludedObjects, setExcludedObjects] = useState<number[]>([])
	const [expandedStates, setExpandedStates] = useState<Record<string, boolean>>({})
	const [isSelectorDialogOpen, setIsSelectorDialogOpen] = useState(false)
	const [isMerging, setIsMerging] = useState(false)

	const hasAddedAnalyses = selectedAnalyses.length > 0
	const canMerge = selectedAnalyses.length >= 2

	useEffect(() => {
		const fetchObjectData = async () => {
			if (selectedAnalyses.length === 0) {
				return
			}

			setLoadingAnalyses(true)
			const results = await Promise.allSettled(
				selectedAnalyses.map((analysis) => fetchAnalysisObjects(analysis.id))
			)

			const newObjectData: Record<string, KalibriObject[]> = {}
			const failedIds: string[] = []
			let firstReason: unknown

			selectedAnalyses.forEach((analysis, i) => {
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

			setAllObjectsData(newObjectData)

			if (failedIds.length > 0) {
				if (failedIds.length === selectedAnalyses.length) {
					showError(getUserFacingErrorMessage(firstReason))
				} else {
					showError(`Не удалось загрузить объекты для анализов: ${failedIds.join(", ")}.`)
				}
			}

			setLoadingAnalyses(false)
		}

		void fetchObjectData()
	}, [selectedAnalyses, showError])

	const fetchAnalysisObjectsIfNeeded = useCallback(
		async (analysisId: string) => {
			if (allObjectsData[analysisId] || loadingObjectImages[analysisId]) {
				return
			}

			setLoadingObjectImages((prev) => ({ ...prev, [analysisId]: true }))
			try {
				const objects = await fetchAnalysisObjects(analysisId)
				setAllObjectsData((prev) => ({ ...prev, [analysisId]: objects }))
			} catch (error) {
				log.devError(`Failed to fetch objects for analysis ${analysisId}:`, error)
				setAllObjectsData((prev) => ({ ...prev, [analysisId]: [] }))
			} finally {
				setLoadingObjectImages((prev) => ({ ...prev, [analysisId]: false }))
			}
		},
		[allObjectsData, loadingObjectImages]
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

	const handleToggleExclude = useCallback((objectId: number) => {
		setExcludedObjects((prev) =>
			prev.includes(objectId) ? prev.filter((id) => id !== objectId) : [...prev, objectId]
		)
	}, [])

	const handleAddAnalysis = useCallback((analysis: Analysis) => {
		setSelectedAnalyses((prev) => {
			if (prev.some((a) => a.id === analysis.id)) {
				return prev
			}
			return [...prev, analysis]
		})
	}, [])

	const handleRemoveAnalysis = useCallback(
		(analysisId: string) => {
			const objectsInAnalysis = allObjectsData[analysisId] || []
			const objectIdsInAnalysis = new Set(objectsInAnalysis.map((o) => o.id))
			setSelectedAnalyses((prev) => prev.filter((analysis) => analysis.id !== analysisId))
			setAllObjectsData((prev) => {
				const newObjectsData = { ...prev }
				delete newObjectsData[analysisId]
				return newObjectsData
			})
			setLoadingObjectImages((prev) => {
				const newLoadingStates = { ...prev }
				delete newLoadingStates[analysisId]
				return newLoadingStates
			})
			setExcludedObjects((prev) => prev.filter((id) => !objectIdsInAnalysis.has(id)))
			setExpandedStates((prev) => {
				const newExpandedStates = { ...prev }
				delete newExpandedStates[analysisId]
				return newExpandedStates
			})
		},
		[allObjectsData]
	)

	const handleRemoveAllAnalyses = useCallback(() => {
		setSelectedAnalyses([])
		setAllObjectsData({})
		setLoadingObjectImages({})
		setExcludedObjects([])
		setExpandedStates({})
	}, [])

	const handleMergeAnalyses = useCallback(async () => {
		if (!canMerge) return
		if (!userId) {
			showError("Требуется авторизация")
			return
		}

		try {
			setIsMerging(true)
			const analysisIds = selectedAnalyses.map((a) => a.id)
			const result = await mergeAnalyses(analysisIds)
			showSuccess(result.message ?? "Успешно объединено")
		} catch (err) {
			showError(getUserFacingErrorMessage(err))
		} finally {
			setIsMerging(false)
		}
	}, [canMerge, selectedAnalyses, showError, showSuccess, userId])

	return (
		<div className="flex flex-col">
			<header className="sticky top-0 z-50 w-full border-b bg-base-100 border-base-200">
				<div className="p-2 w-full">
					<div className="flex justify-between">
						<button
							type="button"
							onClick={handleMergeAnalyses}
							className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm flex-1"
							disabled={!canMerge || isMerging}
						>
							{isMerging ? (
								<span className="loading loading-spinner loading-sm" />
							) : (
								<>
									<FaCodeBranch className="mb-1 w-4 h-4" />
									<span className="text-xs">Объединить</span>
								</>
							)}
						</button>
						<button
							type="button"
							onClick={() => setIsSelectorDialogOpen(true)}
							className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm flex-1"
						>
							<FaPlus className="mb-1 w-4 h-4" />
							<span className="text-xs">Добавить анализ</span>
						</button>

						<button
							type="button"
							onClick={handleRemoveAllAnalyses}
							className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm flex-1"
							disabled={!hasAddedAnalyses}
						>
							<RiResetLeftLine className="mb-1 w-4 h-4" />
							<span className="text-xs">Сбросить</span>
						</button>
					</div>
				</div>
			</header>

			<main className="container px-2 py-2 mx-auto">
				<div className="space-y-4">
					{loadingAnalyses && (
						<div className="py-8 text-center">
							<p className="text-sm text-gray-500">Загрузка анализов...</p>
						</div>
					)}

					{!loadingAnalyses && selectedAnalyses.length > 0 ? (
						<div className="space-y-3">
							{selectedAnalyses.map((analysis) => {
								const analysisId = analysis.id

								return (
									<AnalysisItem
										key={analysisId}
										analysis={analysis}
										isExpanded={!!expandedStates[analysisId]}
										onToggleExpand={toggleExpand}
										onRemove={handleRemoveAnalysis}
										excludedObjects={excludedObjects}
										onToggleExclude={handleToggleExclude}
									/>
								)
							})}
						</div>
					) : (
						!loadingAnalyses && (
							<div className="py-8 text-center">
								<p className="mb-2 font-bold text-base-content">Анализы не выбраны</p>
								<p className="text-sm text-base-content">Добавьте анализы для объединения</p>
							</div>
						)
					)}
				</div>
			</main>

			<AnalysisSelectorDialog
				isOpen={isSelectorDialogOpen}
				onClose={() => setIsSelectorDialogOpen(false)}
				selectedAnalysisIds={selectedAnalyses.map((analysis) => analysis.id)}
				onAddAnalysis={handleAddAnalysis}
				onRemoveAnalysis={handleRemoveAnalysis}
				onRemoveAllAnalyses={handleRemoveAllAnalyses}
				hasAddedAnalyses={hasAddedAnalyses}
				onOpenCreateDialog={() => {
					setIsSelectorDialogOpen(false)
					void navigate({ to: "/analysis/create", search: { openRequest: undefined } })
				}}
			/>
		</div>
	)
}
