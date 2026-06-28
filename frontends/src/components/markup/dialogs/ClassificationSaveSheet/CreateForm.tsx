import { useQuery } from "@tanstack/react-query"
import type React from "react"
import { useEffect, useId, useMemo } from "react"
import { AiOutlineLoading3Quarters } from "react-icons/ai"
import { FaSave } from "react-icons/fa"
import type { ParameterGroup } from "@/api/correlation/types"
import { getAllProducts } from "@/api/product"
import type { Product } from "@/api/product/types"
import { queryKeys } from "@/api/queryKeys"
import { ModalSelect } from "@/components/common/ui/ModalSelect"

interface NewClassificationForm {
	name: string
	product?: Product
	selectedParams?: Set<ParameterGroup>
}

interface CreateFormProps {
	form: NewClassificationForm
	onFormChange: (form: NewClassificationForm) => void
	onSave: () => void
	onCancel: () => void
	isLoading: boolean
}

const PARAMETER_OPTIONS: {
	id: ParameterGroup
	label: string
	description: string
}[] = [
	{
		id: "all",
		label: "Все параметры",
		description: "Полный анализ всех доступных параметров",
	},
	{ id: "color", label: "Цвет", description: "Анализ цветовых характеристик" },
	{
		id: "geometry",
		label: "Геометрия",
		description: "Анализ геометрических параметров",
	},
	{
		id: "median",
		label: "Медиана",
		description: "Медианные значения параметров",
	},
]

const DEFAULT_SELECTED_PARAMS = new Set<ParameterGroup>(["all"])
const MIN_NAME_LENGTH = 3

const CreateForm: React.FC<CreateFormProps> = ({
	form,
	onFormChange,
	onSave,
	onCancel,
	isLoading,
}) => {
	const nameId = useId()
	const productId = useId()
	const selectedParams = form.selectedParams || DEFAULT_SELECTED_PARAMS

	const { data: productsData, isLoading: loadingProducts } = useQuery({
		queryKey: queryKeys.products,
		queryFn: getAllProducts,
		staleTime: 1000 * 60 * 5,
		gcTime: 1000 * 60 * 15,
	})
	const products: Product[] = productsData || []
	const productOptions = useMemo(
		() =>
			products
				.map((product) => ({
					value: String(product.id),
					label: product.name,
				}))
				.sort((left, right) => left.label.localeCompare(right.label, "ru")),
		[products]
	)

	useEffect(() => {
		if (!form.selectedParams || form.selectedParams.size === 0) {
			onFormChange({ ...form, selectedParams: new Set(["all"]) })
		}
	}, [form, form.selectedParams, onFormChange])

	const handleParameterToggle = (paramId: ParameterGroup) => {
		let newSelectedParams: Set<ParameterGroup>
		if (paramId === "all") {
			if (selectedParams.has("all")) {
				// Unselect all
				newSelectedParams = new Set()
			} else {
				// Select only 'all'
				newSelectedParams = new Set(["all"])
			}
		} else {
			newSelectedParams = new Set(selectedParams)
			if (newSelectedParams.has(paramId)) {
				newSelectedParams.delete(paramId)
			} else {
				newSelectedParams.add(paramId)
			}
			// If any other is selected, unselect 'all'
			if (newSelectedParams.has(paramId)) {
				newSelectedParams.delete("all")
			}
			// If none left, reselect all
			if (newSelectedParams.size === 0) {
				newSelectedParams = new Set(["all"])
			}
		}
		onFormChange({ ...form, selectedParams: newSelectedParams })
	}

	const isNameValid = form.name.trim().length >= MIN_NAME_LENGTH
	const isFormValid = isNameValid && !!form.product && selectedParams.size > 0

	return (
		<div className="flex flex-col h-full bg-base-100">
			<div className="overflow-y-auto flex-1 px-4 py-2 space-y-6">
				<fieldset className="fieldset">
					<legend className="fieldset-legend">Основная информация</legend>
					<div className="space-y-4">
						<div className="">
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
						<div>
							<label htmlFor={productId} className="label">
								<span className="label-text">Продукт</span>
							</label>
							<ModalSelect
								id={productId}
								title="Продукт"
								placeholder="Выберите продукт"
								options={productOptions}
								value={form.product?.id != null ? String(form.product.id) : ""}
								onChange={(v) => {
									const selected = v === "" ? undefined : products.find((p) => String(p.id) === v)
									onFormChange({ ...form, product: selected })
								}}
								disabled={loadingProducts}
							/>
						</div>
					</div>
				</fieldset>

				<fieldset className="fieldset">
					<legend className="fieldset-legend">Параметры корреляции</legend>
					<div className="space-y-3">
						<p className="text-sm text-base-content/70">
							Выберите группы параметров для анализа корреляции:
						</p>
						<div className="space-y-2">
							{PARAMETER_OPTIONS.map((param) => (
								<div key={param.id} className="form-control">
									<label className="label cursor-pointer justify-start gap-3">
										<input
											type="checkbox"
											className="checkbox checkbox-primary"
											checked={selectedParams.has(param.id)}
											onChange={() => handleParameterToggle(param.id)}
										/>
										<div className="flex-1">
											<span className="label-text font-medium">{param.label}</span>
											<p className="text-xs text-base-content/60 mt-1">{param.description}</p>
										</div>
									</label>
								</div>
							))}
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
						disabled={!isFormValid || isLoading}
					>
						{isLoading ? (
							<>
								<AiOutlineLoading3Quarters className="animate-spin mr-2" />
								Расчет корреляции...
							</>
						) : (
							<>
								<FaSave className="mr-2" />
								Сохранить
							</>
						)}
					</button>
				</div>
			</div>
		</div>
	)
}

export default CreateForm
export type { NewClassificationForm, CreateFormProps }
