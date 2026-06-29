import { useQuery } from "@tanstack/react-query"
import { AnimatePresence, motion } from "framer-motion"
import { useEffect, useId, useMemo, useState } from "react"
import { FiArrowDown, FiArrowLeft, FiArrowUp, FiEdit3, FiPlus, FiTrash2 } from "react-icons/fi"
import { PiEmpty } from "react-icons/pi"
import { createClassification, getClassification, updateClassification } from "@/api/classification"
import type {
	Classification,
	Condition,
	Fraction,
	Param,
	Product,
	SaveCompleteClassificationRequest,
	SaveFractionRequest,
} from "@/api/classification/types"
import { getParams } from "@/api/params"
import type { ClassificationParam } from "@/api/params/types"
import { getAllProducts } from "@/api/product"
import { queryKeys } from "@/api/queryKeys"
import SaveClassificationSheet from "@/components/classification/dialogs/SaveClassificationSheet"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import EditFractionAlert from "@/components/markup/dialogs/EditFractionAlert"
import RemoveFractionAlert from "@/components/markup/dialogs/RemoveFractionAlert"
import { PARAMETER_GROUPS } from "@/constants"
import { useAlert } from "@/hooks/useAlert"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"

function mapFractionsToSaveRequest(fractions: Fraction[]): SaveFractionRequest[] {
	return fractions.map((fraction, fractionIndex) => ({
		name: fraction.name,
		order_index: fractionIndex,
		conditions: fraction.conditions.map((condition, conditionIndex) => ({
			name: condition.name,
			operator: condition.operator as "<" | "<=" | "==" | ">=" | ">" | "!=",
			connection: condition.connection,
			order_index: conditionIndex,
			params: condition.params.map((param) => ({
				name: param.name,
				operator: param.operator as "<" | "<=" | "==" | ">=" | ">" | "!=",
				value: param.value ?? 0,
			})),
		})),
	}))
}

interface ClassificationConstructorProps {
	initialData?: Classification
	isCopyMode?: boolean
	onSaveSuccess?: () => void
	fractionsData?: Fraction[]
}

function ClassificationConstructor({
	initialData,
	isCopyMode = false,
	onSaveSuccess,
	fractionsData,
}: ClassificationConstructorProps) {
	const { showSuccess, showError } = useAlert()

	const paramsQuery = useQuery({
		queryKey: queryKeys.classificationParams,
		queryFn: getParams,
		staleTime: 1000 * 60 * 15,
		gcTime: 1000 * 60 * 15,
	})

	const parameterOptions: string[] = paramsQuery.data?.map((p: ClassificationParam) => p.name) ?? []

	const groupedParameterOptions = PARAMETER_GROUPS.map((group) => ({
		label: group.label,
		options: group.options.filter((opt) => parameterOptions.includes(opt.value)),
	})).filter((group) => group.options.length > 0)

	const productSelectId = useId()
	const paramSelectPrefix = useId()
	const operatorSelectPrefix = useId()

	const classificationId = initialData?.id
	const {
		data: completeData,
		isLoading: isClassificationLoading,
		isError: isClassificationError,
	} = useQuery({
		queryKey: queryKeys.completeClassification(classificationId ?? ""),
		queryFn: () =>
			classificationId ? getClassification(classificationId) : Promise.resolve(undefined),
		enabled: !!classificationId && !isCopyMode,
	})

	const [fractions, setFractions] = useState<Fraction[]>([])
	const [selectedProduct, setSelectedProduct] = useState<Product | undefined>(initialData?.product)
	const [isEditModalOpen, setIsEditModalOpen] = useState(false)
	const [currentEditingFractionName, setCurrentEditingFractionName] = useState("")
	const [fractionIdToEdit, setFractionIdToEdit] = useState<string | null>(null)

	const [isRemoveFractionModalOpen, setIsRemoveFractionModalOpen] = useState(false)
	const [fractionToDelete, setFractionToDelete] = useState<Fraction | null>(null)

	const [isSaveModalOpen, setIsSaveModalOpen] = useState(false)
	const [classificationName, setClassificationName] = useState<string>(
		isCopyMode ? `${initialData?.name || ""} (копия)` : initialData?.name || ""
	)
	const [displayTitle, setDisplayTitle] = useState<string>(
		isCopyMode ? `${initialData?.name || ""} (копия)` : initialData?.name || ""
	)
	const [isEditMode, setIsEditMode] = useState<boolean>(!!initialData && !isCopyMode)

	useEffect(() => {
		if (completeData && !isCopyMode) {
			setFractions(completeData.fractions || [])
			setSelectedProduct(completeData.classification.product)
			setClassificationName(completeData.classification.name)
			setDisplayTitle(completeData.classification.name)
			setIsEditMode(true)
		}
	}, [completeData, isCopyMode])

	useEffect(() => {
		if (isCopyMode && initialData && fractionsData) {
			const sanitizeFractions = (items: Fraction[]) =>
				items.map(({ id: _fid, conditions, ...fractionRest }) => ({
					...fractionRest,
					id: `f-${crypto.randomUUID()}`,
					conditions: conditions.map(({ id: _cid, params, ...condRest }: Condition) => ({
						...condRest,
						id: `g-${crypto.randomUUID()}`,
						params: params.map(({ id: _pid, ...paramRest }: Param) => ({
							...paramRest,
							id: `c-${crypto.randomUUID()}`,
						})),
					})),
				}))
			setFractions(sanitizeFractions(fractionsData))
			setSelectedProduct(initialData.product)
			setClassificationName(`${initialData.name || ""} (копия)`)
			setDisplayTitle(`${initialData.name || ""} (копия)`)
			setIsEditMode(false)
		}
	}, [isCopyMode, initialData, fractionsData])

	const createNewParam = (): Param => ({
		id: `c-${crypto.randomUUID()}`,
		name: groupedParameterOptions[0]?.options[0]?.value ?? parameterOptions[0] ?? "",
		operator: "<",
		value: 0,
	})

	const createNewCondition = (): Condition => ({
		id: `g-${crypto.randomUUID()}`,
		name: "Новое условие",
		params: [createNewParam()],
		operator: "AND",
		connection: "AND",
		order_index: 0,
	})

	const addFraction = () => {
		const defaultName = `Fraction ${fractions.length + 1}`

		const newFraction: Fraction = {
			id: `f-${crypto.randomUUID()}`,
			name: defaultName,
			conditions: [createNewCondition()],
			order_index: 0,
		}

		setFractions([...fractions, newFraction])
	}

	const deleteFraction = (fractionId: string) => {
		setFractions(fractions.filter((f) => f.id !== fractionId))
	}

	const updateFraction = (fractionId: string, updates: Partial<Fraction>) => {
		setFractions(fractions.map((f) => (f.id === fractionId ? { ...f, ...updates } : f)))
	}

	const moveFractionUp = (fractionId: string) => {
		setFractions((currentFractions) => {
			const index = currentFractions.findIndex((f) => f.id === fractionId)
			if (index <= 0) return currentFractions

			const newFractions = [...currentFractions]
			;[newFractions[index - 1], newFractions[index]] = [
				newFractions[index],
				newFractions[index - 1],
			]
			return newFractions
		})
	}

	const moveFractionDown = (fractionId: string) => {
		setFractions((currentFractions) => {
			const index = currentFractions.findIndex((f) => f.id === fractionId)
			if (index === -1 || index >= currentFractions.length - 1) return currentFractions // Cannot move down if it's the last or doesn't exist

			const newFractions = [...currentFractions]
			;[newFractions[index + 1], newFractions[index]] = [
				newFractions[index],
				newFractions[index + 1],
			]
			return newFractions
		})
	}

	const addConditionGroup = (fractionId: string) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		if (!fraction) return

		const newCondition = createNewCondition()
		updateFraction(fractionId, {
			conditions: [...fraction.conditions, newCondition],
		})
	}

	const deleteConditionGroup = (fractionId: string, groupId: string) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		if (!fraction) return

		updateFraction(fractionId, {
			conditions: fraction.conditions.filter((g) => g.id !== groupId),
		})
	}

	const updateConditionGroup = (
		fractionId: string,
		groupId: string,
		updates: Partial<Condition>
	) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		if (!fraction) return

		updateFraction(fractionId, {
			conditions: fraction.conditions.map((g) => (g.id === groupId ? { ...g, ...updates } : g)),
		})
	}

	const addParam = (fractionId: string, groupId: string) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		const group = fraction?.conditions.find((g) => g.id === groupId)
		if (!group) return

		updateConditionGroup(fractionId, groupId, {
			params: [...group.params, createNewParam()],
		})
	}

	const deleteParam = (fractionId: string, groupId: string, paramId: string) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		const group = fraction?.conditions.find((g) => g.id === groupId)
		if (!group) return

		updateConditionGroup(fractionId, groupId, {
			params: group.params.filter((p) => p.id !== paramId),
		})
	}

	const updateParam = (
		fractionId: string,
		groupId: string,
		paramId: string,
		paramUpdates: Partial<Param>
	) => {
		const fraction = fractions.find((f) => f.id === fractionId)
		const group = fraction?.conditions.find((g) => g.id === groupId)
		if (!group) return

		updateConditionGroup(fractionId, groupId, {
			params: group.params.map((p) => (p.id === paramId ? { ...p, ...paramUpdates } : p)),
		})
	}

	const handleOpenEditModal = (fraction: Fraction) => {
		setFractionIdToEdit(fraction.id)
		setCurrentEditingFractionName(fraction.name)
		setIsEditModalOpen(true)
	}

	const handleConfirmEditModal = () => {
		if (fractionIdToEdit) {
			updateFraction(fractionIdToEdit, { name: currentEditingFractionName })
		}
		setIsEditModalOpen(false)
		setFractionIdToEdit(null)
	}

	// Validation function to check if save should be enabled
	const isSaveEnabled = () => {
		// Check if product is selected
		if (!selectedProduct) return false

		// Check if there's at least one fraction
		if (fractions.length === 0) return false

		// Check that ALL conditions in ALL fractions are properly filled
		const allConditionsFilled = fractions.every((fraction) =>
			fraction.conditions.every((condition) =>
				condition.params.every(
					(param) =>
						param.name &&
						condition.operator &&
						param.value !== undefined &&
						param.value !== null &&
						!Number.isNaN(param.value)
				)
			)
		)

		return allConditionsFilled
	}

	const handleSave = () => {
		setIsSaveModalOpen(true)
	}

	const handleConfirmSave = async () => {
		try {
			if (!selectedProduct) {
				showError("Необходимо выбрать продукт")
				return
			}
			if (isEditMode && initialData?.id) {
				const updateData: SaveCompleteClassificationRequest = {
					name: classificationName,
					product: selectedProduct,
					fractions: mapFractionsToSaveRequest(fractions),
					is_public: false,
				}

				await updateClassification(initialData.id, updateData)
				showSuccess("Классификация успешно обновлена!")
			} else {
				// Create new classification (both for new and copy mode)
				const createData: SaveCompleteClassificationRequest = {
					name: classificationName,
					product: selectedProduct,
					fractions: mapFractionsToSaveRequest(fractions),
					is_public: false,
				}

				await createClassification(createData)

				if (isCopyMode) {
					showSuccess("Классификация успешно скопирована!")
				} else {
					showSuccess("Классификация успешно создана!")
				}

				// Switch to edit mode after successful creation
				setIsEditMode(true)
			}

			// Update the display title with the saved name
			setDisplayTitle(classificationName)
			setIsSaveModalOpen(false)

			if (onSaveSuccess) onSaveSuccess()
		} catch (error) {
			log.devError("Error saving classification:", error)
			showError(getUserFacingErrorMessage(error))
		}
	}

	const handleDeleteFraction = (fraction: Fraction) => {
		setFractionToDelete(fraction)
		setIsRemoveFractionModalOpen(true)
	}

	const confirmDeleteFraction = () => {
		if (fractionToDelete) {
			deleteFraction(fractionToDelete.id)
		}
		setIsRemoveFractionModalOpen(false)
		setFractionToDelete(null)
	}

	const cancelDeleteFraction = () => {
		setIsRemoveFractionModalOpen(false)
		setFractionToDelete(null)
	}

	const productsQuery = useQuery({
		queryKey: queryKeys.products,
		queryFn: getAllProducts,
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 15, // 15 minutes
	})
	const productOptions = useMemo(
		() =>
			(productsQuery.data ?? [])
				.map((option) => ({
					value: String(option.id),
					label: option.name,
				}))
				.sort((left, right) => left.label.localeCompare(right.label, "ru")),
		[productsQuery.data]
	)

	// Показываем лоадер, если идёт загрузка полной классификации
	if (isClassificationLoading && !isCopyMode) {
		return (
			<div className="p-8 text-center">
				<span className="loading loading-lg loading-spinner" />
			</div>
		)
	}
	if (isClassificationError && !isCopyMode) {
		return <div className="p-8 text-center text-error">Ошибка при загрузке классификации</div>
	}

	return (
		<>
			<header className="sticky top-0 z-50 px-2 py-2 space-y-2 w-full bg-base-100">
				<div className="flex justify-between items-center">
					<h1 className="text-2xl font-bold truncate">
						{isEditMode
							? displayTitle || "Классификация"
							: isCopyMode
								? "Копирование классификации"
								: "Новая классификация"}
					</h1>
					<button
						type="button"
						className="btn btn-ghost btn-circle"
						onClick={() => window.history.back()}
					>
						<FiArrowLeft className="w-5 h-5 font-bold" />
					</button>
				</div>

				<div className="form-control">
					<ModalSelect
						id={productSelectId}
						title="Продукт"
						placeholder="Выберите продукт"
						options={productOptions}
						value={selectedProduct?.id != null ? String(selectedProduct.id) : ""}
						onChange={(v) =>
							setSelectedProduct(
								v === "" ? undefined : productsQuery.data?.find((p) => String(p.id) === v)
							)
						}
						disabled={productsQuery.isLoading}
						size="sm"
					/>
				</div>
			</header>

			<div className="p-2 pb-16 space-y-4 w-full">
				<div className="space-y-3">
					<AnimatePresence>
						{fractions.map((fraction, fractionIndex) => (
							<motion.div
								key={fraction.id}
								layoutId={fraction.id}
								layout
								initial={{ opacity: 0, y: 20 }}
								animate={{ opacity: 1, y: 0 }}
								exit={{ opacity: 0, y: -20 }}
								transition={{ duration: 0.2 }}
								className="border collapse collapse-arrow border-base-200 bg-base-100"
							>
								<input type="checkbox" defaultChecked className="peer" />
								<div className="collapse-title pr-12! min-w-0">
									<div className="flex justify-between items-center min-w-0">
										<div className="flex-1 mr-4 min-w-0">
											<span className="block font-semibold truncate">{fraction.name}</span>
										</div>

										<div className="flex z-20 shrink-0 gap-1 items-center">
											<button
												type="button"
												onClick={(e) => {
													e.preventDefault()
													e.stopPropagation()
													moveFractionUp(fraction.id)
												}}
												disabled={fractionIndex === 0}
												className="btn btn-ghost btn-sm btn-circle"
											>
												<FiArrowUp className="w-4 h-4" />
											</button>
											<button
												type="button"
												onClick={(e) => {
													e.preventDefault()
													e.stopPropagation()
													moveFractionDown(fraction.id)
												}}
												disabled={fractionIndex === fractions.length - 1}
												className="btn btn-ghost btn-sm btn-circle"
											>
												<FiArrowDown className="w-4 h-4" />
											</button>
											<button
												type="button"
												onClick={(e) => {
													e.preventDefault()
													e.stopPropagation()
													handleOpenEditModal(fraction)
												}}
												className="btn btn-ghost btn-sm btn-circle"
											>
												<FiEdit3 className="w-4 h-4" />
											</button>
											<button
												type="button"
												onClick={(e) => {
													e.preventDefault()
													e.stopPropagation()
													handleDeleteFraction(fraction)
												}}
												className="btn btn-ghost btn-sm btn-circle"
											>
												<FiTrash2 className="w-4 h-4" />
											</button>
										</div>
									</div>
								</div>

								<div className="collapse-content">
									<div className="space-y-2">
										{fraction.conditions.map((condition, groupIndex) => (
											<div key={condition.id}>
												{groupIndex > 0 && (
													<div className="flex justify-center my-2">
														<button
															type="button"
															className="w-full btn btn-sm"
															onClick={() =>
																updateConditionGroup(fraction.id, condition.id, {
																	connection: condition.connection === "AND" ? "OR" : "AND",
																})
															}
														>
															{condition.connection === "AND" ? "И" : "ИЛИ"}
														</button>
													</div>
												)}
												<div className="rounded-md border border-base-300">
													<div className="p-2 rounded-t-sm border-b border-base-300 bg-base-200">
														<div className="flex gap-4 justify-between items-center">
															<span className="text-sm font-medium text-base-content">
																{condition.name}
															</span>
															<div className="flex gap-2 items-center">
																<label className="swap swap-flip">
																	<input
																		type="checkbox"
																		checked={condition.operator === "OR"}
																		onChange={() =>
																			updateConditionGroup(fraction.id, condition.id, {
																				operator: condition.operator === "AND" ? "OR" : "AND",
																			})
																		}
																		title="И/ИЛИ"
																	/>
																	<div className="swap-on btn-primary btn btn-sm" title="ИЛИ">
																		ИЛИ
																	</div>
																	<div className="swap-off btn-primary btn btn-sm" title="И">
																		И
																	</div>
																</label>
																<button
																	type="button"
																	onClick={() => deleteConditionGroup(fraction.id, condition.id)}
																	className={`btn btn-xs btn-ghost btn-circle ${
																		fraction.conditions.length === 1
																			? "text-base-content/30 cursor-not-allowed"
																			: "text-primary hover:text-primary-focus"
																	}`}
																	disabled={fraction.conditions.length === 1}
																	title={
																		fraction.conditions.length === 1
																			? "Нельзя удалить последнюю группу"
																			: "Удалить группу"
																	}
																>
																	<FiTrash2 className="w-4 h-4" />
																</button>
															</div>
														</div>
													</div>

													<div className="flex items-start p-2 space-x-2">
														<div className="grow space-y-2">
															{condition.params.map((param, _) => (
																<div key={param.id} className="space-y-2">
																	<div className="flex items-center p-2 space-x-2 rounded-xl bg-base-100 border-base-200">
																		<button
																			type="button"
																			onClick={() =>
																				deleteParam(fraction.id, condition.id, param.id)
																			}
																			className="text-primary btn btn-xs btn-ghost btn-circle"
																			disabled={condition.params.length === 1}
																		>
																			<FiTrash2 className="w-4 h-4" />
																		</button>

																		<div className="w-1/3 min-w-0">
																			{groupedParameterOptions.length > 0 ? (
																				<ModalSelect
																					id={`${paramSelectPrefix}-${param.id}`}
																					title="Параметр"
																					placeholder="Параметр"
																					groupedOptions={groupedParameterOptions.map((group) => ({
																						label: group.label,
																						options: group.options.map((opt) => ({
																							value: opt.value,
																							label: `${opt.value} - ${opt.label}`,
																						})),
																					}))}
																					value={param.name}
																					onChange={(v) =>
																						updateParam(fraction.id, condition.id, param.id, {
																							name: v,
																						})
																					}
																					clearable={false}
																					size="sm"
																				/>
																			) : (
																				<ModalSelect
																					id={`${paramSelectPrefix}-${param.id}`}
																					title="Параметр"
																					placeholder="Параметр"
																					options={parameterOptions.map((name) => ({
																						value: name,
																						label: name,
																					}))}
																					value={param.name}
																					onChange={(v) =>
																						updateParam(fraction.id, condition.id, param.id, {
																							name: v,
																						})
																					}
																					clearable={false}
																					size="sm"
																				/>
																			)}
																		</div>

																		<div className="min-w-0 flex-1">
																			<ModalSelect
																				id={`${operatorSelectPrefix}-${param.id}`}
																				title="Оператор"
																				placeholder="Оператор"
																				options={(["<", "<=", "==", ">=", ">", "!="] as const).map(
																					(op) => ({ value: op, label: op })
																				)}
																				value={param.operator}
																				onChange={(v) =>
																					updateParam(fraction.id, condition.id, param.id, {
																						operator: v,
																					})
																				}
																				clearable={false}
																				size="sm"
																			/>
																		</div>
																		<input
																			type="number"
																			step="0.01"
																			className="flex-1 input input-sm input-bordered [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
																			placeholder="0.00"
																			value={param.value ?? ""}
																			onKeyDown={(e) => {
																				// Allow only digits, comma, dot, minus and control keys
																				if (e.key.length === 1 && !/[0-9,.-]/.test(e.key)) {
																					e.preventDefault()
																				}
																			}}
																			onChange={(e) =>
																				updateParam(fraction.id, condition.id, param.id, {
																					value:
																						e.target.value === ""
																							? null
																							: Number(e.target.value.replace(/,/g, ".")),
																				})
																			}
																		/>
																	</div>
																</div>
															))}

															<button
																type="button"
																onClick={() => addParam(fraction.id, condition.id)}
																className="w-full btn btn-sm btn-primary"
															>
																<FiPlus className="w-4 h-4" />
																Добавить условие
															</button>
														</div>
													</div>
												</div>
											</div>
										))}

										<button
											type="button"
											onClick={() => addConditionGroup(fraction.id)}
											className="w-full text-base-content btn btn-sm"
										>
											<FiPlus className="w-4 h-4" />
											Добавить группу
										</button>
									</div>
								</div>
							</motion.div>
						))}
					</AnimatePresence>
				</div>

				{fractions.length === 0 && (
					<div className="py-12 text-center text-base-content/70">
						<PiEmpty className="mx-auto mb-4 w-12 h-12 text-base-content/50" />
						<p>Список фракций пуст</p>
						<p className="text-sm">Добавьте первую фракцию используя кнопку внизу экрана</p>
					</div>
				)}
			</div>

			<footer className="fixed right-0 bottom-0 left-0 z-50 p-2 bg-base-100">
				<div className="flex gap-2">
					<button type="button" className="flex-1 btn btn-outline" onClick={addFraction}>
						<FiPlus className="mr-2" />
						Новая фракция
					</button>
					<button
						type="button"
						className="flex-1 btn btn-primary"
						onClick={handleSave}
						disabled={!isSaveEnabled()}
					>
						Сохранить
					</button>
				</div>
			</footer>
			<EditFractionAlert
				isOpen={isEditModalOpen}
				onClose={() => setIsEditModalOpen(false)}
				onConfirm={handleConfirmEditModal}
				newFractionName={currentEditingFractionName}
				setNewFractionName={setCurrentEditingFractionName}
			/>
			<SaveClassificationSheet
				isOpen={isSaveModalOpen}
				onClose={() => setIsSaveModalOpen(false)}
				onConfirm={handleConfirmSave}
				classificationName={classificationName}
				setClassificationName={setClassificationName}
				isEditing={isEditMode}
				isCopying={isCopyMode && !isEditMode}
			/>
			<RemoveFractionAlert
				isOpen={isRemoveFractionModalOpen}
				onClose={cancelDeleteFraction}
				onConfirm={confirmDeleteFraction}
				fractionName={fractionToDelete?.name}
			/>
		</>
	)
}

export default ClassificationConstructor
