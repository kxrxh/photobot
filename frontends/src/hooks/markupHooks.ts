import { useQuery } from "@tanstack/react-query"
import { useCallback, useEffect, useRef, useState } from "react"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import type { WeedAnalysisObjects } from "@/api/catalog/types"
import { useAlert } from "@/hooks/useAlert"
import type { FractionType } from "./useFractions"

export const useModalManager = () => {
	const [activeModal, setActiveModal] = useState<ModalType | null>(null)
	const [modalData, setModalData] = useState<Record<string, unknown>>({})
	const clearModalDataTimeoutRef = useRef<number | null>(null)

	const openModal = useCallback((modalName: ModalType, data: Record<string, unknown> = {}) => {
		if (clearModalDataTimeoutRef.current) {
			clearTimeout(clearModalDataTimeoutRef.current)
			clearModalDataTimeoutRef.current = null
		}
		setActiveModal(modalName)
		setModalData(data)
	}, [])

	const closeModal = useCallback(() => {
		setActiveModal(null)
		if (clearModalDataTimeoutRef.current) {
			clearTimeout(clearModalDataTimeoutRef.current)
		}
		clearModalDataTimeoutRef.current = window.setTimeout(() => {
			clearModalDataTimeoutRef.current = null
			setModalData({})
		}, 300)
	}, [])

	useEffect(() => {
		return () => {
			if (clearModalDataTimeoutRef.current) {
				clearTimeout(clearModalDataTimeoutRef.current)
			}
		}
	}, [])

	return {
		activeModal,
		modalData,
		openModal,
		closeModal,
	}
}

export const useControlModeManager = (
	fractions: FractionType[],
	moveObject: (objectId: number, sourceFractionId: string, targetFractionId: string) => void,
	moveMultipleObjects: (
		objects: { objectId: number; sourceFractionId: string }[],
		targetFractionId: string
	) => void
) => {
	const [isActive, setIsActive] = useState(false)
	const [selectedObjects, setSelectedObjects] = useState<
		{ objectId: number; fractionId: string }[]
	>([])
	const [targetFractionId, setTargetFractionId] = useState("")
	const { showInfo, showSuccess } = useAlert()

	const enter = useCallback(() => setIsActive(true), [])
	const exit = useCallback(() => {
		setIsActive(false)
		setSelectedObjects([])
		setTargetFractionId("")
	}, [])

	const toggleObjectSelection = useCallback(
		(objectId: number, fractionId: string) => {
			if (!isActive) return
			setSelectedObjects((prev) =>
				prev.some((o) => o.objectId === objectId)
					? prev.filter((o) => o.objectId !== objectId)
					: [...prev, { objectId, fractionId }]
			)
		},
		[isActive]
	)

	const isObjectSelected = useCallback(
		(objectId: number) => selectedObjects.some((o) => o.objectId === objectId),
		[selectedObjects]
	)

	const toggleFractionSelection = useCallback(
		(fractionId: string) => {
			if (!isActive) return

			const fraction = fractions.find((f) => f.id === fractionId)
			if (!fraction) return

			const fractionObjects = fraction.objects.map((obj) => ({
				objectId: obj.id,
				fractionId,
			}))

			const allSelected = fractionObjects.every((obj) =>
				selectedObjects.some((selected) => selected.objectId === obj.objectId)
			)

			setSelectedObjects((prev) => {
				if (allSelected) {
					return prev.filter(
						(selected) => !fractionObjects.some((obj) => obj.objectId === selected.objectId)
					)
				} else {
					const newSelections = fractionObjects.filter(
						(obj) => !prev.some((selected) => selected.objectId === obj.objectId)
					)
					return [...prev, ...newSelections]
				}
			})
		},
		[isActive, fractions, selectedObjects]
	)

	const moveSelectedObjects = useCallback(() => {
		if (!targetFractionId || selectedObjects.length === 0) return

		const objectsToMove = selectedObjects.filter((o) => o.fractionId !== targetFractionId)

		if (objectsToMove.length === 0) {
			showInfo("Объекты уже находятся в выбранной фракции")
			return
		}

		if (objectsToMove.length === 1) {
			const { objectId, fractionId } = objectsToMove[0]
			moveObject(objectId, fractionId, targetFractionId)
		} else {
			moveMultipleObjects(
				objectsToMove.map(({ objectId, fractionId }) => ({
					objectId,
					sourceFractionId: fractionId,
				})),
				targetFractionId
			)
		}

		const targetFractionName = fractions.find((f) => f.id === targetFractionId)?.name
		showSuccess(`Перемещено ${objectsToMove.length} объектов в "${targetFractionName}"`)
		setSelectedObjects([])
	}, [
		selectedObjects,
		targetFractionId,
		fractions,
		moveObject,
		moveMultipleObjects,
		showInfo,
		showSuccess,
	])

	return {
		isActive,
		enter,
		exit,
		selectedObjects,
		targetFractionId,
		setTargetFractionId,
		toggleObjectSelection,
		toggleFractionSelection,
		isObjectSelected,
		moveSelectedObjects,
		hasSelectedObjects: selectedObjects.length > 0,
	}
}

export const useDataSourceManager = (
	addObjectsToFraction: (
		fractionId: string,
		objects: KalibriObject[],
		source: { type: "analysis" | "catalog"; sourceId: number | string }
	) => void,
	setFractions: React.Dispatch<React.SetStateAction<FractionType[]>>,
	targetFractionId: string
) => {
	const { showError } = useAlert()
	const [analysisToAdd, setAnalysisToAdd] = useState<Analysis | null>(null)
	const [catalogQueue, setCatalogQueue] = useState<
		{ name: string; analysisObjects: WeedAnalysisObjects }[]
	>([])
	const [currentCatalogItem, setCurrentCatalogItem] = useState<{
		name: string
		analysisObjects: WeedAnalysisObjects
	} | null>(null)

	const {
		data: analysisData,
		isLoading: isAnalysisLoading,
		isSuccess: isAnalysisSuccess,
		error: analysisError,
	} = useQuery({
		queryKey: queryKeys.analyses.objects(analysisToAdd?.id ?? ""),
		queryFn: () => {
			if (!analysisToAdd) return Promise.resolve([])
			return fetchAnalysisObjects(analysisToAdd.id)
		},
		enabled: !!analysisToAdd,
	})

	useEffect(() => {
		if (isAnalysisSuccess && analysisData && analysisToAdd) {
			addObjectsToFraction(targetFractionId, analysisData, {
				type: "analysis",
				sourceId: analysisToAdd.id,
			})
			setAnalysisToAdd(null)
		} else if (analysisError) {
			showError("Не удалось загрузить объекты анализа")
			setAnalysisToAdd(null)
		}
	}, [
		isAnalysisSuccess,
		analysisData,
		analysisToAdd,
		analysisError,
		addObjectsToFraction,
		targetFractionId,
		showError,
	])

	const {
		data: catalogData,
		isLoading: isCatalogLoading,
		isSuccess: isCatalogSuccess,
		error: catalogError,
	} = useQuery({
		queryKey: queryKeys.catalog.objects(String(currentCatalogItem?.analysisObjects.id ?? "")),
		queryFn: async () => {
			if (!currentCatalogItem) return []
			const { analyses_ids, excluded_objects } = currentCatalogItem.analysisObjects
			const allObjects = await Promise.all(analyses_ids.map(fetchAnalysisObjects))
			return allObjects.flat().filter((obj: KalibriObject) => !excluded_objects.includes(obj.id))
		},
		enabled: !!currentCatalogItem,
	})

	useEffect(() => {
		if (isCatalogSuccess && catalogData && currentCatalogItem) {
			const { name, analysisObjects } = currentCatalogItem
			const source = { type: "catalog" as const, sourceId: analysisObjects.id }

			if (targetFractionId !== "0") {
				addObjectsToFraction(targetFractionId, catalogData, source)
			} else {
				const newFractionId = crypto.randomUUID()
				const newFraction: FractionType = {
					id: newFractionId,
					name,
					objects: catalogData.map((obj) => ({
						...obj,
						source,
						file: getObjectImageUrl(obj) ?? obj.file ?? "",
						classificationState: "auto" as const,
					})),
				}
				setFractions((prev) => [...prev, newFraction])
			}
			setCurrentCatalogItem(null)
		} else if (catalogError) {
			showError("Не удалось загрузить объекты каталога")
			setCurrentCatalogItem(null)
		}
	}, [
		isCatalogSuccess,
		catalogData,
		currentCatalogItem,
		catalogError,
		addObjectsToFraction,
		setFractions,
		targetFractionId,
		showError,
	])

	useEffect(() => {
		if (catalogQueue.length > 0 && !currentCatalogItem && !isCatalogLoading) {
			const nextItem = catalogQueue[0]
			setCurrentCatalogItem(nextItem)
			setCatalogQueue((prev) => prev.slice(1))
		}
	}, [catalogQueue, currentCatalogItem, isCatalogLoading])

	return {
		isLoading: isAnalysisLoading || isCatalogLoading,
		addAnalysis: setAnalysisToAdd,
		addCatalogItemsToQueue: (items: { name: string; analysisObjects: WeedAnalysisObjects }[]) => {
			setCatalogQueue((prev) => [...prev, ...items])
		},
		isQueueProcessing: !!currentCatalogItem || catalogQueue.length > 0,
	}
}

export type ModalType =
	| "removeFraction"
	| "resetFractions"
	| "editFraction"
	| "analysisSelector"
	| "catalogSelector"
	| "loadModal"
	| "markupModal"
	| "markupSelector"
	| "classificationSave"
	| "markupSave"
	| "fractionStats"
	| "fractionObjects"
	| "fractionClassificationRule"
	| "reclassificationConfirm"

import { queryKeys } from "@/api/queryKeys"
import { getObjectImageUrl } from "@/utils/image"
