import type React from "react"
import { useId } from "react"

interface SaveClassificationSheetProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	classificationName: string
	setClassificationName: (name: string) => void
	isEditing?: boolean
	isCopying?: boolean
}

const SaveClassificationSheet: React.FC<SaveClassificationSheetProps> = ({
	isOpen,
	onClose,
	onConfirm,
	classificationName,
	setClassificationName,
	isEditing = false,
	isCopying = false,
}) => {
	const modalId = useId()

	return (
		<dialog id={modalId} className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}>
			<div className="modal-box">
				<h3 className="text-lg font-bold">
					{isEditing
						? "Сохранить изменения"
						: isCopying
							? "Копировать классификацию"
							: "Сохранить классификацию"}
				</h3>
				<p className="py-4">
					{isEditing
						? "Подтвердите название классификации:"
						: isCopying
							? "Введите название для копии классификации:"
							: "Введите название для новой классификации:"}
				</p>
				<input
					type="text"
					placeholder="Название классификации"
					className="w-full input input-bordered input-primary"
					value={classificationName}
					onChange={(e) => setClassificationName(e.target.value)}
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
							disabled={!classificationName.trim()}
						>
							Сохранить
						</button>
					</form>
				</div>
			</div>
		</dialog>
	)
}

export default SaveClassificationSheet
