import type React from "react"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import type { FractionType } from "@/hooks/useFractions"
import { ObjectImage } from "../components/ObjectImage"

interface FractionObjectsSheetProps {
	isOpen: boolean
	onClose: () => void
	fraction: FractionType | null
	isControlModeActive: boolean
	isObjectSelected: (id: number) => boolean
	onObjectClick: (id: number, fractionId: string) => void
}

const EmptyState: React.FC = () => (
	<div className="p-6 text-center">
		<p className="text-base-content/70">В этой фракции нет объектов.</p>
	</div>
)

const FractionObjectsSheet: React.FC<FractionObjectsSheetProps> = ({
	isOpen,
	onClose,
	fraction,
	isControlModeActive,
	isObjectSelected,
	onObjectClick,
}) => {
	if (!isOpen || !fraction) {
		return null
	}

	const objectsCount = fraction.objects.length

	return (
		<div className="modal modal-open">
			<div className="modal-box max-w-4xl h-128 flex flex-col">
				<div className="flex items-center justify-between mb-4 shrink-0 gap-3">
					<div className="min-w-0">
						<h3 className="text-lg font-bold">Объекты фракции</h3>
						<p className="text-sm text-base-content/70">
							{fraction.name} • {objectsCount} объектов
						</p>
					</div>
					<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
				</div>

				<div className="flex-1 overflow-y-auto">
					{objectsCount === 0 ? (
						<EmptyState />
					) : (
						<div className="grid gap-2 grid-cols-4 sm:grid-cols-4 md:grid-cols-6 lg:grid-cols-8">
							{fraction.objects.map((obj) => (
								<ObjectImage
									key={obj.id}
									object={obj}
									isControlModeActive={isControlModeActive}
									isSelected={isObjectSelected(obj.id)}
									onClick={() => onObjectClick(obj.id, fraction.id)}
								/>
							))}
						</div>
					)}
				</div>

				<div className="modal-action shrink-0">
					<button type="button" className="btn btn-block" onClick={onClose}>
						Закрыть
					</button>
				</div>
			</div>
		</div>
	)
}

export default FractionObjectsSheet
