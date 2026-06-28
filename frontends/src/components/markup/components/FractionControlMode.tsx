import { memo } from "react"
import type { FractionType } from "@/hooks/useFractions"
import { GRID_CLASSES } from "@/utils/markupUtils"
import { ObjectImage } from "./ObjectImage"

interface FractionControlItemProps {
	fraction: FractionType
	isObjectSelected: (id: number) => boolean
	onObjectClick: (id: number, fractionId: string) => void
	onToggleFractionSelection: (fractionId: string) => void
}

const FractionControlItem = memo(
	({
		fraction,
		isObjectSelected,
		onObjectClick,
		onToggleFractionSelection,
	}: FractionControlItemProps) => {
		const selectedCount = fraction.objects.filter((obj) => isObjectSelected(obj.id)).length
		const allSelected = fraction.objects.length > 0 && selectedCount === fraction.objects.length

		return (
			<div className="collapse collapse-arrow border border-base-300">
				<input type="checkbox" />
				<div className="collapse-title">
					<div className="flex overflow-hidden items-center text-xl font-medium">
						<span className="flex-1 mr-2 truncate">{fraction.name}</span>
						<div className="mr-2 badge badge-neutral badge-sm">{fraction.objects.length}</div>
					</div>
				</div>
				<div className="collapse-content">
					{fraction.objects.length > 0 && (
						<div className="flex p-2 pb-2 relative z-10">
							<button
								type="button"
								className="btn btn-sm btn-outline btn-primary w-full"
								onClick={(e) => {
									e.stopPropagation()
									onToggleFractionSelection(fraction.id)
								}}
							>
								{allSelected ? "Снять выделение" : "Выделить все"}
							</button>
						</div>
					)}
					<div className={`${GRID_CLASSES} p-2 ${fraction.objects.length > 0 ? "pt-0" : ""}`}>
						{fraction.objects.map((obj) => (
							<ObjectImage
								key={obj.id}
								object={obj}
								isSelected={isObjectSelected(obj.id)}
								isControlModeActive={true}
								onClick={() => onObjectClick(obj.id, fraction.id)}
							/>
						))}
					</div>
				</div>
			</div>
		)
	}
)

interface FractionControlModeProps {
	fractions: FractionType[]
	isObjectSelected: (id: number) => boolean
	onObjectClick: (id: number, fractionId: string) => void
	onToggleFractionSelection: (fractionId: string) => void
}

const FractionControlMode = memo(
	({
		fractions,
		isObjectSelected,
		onObjectClick,
		onToggleFractionSelection,
	}: FractionControlModeProps) => (
		<main className="container px-2 py-2 mx-auto">
			<div className="space-y-2 w-full">
				{fractions.map((fraction) => (
					<FractionControlItem
						key={fraction.id}
						fraction={fraction}
						isObjectSelected={isObjectSelected}
						onObjectClick={onObjectClick}
						onToggleFractionSelection={onToggleFractionSelection}
					/>
				))}
			</div>
		</main>
	)
)

export default FractionControlMode
