import { useEffect, useId, useState } from "react"
import { FaPlus, FaTrash } from "react-icons/fa"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import { PARAMETER_GROUPS } from "@/constants"
import type { ClassificationRule, ClassificationRules } from "@/hooks/useFractions"

interface FractionClassificationRuleSheetProps {
	isOpen: boolean
	onClose: () => void
	onConfirm: (rules?: ClassificationRules) => void
	fractionName: string
	initialRules?: ClassificationRules
}

// Create grouped parameter options from PARAMETER_GROUPS
// Map parameter group values to the actual parameter names used in the system
const PARAMETER_VALUE_MAPPING: Record<string, string> = {
	L: "l",
	W: "w",
	Pr: "pr",
	Sq: "sq",
	"L/M": "l_m", // Assuming this maps to some relative length field
	"W/M": "w_m", // Assuming this maps to some relative width field
	"L/W": "l_w",
	Solid: "solid",
	"L/Avg": "l_avg", // Assuming this maps to some average length field
	"W/Avg": "w_avg", // Assuming this maps to some average width field
	R: "r",
	G: "g",
	B: "b",
	Min_R: "min_r",
	Min_G: "min_g",
	Min_B: "min_b",
	Max_R: "max_r",
	Max_G: "max_g",
	Max_B: "max_b",
	median_R: "m_r",
	median_G: "m_g",
	median_B: "m_b",
	"R/M": "r_m", // Assuming relative to median
	"G/M": "g_m",
	"B/M": "b_m",
	"R/Avg": "r_avg",
	"G/Avg": "g_avg",
	"B/Avg": "b_avg",
	Brt: "brt",
	"Brt/M": "brt_m",
	"Brt/Avg": "brt_avg",
	H: "h",
	S: "s",
	V: "v",
	Min_H: "min_h",
	Min_S: "min_s",
	Min_V: "min_v",
	Max_H: "max_h",
	Max_S: "max_s",
	Max_V: "max_v",
	median_H: "m_h",
	median_S: "m_s",
	median_V: "m_v",
	"H/M": "h_m",
	"S/M": "s_m",
	"V/M": "v_m",
	"H/Avg": "h_avg",
	"S/Avg": "s_avg",
	"V/Avg": "v_avg",
	"Sq/SqCrl": "sq_sqcrl",
	Hu1: "hu1",
	Hu2: "hu2",
	Hu3: "hu3",
	Hu4: "hu4",
	Hu5: "hu5",
	Hu6: "hu6",
}

// Create grouped options with mapped values
const GROUPED_PARAMETER_OPTIONS = PARAMETER_GROUPS.map((group) => ({
	label: group.label,
	options: group.options.map((opt) => ({
		value: PARAMETER_VALUE_MAPPING[opt.value] || opt.value.toLowerCase(),
		label: opt.label,
	})),
})).filter((group) => group.options.length > 0)

const OPERATOR_OPTIONS = [
	{ value: "<", label: "<" },
	{ value: "<=", label: "≤" },
	{ value: "==", label: "=" },
	{ value: ">=", label: "≥" },
	{ value: ">", label: ">" },
	{ value: "!=", label: "≠" },
] as const

const LOGIC_OPTIONS = [
	{ value: "AND", label: "И (все правила должны выполняться)" },
	{ value: "OR", label: "ИЛИ (хотя бы одно правило должно выполняться)" },
] as const

function FractionClassificationRuleSheet({
	isOpen,
	onClose,
	onConfirm,
	fractionName,
	initialRules,
}: FractionClassificationRuleSheetProps) {
	const [rules, setRules] = useState<ClassificationRule[]>(initialRules?.rules || [])
	const [logic, setLogic] = useState<"AND" | "OR">(initialRules?.logic || "AND")

	const logicId = useId()
	const ruleParamPrefix = useId()
	const ruleOperatorPrefix = useId()

	const groupedParameterModalOptions = GROUPED_PARAMETER_OPTIONS.map((group) => ({
		label: group.label,
		options: group.options.map((opt) => ({ value: opt.value, label: opt.label })),
	}))

	useEffect(() => {
		if (isOpen) {
			setRules(initialRules?.rules || [])
			setLogic(initialRules?.logic || "AND")
		}
	}, [isOpen, initialRules])

	const handleAddRule = () => {
		const newRule: ClassificationRule = {
			parameter: "",
			operator: "<",
			value: 0,
		}
		setRules((prev) => [...prev, newRule])
	}

	const handleRemoveRule = (index: number) => {
		setRules((prev) => prev.filter((_, i) => i !== index))
	}

	const handleConfirm = () => {
		if (rules.length > 0) {
			const classificationRules: ClassificationRules = {
				rules,
				logic,
			}
			onConfirm(classificationRules)
		} else {
			onConfirm(undefined) // Remove all rules
		}
		onClose()
	}

	if (!isOpen) return null

	return (
		<div className="modal modal-open">
			<div className="modal-box max-w-4xl h-128 flex flex-col">
				<div className="flex items-center justify-between mb-4 shrink-0 gap-3">
					<h3 className="min-w-0 text-lg font-bold">Правила классификации для "{fractionName}"</h3>
					<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
				</div>

				{rules.length > 1 && (
					<div className="mb-4 flex w-full min-w-0 shrink-0 flex-col gap-1.5">
						<label className="text-sm font-medium text-base-content/85" htmlFor={logicId}>
							Логика объединения
						</label>
						<ModalSelect
							id={logicId}
							title="Логика объединения"
							placeholder="Логика объединения"
							options={LOGIC_OPTIONS.map((opt) => ({ value: opt.value, label: opt.label }))}
							value={logic}
							onChange={(v) => setLogic(v === "OR" ? "OR" : "AND")}
							clearable={false}
							size="sm"
						/>
					</div>
				)}

				<div className="flex-1 space-y-2 overflow-y-auto">
					{rules.map((rule, index) => (
						<div
							key={`${rule.parameter}-${rule.operator}-${index}`}
							className="flex items-center p-2 space-x-2 rounded-xl bg-base-100 border border-base-200"
						>
							<button
								type="button"
								onClick={() => handleRemoveRule(index)}
								className="text-primary btn btn-xs btn-ghost btn-circle"
							>
								<FaTrash className="w-4 h-4" />
							</button>

							<div className="w-1/3 min-w-0">
								<ModalSelect
									id={`${ruleParamPrefix}-${index}`}
									title="Параметр"
									placeholder="Выберите параметр"
									groupedOptions={groupedParameterModalOptions}
									value={rule.parameter}
									onChange={(v) => {
										const updatedRules = [...rules]
										updatedRules[index] = { ...rule, parameter: v }
										setRules(updatedRules)
									}}
									size="sm"
								/>
							</div>

							<div className="min-w-0 flex-1">
								<ModalSelect
									id={`${ruleOperatorPrefix}-${index}`}
									title="Оператор"
									placeholder="Оператор"
									options={OPERATOR_OPTIONS.map((opt) => ({
										value: opt.value,
										label: opt.label,
									}))}
									value={rule.operator}
									onChange={(v) => {
										const updatedRules = [...rules]
										updatedRules[index] = {
											...rule,
											operator: v as ClassificationRule["operator"],
										}
										setRules(updatedRules)
									}}
									clearable={false}
									size="sm"
								/>
							</div>

							<input
								type="number"
								step="0.01"
								min="-999999"
								max="999999"
								className="flex-1 input input-sm input-bordered [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
								placeholder="0.00"
								value={rule.value}
								onChange={(e) => {
									const updatedRules = [...rules]
									updatedRules[index] = { ...rule, value: Number(e.target.value) || 0 }
									setRules(updatedRules)
								}}
							/>
						</div>
					))}

					{rules.length === 0 && (
						<div className="text-center py-8 text-base-content/60">
							Нет правил классификации. Добавьте первое правило, чтобы начать распределение объектов
							по фракциям.
						</div>
					)}
				</div>

				<div className="modal-action flex-col gap-2 shrink-0">
					<div className="w-full">
						<button type="button" onClick={handleAddRule} className="btn btn-sm btn-accent w-full">
							<FaPlus className="mr-2" />
							Добавить правило
						</button>
					</div>
					<div className="flex gap-2">
						<button type="button" onClick={onClose} className="btn flex-1">
							Отмена
						</button>
						<button
							type="button"
							onClick={handleConfirm}
							className="btn btn-primary flex-1"
							disabled={rules.some((rule) => !rule.parameter)}
						>
							Применить
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}

export default FractionClassificationRuleSheet
