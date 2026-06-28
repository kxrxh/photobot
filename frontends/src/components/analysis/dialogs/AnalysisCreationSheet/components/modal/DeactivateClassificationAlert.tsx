import type React from "react"
import { useEffect, useRef } from "react"

interface DeactivateClassificationAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	isPending: boolean
}

const DeactivateClassificationAlert: React.FC<DeactivateClassificationAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	isPending,
}) => {
	const deactivateModalRef = useRef<HTMLDialogElement>(null)

	// Handle dialog open/close
	useEffect(() => {
		if (isOpen) {
			deactivateModalRef.current?.showModal()
		} else {
			deactivateModalRef.current?.close()
		}
	}, [isOpen])

	const handleCancel = () => {
		onClose()
	}

	const handleConfirm = () => {
		onConfirm()
	}

	return (
		<dialog
			ref={deactivateModalRef}
			className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}
		>
			<div className="modal-box">
				<h3 className="text-lg font-bold mb-2">Подтверждение деактивации</h3>
				<p className="text-sm text-base-content/70 mb-6">
					Вы уверены, что хотите деактивировать активную классификацию?
				</p>
				<div className="modal-action">
					<form method="dialog" className="flex flex-row w-full gap-3">
						<button
							className="btn flex-1"
							onClick={handleCancel}
							type="button"
							disabled={isPending}
						>
							Отмена
						</button>
						<button
							className="btn btn-primary flex-1"
							onClick={handleConfirm}
							type="button"
							disabled={isPending}
						>
							{isPending ? (
								<span className="loading loading-spinner loading-sm"></span>
							) : (
								"Деактивировать"
							)}
						</button>
					</form>
				</div>
			</div>
			<form
				method="dialog"
				className="modal-backdrop"
				onClick={handleCancel}
				onKeyDown={(e) => {
					if (e.key === "Enter" || e.key === " ") {
						e.preventDefault()
						handleCancel()
					}
				}}
			>
				<button type="button">close</button>
			</form>
		</dialog>
	)
}

export default DeactivateClassificationAlert
