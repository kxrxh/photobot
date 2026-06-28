import { useEffect, useId, useRef, useState } from "react"
import { FaExclamationTriangle } from "react-icons/fa"

export type ReclassificationAction = "preserve_manual" | "reclassify_all" | "cancel"

export interface ReclassificationModalData {
	manualObjectsCount: number
	actionType: "adding_analysis" | "changing_rules" | "manual_reclassify"
	affectedFractions: string[]
	newObjectsCount?: number
}

interface ReclassificationConfirmAlertProps {
	isOpen: boolean
	onConfirm: (action: ReclassificationAction) => void
	modalData: ReclassificationModalData
}

const ReclassificationConfirmAlert: React.FC<ReclassificationConfirmAlertProps> = ({
	isOpen,
	onConfirm,
	modalData,
}) => {
	const modalId = useId()
	const dialogRef = useRef<HTMLDialogElement>(null)
	const [selectedAction, setSelectedAction] = useState<ReclassificationAction>("preserve_manual")

	useEffect(() => {
		const dialog = dialogRef.current
		if (dialog) {
			if (isOpen) {
				dialog.showModal()
			} else {
				dialog.close()
			}
		}
	}, [isOpen])

	useEffect(() => {
		const dialog = dialogRef.current
		if (!dialog) return

		const handleClose = () => {
			if (isOpen) {
				// If the dialog was closed by clicking outside/backdrop, treat as cancel
				onConfirm("cancel")
			}
		}

		dialog.addEventListener("close", handleClose)
		return () => {
			dialog.removeEventListener("close", handleClose)
		}
	}, [isOpen, onConfirm])

	const getActionDescription = () => {
		switch (modalData.actionType) {
			case "adding_analysis":
				return `При добавлении ${modalData.newObjectsCount || 0} новых объектов будут затронуты ${modalData.manualObjectsCount} объектов, размещенных вручную.`
			case "changing_rules":
				return `Изменение правил классификации затронет ${modalData.manualObjectsCount} объектов, размещенных вручную.`
			case "manual_reclassify":
				return `Пересортировка всех объектов затронет ${modalData.manualObjectsCount} объектов, размещенных вручную.`
			default:
				return ""
		}
	}

	// Don't render if not open or invalid data
	if (!isOpen || !modalData || !modalData.affectedFractions) {
		return null
	}

	const handleConfirm = (action: ReclassificationAction) => {
		onConfirm(action)
	}

	return (
		<dialog
			ref={dialogRef}
			id={`reclassification_confirm_modal_${modalId}`}
			className="modal backdrop-blur-xs"
		>
			<div className="modal-box max-w-sm mx-4">
				<h3 className="text-base font-bold mb-3 flex items-center gap-2">
					<FaExclamationTriangle className="text-warning w-4 h-4" />
					Ручное размещение объектов
				</h3>
				<div className="space-y-3">
					<p className="text-sm">{getActionDescription()}</p>

					<div className="space-y-2">
						<p className="text-sm font-medium">Выберите действие:</p>

						<div className="space-y-2">
							<label className="flex items-start gap-3 cursor-pointer p-2 rounded-lg hover:bg-base-200 transition-colors">
								<input
									type="radio"
									name="reclassification-option"
									className="radio radio-primary radio-sm mt-0.5"
									checked={selectedAction === "preserve_manual"}
									onChange={() => setSelectedAction("preserve_manual")}
								/>
								<div className="flex-1">
									<div className="font-medium text-sm">Сохранить размещение</div>
									<div className="text-xs text-base-content/70 mt-0.5">
										Вручную размещенные объекты останутся на местах
									</div>
								</div>
							</label>

							<label className="flex items-start gap-3 cursor-pointer p-2 rounded-lg hover:bg-base-200 transition-colors">
								<input
									type="radio"
									name="reclassification-option"
									className="radio radio-warning radio-sm mt-0.5"
									checked={selectedAction === "reclassify_all"}
									onChange={() => setSelectedAction("reclassify_all")}
								/>
								<div className="flex-1">
									<div className="font-medium text-sm">Переместить все объекты</div>
									<div className="text-xs text-base-content/70 mt-0.5">
										Все объекты будут отсортированы по правилам
									</div>
								</div>
							</label>
						</div>
					</div>

					{modalData.affectedFractions && modalData.affectedFractions.length > 0 && (
						<div className="text-xs bg-info/10 text-info rounded-lg p-2 border border-info/20">
							Фракции: {modalData.affectedFractions.join(", ")}
						</div>
					)}
				</div>

				<div className="modal-action flex gap-2 mt-4">
					<button
						className="flex-1 btn btn-sm"
						onClick={() => handleConfirm("cancel")}
						type="button"
					>
						Отмена
					</button>
					<button
						className="flex-1 btn btn-primary btn-sm"
						onClick={() => handleConfirm(selectedAction)}
						type="button"
					>
						Применить
					</button>
				</div>
			</div>
		</dialog>
	)
}

export default ReclassificationConfirmAlert
