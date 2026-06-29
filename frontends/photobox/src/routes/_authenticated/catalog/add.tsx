import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect, useMemo, useState } from "react"
import { FaRegFileAlt } from "react-icons/fa"
import { IoImageOutline, IoSettingsOutline } from "react-icons/io5"
import { MdOutlineAnalytics } from "react-icons/md"
import { applyProposal, createProposal, uploadProposalImage } from "@/api/catalog"
import type { ClassificationsResponse } from "@/api/catalog/classifications"
import { queryKeys } from "@/api/queryKeys"
import {
	canEditCatalogItem,
	isModeratorRole,
} from "@/components/catalog/components/CatalogProposalsShared"
import AnalysisTab from "@/components/catalog/forms/AnalysisTab"
import CharacteristicsTab from "@/components/catalog/forms/CharacteristicsTab"
import GeneralInfoTab from "@/components/catalog/forms/GeneralInfoTab"
import PhotosTab from "@/components/catalog/forms/PhotosTab"
import useItemState from "@/components/catalog/hooks/useItemState"
import { CatalogItemFormLayout } from "@/components/catalog/layout/CatalogItemFormLayout"
import { Button } from "@/components/common/ui/Button"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { useClassifications } from "@/hooks/useClassifications"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import { resetScroll } from "@/utils/scroll"
import { convertToBackendStatistics } from "@/utils/stats"

enum Tab {
	GENERAL = "Основное",
	CHARACTERISTICS = "Характеристики",
	ANALYTICS = "Анализы",
	PHOTOS = "Фото",
}

const TABS = [
	{ id: Tab.GENERAL, label: Tab.GENERAL, icon: <FaRegFileAlt /> },
	{ id: Tab.CHARACTERISTICS, label: Tab.CHARACTERISTICS, icon: <IoSettingsOutline /> },
	{ id: Tab.ANALYTICS, label: Tab.ANALYTICS, icon: <MdOutlineAnalytics /> },
	{ id: Tab.PHOTOS, label: Tab.PHOTOS, icon: <IoImageOutline /> },
] as const

function useClassificationHierarchy(
	classifications: ClassificationsResponse | undefined,
	selectedPrimary: string | null,
	selectedSecondary: string | null
) {
	return useMemo(() => {
		if (!classifications) {
			return { primary: [], secondary: [], tertiary: [] }
		}
		const primary = Object.entries(classifications.main_groups).map(([id, name]) => ({ id, name }))
		const secondary =
			selectedPrimary != null
				? Object.keys(classifications.hierarchy[selectedPrimary] ?? {}).map((id) => ({
						id,
						name: classifications.main_subgroups[id],
					}))
				: []
		const tertiary =
			selectedPrimary != null && selectedSecondary != null
				? (classifications.hierarchy[selectedPrimary]?.[selectedSecondary] ?? []).map((id) => ({
						id,
						name: classifications.subgroups[id],
					}))
				: []
		return { primary, secondary, tertiary }
	}, [classifications, selectedPrimary, selectedSecondary])
}

function CatalogAddPage() {
	const navigate = useNavigate({ from: "/catalog/add" })
	const queryClient = useQueryClient()
	const { showError } = useAlert()
	const { roles } = useAuth()
	const canEdit = canEditCatalogItem(roles)
	useEffect(() => {
		if (!canEdit) {
			navigate({ to: "/catalog", replace: true })
		}
	}, [canEdit, navigate])

	const [activeTab, setActiveTab] = useState<Tab>(Tab.GENERAL)
	const [isSubmitting, setIsSubmitting] = useState(false)
	const [selectedPrimary, setSelectedPrimary] = useState<string | null>(null)
	const [selectedSecondary, setSelectedSecondary] = useState<string | null>(null)
	const [selectedTertiary, setSelectedTertiary] = useState<string | null>(null)

	const {
		itemState,
		setName,
		setDescription,
		setHarmfulness,
		setPhotos,
		addAnalysis,
		removeAnalysis,
		setAllObjectsData,
		toggleExcludeObject,
	} = useItemState()
	const { data: classifications } = useClassifications()

	const {
		primary: primaryClassifications,
		secondary: secondaryClassifications,
		tertiary: tertiaryClassifications,
	} = useClassificationHierarchy(classifications, selectedPrimary, selectedSecondary)

	useEffect(() => {
		if (selectedPrimary != null) {
			setSelectedSecondary(null)
			setSelectedTertiary(null)
		}
	}, [selectedPrimary])

	useEffect(() => {
		if (selectedSecondary != null) setSelectedTertiary(null)
	}, [selectedSecondary])

	const handleCancel = () => navigate({ to: "/catalog", replace: true })

	const handleSave = async () => {
		if (isSubmitting) return
		if (!itemState.name.trim()) {
			showError("Укажите название")
			return
		}
		if (!selectedPrimary || !selectedSecondary) {
			showError("Пожалуйста, выберите классификации")
			return
		}
		if (isModeratorRole(roles) && !selectedTertiary) {
			showError("Пожалуйста, выберите классификации")
			return
		}
		setIsSubmitting(true)
		try {
			const analysisIds = itemState.analyses.map((a) => a.id)
			const excludedObjects = itemState.excludedObjects
			const statistics =
				convertToBackendStatistics(itemState.allObjectsData, excludedObjects) ?? undefined

			const proposal = await createProposal({
				name: itemState.name,
				description: itemState.description,
				harmfulness: itemState.harmfulness,
				main_group: selectedPrimary ?? undefined,
				main_subgroup: selectedSecondary ?? undefined,
				subgroup: selectedTertiary ?? undefined,
				analysis_ids: analysisIds.length > 0 ? analysisIds : undefined,
				excluded_objects: excludedObjects.length > 0 ? excludedObjects : undefined,
				statistics,
			})

			const photosToUpload = itemState.photos.filter((p): p is File => p instanceof File)
			for (const file of photosToUpload) {
				try {
					await uploadProposalImage(proposal.id, file)
				} catch (err) {
					log.devError("Failed to upload proposal images", err)
					showError("Предложение отправлено, но не удалось загрузить фото.")
					break
				}
			}

			if (isModeratorRole(roles)) {
				try {
					const applied = await applyProposal(proposal.id)
					queryClient.invalidateQueries({ queryKey: queryKeys.catalog.items })
					queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
					if (applied.applied_weed_id != null) {
						navigate({
							to: "/catalog/$catalogItemId",
							params: { catalogItemId: String(applied.applied_weed_id) },
							replace: true,
						})
					} else {
						navigate({ to: "/catalog", replace: true })
					}
				} catch (applyErr) {
					log.devError("Failed to apply proposal", applyErr)
					showError(
						"Предложение создано, но не удалось добавить в каталог. Попробуйте применить его вручную."
					)
					queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
					navigate({ to: "/catalog/proposals", replace: true })
				}
			} else {
				queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
				navigate({ to: "/catalog/proposals", replace: true })
			}
		} catch (err) {
			log.devError("Failed to create proposal", err)
			showError(getUserFacingErrorMessage(err))
		} finally {
			setIsSubmitting(false)
		}
	}
	const isSavePending = isSubmitting

	const renderTabContent = () => {
		switch (activeTab) {
			case Tab.GENERAL:
				return (
					<GeneralInfoTab
						productName={itemState.name}
						onNameChange={setName}
						description={itemState.description}
						onDescriptionChange={setDescription}
						harmfulness={itemState.harmfulness}
						onHarmfulnessChange={setHarmfulness}
						primaryClassifications={primaryClassifications}
						secondaryClassifications={secondaryClassifications}
						tertiaryClassifications={tertiaryClassifications}
						selectedPrimaryClassification={selectedPrimary}
						onPrimaryClassificationChange={(id) => {
							setSelectedPrimary(id)
							setSelectedSecondary(null)
						}}
						selectedSecondaryClassification={selectedSecondary}
						onSecondaryClassificationChange={setSelectedSecondary}
						selectedTertiaryClassification={selectedTertiary}
						onTertiaryClassificationChange={setSelectedTertiary}
					/>
				)
			case Tab.CHARACTERISTICS:
				return (
					<CharacteristicsTab
						analyses={itemState.analyses}
						allObjectsData={itemState.allObjectsData}
						excludedObjects={itemState.excludedObjects}
						onSelectAnalysis={() => {
							setActiveTab(Tab.ANALYTICS)
							resetScroll()
						}}
					/>
				)
			case Tab.ANALYTICS:
				return (
					<AnalysisTab
						selectedAnalyses={itemState.analyses}
						onAddAnalysis={addAnalysis}
						onRemoveAnalysis={removeAnalysis}
						allObjectsData={itemState.allObjectsData}
						excludedObjects={itemState.excludedObjects}
						onToggleExclude={toggleExcludeObject}
						onObjectsDataChange={setAllObjectsData}
					/>
				)
			case Tab.PHOTOS:
				return <PhotosTab photos={itemState.photos} onPhotosChange={setPhotos} />
		}
	}

	if (!canEdit) {
		return null
	}

	return (
		<CatalogItemFormLayout
			title="Новая заявка"
			subtitle={
				isModeratorRole(roles)
					? "Создание записи в каталоге"
					: "Заполните карточку и отправьте на проверку"
			}
			onBack={handleCancel}
			tabs={[...TABS]}
			activeTabId={activeTab}
			onTabChange={(id) => {
				setActiveTab(id as Tab)
				resetScroll()
			}}
			footerActions={
				<Button
					type="button"
					fullWidth
					onClick={handleSave}
					disabled={
						isSavePending ||
						!itemState.name.trim() ||
						!selectedPrimary ||
						!selectedSecondary ||
						(isModeratorRole(roles) && !selectedTertiary)
					}
				>
					{isSavePending ? "Отправка..." : "Отправить"}
				</Button>
			}
		>
			{renderTabContent()}
		</CatalogItemFormLayout>
	)
}

export const Route = createFileRoute("/_authenticated/catalog/add")({
	component: CatalogAddPage,
})
