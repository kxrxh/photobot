import { useCallback, useEffect, useState } from "react"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis } from "@/api/analysis/types"
import { fetchWeedAnalysisObjects } from "@/api/catalog"
import type { WeedListItem } from "@/api/catalog/types"
import { createMarkup, getMarkup, updateMarkup } from "@/api/markup"
import type { Markup, SaveMarkup } from "@/api/markup/types"
import Loading from "@/components/common/ui/Loading"
import AppModals from "@/components/markup/components/AppModals"
import FractionControlMode from "@/components/markup/components/FractionControlMode"
import { FractionList } from "@/components/markup/components/MarkupComponents"
import { MarkupHeader } from "@/components/markup/headers/MarkupHeaders"
import { useControlModeManager, useDataSourceManager, useModalManager } from "@/hooks/markupHooks"
import { useAlert } from "@/hooks/useAlert"
import { type FractionType, type KalibriObjectWithSource, useFractions } from "@/hooks/useFractions"
import type { SelectedAnalysis } from "@/routes/_authenticated/markup.tsx"
import { getObjectImageUrl } from "@/utils/image"
import { log } from "@/utils/log"
import { createObjectLookup, DEFAULT_FRACTION_ID } from "@/utils/markupUtils"
import { truncate } from "@/utils/text"

interface FractionModalData {
	id: string
	name: string
}

export default function MarkupPage() {
	const {
		fractions,
		setFractions,
		addNewFraction,
		removeFraction,
		clearFractions,
		editFractionName,
		moveObjectToFraction,
		moveMultipleObjectsToFraction,
		addObjectsToFraction,
		removeObjectsByAnalysisId,
		removeObjectsByCatalogItemId,
		setClassificationRules,
		handleReclassificationChoice,
		pendingReclassification,
	} = useFractions()
	const { showInfo, showSuccess, showError } = useAlert()
	const modalManager = useModalManager()

	const [selectedAnalyses, setSelectedAnalyses] = useState<SelectedAnalysis[]>([])
	const [selectedCatalogItems, setSelectedCatalogItems] = useState<WeedListItem[]>([])
	const [targetFractionForAdd, setTargetFractionForAdd] = useState(DEFAULT_FRACTION_ID)
	const [isLoadingMarkup, setIsLoadingMarkup] = useState(false)
	const [newFractionName, setNewFractionName] = useState("")
	const [wasModalOpen, setWasModalOpen] = useState(false)

	const dataSourceManager = useDataSourceManager(
		addObjectsToFraction,
		setFractions,
		targetFractionForAdd
	)
	const controlMode = useControlModeManager(
		fractions,
		moveObjectToFraction,
		moveMultipleObjectsToFraction
	)

	useEffect(() => {
		if (pendingReclassification) {
			modalManager.openModal("reclassificationConfirm", pendingReclassification.modalData)
			setWasModalOpen(true)
		} else if (wasModalOpen) {
			modalManager.closeModal()
			setWasModalOpen(false)
		}
	}, [pendingReclassification, modalManager, wasModalOpen])

	const handleRemoveFraction = useCallback(() => {
		const { fractionToRemove } = modalManager.modalData as {
			fractionToRemove: FractionModalData
		}
		if (!fractionToRemove) return
		removeFraction(fractionToRemove.id)
		showInfo(`Фракция "${truncate.medium(fractionToRemove.name)}" удалена`)
		modalManager.closeModal()
	}, [removeFraction, showInfo, modalManager])

	const handleEditFraction = useCallback(() => {
		const { fractionToEdit } = modalManager.modalData as {
			fractionToEdit: FractionModalData
		}
		if (!fractionToEdit) return
		editFractionName(fractionToEdit.id, newFractionName)
		modalManager.closeModal()
	}, [editFractionName, newFractionName, modalManager])

	const handleResetFractions = useCallback(() => {
		clearFractions()
		setSelectedAnalyses([])
		setSelectedCatalogItems([])
		showSuccess("Все фракции сброшены")
		modalManager.closeModal()
	}, [clearFractions, showSuccess, modalManager])

	const handleAddAnalysis = useCallback(
		(analysis: Analysis) => {
			if (selectedAnalyses.some((a) => a.id === analysis.id)) return
			setSelectedAnalyses((prev) => [...prev, { id: analysis.id }])
			dataSourceManager.addAnalysis(analysis)
		},
		[selectedAnalyses, dataSourceManager]
	)

	const handleRemoveAnalysis = useCallback(
		(analysisId: string) => {
			setSelectedAnalyses((prev) => prev.filter((a) => a.id !== analysisId))
			removeObjectsByAnalysisId(analysisId)
		},
		[removeObjectsByAnalysisId]
	)

	const handleRemoveAllAnalyses = useCallback(() => {
		for (const a of selectedAnalyses) {
			removeObjectsByAnalysisId(a.id)
		}
		setSelectedAnalyses([])
	}, [selectedAnalyses, removeObjectsByAnalysisId])

	const handleConfirmCatalogSelection = useCallback(
		async (newSelectedItems: WeedListItem[]) => {
			const currentIds = new Set(selectedCatalogItems.map((i) => i.id))
			const newIds = new Set(newSelectedItems.map((i) => i.id))

			const itemsToAdd = newSelectedItems.filter((i) => !currentIds.has(i.id))
			const itemsToRemove = selectedCatalogItems.filter((i) => !newIds.has(i.id))

			setSelectedCatalogItems(newSelectedItems)

			for (const item of itemsToRemove) {
				removeObjectsByCatalogItemId(item.id)
			}

			if (itemsToAdd.length > 0) {
				try {
					const itemsWithAnalysisObjects = await Promise.all(
						itemsToAdd.map(async (item) => ({
							name: item.name,
							analysisObjects: await fetchWeedAnalysisObjects(item.id),
						}))
					)

					dataSourceManager.addCatalogItemsToQueue(itemsWithAnalysisObjects)
				} catch {
					showError("Не удалось получить объекты для выбранных элементов каталога")
				}
			}
		},
		[selectedCatalogItems, dataSourceManager, removeObjectsByCatalogItemId, showError]
	)

	const handleSaveMarkup = useCallback(
		async (markup: SaveMarkup, id?: string) => {
			try {
				const promise = id ? updateMarkup(id, markup) : createMarkup(markup)
				await promise
				const message = id ? "обновлена" : "сохранена"
				showSuccess(`Разметка "${truncate.medium(markup.name)}" успешно ${message}!`)
			} catch (error) {
				showError("Ошибка сохранения разметки")
				log.error("Markup save failed:", error)
			}
		},
		[showSuccess, showError]
	)

	const handleOpenMarkupSaveDialog = useCallback(() => {
		if (fractions.length === 0) {
			showError("Невозможно сохранить пустую разметку")
			return
		}
		modalManager.openModal("markupSave")
	}, [fractions.length, showError, modalManager])

	const handleSelectMarkup = useCallback(
		async (markup: Markup) => {
			setIsLoadingMarkup(true)
			modalManager.closeModal()
			try {
				const detailedMarkup = await getMarkup(markup.id.toString())
				const { fractions: markupFractions, analyses_ids } = detailedMarkup

				const analysisObjectsMap = Object.fromEntries(
					await Promise.all(
						analyses_ids.map(async (id) => [
							id.toString(),
							await fetchAnalysisObjects(id.toString()),
						])
					)
				)
				setSelectedAnalyses(analyses_ids.map((id) => ({ id: id.toString() })))

				const objectLookup = createObjectLookup(analysisObjectsMap, analyses_ids)
				const transformedFractions = markupFractions.map((f) => ({
					id: f.id,
					name: f.name,
					objects: f.object_ids
						.map((objId) => {
							const obj = objectLookup[objId]
							if (!obj) return null

							return {
								...obj,
								file: getObjectImageUrl(obj) ?? obj.file ?? "",
								source: {
									type: "analysis" as const,
									sourceId: obj.analysisId,
								},
								classificationState: "auto" as const,
							}
						})
						.filter(Boolean) as KalibriObjectWithSource[],
				}))

				const otherObjects = transformedFractions.find((f) => f.name === "Прочее")?.objects || []
				const finalFractions: FractionType[] = [
					{ id: "0", name: "Прочее", objects: otherObjects },
					...transformedFractions.filter((f) => f.name !== "Прочее"),
				]

				setFractions(finalFractions)
				showSuccess(`Разметка "${truncate.medium(markup.name)}" успешно загружена!`)
			} catch (error) {
				showError(`Ошибка загрузки разметки, ${error}`)
				log.error("Markup load failed:", error)
			} finally {
				setIsLoadingMarkup(false)
			}
		},
		[setFractions, showSuccess, showError, modalManager]
	)

	const handleSetClassificationRules = useCallback(
		(rules?: import("@/hooks/useFractions").ClassificationRules) => {
			const { fractionToEdit } = modalManager.modalData as {
				fractionToEdit: FractionModalData
			}
			if (fractionToEdit) {
				setClassificationRules(fractionToEdit.id, rules)
			}
		},
		[setClassificationRules, modalManager.modalData]
	)

	const openLoadModal = (fractionId: string = DEFAULT_FRACTION_ID) => {
		setTargetFractionForAdd(fractionId)
		modalManager.openModal("loadModal")
	}
	const openRemoveModal = (fraction: FractionModalData) =>
		modalManager.openModal("removeFraction", { fractionToRemove: fraction })
	const openEditModal = (fraction: FractionModalData) => {
		setNewFractionName(fraction.name)
		modalManager.openModal("editFraction", { fractionToEdit: fraction })
	}
	const openClassificationRuleModal = (fraction: FractionModalData) =>
		modalManager.openModal("fractionClassificationRule", {
			fractionToEdit: fraction,
		})
	const openStatsModal = (fraction: FractionType) =>
		modalManager.openModal("fractionStats", {
			selectedFractionForStats: fraction,
		})
	const openObjectsModal = (fraction: FractionType) =>
		modalManager.openModal("fractionObjects", {
			selectedFractionForObjects: fraction,
		})

	const isLoading =
		isLoadingMarkup || dataSourceManager.isLoading || dataSourceManager.isQueueProcessing

	return (
		<div>
			{isLoading && (
				<div className="fixed inset-0 z-100 flex items-center justify-center bg-base-100/80">
					<Loading />
				</div>
			)}

			<MarkupHeader
				controlMode={controlMode}
				fractions={fractions}
				modalManager={modalManager}
				addNewFraction={addNewFraction}
			/>

			{controlMode.isActive ? (
				<FractionControlMode
					fractions={fractions}
					isObjectSelected={controlMode.isObjectSelected}
					onObjectClick={controlMode.toggleObjectSelection}
					onToggleFractionSelection={controlMode.toggleFractionSelection}
				/>
			) : (
				<FractionList
					fractions={fractions}
					isControlModeActive={controlMode.isActive}
					onOpenLoadModal={openLoadModal}
					onOpenStatsModal={openStatsModal}
					onOpenObjectsModal={openObjectsModal}
					onOpenEditModal={openEditModal}
					onOpenRemoveModal={openRemoveModal}
					onOpenClassificationRuleModal={openClassificationRuleModal}
				/>
			)}

			<AppModals
				modals={modalManager}
				handlers={{
					handleRemoveFraction,
					handleResetFractions,
					handleEditFraction,
					setNewFractionName,
					handleAddAnalysis,
					handleRemoveAnalysis,
					handleRemoveAllAnalyses,
					handleConfirmCatalogSelection,
					handleSelectMarkup,
					handleSaveMarkup,
					handleOpenMarkupSaveDialog,
					handleSetClassificationRules,
					handleReclassificationChoice,
				}}
				data={{
					fractions,
					selectedAnalyses,
					selectedCatalogItems,
					newFractionName,
					reclassificationModalData: pendingReclassification?.modalData,
					...(modalManager.modalData as object),
				}}
				controlModeProps={{
					isControlModeActive: controlMode.isActive,
					isObjectSelected: controlMode.isObjectSelected,
					onObjectClick: controlMode.toggleObjectSelection,
				}}
				addObjectsToFraction={addObjectsToFraction}
			/>
		</div>
	)
}
