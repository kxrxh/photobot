import type React from "react"
import { useId } from "react"
import { FaSave } from "react-icons/fa"

interface NewMarkupForm {
	name: string
}

interface CreateFormProps {
	form: NewMarkupForm
	onFormChange: (form: NewMarkupForm) => void
	onSave: () => void
	onCancel: () => void
}

const CreateForm: React.FC<CreateFormProps> = ({ form, onFormChange, onSave, onCancel }) => {
	const isFormValid = form.name.trim().length > 0
	const nameId = useId()

	return (
		<div className="flex flex-col h-full bg-base-100">
			<div className="overflow-y-auto flex-1 px-4 py-2 space-y-6">
				<fieldset className="fieldset">
					<legend className="fieldset-legend">Основная информация</legend>
					<div className="space-y-4">
						<div>
							<label htmlFor={nameId} className="label">
								<span className="label-text">Название разметки</span>
							</label>
							<input
								id={nameId}
								type="text"
								value={form.name}
								onChange={(e) => onFormChange({ ...form, name: e.target.value })}
								placeholder="Введите название разметки"
								className="w-full input input-bordered"
							/>
						</div>
					</div>
				</fieldset>
			</div>
			<div className="sticky bottom-0 z-10 p-2 bg-base-100">
				<div className="flex relative gap-2 w-full">
					<button type="button" className="flex-1 btn" onClick={onCancel}>
						Отмена
					</button>
					<button
						type="button"
						className="flex-1 btn btn-primary"
						onClick={onSave}
						disabled={!isFormValid}
					>
						<FaSave className="mr-2" />
						Сохранить
					</button>
				</div>
			</div>
		</div>
	)
}

export default CreateForm
export type { NewMarkupForm, CreateFormProps }
