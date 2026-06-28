import { useId } from "react"
import { truncate } from "@/utils/text"

interface RemoveFractionAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	fractionName?: string
}

const RemoveFractionAlert: React.FC<RemoveFractionAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	fractionName,
}) => {
	return (
		<dialog
			id={`remove_fraction_modal_${useId()}`}
			className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}
		>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Подтверждение удаления</h3>
				<div className="py-4">
					<p className="space-x-1">
						Вы уверены, что хотите удалить фракцию
						{fractionName && (
							<span className="font-semibold text-primary"> "{truncate.medium(fractionName)}"</span>
						)}
						?
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

export default RemoveFractionAlert
