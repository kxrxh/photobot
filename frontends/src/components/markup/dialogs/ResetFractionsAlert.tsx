import type React from "react"
import { Button } from "@/components/common/ui/Button"
import { Dialog } from "@/components/common/ui/Dialog"

interface ResetFractionsAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
}

const ResetFractionsAlert: React.FC<ResetFractionsAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
}) => {
	return (
		<Dialog
			open={isOpen}
			onClose={onClose}
			title="Подтверждение сброса"
			footer={
				<div className="flex flex-row gap-2 w-full">
					<Button className="flex-1" variant="ghost" onClick={onClose}>
						Отмена
					</Button>
					<Button className="flex-1" variant="primary" onClick={onConfirm}>
						Сбросить
					</Button>
				</div>
			}
		>
			<p className="text-sm text-base-content/80">Вы уверены, что хотите сбросить все фракции?</p>
		</Dialog>
	)
}

export default ResetFractionsAlert
