import { useId } from "react"

export default function ConfirmDeleteAlert({
	isOpen,
	onCancel,
	onConfirm,
}: {
	isOpen: boolean
	onCancel: () => void
	onConfirm: () => void
}) {
	const dialogId = useId()
	return (
		<dialog id={dialogId} className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Подтверждение удаления</h3>
				<div className="py-4">
					<p className="space-x-1">Вы уверены, что хотите удалить заметку?</p>
				</div>
				<div className="modal-action">
					<form method="dialog" className="flex flex-row gap-2 w-full">
						<button className="flex-1 btn" onClick={onCancel} type="button">
							Отмена
						</button>
						<button className="flex-1 btn btn-primary" onClick={onConfirm} type="button">
							Удалить
						</button>
					</form>
				</div>
			</div>
		</dialog>
	)
}
