import type React from "react"
import { useId } from "react"

interface EditProductSheetProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	newProductName: string
	setNewProductName: (name: string) => void
}

const EditProductSheet: React.FC<EditProductSheetProps> = ({
	isOpen,
	onClose,
	onConfirm,
	newProductName,
	setNewProductName,
}) => {
	return (
		<dialog
			id={`edit_fraction_modal_${useId()}`}
			className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}
		>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Редактировать продукт</h3>
				<p className="py-4">Введите новое название для продукта:</p>
				<input
					type="text"
					placeholder="Название продукта"
					className="w-full input input-bordered input-primary"
					value={newProductName}
					onChange={(e) => setNewProductName(e.target.value)}
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
							disabled={!newProductName.trim()}
						>
							Сохранить
						</button>
					</form>
				</div>
			</div>
		</dialog>
	)
}

export default EditProductSheet
