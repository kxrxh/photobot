import type React from "react"

const SAVE_PARAMETERS = [
	{ id: "all", label: "Все параметры" },
	{ id: "color", label: "Цвет" },
	{ id: "geometry", label: "Геометрия" },
	{ id: "median", label: "Медианы" },
]

const ParameterSelector: React.FC<{
	classificationId: string
	selectedParams: Set<string>
	onToggleParameter: (paramId: string) => void
}> = ({ classificationId, selectedParams, onToggleParameter }) => (
	<div className="space-y-3">
		<h4 className="text-sm font-medium text-base-content/80">Параметры сохранения:</h4>
		<div className="space-y-2">
			{SAVE_PARAMETERS.map((param) => {
				const isChecked = selectedParams.has(param.id)
				return (
					<div key={param.id} className="flex gap-2 items-center">
						<input
							type="checkbox"
							id={`param-${param.id}-${classificationId}`}
							className="checkbox checkbox-primary checkbox-sm"
							checked={isChecked}
							onChange={() => onToggleParameter(param.id)}
						/>
						<label
							htmlFor={`param-${param.id}-${classificationId}`}
							className="text-sm cursor-pointer text-base-content/80"
						>
							{param.label}
						</label>
					</div>
				)
			})}
		</div>
	</div>
)

export default ParameterSelector
