import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { FaRegFileAlt } from "react-icons/fa"
import { IoImageOutline, IoSettingsOutline } from "react-icons/io5"
import { MdOutlineAnalytics } from "react-icons/md"
import {
	createProposal,
	deleteProposalImage,
	deleteWeed,
	fetchWeedDetails,
	updateWeed,
	uploadProposalImage,
} from "@/api/catalog"
import { addCoffeeImage, deleteCoffeeImage } from "@/api/catalog/images"
import type { WeedImage, WeedStatistics } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import {
	canEditCatalogItem,
	isModeratorRole,
} from "@/components/catalog/components/CatalogProposalsShared"
import AnalysisTab from "@/components/catalog/forms/AnalysisTab"
import CharacteristicsTab from "@/components/catalog/forms/CharacteristicsTab"
import GeneralInfoTab from "@/components/catalog/forms/GeneralInfoTab"
import PhotosTab from "@/components/catalog/forms/PhotosTab"
import RemoveCoffeeAlert from "@/components/catalog/forms/RemoveCoffeeAlert"
import useItemState from "@/components/catalog/hooks/useItemState"
import { CatalogItemFormLayout } from "@/components/catalog/layout/CatalogItemFormLayout"
import { Button } from "@/components/common/ui/Button"
import Loading from "@/components/common/ui/Loading"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { useAnalysesWithObjects } from "@/hooks/useAnalysesWithObjects"
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

export const Route = createFileRoute("/_authenticated/catalog/$catalogItemId/edit")({
	component: CatalogItemEdit,
})

function CatalogItemEdit() {
	const { catalogItemId } = Route.useParams()
	const navigate = Route.useNavigate()
	const queryClient = useQueryClient()
	const [activeTab, setActiveTab] = useState<Tab>(Tab.GENERAL)
	const {
		itemState,
		setName,
		setDescription,
		setHarmfulness,
		setPhotos,
		addAnalysis,
		removeAnalysis,
		setItemState,
		setAllObjectsData,
		toggleExcludeObject,
	} = useItemState()
	const { showError } = useAlert()
	const { roles } = useAuth()
	const canEdit = canEditCatalogItem(roles)
	const isModerator = isModeratorRole(roles)
	const { data: classifications } = useClassifications()

	const { data: weedDetails, isLoading: isLoadingWeed } = useQuery({
		queryKey: queryKeys.catalog.weedDetails(catalogItemId),
		queryFn: () => fetchWeedDetails(Number(catalogItemId)),
		enabled: !!catalogItemId && canEdit,
	})
	const {
		analyses: analysesData,
		objectsData: initialObjectsData,
		isFetching: isAnalysesFetching,
	} = useAnalysesWithObjects(catalogItemId, weedDetails?.analyses ?? [], {
		enabled: !!weedDetails && (weedDetails.analyses?.length ?? 0) > 0,
	})
	const isLoading = isLoadingWeed || !classifications

	const [selectedPrimary, setSelectedPrimary] = useState<string | null>(null)
	const [selectedSecondary, setSelectedSecondary] = useState<string | null>(null)
	const [selectedTertiary, setSelectedTertiary] = useState<string | null>(null)

	useEffect(() => {
		if (!canEdit) {
			navigate({ to: "/catalog", replace: true })
		}
	}, [canEdit, navigate])
	const primaryClassifications = classifications
		? Object.entries(classifications.main_groups).map(([id, name]) => ({
				id,
				name,
			}))
		: []
	const secondaryClassifications =
		classifications && selectedPrimary
			? Object.keys(classifications.hierarchy[selectedPrimary] || {}).map((id) => ({
					id,
					name: classifications.main_subgroups[id],
				}))
			: []
	const tertiaryClassifications =
		classifications && selectedPrimary && selectedSecondary
			? (classifications.hierarchy[selectedPrimary]?.[selectedSecondary] || []).map((id) => ({
					id,
					name: classifications.subgroups[id],
				}))
			: []

	const [originalImages, setOriginalImages] = useState<WeedImage[]>([])

	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)

	const updateCoffeeMutation = useMutation({
		mutationFn: async () => {
			if (!selectedPrimary || !selectedSecondary) {
				throw new Error("Пожалуйста, выберите классификации")
			}

			const coffeeId = Number(catalogItemId)

			const statistics =
				convertToBackendStatistics(itemState.allObjectsData, itemState.excludedObjects) || undefined

			const analysisIds = itemState.analyses.map((analysis) => analysis.id)

			await updateWeed(coffeeId, {
				name: itemState.name,
				description: itemState.description,
				main_group: selectedPrimary || undefined,
				main_subgroup: selectedSecondary || undefined,
				subgroup: selectedTertiary || undefined,
				latin_name: weedDetails?.latin_name ?? "",
				is_quarantine: false,
				harmfulness: itemState.harmfulness,
				statistics: statistics as WeedStatistics,
				analysis_ids: analysisIds,
				excluded_objects: itemState.excludedObjects,
			})

			const currentImageUrls = itemState.photos.filter((p): p is string => typeof p === "string")

			const deletePromises = originalImages
				.filter((img: WeedImage) => !currentImageUrls.includes(img.url))
				.map((img) => deleteCoffeeImage(coffeeId, img.id))

			const addPromises = itemState.photos
				.map((photo) => {
					if (photo instanceof File) {
						return addCoffeeImage(coffeeId, photo)
					}
					return null
				})
				.filter(Boolean) as Promise<unknown>[]

			await Promise.all([...deletePromises, ...addPromises])
		},
		onSuccess: async () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.items })
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.weedDetails(catalogItemId) })
			navigate({ to: "/catalog", replace: true })
		},
		onError: (err) => {
			log.devError("Failed to update coffee", err)
			showError(getUserFacingErrorMessage(err))
		},
	})

	const deleteCoffeeMutation = useMutation({
		mutationFn: async () => {
			const coffeeId = Number(catalogItemId)
			await deleteWeed(coffeeId)
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.items })
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.weedDetails(catalogItemId) })
			navigate({ to: "/catalog", replace: true })
		},
		onError: (err) => {
			log.devError("Failed to delete coffee", err)
			showError(getUserFacingErrorMessage(err))
		},
	})

	const createProposalMutation = useMutation({
		mutationFn: async () => {
			if (!selectedPrimary || !selectedSecondary) {
				throw new Error("Пожалуйста, выберите классификации")
			}
			const statistics =
				convertToBackendStatistics(itemState.allObjectsData, itemState.excludedObjects) || undefined
			const proposal = await createProposal({
				target_weed_id: Number(catalogItemId),
				name: itemState.name,
				description: itemState.description,
				harmfulness: itemState.harmfulness || undefined,
				main_group: selectedPrimary ?? undefined,
				main_subgroup: selectedSecondary ?? undefined,
				subgroup: selectedTertiary ?? undefined,
				analysis_ids:
					itemState.analyses.length > 0 ? itemState.analyses.map((a) => a.id) : undefined,
				excluded_objects:
					itemState.excludedObjects.length > 0 ? itemState.excludedObjects : undefined,
				statistics,
			})
			const keptUrls = new Set(
				itemState.photos.filter((p): p is string => typeof p === "string") as string[]
			)
			const proposalImages = proposal.images ?? []
			const pendingByUrl = new Map(proposalImages.map((img) => [img.url, img.id]))
			for (const orig of originalImages) {
				if (!keptUrls.has(orig.url)) {
					const pendingId = pendingByUrl.get(orig.url)
					if (pendingId !== undefined) {
						await deleteProposalImage(proposal.id, pendingId)
					}
				}
			}
			const photosToUpload = itemState.photos.filter((p): p is File => p instanceof File)
			for (const file of photosToUpload) {
				try {
					await uploadProposalImage(proposal.id, file)
				} catch (err) {
					log.devError("Failed to upload proposal image", err)
					showError("Предложение создано, но не удалось загрузить одно из фото.")
					break
				}
			}
			return proposal
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
			navigate({ to: "/catalog/proposals", replace: true })
		},
		onError: (err) => {
			log.devError("Failed to create proposal", err)
			showError(getUserFacingErrorMessage(err))
		},
	})

	useEffect(() => {
		if (!weedDetails || !classifications) return
		setOriginalImages(weedDetails.images)
		setItemState((prev) => ({
			...prev,
			name: weedDetails.name,
			description: weedDetails.description ?? "",
			harmfulness: weedDetails.harmfulness ?? "",
			photos: weedDetails.images.map((img) => img.url),
			excludedObjects: weedDetails.statistics?.excluded_objects || [],
		}))
		const primaryCode =
			Object.entries(classifications.main_groups).find(
				([, name]) => name === (weedDetails.main_group ?? "")
			)?.[0] ?? null
		const secondaryCode =
			Object.entries(classifications.main_subgroups).find(
				([, name]) => name === (weedDetails.main_subgroup ?? "")
			)?.[0] ?? null
		const tertiaryCode =
			Object.entries(classifications.subgroups).find(
				([, name]) => name === (weedDetails.subgroup ?? "")
			)?.[0] ?? null
		setSelectedPrimary(primaryCode)
		setSelectedSecondary(secondaryCode)
		setSelectedTertiary(tertiaryCode)
	}, [weedDetails, classifications, setItemState])

	useEffect(() => {
		const hasInitialAnalyses = (weedDetails?.analyses?.length ?? 0) > 0
		if (hasInitialAnalyses && analysesData) {
			setItemState((prev) => ({
				...prev,
				analyses: analysesData,
			}))
		}
	}, [analysesData, setItemState, weedDetails?.analyses])

	useEffect(() => {
		const hasInitialAnalyses = (weedDetails?.analyses?.length ?? 0) > 0
		if (hasInitialAnalyses && initialObjectsData) {
			setAllObjectsData(initialObjectsData)
		}
	}, [initialObjectsData, setAllObjectsData, weedDetails?.analyses])

	const handleCancel = () => {
		navigate({ to: "/catalog", replace: true })
	}

	const handleSave = () => {
		if (!selectedPrimary || !selectedSecondary) {
			showError("Пожалуйста, выберите классификации")
			return
		}
		if (isModerator) {
			if (!selectedTertiary) {
				showError("Пожалуйста, выберите классификации")
				return
			}
			updateCoffeeMutation.mutate()
		} else {
			createProposalMutation.mutate()
		}
	}

	const handleDeleteClick = () => {
		setIsDeleteModalOpen(true)
	}

	const handleDeleteConfirm = () => {
		setIsDeleteModalOpen(false)
		deleteCoffeeMutation.mutate()
	}

	const handleDeleteCancel = () => {
		setIsDeleteModalOpen(false)
	}

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
							setSelectedTertiary(null)
						}}
						selectedSecondaryClassification={selectedSecondary}
						onSecondaryClassificationChange={(id) => {
							setSelectedSecondary(id)
							setSelectedTertiary(null)
						}}
						selectedTertiaryClassification={selectedTertiary}
						onTertiaryClassificationChange={setSelectedTertiary}
						onDelete={isModerator ? handleDeleteClick : undefined}
					/>
				)
			case Tab.CHARACTERISTICS:
				return (
					<CharacteristicsTab
						analyses={itemState.analyses}
						onSelectAnalysis={() => {
							setActiveTab(Tab.ANALYTICS)
							resetScroll()
						}}
						allObjectsData={itemState.allObjectsData}
						excludedObjects={itemState.excludedObjects}
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
						isParentLoading={isAnalysesFetching}
					/>
				)
			case Tab.PHOTOS:
				return <PhotosTab photos={itemState.photos} onPhotosChange={setPhotos} />
		}
	}

	const tabs = [
		{ id: Tab.GENERAL, label: Tab.GENERAL, icon: <FaRegFileAlt /> },
		{ id: Tab.CHARACTERISTICS, label: Tab.CHARACTERISTICS, icon: <IoSettingsOutline /> },
		{ id: Tab.ANALYTICS, label: Tab.ANALYTICS, icon: <MdOutlineAnalytics /> },
		{ id: Tab.PHOTOS, label: Tab.PHOTOS, icon: <IoImageOutline /> },
	]

	if (!canEdit) {
		return null
	}

	if (isLoading) {
		return (
			<div className="flex min-h-0 flex-1 items-center justify-center bg-base-100">
				<Loading />
			</div>
		)
	}

	const pageTitle = itemState.name.trim() || weedDetails?.name || "Редактирование"
	const pageSubtitle = isModerator
		? "Изменения сохраняются в каталог"
		: "Будет создана заявка на изменение позиции"
	const isSavePending = isModerator
		? updateCoffeeMutation.isPending
		: createProposalMutation.isPending

	return (
		<>
			<CatalogItemFormLayout
				title={pageTitle}
				subtitle={pageSubtitle}
				onBack={handleCancel}
				tabs={tabs}
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
							!itemState.name.trim() || !selectedPrimary || !selectedSecondary || isSavePending
						}
					>
						{isSavePending
							? isModerator
								? "Сохранение..."
								: "Отправка..."
							: isModerator
								? "Сохранить"
								: "Отправить"}
					</Button>
				}
			>
				{renderTabContent()}
			</CatalogItemFormLayout>

			<RemoveCoffeeAlert
				isOpen={isDeleteModalOpen}
				onClose={handleDeleteCancel}
				onConfirm={handleDeleteConfirm}
				coffeeName={itemState.name}
			/>
		</>
	)
}
