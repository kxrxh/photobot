import { useQuery, useQueryClient } from "@tanstack/react-query"
import type React from "react"
import { useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import {
	createClassification,
	getClassification,
	getClassifications,
	updateClassification,
} from "@/api/classification"
import type {
	Classification,
	SaveCompleteClassificationRequest,
	SaveParamRequest,
} from "@/api/classification/types"
import { calculateCorrelation } from "@/api/correlation"
import type { CorrelationRequest, ParameterGroup } from "@/api/correlation/types"
import { queryKeys } from "@/api/queryKeys"
import LoadingSkeleton from "@/components/common/ui/LoadingSkeleton"
import SearchAndCreateBar from "@/components/common/ui/SearchAndCreateBar"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import { useAlert } from "@/hooks/useAlert"
import type { FractionType } from "@/hooks/useFractions"
import { convertAttribute } from "@/utils/convertAttributes"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import ClassificationCard from "./ClassificationCard"
import CreateForm from "./CreateForm"
import { useDialogState, useParameterToggle } from "./dialogHooks"

interface ClassificationSaveSheetProps {
	isOpen: boolean
	onClose: () => void
	fractions: FractionType[]
	currentClassification?: Classification
}

const ClassificationSaveSheet: React.FC<ClassificationSaveSheetProps> = ({
	isOpen,
	onClose,
	fractions,
}) => {
	const dialogRef = useRef<HTMLDialogElement>(null)
	const classificationsListRef = useRef<HTMLDivElement>(null)
	const queryClient = useQueryClient()
	const { showSuccess, showError } = useAlert()
	const {
		searchTerm,
		setSearchTerm,
		selectedClassification,
		setSelectedClassification,
		showCreateForm,
		setShowCreateForm,
		newClassification,
		setNewClassification,
		expandedCards,
		setExpandedCards,
		selectedParams,
		setSelectedParams,
	} = useDialogState(isOpen, [])

	const { toggleParameter } = useParameterToggle(setSelectedParams)

	const [debouncedSearchTerm, setDebouncedSearchTerm] = useState(searchTerm)
	const [isCalculatingCorrelation, setIsCalculatingCorrelation] = useState(false)

	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedSearchTerm(searchTerm)
		}, 300)
		return () => {
			clearTimeout(handler)
		}
	}, [searchTerm])

	const { data: classificationsData, isLoading: loadingClassifications } = useQuery({
		queryKey: queryKeys.classifications.search(debouncedSearchTerm),
		queryFn: () =>
			getClassifications(debouncedSearchTerm ? { name: debouncedSearchTerm } : undefined),
		enabled: isOpen,
		staleTime: 5 * 60_000,
		gcTime: 15 * 60_000,
	})
	const classifications: Classification[] = classificationsData?.classifications || []

	useEffect(() => {
		if (dialogRef.current) {
			if (isOpen) {
				dialogRef.current.showModal()
				if (classificationsListRef.current) {
					classificationsListRef.current.scrollTop = 0
				}
			} else {
				dialogRef.current.close()
			}
		}
	}, [isOpen])

	useEffect(() => {
		const dialog = dialogRef.current
		if (!dialog) return

		const handleDialogClose = () => {
			onClose()
		}
		dialog.addEventListener("close", handleDialogClose)

		return () => {
			dialog.removeEventListener("close", handleDialogClose)
		}
	}, [onClose])

	const toggleCardExpansion = (classificationId: string) => {
		setExpandedCards((prev) => {
			const newSet = new Set(prev)
			if (newSet.has(classificationId)) {
				newSet.delete(classificationId)
			} else {
				newSet.add(classificationId)
			}
			return newSet
		})
	}

	const handleCreateNew = () => {
		setShowCreateForm(true)
		setSelectedClassification(null)
		setNewClassification({ name: "", product: undefined })
	}

	const handleCancelCreate = () => {
		setShowCreateForm(false)
		setNewClassification({
			name: "",
			product: undefined,
		})
	}

	const handleReplaceClassification = async (classification: Classification) => {
		const parameterGroups = Array.from(selectedParams[classification.id] || []).filter(
			(p): p is ParameterGroup => ["color", "geometry", "median", "all"].includes(p)
		)
		if (parameterGroups.length === 0) {
			showError("Пожалуйста, выберите хотя бы один параметр для сохранения")
			return
		}
		try {
			setIsCalculatingCorrelation(true)
			const complete = await getClassification(String(classification.id))
			const fullClassification = complete.classification
			const parameterGroups = Array.from(selectedParams[classification.id] || []).filter(
				(p): p is ParameterGroup => ["color", "geometry", "median", "all"].includes(p)
			)

			const correlationRequest: CorrelationRequest = {
				fractions: fractions
					.filter((fraction) => fraction.id !== "0")
					.map((fraction) => ({
						name: fraction.name,
						object_ids: fraction.objects.map((obj) => obj.id),
					})),
				parameter_groups: parameterGroups,
			}
			const correlationFractions = await calculateCorrelation(correlationRequest)

			const saveRequest: SaveCompleteClassificationRequest = {
				name: fullClassification.name,
				is_public: fullClassification.is_public,
				product: fullClassification.product,
				fractions: correlationFractions.map((fraction, index) => ({
					name: fraction.name,
					order_index: index,
					conditions: fraction.conditions.map((condition, conditionIndex) => ({
						params: [
							{
								name: convertAttribute(condition.attribute),
								operator: condition.operator as SaveParamRequest["operator"],
								value: condition.value,
							},
						],
						name: `Новое условие ${conditionIndex + 1}`,
						operator: "AND",
						connection: "AND",
						order_index: conditionIndex,
					})),
				})),
			}

			await updateClassification(String(fullClassification.id), saveRequest)
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Классификация успешно обновлена")
			onClose()
		} catch (error) {
			log.error("Failed to update or correlate classification:", error)
			showError(getUserFacingErrorMessage(error))
		} finally {
			setIsCalculatingCorrelation(false)
		}
	}

	const handleSaveNew = async () => {
		if (newClassification.name.trim().length < 3) {
			showError("Название классификации должно быть не менее 3 символов")
			return
		}
		if (!newClassification.product) {
			showError("Пожалуйста, выберите продукт")
			return
		}
		try {
			setIsCalculatingCorrelation(true)
			const parameterGroups: ParameterGroup[] = Array.from(
				newClassification.selectedParams || []
			).filter((p): p is ParameterGroup =>
				["color", "geometry", "median", "all"].includes(p as ParameterGroup)
			)

			const correlationRequest: CorrelationRequest = {
				fractions: fractions
					.filter((fraction) => fraction.id !== "0")
					.map((fraction) => ({
						name: fraction.name,
						object_ids: fraction.objects.map((obj) => obj.id),
					})),
				parameter_groups: parameterGroups,
			}
			const correlationFractions = await calculateCorrelation(correlationRequest)

			const saveRequest: SaveCompleteClassificationRequest = {
				name: newClassification.name,
				is_public: false,
				product: newClassification.product,
				fractions: [],
			}

			for (const [index, fraction] of correlationFractions.entries()) {
				saveRequest.fractions.push({
					name: fraction.name,
					order_index: index,
					conditions: fraction.conditions.map((condition, conditionIndex) => ({
						params: [
							{
								name: convertAttribute(condition.attribute),
								operator: condition.operator as SaveParamRequest["operator"],
								value: condition.value,
							},
						],
						name: `Новое условие ${conditionIndex + 1}`,
						operator: "AND",
						connection: "AND",
						order_index: conditionIndex,
					})),
				})
			}

			await createClassification(saveRequest)
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Классификация успешно создана")
			onClose()
		} catch (error) {
			log.error("Failed to create or correlate classification:", error)
			showError(getUserFacingErrorMessage(error))
		} finally {
			setIsCalculatingCorrelation(false)
		}
	}

	return createPortal(
		<dialog ref={dialogRef} className="modal">
			<div className="flex fixed inset-0 z-50 justify-center items-center">
				<div className="flex relative flex-col w-full h-full bg-base-100">
					<div className="flex sticky top-0 z-10 justify-between items-center p-2 border-b bg-base-100 border-base-200">
						<h2 className="text-xl font-bold">
							{showCreateForm ? "Новая классификация" : "Заменить классификацию"}
						</h2>
						<div className="flex gap-4 items-center">
							<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
						</div>
					</div>

					{showCreateForm ? (
						<CreateForm
							form={newClassification}
							onFormChange={setNewClassification}
							onSave={handleSaveNew}
							onCancel={handleCancelCreate}
							isLoading={isCalculatingCorrelation}
						/>
					) : (
						<>
							<SearchAndCreateBar
								searchTerm={searchTerm}
								onSearchChange={setSearchTerm}
								onCreateNew={handleCreateNew}
								searchPlaceholder="Поиск классификации..."
							/>

							<div className="overflow-y-auto flex-1 p-2 space-y-4" ref={classificationsListRef}>
								{loadingClassifications ? (
									<LoadingSkeleton itemCount={5} />
								) : (
									<>
										{classifications
											.sort((a, b) => {
												if (a.id === selectedClassification?.id) return -1
												if (b.id === selectedClassification?.id) return 1
												return a.is_public === b.is_public ? 0 : a.is_public ? -1 : 1
											})
											.map((classification) => (
												<ClassificationCard
													key={classification.id}
													classification={classification}
													isSelected={selectedClassification?.id === classification.id}
													isExpanded={expandedCards.has(classification.id)}
													selectedParams={selectedParams[classification.id] || new Set()}
													isLoading={isCalculatingCorrelation}
													onToggleExpansion={() => toggleCardExpansion(classification.id)}
													onToggleParameter={(paramId) =>
														toggleParameter(classification.id, paramId)
													}
													onReplace={() => handleReplaceClassification(classification)}
												/>
											))}

										{classifications.length === 0 && (
											<div className="p-8 text-center text-base-content/70">
												Нет классификаций, соответствующих поиску.
											</div>
										)}
									</>
								)}
							</div>
						</>
					)}
				</div>
			</div>
		</dialog>,
		document.body
	)
}

export default ClassificationSaveSheet
