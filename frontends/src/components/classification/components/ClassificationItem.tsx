import { motion } from "framer-motion"
import { useEffect } from "react"
import { MdContentCopy, MdDelete, MdEdit, MdMoreVert, MdPause, MdPlayArrow } from "react-icons/md"
import { PiGlobe, PiGlobeXLight } from "react-icons/pi"
import type { Classification } from "@/api/classification/types"

interface ClassificationItemProps {
	classification: Classification
	hasSuperRights: boolean
	onEdit: (classification: Classification) => void
	onCopy: (classification: Classification) => void
	onDelete: (classification: Classification) => void
	onCheck: (classification: Classification) => void
	onTogglePublic: (classification: Classification) => void
	isSelected: boolean
}

function ClassificationItem({
	classification,
	hasSuperRights,
	onEdit,
	onCopy,
	onDelete,
	onCheck,
	onTogglePublic,
	isSelected,
}: ClassificationItemProps) {
	const canEditOrDelete = !classification.is_public || hasSuperRights

	useEffect(() => {
		if (isSelected) {
			window.scrollTo({ top: 0, behavior: "smooth" })
		}
	}, [isSelected])

	return (
		<motion.div
			layout
			initial={{ opacity: 0, y: 20 }}
			animate={{ opacity: 1, y: 0 }}
			exit={{ opacity: 0, y: -20 }}
			transition={{ type: "spring", stiffness: 200, damping: 25 }}
			className={`card card-border p-4 ${isSelected ? "ring-2 ring-primary" : ""}`}
		>
			<div className="flex flex-col gap-3 p-0 card-body">
				<div className="flex gap-4 justify-between items-center">
					<div className="flex overflow-hidden flex-grow gap-2 items-center">
						<h3 className="text-base font-semibold truncate card-title">{classification.name}</h3>
						{classification.is_public ? (
							<PiGlobe className="text-lg shrink-0" title="Публичная" />
						) : null}
					</div>

					<div className="flex gap-1 items-center shrink-0">
						<div className="dropdown dropdown-end">
							<button
								className="btn btn-ghost btn-sm btn-square"
								title="Больше действий"
								type="button"
								tabIndex={0}
							>
								<MdMoreVert className="text-lg" />
							</button>
							<ul className="dropdown-content menu bg-base-100 rounded-box z-[1] w-52 p-2 shadow">
								{canEditOrDelete && (
									<li>
										<button
											type="button"
											onClick={() => onEdit(classification)}
											className="justify-start w-full"
										>
											<MdEdit className="text-lg" /> Редактировать
										</button>
									</li>
								)}
								<li>
									<button
										type="button"
										onClick={() => onCopy(classification)}
										className="justify-start w-full"
									>
										<MdContentCopy className="text-lg" /> Копировать
									</button>
								</li>
								{canEditOrDelete && (
									<li>
										<button
											type="button"
											onClick={() => onDelete(classification)}
											className="justify-start w-full text-primary"
										>
											<MdDelete className="text-lg" /> Удалить
										</button>
									</li>
								)}
							</ul>
						</div>
					</div>
				</div>
				<div className="flex flex-wrap gap-y-1 gap-x-2 items-center text-sm text-base-content/70">
					<div className="flex gap-1 items-center">
						Продукт:
						<span className="badge badge-outline badge-sm">{classification.product.name}</span>
					</div>
					<div className="flex gap-1 items-center">
						Обновлено:
						<span className="badge badge-outline badge-sm">
							{new Date(classification.updated_at).toLocaleDateString("ru-RU", {
								day: "2-digit",
								month: "2-digit",
								year: "2-digit",
								hour: "2-digit",
								minute: "2-digit",
							})}
						</span>
					</div>
				</div>
				{hasSuperRights && (
					<div className="justify-end card-actions">
						<button
							type="button"
							className={`btn btn-sm w-full sm:w-auto ${classification.is_public ? "btn-accent" : "btn-neutral"}`}
							onClick={() => onTogglePublic(classification)}
						>
							{classification.is_public ? (
								<>
									<PiGlobeXLight className="mr-1 text-lg" />
									Снять с публикации
								</>
							) : (
								<>
									<PiGlobe className="mr-1 text-lg" />
									Опубликовать
								</>
							)}
						</button>
					</div>
				)}
				<button
					type="button"
					className={`w-full btn btn-sm sm:w-auto ${isSelected ? "btn-secondary" : "btn-primary"}`}
					onClick={() => onCheck(classification)}
				>
					{isSelected ? (
						<>
							<MdPause className="mr-1 text-lg" />
							Деактивировать
						</>
					) : (
						<>
							<MdPlayArrow className="mr-1 text-lg" />
							Активировать
						</>
					)}
				</button>
			</div>
		</motion.div>
	)
}

export default ClassificationItem
