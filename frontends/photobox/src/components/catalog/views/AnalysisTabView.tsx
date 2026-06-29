import { useCallback, useEffect, useRef, useState } from "react"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import AnalysisItem from "@/components/catalog/forms/AnalysisItem"
import { useAlert } from "@/hooks/useAlert"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"

interface AnalysisTabViewProps {
	selectedAnalyses?: Analysis[]
	allObjectsData: Record<string, KalibriObject[]>
	onObjectsDataChange: (data: Record<string, KalibriObject[]>) => void
	excludedObjects?: number[]
}

export default function AnalysisTabView({
	selectedAnalyses = [],
	allObjectsData,
	onObjectsDataChange,
	excludedObjects = [],
}: AnalysisTabViewProps) {
	const { showError } = useAlert()
	const [loadingAnalyses, setLoadingAnalyses] = useState(false)
	const [loadingObjectImages, setLoadingObjectImages] = useState<Record<string, boolean>>({})
	const [expandedStates, setExpandedStates] = useState<Record<string, boolean>>({})

	const prevAnalysisIdsRef = useRef<Set<string>>(new Set())

	useEffect(() => {
		const currentAnalysisIds = new Set(selectedAnalyses.map((a) => a.id))
		const previousAnalysisIds = prevAnalysisIdsRef.current

		const removedAnalysisIds = Array.from(previousAnalysisIds).filter(
			(id) => !currentAnalysisIds.has(id)
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

		const shouldFetchData = selectedAnalyses.length > 0 && previousAnalysisIds.size === 0

		if (shouldFetchData) {
			const fetchObjectData = async () => {
				const analysesToFetch = selectedAnalyses.filter((analysis) => !allObjectsData[analysis.id])

				if (analysesToFetch.length === 0) {
					setLoadingAnalyses(false)
					return
				}

				setLoadingAnalyses(true)
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

				setLoadingAnalyses(false)
			}
			void fetchObjectData()
		} else if (selectedAnalyses.length === 0) {
			setLoadingAnalyses(false)
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

	return (
		<div className="space-y-2 px-4 pb-8 animate-fadeIn">
			<div className="pt-6 pb-4">
				<h1 className="text-2xl font-bold text-center">Связанные анализы</h1>
				{excludedObjects.length > 0 && (
					<p className="mt-2 text-sm text-center opacity-70">
						Исключенных объектов: {excludedObjects.length}
					</p>
				)}
			</div>

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
									excludedObjects={excludedObjects}
								/>
							)
						})}
					</div>
				) : (
					!loadingAnalyses && (
						<div className="p-4 text-center text-base-content/60">Нет привязанных анализов.</div>
					)
				)}
			</div>
		</div>
	)
}
