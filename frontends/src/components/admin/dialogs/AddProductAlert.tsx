import type React from "react"
import { useId } from "react"
import { Button } from "@/components/common/ui/Button"
import { Dialog } from "@/components/common/ui/Dialog"
import { FormField } from "@/components/common/ui/FormField"
import { Input } from "@/components/common/ui/Input"

interface AddProductAlertProps {
	isOpen: boolean
	onClose: () => void
	newProductName: string
	setNewProductName: (name: string) => void
	onConfirm: () => void
}

const AddProductAlert: React.FC<AddProductAlertProps> = ({
	isOpen,
	onClose,
	onConfirm,
	newProductName,
	setNewProductName,
}) => {
	const inputId = useId()

	return (
		<Dialog
			open={isOpen}
			onClose={onClose}
			title="Добавить продукт"
			size="md"
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
			<p className="text-sm text-base-content/80">Введите название для продукта:</p>
			<FormField id={inputId} label="Название" required>
				<Input
					id={inputId}
					type="text"
					placeholder="Название продукта"
					value={newProductName}
					onChange={(e) => setNewProductName(e.target.value)}
					required
					size="sm"
				/>
			</FormField>
		</Dialog>
	)
}

export default AddProductAlert
