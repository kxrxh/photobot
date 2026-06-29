import { useNavigate } from "@tanstack/react-router"
import { memo } from "react"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import type { WeedListItem } from "@/api/catalog/types"
import type { Markup, SaveMarkup } from "@/api/markup/types"
import AnalysisSelectorSheet from "@/components/analysis/selectors/AnalysisSelector"
import CatalogSelectorSheet from "@/components/catalog/selectors/CatalogSelector"
import ClassificationSaveSheet from "@/components/markup/dialogs/ClassificationSaveSheet"
import EditFractionAlert from "@/components/markup/dialogs/EditFractionAlert"
import FractionClassificationRuleSheet from "@/components/markup/dialogs/FractionClassificationRuleSheet"
import FractionObjectsSheet from "@/components/markup/dialogs/FractionObjectsSheet"
import FractionStatsSheet from "@/components/markup/dialogs/FractionStatsSheet"
import LoadActionSheet from "@/components/markup/dialogs/LoadActionSheet"
import MarkupActionSheet from "@/components/markup/dialogs/MarkupActionSheet"
import MarkupSaveSheet from "@/components/markup/dialogs/MarkupSaveSheet"
import MarkupSelectorSheet from "@/components/markup/dialogs/MarkupSelectorSheet"
import ReclassificationConfirmAlert, {
	type ReclassificationModalData,
} from "@/components/markup/dialogs/ReclassificationConfirmAlert"
import RemoveFractionAlert from "@/components/markup/dialogs/RemoveFractionAlert"
import ResetFractionsAlert from "@/components/markup/dialogs/ResetFractionsAlert"
import type { FractionType } from "@/hooks/useFractions"
import type { SelectedAnalysis } from "@/routes/_authenticated/markup"

type ModalType =
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

interface FractionModalData {
	id: string
	name: string
}

type ModalManager = {
	activeModal: ModalType | null
	modalData: Record<string, unknown>
	openModal: (modalName: ModalType, data?: Record<string, unknown>) => void
	closeModal: () => void
}

const AppModals = memo(
	({
		modals,
		handlers,
		data,
		controlModeProps,
	}: {
		modals: ModalManager
		handlers: {
			handleRemoveFraction: () => void
			handleResetFractions: () => void
			handleEditFraction: () => void
			setNewFractionName: (name: string) => void
			handleAddAnalysis: (analysis: Analysis) => void
			handleRemoveAnalysis: (id: string) => void
			handleRemoveAllAnalyses: () => void
			handleConfirmCatalogSelection: (selectedItems: WeedListItem[]) => Promise<void>
			handleSelectMarkup: (markup: Markup) => Promise<void>
			handleSaveMarkup: (markup: SaveMarkup, id?: string) => Promise<void>
			handleOpenMarkupSaveDialog: () => void
			handleSetClassificationRules: (
				rules?: import("@/hooks/useFractions").ClassificationRules
			) => void
			handleReclassificationChoice: (
				action: import("@/components/markup/dialogs/ReclassificationConfirmAlert").ReclassificationAction
			) => void
		}
		data: {
			fractions: FractionType[]
			selectedAnalyses: SelectedAnalysis[]
			selectedCatalogItems: WeedListItem[]
			fractionToEdit?: FractionModalData
			fractionToRemove?: FractionModalData
			newFractionName: string
			selectedFractionForStats?: FractionType
			selectedFractionForObjects?: FractionType
			reclassificationModalData?: ReclassificationModalData
		}
		controlModeProps: {
			isControlModeActive: boolean
			isObjectSelected: (id: number) => boolean
			onObjectClick: (id: number, fractionId: string) => void
		}
		addObjectsToFraction: (
			fractionId: string,
			objects: KalibriObject[],
			source: { type: "analysis" | "catalog"; sourceId: number | string }
		) => void
	}) => {
		const navigate = useNavigate()
		const { activeModal, closeModal, openModal } = modals
		const {
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
		} = handlers

		return (
			<>
				<RemoveFractionAlert
					isOpen={activeModal === "removeFraction"}
					onClose={closeModal}
					onConfirm={handleRemoveFraction}
					fractionName={data.fractionToRemove?.name}
				/>
				<ResetFractionsAlert
					isOpen={activeModal === "resetFractions"}
					onClose={closeModal}
					onConfirm={handleResetFractions}
				/>
				<EditFractionAlert
					isOpen={activeModal === "editFraction"}
					onClose={closeModal}
					onConfirm={handleEditFraction}
					newFractionName={data.newFractionName}
					setNewFractionName={setNewFractionName}
				/>
				<LoadActionSheet
					isOpen={activeModal === "loadModal"}
					onClose={closeModal}
					openAnalysisModal={() => openModal("analysisSelector")}
					openCatalogSelectorModal={() => openModal("catalogSelector")}
					openPhotoModal={() => {
						closeModal()
						void navigate({ to: "/analysis/create", search: { openRequest: undefined } })
					}}
				/>
				<AnalysisSelectorSheet
					isOpen={activeModal === "analysisSelector"}
					onClose={closeModal}
					selectedAnalysisIds={data.selectedAnalyses.map((a) => a.id)}
					onAddAnalysis={handleAddAnalysis}
					onRemoveAnalysis={handleRemoveAnalysis}
					onRemoveAllAnalyses={handleRemoveAllAnalyses}
					hasAddedAnalyses={data.selectedAnalyses.length > 0}
					onOpenCreateDialog={() => {
						closeModal()
						void navigate({ to: "/analysis/create", search: { openRequest: undefined } })
					}}
				/>
				<CatalogSelectorSheet
					isOpen={activeModal === "catalogSelector"}
					onClose={closeModal}
					initialSelectedItems={data.selectedCatalogItems}
					onConfirm={handleConfirmCatalogSelection}
				/>
				<MarkupActionSheet
					isOpen={activeModal === "markupModal"}
					onClose={closeModal}
					onOpenMarkupSelector={() => openModal("markupSelector")}
					onOpenSaveDialog={handleOpenMarkupSaveDialog}
				/>
				<MarkupSelectorSheet
					isOpen={activeModal === "markupSelector"}
					onClose={closeModal}
					onSelectMarkup={handleSelectMarkup}
				/>
				<MarkupSaveSheet
					isOpen={activeModal === "markupSave"}
					onClose={closeModal}
					onSave={handleSaveMarkup}
					fractions={data.fractions}
					analyses={data.selectedAnalyses}
				/>
				<ClassificationSaveSheet
					isOpen={activeModal === "classificationSave"}
					onClose={closeModal}
					fractions={data.fractions}
				/>
				<FractionStatsSheet
					isOpen={activeModal === "fractionStats"}
					onClose={closeModal}
					fraction={data.selectedFractionForStats ?? null}
				/>
				<FractionObjectsSheet
					isOpen={activeModal === "fractionObjects"}
					onClose={closeModal}
					fraction={data.selectedFractionForObjects ?? null}
					isControlModeActive={controlModeProps.isControlModeActive}
					isObjectSelected={controlModeProps.isObjectSelected}
					onObjectClick={controlModeProps.onObjectClick}
				/>
				<FractionClassificationRuleSheet
					isOpen={activeModal === "fractionClassificationRule"}
					onClose={closeModal}
					onConfirm={handleSetClassificationRules}
					fractionName={data.fractionToEdit?.name || ""}
					initialRules={
						data.fractionToEdit
							? data.fractions.find((f) => f.id === data.fractionToEdit?.id)?.classificationRules
							: undefined
					}
				/>
				<ReclassificationConfirmAlert
					isOpen={activeModal === "reclassificationConfirm"}
					onConfirm={handlers.handleReclassificationChoice}
					modalData={
						(data.reclassificationModalData as ReclassificationModalData) ||
						(modals.modalData as unknown as ReclassificationModalData)
					}
				/>
			</>
		)
	}
)

export default AppModals
