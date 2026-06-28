import type React from "react"
import { useId } from "react"

interface EditFractionAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	newFractionName: string
	setNewFractionName: (name: string) => void
}

const EditFractionAlert: React.FC<EditFractionAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	newFractionName,
	setNewFractionName,
}) => {
	const modalId = useId()

	return (
		<dialog id={modalId} className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Редактировать фракцию</h3>
				<p className="py-4">Введите новое название для фракции:</p>
				<input
					type="text"
					placeholder="Название фракции"
					className="w-full input input-bordered input-primary"
					value={newFractionName}
					onChange={(e) => setNewFractionName(e.target.value)}
					required
				/>
				<div className="modal-action">
					<form method="dialog" className="flex flex-row gap-2 w-full">
						<button className="flex-1 btn" onClick={onClose} type="button">
							Отмена
						</button>
						<button
							className="flex-1 btn btn-primary"
							onClick={onConfirm}
							type="button"
							disabled={!newFractionName.trim()}
						>
							Сохранить
						</button>
					</form>
				</div>
			</div>
		</dialog>
	)
}

export default EditFractionAlert
