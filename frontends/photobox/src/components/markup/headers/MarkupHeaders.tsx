import { memo, useId } from "react"
import { FaInfoCircle, FaList, FaPlus, FaSave } from "react-icons/fa"
import { RiResetLeftLine } from "react-icons/ri"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { SheetHeaderBackButton } from "@/components/common/ui/SheetHeaderActions"
import type { ModalType } from "@/hooks/markupHooks"
import type { FractionType } from "@/hooks/useFractions"

interface NormalModeHeaderProps {
	onSave: () => void
	onLoad: () => void
	onMarkup: () => void
	onReset: () => void
	onEnterControlMode: () => void
	onAddNewFraction: () => void
}

const NormalModeHeader = memo(
	({
		onSave,
		onLoad,
		onMarkup,
		onReset,
		onEnterControlMode,
		onAddNewFraction,
	}: NormalModeHeaderProps) => {
		return (
			<div className="space-y-2">
				<div className="grid grid-cols-4 gap-2">
					<button
						type="button"
						className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm"
						onClick={onSave}
					>
						<FaSave className="mb-1 w-4 h-4" />
						<span className="text-xs leading-none text-center">Сохранить</span>
					</button>
					<button
						type="button"
						className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm"
						onClick={onLoad}
					>
						<FaPlus className="mb-1 w-4 h-4 transition-colors" />
						<span className="text-xs font-medium">Загрузить</span>
					</button>
					<button
						type="button"
						className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm"
						onClick={onMarkup}
					>
						<FaList className="mb-1 w-4 h-4" />
						<span className="text-xs">Разметка</span>
					</button>
					<button
						type="button"
						className="flex flex-col justify-center items-center p-1 h-16 btn btn-ghost btn-sm"
						onClick={onReset}
					>
						<RiResetLeftLine className="mb-1 w-4 h-4" />
						<span className="text-xs">Сбросить</span>
					</button>
				</div>
				<div className="grid grid-cols-2 gap-2">
					<button type="button" className="btn btn-sm" onClick={onEnterControlMode}>
						Упр. объектами
					</button>
					<button type="button" className="btn btn-primary btn-sm" onClick={onAddNewFraction}>
						Новая фракция
					</button>
				</div>
			</div>
		)
	}
)

interface ControlModeHeaderProps {
	selectedCount: number
	targetFractionId: string
	onTargetFractionChange: (id: string) => void
	fractions: FractionType[]
	onMove: () => void
	canMove: boolean
	onExit: () => void
}

const ControlModeHeader = memo(
	({
		selectedCount,
		targetFractionId,
		onTargetFractionChange,
		fractions,
		onMove,
		canMove,
		onExit,
	}: ControlModeHeaderProps) => {
		const targetFractionSelectId = useId()
		return (
			<div className="space-y-3">
				<div className="flex justify-between items-center">
					<div className="flex gap-2 items-center">
						<h2 className="text-xl font-bold">Перемещение объектами</h2>
						{selectedCount > 0 && (
							<div className="badge badge-primary badge-sm">{selectedCount}</div>
						)}
					</div>
					<SheetHeaderBackButton
						onClick={onExit}
						aria-label="Выйти из режима перемещения"
						title="Выйти из режима перемещения"
					/>
				</div>
				{selectedCount > 0 ? (
					<div className="">
						<div className="flex gap-2">
							<div className="max-w-xs min-w-0 flex-1">
								<ModalSelect
									id={targetFractionSelectId}
									title="Целевая фракция"
									placeholder="Куда переместить?"
									options={fractions.map((f) => ({ value: f.id, label: f.name }))}
									value={targetFractionId}
									onChange={onTargetFractionChange}
									clearable={false}
									size="sm"
								/>
							</div>
							<button
								type="button"
								onClick={onMove}
								disabled={!canMove}
								className="btn btn-primary btn-sm"
							>
								Переместить
							</button>
						</div>
					</div>
				) : (
					<div className="flex gap-2 items-center p-4 w-full badge badge-primary">
						<FaInfoCircle className="w-4 h-4 shrink-0" />
						<span className="text-sm">Выберите объекты для перемещения</span>
					</div>
				)}
			</div>
		)
	}
)

interface MarkupHeaderProps {
	controlMode: {
		isActive: boolean
		selectedObjects: { objectId: number; fractionId: string }[]
		targetFractionId: string
		setTargetFractionId: (id: string) => void
		enter: () => void
		exit: () => void
		moveSelectedObjects: () => void
		hasSelectedObjects: boolean
	}
	fractions: FractionType[]
	modalManager: {
		openModal: (modalName: ModalType) => void
	}
	addNewFraction: () => void
}

export const MarkupHeader = memo((props: MarkupHeaderProps) => {
	const { controlMode, fractions, modalManager, addNewFraction } = props

	return (
		<header className="sticky top-0 z-50 w-full border-b bg-base-100 border-base-200">
			<div className="p-2 w-full">
				{controlMode.isActive ? (
					<ControlModeHeader
						selectedCount={controlMode.selectedObjects.length}
						targetFractionId={controlMode.targetFractionId}
						onTargetFractionChange={controlMode.setTargetFractionId}
						fractions={fractions}
						onMove={controlMode.moveSelectedObjects}
						canMove={!!controlMode.targetFractionId && controlMode.hasSelectedObjects}
						onExit={controlMode.exit}
					/>
				) : (
					<NormalModeHeader
						onSave={() => modalManager.openModal("classificationSave")}
						onLoad={() => modalManager.openModal("loadModal")}
						onMarkup={() => modalManager.openModal("markupModal")}
						onReset={() => modalManager.openModal("resetFractions")}
						onEnterControlMode={controlMode.enter}
						onAddNewFraction={addNewFraction}
					/>
				)}
			</div>
		</header>
	)
})
