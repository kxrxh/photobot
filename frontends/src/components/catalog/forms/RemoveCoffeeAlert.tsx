import type React from "react"
import { useId } from "react"
import { truncate } from "@/utils/text"

interface RemoveCoffeeAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	coffeeName?: string
}

const RemoveCoffeeAlert: React.FC<RemoveCoffeeAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	coffeeName,
}) => {
	const dialogId = useId()
	return (
		<dialog id={dialogId} className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Подтверждение удаления</h3>
				<div className="py-4">
					<p className="space-x-1">
						Вы уверены, что хотите удалить запись
						{coffeeName && (
							<span className="font-semibold text-primary"> "{truncate.medium(coffeeName)}"</span>
						)}
						? Это действие нельзя отменить.
					</p>
				</div>
				<div className="modal-action">
					<form method="dialog" className="flex flex-row gap-2 w-full">
						<button className="flex-1 btn" onClick={onClose} type="button">
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

export default RemoveCoffeeAlert
