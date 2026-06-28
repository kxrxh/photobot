import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { AnimatePresence, motion } from "framer-motion"
import { useEffect, useId, useState } from "react"
import { FaTrash } from "react-icons/fa"
import { MdAdd, MdFilterList } from "react-icons/md"
import {
	deleteClassification,
	deleteUserActiveClassification,
	getClassification,
	getClassifications,
	makeClassificationPrivate,
	makeClassificationPublic,
	setUserActiveClassification,
} from "@/api/classification"
import type { Classification, Fraction } from "@/api/classification/types"
import { getAllProducts } from "@/api/product"
import { queryKeys } from "@/api/queryKeys"
import ClassificationItem from "@/components/classification/components/ClassificationItem"
import DeleteClassificationAlert from "@/components/classification/dialogs/DeleteClassificationAlert"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import { PageType } from "@/routes/_authenticated/classification"
import { getUserFacingErrorMessage } from "@/utils/errors"

interface ClassificationListProps {
	setCurrentPage: (
		page: PageType,
		classification?: Classification,
		isCopyMode?: boolean,
		options?: { onSaveSuccess?: () => void; fractions?: Fraction[] }
	) => void
}

const ClassificationList = ({ setCurrentPage }: ClassificationListProps) => {
	const [searchTerm, setSearchTerm] = useState("")
	const [productFilter, setProductFilter] = useState<string | undefined>(undefined)
	const [showFilters, setShowFilters] = useState(false)
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
	const [classificationToDelete, setClassificationToDelete] = useState<Classification | null>(null)

	const productFilterId = useId()
	const [selectedClassificationId, setSelectedClassificationId] = useState<string | null>(null)

	const { roles } = useAuth()
	const hasGlobalEditPermission = roles.has("admin") || roles.has("classification_editor")

	const classificationsQuery = useQuery({
		queryKey: queryKeys.classifications.search(searchTerm, productFilter),
		queryFn: () =>
			getClassifications({
				name: searchTerm,
				product_id: productFilter,
			}),
		select: (data) => ({
			classifications: data.classifications,
			activeClassification: data.active_classification ?? undefined,
		}),
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 15, // 15 minutes
	})

	const productsQuery = useQuery({
		queryKey: queryKeys.products,
		queryFn: getAllProducts,
		staleTime: 1000 * 60 * 5, // 5 minutes
		gcTime: 1000 * 60 * 15, // 15 minutes
		enabled: showFilters,
	})

	const queryClient = useQueryClient()
	const { showSuccess, showError } = useAlert()

	const deleteMutation = useMutation({
		mutationFn: (id: string) => deleteClassification(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Классификация удалена")
		},
		onError: () => {
			showError("Ошибка при удалении классификации")
		},
	})

	const publicMutation = useMutation({
		mutationFn: (id: string) => makeClassificationPublic(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Классификация опубликована")
		},
		onError: () => {
			showError("Ошибка при публикации классификации")
		},
	})

	const privateMutation = useMutation({
		mutationFn: (id: string) => makeClassificationPrivate(id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Классификация снята с публикации")
		},
		onError: () => {
			showError("Ошибка при снятии с публикации")
		},
	})

	const activateMutation = useMutation({
		mutationFn: (id: string) => setUserActiveClassification(id),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: queryKeys.classifications.all,
			})
			showSuccess("Классификация активирована")
		},
		onError: () => {
			showError("Ошибка при активации классификации")
		},
	})

	const deactivateMutation = useMutation({
		mutationFn: () => deleteUserActiveClassification(),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: queryKeys.classifications.all,
			})
			showSuccess("Классификация деактивирована")
		},
		onError: () => {
			showError("Ошибка при деактивации классификации")
		},
	})

	const clearFilters = () => {
		setProductFilter(undefined)
		setShowFilters(false)
	}

	const classifications = classificationsQuery.data?.classifications ?? []
	const activeClassification = classificationsQuery.data?.activeClassification

	useEffect(() => {
		if (activeClassification) {
			setSelectedClassificationId(activeClassification.id)
		} else {
			setSelectedClassificationId(null)
		}
	}, [activeClassification])

	const handleEdit = (classification: Classification) => {
		setCurrentPage(PageType.CONSTRUCTOR, { id: classification.id } as Classification, false, {
			onSaveSuccess: () => {
				queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
				setCurrentPage(PageType.LIST)
			},
		})
	}

	const handleCopy = async (classification: Classification) => {
		try {
			const response = await getClassification(classification.id)
			setCurrentPage(PageType.COPY, response.classification, true, {
				fractions: response.fractions,
			})
		} catch (err) {
			showError(getUserFacingErrorMessage(err))
		}
	}

	const handleDelete = (classification: Classification) => {
		setClassificationToDelete(classification)
		setIsDeleteModalOpen(true)
	}

	const confirmDelete = () => {
		if (classificationToDelete) {
			deleteMutation.mutate(classificationToDelete.id)
		}
		setIsDeleteModalOpen(false)
		setClassificationToDelete(null)
	}

	const handleCheck = (classification: Classification) => {
		setSelectedClassificationId((prevId) => {
			if (prevId === classification.id) {
				deactivateMutation.mutate()
				return null
			}
			activateMutation.mutate(classification.id)
			return classification.id
		})
	}

	const handleTogglePublic = (classification: Classification) => {
		if (classification.is_public) {
			privateMutation.mutate(classification.id)
		} else {
			publicMutation.mutate(classification.id)
		}
	}

	const handleCreate = () => {
		setCurrentPage(PageType.CONSTRUCTOR, undefined, false, {
			onSaveSuccess: () => {
				queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
				setCurrentPage(PageType.LIST)
			},
		})
	}

	return (
		<>
			<div className="sticky top-0 z-10 border-b shadow-sm bg-base-100 border-base-200">
				<div className="flex justify-between items-center p-2">
					<div className="flex gap-3 items-center w-full">
						<input
							type="text"
							placeholder="Поиск классификации..."
							className="flex-1 input input-bordered input-sm"
							value={searchTerm}
							onChange={(e) => setSearchTerm(e.target.value)}
						/>
						<button
							type="button"
							className={`btn btn-sm btn-square ${showFilters ? "btn-primary" : "btn-ghost"}`}
							onClick={() => setShowFilters(!showFilters)}
							title="Фильтры"
						>
							<MdFilterList className="text-lg" />
						</button>
						<button
							type="button"
							className="flex gap-2 items-center btn btn-primary btn-sm shrink-0"
							onClick={handleCreate}
						>
							<MdAdd className="text-lg" />
							Создать
						</button>
					</div>
				</div>

				<AnimatePresence>
					{showFilters && (
						<motion.div
							initial={{ height: 0, opacity: 0 }}
							animate={{ height: "auto", opacity: 1 }}
							exit={{ height: 0, opacity: 0 }}
							transition={{ duration: 0.2 }}
							className="overflow-hidden border-t border-base-200"
						>
							<div className="flex gap-3 items-center p-2">
								<div className="flex flex-1 gap-2 items-center">
									<label
										htmlFor={productFilterId}
										className="text-sm font-medium text-base-content/70"
									>
										Продукт:
									</label>
									<ModalSelect
										id={productFilterId}
										title="Продукт"
										placeholder="Все продукты"
										options={
											productsQuery.data?.map((product) => ({
												value: String(product.id),
												label: product.name,
											})) ?? []
										}
										value={productFilter ?? ""}
										onChange={(v) => setProductFilter(v === "" ? undefined : v)}
										disabled={productsQuery.isLoading}
										size="sm"
									/>
								</div>
								{productFilter && (
									<button
										type="button"
										className="btn btn-sm btn-ghost text-base-content/70 gap-1"
										onClick={clearFilters}
									>
										<FaTrash className="mr-1" />
										Сбросить
									</button>
								)}
							</div>
						</motion.div>
					)}
				</AnimatePresence>
			</div>

			<div className="p-2 space-y-2">
				{classificationsQuery.isLoading && (
					<div className="p-8 text-center">
						<span className="loading loading-lg loading-spinner" />
					</div>
				)}
				{classificationsQuery.isError && (
					<div className="p-8 text-center text-error">
						Ошибка при загрузке классификаций.{" "}
						{getUserFacingErrorMessage(classificationsQuery.error)}
					</div>
				)}
				{classificationsQuery.isSuccess && (
					<>
						{[...classifications]
							.sort((a, b) => {
								if (a.id === selectedClassificationId) return -1
								if (b.id === selectedClassificationId) return 1
								return a.is_public === b.is_public ? 0 : a.is_public ? -1 : 1
							})
							.map((classification) => (
								<ClassificationItem
									key={classification.id}
									classification={classification}
									hasSuperRights={hasGlobalEditPermission}
									onEdit={handleEdit}
									onCopy={handleCopy}
									onDelete={handleDelete}
									onCheck={handleCheck}
									onTogglePublic={handleTogglePublic}
									isSelected={classification.id === selectedClassificationId}
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
			<DeleteClassificationAlert
				isOpen={isDeleteModalOpen}
				onClose={() => {
					setIsDeleteModalOpen(false)
					setClassificationToDelete(null)
				}}
				onConfirm={confirmDelete}
				classificationName={classificationToDelete?.name}
			/>
		</>
	)
}

export default ClassificationList
