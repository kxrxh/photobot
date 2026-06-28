import type React from "react"
import { Button } from "@/components/common/ui/Button"
import { Dialog } from "@/components/common/ui/Dialog"

interface DeleteClassificationAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	classificationName?: string
}

const DeleteClassificationAlert: React.FC<DeleteClassificationAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	classificationName,
}) => {
	return (
		<Dialog
			open={isOpen}
			onClose={onClose}
			title="Подтверждение удаления"
			footer={
				<div className="flex flex-row gap-2 w-full">
					<Button className="flex-1" variant="ghost" onClick={onClose}>
						Отмена
					</Button>
					<Button className="flex-1" variant="primary" onClick={onConfirm}>
						Удалить
					</Button>
				</div>
			}
		>
			<p className="text-sm text-base-content/80">
				Вы уверены, что хотите удалить классификацию "{classificationName}"?
			</p>
		</Dialog>
	)
}

export default DeleteClassificationAlert
