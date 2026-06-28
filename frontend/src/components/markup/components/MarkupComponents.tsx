import { AnimatePresence, motion } from "framer-motion"
import { memo } from "react"
import { FaChartBar, FaEdit, FaPlus, FaTh, FaTrash } from "react-icons/fa"
import { MdMoreVert } from "react-icons/md"
import { RiRobotLine } from "react-icons/ri"
import type { FractionType } from "@/hooks/useFractions"

interface FractionModalData {
	id: string
	name: string
}

interface FractionItemProps {
	fraction: FractionType
	isControlModeActive: boolean
	onOpenLoadModal: (fractionId?: string) => void
	onOpenStatsModal: (fraction: FractionType) => void
	onOpenObjectsModal: (fraction: FractionType) => void
	onOpenEditModal: (fraction: FractionModalData) => void
	onOpenRemoveModal: (fraction: FractionModalData) => void
	onOpenClassificationRuleModal: (fraction: FractionModalData) => void
}

const FractionItem = memo(
	({
		fraction,
		isControlModeActive,
		onOpenLoadModal,
		onOpenStatsModal,
		onOpenObjectsModal,
		onOpenEditModal,
		onOpenRemoveModal,
		onOpenClassificationRuleModal,
	}: FractionItemProps) => (
		<motion.div
			layout
			initial={{ opacity: 0, y: 20 }}
			animate={{ opacity: 1, y: 0 }}
			exit={{ opacity: 0, y: -20 }}
			transition={{ type: "spring", stiffness: 200, damping: 25 }}
			className="card card-border p-4"
		>
			<div className="flex flex-col gap-3 p-0 card-body">
				<div className="flex gap-4 justify-between items-start">
					<div className="flex overflow-hidden grow gap-3 items-start min-w-0">
						<h3 className="text-lg font-bold truncate card-title leading-tight">{fraction.name}</h3>
						<div className="flex gap-1 items-center shrink-0">
							{fraction.classificationRules && (
								<RiRobotLine
									className="text-success text-base"
									title="Правило классификации установлено"
								/>
							)}
						</div>
					</div>

					<div className="flex gap-1 items-center shrink-0">
						{!isControlModeActive && fraction.id !== "0" && (
							<div className="dropdown dropdown-end">
								<button
									className="btn btn-ghost btn-sm btn-square"
									title="Больше действий"
									type="button"
									tabIndex={0}
								>
									<MdMoreVert className="text-lg" />
								</button>
								<ul className="dropdown-content menu bg-base-100 rounded-box z-1 w-52 p-2 shadow">
									<li>
										<button
											type="button"
											onClick={() => onOpenEditModal(fraction)}
											className="justify-start w-full"
										>
											<FaEdit className="text-lg" /> Редактировать
										</button>
									</li>
									<li>
										<button
											type="button"
											onClick={() => onOpenStatsModal(fraction)}
											className="justify-start w-full"
										>
											<FaChartBar className="text-lg" /> Статистика
										</button>
									</li>
									<li>
										<button
											type="button"
											onClick={() => onOpenRemoveModal(fraction)}
											className="justify-start w-full text-error"
										>
											<FaTrash className="text-lg" /> Удалить
										</button>
									</li>
								</ul>
							</div>
						)}
					</div>
				</div>

				<div className="flex flex-wrap gap-y-2 gap-x-4 items-center text-sm text-base-content/70">
					<div className="flex gap-2 items-center">
						<span className="font-medium">Объекты:</span>
						<span className="badge badge-neutral badge-sm">{fraction.objects.length}</span>
					</div>
				</div>

				{!isControlModeActive && (
					<div className="flex flex-col gap-2 card-actions">
						<button
							type="button"
							className="btn btn-outline btn-sm w-full gap-2"
							onClick={() => onOpenObjectsModal(fraction)}
						>
							<FaTh className="text-base" />
							Объекты
						</button>
						{fraction.id !== "0" && (
							<button
								type="button"
								className="btn btn-accent btn-sm w-full gap-2"
								onClick={() => onOpenClassificationRuleModal(fraction)}
							>
								<RiRobotLine className="text-base" />
								Правила
							</button>
						)}
						<button
							type="button"
							className="btn btn-primary btn-sm w-full gap-2"
							onClick={() => onOpenLoadModal(fraction.id)}
						>
							<FaPlus className="text-base" />
							Добавить объекты
						</button>
					</div>
				)}
			</div>
		</motion.div>
	)
)

interface FractionListProps {
	fractions: FractionType[]
	isControlModeActive: boolean
	onOpenLoadModal: (fractionId?: string) => void
	onOpenStatsModal: (fraction: FractionType) => void
	onOpenObjectsModal: (fraction: FractionType) => void
	onOpenEditModal: (fraction: FractionModalData) => void
	onOpenRemoveModal: (fraction: FractionModalData) => void
	onOpenClassificationRuleModal: (fraction: FractionModalData) => void
}

export const FractionList = memo((props: FractionListProps) => {
	return (
		<main className="container px-2 py-2 mx-auto">
			<div className="space-y-2 w-full">
				<AnimatePresence>
					{props.fractions.map((fraction) => (
						<FractionItem key={fraction.id} fraction={fraction} {...props} />
					))}
				</AnimatePresence>
			</div>
		</main>
	)
})
