import { useQuery } from "@tanstack/react-query"
import { Link } from "@tanstack/react-router"
import type React from "react"
import { useEffect, useId, useRef, useState } from "react"
import { FaExternalLinkAlt, FaGlobe, FaInfoCircle } from "react-icons/fa"
import { getClassifications } from "@/api/classification"
import type { Classification } from "@/api/classification/types"
import { queryKeys } from "@/api/queryKeys"

interface ClassificationSelectionAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: (classificationId: string) => void
	activeClassification: Classification | null | undefined
	isPending: boolean
}

const ClassificationSelectionAlert: React.FC<ClassificationSelectionAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	activeClassification,
	isPending,
}) => {
	const modalId = useId()
	const modalRef = useRef<HTMLDialogElement>(null)
	const [selectedClassificationId, setSelectedClassificationId] = useState<string>("")

	// Fetch classifications for selection modal
	const { data: classificationsData, isLoading: loadingClassifications } = useQuery({
		queryKey: queryKeys.classifications.all,
		queryFn: () => getClassifications(),
		enabled: isOpen,
	})

	const classifications = classificationsData?.classifications ?? []

	// Pre-select active classification when modal opens or active classification changes
	useEffect(() => {
		if (isOpen && activeClassification) {
			setSelectedClassificationId(activeClassification.id)
		}
	}, [isOpen, activeClassification])

	// Handle dialog open/close
	useEffect(() => {
		if (isOpen) {
			modalRef.current?.showModal()
		} else {
			modalRef.current?.close()
			setSelectedClassificationId("")
		}
	}, [isOpen])

	const handleDialogClose = () => {
		onClose()
		setSelectedClassificationId("")
	}

	const handleConfirm = () => {
		if (selectedClassificationId) {
			onConfirm(selectedClassificationId)
		}
	}

	return (
		<dialog ref={modalRef} id={modalId} className="modal" onClose={handleDialogClose}>
			<div className="modal-box">
				<h3 className="font-bold text-lg mb-4">Выберите активную классификацию</h3>

				{loadingClassifications ? (
					<div className="space-y-2">
						<div className="skeleton h-12 w-full"></div>
						<div className="skeleton h-12 w-full"></div>
						<div className="skeleton h-12 w-full"></div>
					</div>
				) : classifications.length === 0 ? (
					<div className="alert alert-info">
						<p className="text-sm">Нет доступных классификаций</p>
					</div>
				) : (
					<div className="space-y-3 max-h-[50vh] overflow-y-auto pr-2">
						{classifications
							.sort((a, b) => {
								if (activeClassification) {
									if (a.id === activeClassification.id) return -1
									if (b.id === activeClassification.id) return 1
								}
								return 0
							})
							.map((classification: Classification) => {
								const isActive = activeClassification?.id === classification.id
								const isSelected = selectedClassificationId === classification.id

								return (
									<label
										key={classification.id}
										className={`
										card card-border
										transition-all duration-200 ease-in-out
										${
											isActive
												? "bg-success/5 border-success/50 border-2 shadow-sm cursor-not-allowed"
												: isSelected
													? "bg-primary/10 border-primary border-2 shadow-md cursor-pointer hover:shadow-lg"
													: "bg-base-100 border border-base-300 shadow-sm cursor-pointer hover:border-primary/60 hover:shadow-md hover:bg-base-200/50"
										}
									`}
									>
										<div className="card-body p-4">
											<div className="flex items-start gap-4">
												<div className="pt-0.5 shrink-0">
													<input
														type="radio"
														name="classification"
														value={classification.id}
														checked={isSelected}
														onChange={(e) => setSelectedClassificationId(e.target.value)}
														className={`radio ${isActive ? "radio-success" : "radio-primary"}`}
														disabled={isActive}
													/>
												</div>
												<div className="flex-1 min-w-0">
													<div className="flex flex-col gap-1.5 sm:flex-row sm:items-start sm:justify-between sm:gap-3 mb-1.5">
														<div className="flex items-center gap-1.5 flex-1 min-w-0">
															{classification.is_public && (
																<FaGlobe
																	className="w-3.5 h-3.5 shrink-0 text-info mt-0.5"
																	title="Публичная классификация"
																/>
															)}
															<h4 className="font-semibold text-sm sm:text-base leading-tight wrap-break-word">
																{classification.name}
															</h4>
														</div>
														<div className="flex items-center gap-1.5 shrink-0">
															{isActive && (
																<span className="badge badge-success badge-sm">Активна</span>
															)}
														</div>
													</div>
													<div className="flex items-center gap-2 flex-wrap">
														<span className="text-xs text-base-content/60 font-medium shrink-0">
															Продукт:
														</span>
														<span className="text-xs text-base-content/80 wrap-break-word">
															{classification.product.name}
														</span>
													</div>
												</div>
											</div>
										</div>
									</label>
								)
							})}
					</div>
				)}

				<div className="mt-4 pt-3 border-t border-base-300">
					<div className="flex flex-col gap-2">
						<div className="flex items-start gap-2">
							<FaInfoCircle className="w-4 h-4 shrink-0 text-info mt-0.5" />
							<div className="flex-1">
								<p className="text-xs text-base-content/70 leading-relaxed">
									Создать или редактировать классификацию можно в разделе{" "}
									<Link
										to="/classification"
										onClick={handleDialogClose}
										className="link link-info font-medium text-info inline-flex items-center gap-1"
									>
										Классификация
										<FaExternalLinkAlt className="w-3 h-3" />
									</Link>
								</p>
							</div>
						</div>
					</div>
				</div>

				<div className="modal-action">
					<form method="dialog" className="flex flex-row w-full gap-3">
						<button
							type="button"
							onClick={handleDialogClose}
							className="btn flex-1"
							disabled={isPending}
						>
							Отмена
						</button>
						<button
							type="button"
							onClick={handleConfirm}
							className="btn btn-primary flex-1"
							disabled={
								!selectedClassificationId ||
								selectedClassificationId === activeClassification?.id ||
								isPending ||
								loadingClassifications
							}
						>
							{isPending ? <span className="loading loading-spinner loading-sm"></span> : "Выбрать"}
						</button>
					</form>
				</div>
			</div>
			<form
				method="dialog"
				className="modal-backdrop"
				onClick={handleDialogClose}
				onKeyDown={(e) => {
					if (e.key === "Enter" || e.key === " ") {
						e.preventDefault()
						handleDialogClose()
					}
				}}
			>
				<button type="button">close</button>
			</form>
		</dialog>
	)
}

export default ClassificationSelectionAlert
