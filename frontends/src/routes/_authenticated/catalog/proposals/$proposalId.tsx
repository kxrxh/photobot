import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { isTimeoutError } from "ky"
import { useEffect, useState } from "react"
import { FaRegFileAlt } from "react-icons/fa"
import { IoImageOutline, IoSettingsOutline } from "react-icons/io5"
import { MdOutlineAnalytics } from "react-icons/md"
import {
	applyProposal,
	deleteProposalImage,
	fetchProposal,
	rejectProposal,
	requestChanges,
	submitProposal,
	updateProposalDraft,
	uploadProposalImage,
} from "@/api/catalog"
import type { ProposalStatus, WeedImage, WeedStatistics } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import { ApiError } from "@/api/types"
import { isModeratorRole } from "@/components/catalog/components/CatalogProposalsShared"
import CatalogProposalDecisionSheet from "@/components/catalog/dialogs/CatalogProposalDecisionSheet"
import AnalysisTab from "@/components/catalog/forms/AnalysisTab"
import CharacteristicsTab from "@/components/catalog/forms/CharacteristicsTab"
import GeneralInfoTab from "@/components/catalog/forms/GeneralInfoTab"
import PhotosTab from "@/components/catalog/forms/PhotosTab"
import useItemState from "@/components/catalog/hooks/useItemState"
import { CatalogItemFormLayout } from "@/components/catalog/layout/CatalogItemFormLayout"
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

export const Route = createFileRoute("/_authenticated/catalog/proposals/$proposalId")({
	component: CatalogItemEdit,
})

function CatalogItemEdit() {
	const { proposalId } = Route.useParams()
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
	const [formReady, setFormReady] = useState(false)

	const [selectedPrimary, setSelectedPrimary] = useState<string | null>(null)
	const [selectedSecondary, setSelectedSecondary] = useState<string | null>(null)
	const [selectedTertiary, setSelectedTertiary] = useState<string | null>(null)
	const [reviewNote, setReviewNote] = useState("")
	const [isDecisionModalOpen, setIsDecisionModalOpen] = useState(false)

	const { showError, showSuccess } = useAlert()
	const { roles, userId } = useAuth()
	const isModerator = isModeratorRole(roles)

	const { data: catalogProposal, isLoading: isProposalLoading } = useQuery({
		queryKey: queryKeys.catalog.proposal(proposalId),
		queryFn: () => fetchProposal(Number(proposalId)),
		refetchOnWindowFocus: false,
	})

	const currentUserId = userId != null ? Number(userId) : null
	const isAuthor =
		!!catalogProposal && currentUserId != null && catalogProposal.request_by === currentUserId
	const canSendToRecheck = isAuthor && catalogProposal?.status === "changes_requested"
	const isSubmittedReadOnly = isAuthor && catalogProposal?.status === "submitted" && !isModerator

	const { data: classifications } = useClassifications()
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

	const updateProposalMutation = useMutation({
		mutationFn: async ({ status, notes }: { status: ProposalStatus; notes?: string }) => {
			const id = Number(proposalId)
			if (status === "applied") {
				return applyProposal(id, notes)
			}
			if (status === "rejected") {
				return rejectProposal(id, notes || "")
			}
			if (status === "changes_requested") {
				return requestChanges(id, notes || "")
			}
			throw new Error(`Unsupported status: ${status}`)
		},
		retry: false,
		onSuccess: async () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
			queryClient.invalidateQueries({
				queryKey: queryKeys.catalog.proposal(proposalId),
			})
			await queryClient.refetchQueries({
				queryKey: queryKeys.catalog.proposals,
				exact: false,
			})
			navigate({ to: "/catalog/proposals", replace: true })
		},
		onError: (err) => {
			log.devError("Failed to update proposal status", err)

			if (isTimeoutError(err)) {
				void queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
				void queryClient.invalidateQueries({
					queryKey: queryKeys.catalog.proposal(proposalId),
				})
				showError(
					"Ответ не пришёл вовремя. Проверьте список заявок — действие могло уже выполниться."
				)
				return
			}

			if (ApiError.isApiError(err) && err.isBadRequest()) {
				const m = err.message.toLowerCase()
				if (m.includes("changes_requested") || m.includes("only request changes on submitted")) {
					void queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
					void queryClient.invalidateQueries({
						queryKey: queryKeys.catalog.proposal(proposalId),
					})
					showError("Статус заявки уже изменён. Обновите список заявок.")
					return
				}
			}

			showError("Не удалось обновить статус заявки")
		},
	})

	const saveDraftAndResubmitMutation = useMutation({
		mutationFn: async () => {
			const id = Number(proposalId)
			if (!selectedPrimary || !selectedSecondary) {
				throw new Error("Пожалуйста, выберите классификации")
			}
			const statistics =
				convertToBackendStatistics(itemState.allObjectsData, itemState.excludedObjects) || undefined
			const draftParams = {
				name: itemState.name.trim(),
				description: itemState.description.trim() || undefined,
				harmfulness: itemState.harmfulness.trim() || undefined,
				main_group: selectedPrimary,
				main_subgroup: selectedSecondary,
				subgroup: selectedTertiary || undefined,
				analysis_ids: itemState.analyses.map((a) => a.id),
				excluded_objects: itemState.excludedObjects,
				statistics: statistics as WeedStatistics | undefined,
			}
			await updateProposalDraft(id, draftParams)

			const currentImageUrls = itemState.photos.filter((p): p is string => typeof p === "string")
			const toDelete = originalImages.filter((img) => !currentImageUrls.includes(img.url))
			for (const img of toDelete) {
				await deleteProposalImage(id, img.id)
			}
			const toAdd = itemState.photos.filter((p): p is File => p instanceof File)
			for (const file of toAdd) {
				await uploadProposalImage(id, file)
			}

			return submitProposal(id)
		},
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.catalog.proposals })
			queryClient.invalidateQueries({
				queryKey: queryKeys.catalog.proposal(proposalId),
			})
			navigate({ to: "/catalog/proposals", replace: true })
		},
		onError: (err) => {
			log.devError("Failed to save draft and resubmit", err)
			showError(getUserFacingErrorMessage(err))
		},
	})

	const saveDraftOnlyMutation = useMutation({
		mutationFn: async () => {
			const id = Number(proposalId)
			if (!selectedPrimary || !selectedSecondary) {
				throw new Error("Пожалуйста, выберите классификации")
			}
			const statistics =
				convertToBackendStatistics(itemState.allObjectsData, itemState.excludedObjects) || undefined
			const draftParams = {
				name: itemState.name.trim(),
				description: itemState.description.trim() || undefined,
				harmfulness: itemState.harmfulness.trim() || undefined,
				main_group: selectedPrimary,
				main_subgroup: selectedSecondary,
				subgroup: selectedTertiary || undefined,
				analysis_ids: itemState.analyses.map((a) => a.id),
				excluded_objects: itemState.excludedObjects,
				statistics: statistics as WeedStatistics | undefined,
			}
			await updateProposalDraft(id, draftParams)

			const currentImageUrls = itemState.photos.filter((p): p is string => typeof p === "string")
			const toDelete = originalImages.filter((img) => !currentImageUrls.includes(img.url))
			for (const img of toDelete) {
				await deleteProposalImage(id, img.id)
			}
			const toAdd = itemState.photos.filter((p): p is File => p instanceof File)
			for (const file of toAdd) {
				await uploadProposalImage(id, file)
			}
		},
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: queryKeys.catalog.proposal(proposalId),
			})
			showSuccess("Черновик сохранён")
		},
		onError: (err) => {
			log.devError("Failed to save draft", err)
			showError(getUserFacingErrorMessage(err))
		},
	})

	// Route param is a proposal id, not a catalog weed id.
	useEffect(() => {
		if (!catalogProposal) {
			setFormReady(false)
			return
		}
		const draft = catalogProposal.draft
		const images = catalogProposal.images ?? []

		setOriginalImages(
			images.map((img) => ({
				id: img.id,
				weed_id: img.pending_weed_id,
				url: img.url,
				is_primary: img.is_primary,
			}))
		)

		setItemState((prev) => ({
			...prev,
			name: draft.name ?? "",
			description: draft.description ?? "",
			harmfulness: draft.harmfulness ?? "",
			photos: images.map((img) => img.url),
			excludedObjects: catalogProposal.statistics?.excluded_objects ?? [],
		}))

		// Draft classification fields may be hierarchy codes or display names.
		let primaryCode: string | null = null
		let secondaryCode: string | null = null
		let tertiaryCode: string | null = null

		if (classifications) {
			const mg = draft.main_group ?? ""
			const msg = draft.main_subgroup ?? ""
			const sg = draft.subgroup ?? ""
			primaryCode =
				classifications.main_groups[mg] !== undefined
					? mg
					: (Object.entries(classifications.main_groups).find(([, name]) => name === mg)?.[0] ??
						null)
			secondaryCode =
				classifications.main_subgroups[msg] !== undefined
					? msg
					: (Object.entries(classifications.main_subgroups).find(([, name]) => name === msg)?.[0] ??
						null)
			tertiaryCode =
				classifications.subgroups[sg] !== undefined
					? sg
					: (Object.entries(classifications.subgroups).find(([, name]) => name === sg)?.[0] ?? null)
		}

		setSelectedPrimary(primaryCode)
		setSelectedSecondary(secondaryCode)
		setSelectedTertiary(tertiaryCode)
		setFormReady(true)
	}, [catalogProposal, classifications, setItemState])

	useEffect(() => {
		if (catalogProposal) {
			setReviewNote((prev) => (prev ? prev : (catalogProposal.review_notes ?? "")))
		}
	}, [catalogProposal])

	const analysisIds = catalogProposal?.analyses ?? []
	const {
		analyses: analysesData,
		objectsData: initialObjectsData,
		isFetching: isAnalysesFetching,
	} = useAnalysesWithObjects(proposalId, analysisIds, {
		enabled: !!catalogProposal && analysisIds.length > 0,
	})

	useEffect(() => {
		setItemState((prev) => ({
			...prev,
			analyses: analysesData,
			allObjectsData: initialObjectsData,
		}))
	}, [analysesData, initialObjectsData, setItemState])

	const handleCancel = () => {
		navigate({ to: "/catalog/proposals", replace: true })
	}

	const submitProposalStatus = (status: ProposalStatus, notes?: string) => {
		updateProposalMutation.mutate({
			status,
			notes: notes?.trim() || undefined,
		})
	}

	const ensureNote = () => {
		if (!reviewNote.trim()) {
			showError("Добавьте комментарий к решению")
			return false
		}
		return true
	}

	const handleApprove = () => {
		if (!ensureNote()) return
		submitProposalStatus("applied", reviewNote)
	}
	const handleRequestChanges = () => {
		if (!ensureNote()) return
		submitProposalStatus("changes_requested", reviewNote)
	}
	const handleReject = () => {
		if (!ensureNote()) return
		submitProposalStatus("rejected", reviewNote)
	}

	const isProcessing =
		updateProposalMutation.isPending ||
		saveDraftAndResubmitMutation.isPending ||
		saveDraftOnlyMutation.isPending
	const closeDecisionModal = () => setIsDecisionModalOpen(false)
	const disableApprove =
		isProcessing ||
		!itemState.name.trim() ||
		!selectedPrimary ||
		!selectedSecondary ||
		!selectedTertiary ||
		!reviewNote.trim()
	const disableNoteActions = isProcessing || !reviewNote.trim()
	const disableSendToRecheck =
		isProcessing || !itemState.name.trim() || !selectedPrimary || !selectedSecondary

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
						readOnly={isSubmittedReadOnly}
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
						readOnly={isSubmittedReadOnly}
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
						readOnly={isSubmittedReadOnly}
					/>
				)
			case Tab.PHOTOS:
				return (
					<PhotosTab
						photos={itemState.photos}
						onPhotosChange={setPhotos}
						readOnly={isSubmittedReadOnly}
					/>
				)
		}
	}

	const tabs = [
		{ id: Tab.GENERAL, label: Tab.GENERAL, icon: <FaRegFileAlt /> },
		{ id: Tab.CHARACTERISTICS, label: Tab.CHARACTERISTICS, icon: <IoSettingsOutline /> },
		{ id: Tab.ANALYTICS, label: Tab.ANALYTICS, icon: <MdOutlineAnalytics /> },
		{ id: Tab.PHOTOS, label: Tab.PHOTOS, icon: <IoImageOutline /> },
	]

	const isPageLoading = !formReady || isProposalLoading

	if (isPageLoading) {
		return (
			<div className="flex min-h-0 flex-1 items-center justify-center bg-base-100">
				<Loading />
			</div>
		)
	}

	return (
		<>
			<CatalogItemFormLayout
				title={itemState.name.trim() || `Заявка #${proposalId}`}
				subtitle={
					isSubmittedReadOnly
						? "На рассмотрении"
						: isModerator
							? "Проверка заявки"
							: "Редактирование черновика"
				}
				onBack={handleCancel}
				backButtonTitle="К списку заявок"
				tabs={tabs}
				activeTabId={activeTab}
				onTabChange={(id) => {
					setActiveTab(id as Tab)
					resetScroll()
				}}
				banner={
					isSubmittedReadOnly ? (
						<div className="mx-auto w-full max-w-lg px-2 pt-3 sm:px-3">
							<p className="rounded-box border border-info/30 bg-info/10 px-3 py-2 text-sm text-base-content">
								Заявка на рассмотрении. Редактирование недоступно до решения модератора.
							</p>
						</div>
					) : null
				}
				footerActions={
					canSendToRecheck ? (
						<div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
							<button
								type="button"
								className="btn btn-primary w-full"
								onClick={() => saveDraftAndResubmitMutation.mutate()}
								disabled={disableSendToRecheck}
							>
								На повторную проверку
							</button>
							<button
								type="button"
								className="btn btn-outline btn-primary w-full sm:col-span-2"
								onClick={() => saveDraftOnlyMutation.mutate()}
								disabled={disableSendToRecheck}
							>
								Сохранить черновик
							</button>
						</div>
					) : isModerator && !canSendToRecheck ? (
						<button
							type="button"
							className="btn btn-primary w-full"
							onClick={() => setIsDecisionModalOpen(true)}
							disabled={isProcessing}
						>
							Принять решение
						</button>
					) : undefined
				}
			>
				{renderTabContent()}
			</CatalogItemFormLayout>
			{isModerator && (
				<CatalogProposalDecisionSheet
					isOpen={isDecisionModalOpen}
					reviewNote={reviewNote}
					onReviewNoteChange={setReviewNote}
					onClose={closeDecisionModal}
					onApprove={handleApprove}
					onRequestChanges={handleRequestChanges}
					onReject={handleReject}
					onCancel={handleCancel}
					disableApprove={disableApprove}
					disableNoteActions={disableNoteActions}
					isProcessing={isProcessing}
				/>
			)}
		</>
	)
}
