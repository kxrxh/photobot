import type React from "react"
import { useId } from "react"
import { Button } from "@/components/common/ui/Button"
import { Dialog } from "@/components/common/ui/Dialog"
import { FormField } from "@/components/common/ui/FormField"
import { Input } from "@/components/common/ui/Input"

interface AddParameterAlertProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: () => void
	newParamName: string
	setNewParamName: (name: string) => void
}

const AddParameterAlert: React.FC<AddParameterAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	newParamName,
	setNewParamName,
}) => {
	const inputId = useId()

	return (
		<Dialog
			open={isOpen}
			onClose={onClose}
			title="Добавить параметр"
			footer={
				<div className="flex flex-row gap-2 w-full">
					<Button className="flex-1" variant="ghost" onClick={onClose}>
						Отмена
					</Button>
					<Button className="flex-1" variant="primary" onClick={onConfirm}>
						Сохранить
					</Button>
				</div>
			}
		>
			<p className="text-sm text-base-content/80">Введите название параметра:</p>
			<FormField id={inputId} label="Название" required>
				<Input
					id={inputId}
					type="text"
					placeholder="Название параметра"
					value={newParamName}
					onChange={(e) => setNewParamName(e.target.value)}
					required
					size="sm"
				/>
			</FormField>
		</Dialog>
	)
}

export default AddParameterAlert
