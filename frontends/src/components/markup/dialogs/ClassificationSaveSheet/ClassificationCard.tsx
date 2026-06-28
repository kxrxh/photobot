import { AnimatePresence, motion } from "framer-motion"
import type React from "react"
import { AiOutlineLoading3Quarters } from "react-icons/ai"
import { FaChevronDown, FaChevronUp } from "react-icons/fa"
import { PiGlobe } from "react-icons/pi"
import type { Classification } from "@/api/classification/types"
import ParameterSelector from "./ParameterSelector"

const ClassificationCard: React.FC<{
	classification: Classification
	isSelected: boolean
	isExpanded: boolean
	selectedParams: Set<string>
	isLoading: boolean
	onToggleExpansion: () => void
	onToggleParameter: (paramId: string) => void
	onReplace: () => void
}> = ({
	classification,
	isSelected,
	isExpanded,
	selectedParams,
	isLoading,
	onToggleExpansion,
	onToggleParameter,
	onReplace,
}) => (
	<div className={`card card-border p-4 ${isSelected ? "ring-2 ring-primary" : ""}`}>
		<div className="flex flex-col gap-3 p-0 card-body">
			<button
				type="button"
				className="flex flex-col gap-3 p-0 w-full text-left bg-transparent border-none cursor-pointer"
				onClick={onToggleExpansion}
				aria-expanded={isExpanded}
			>
				<div className="flex gap-4 justify-between items-center w-full">
					<div className="flex overflow-hidden flex-grow gap-2 items-center">
						<h3 className="text-base font-semibold truncate card-title">{classification.name}</h3>
						{classification.is_public ? (
							<PiGlobe className="text-lg shrink-0" title="Публичная" />
						) : null}
					</div>
					<div className="flex gap-2 items-center shrink-0">
						{isExpanded ? (
							<FaChevronUp className="text-sm" />
						) : (
							<FaChevronDown className="text-sm" />
						)}
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
			</button>
			<AnimatePresence initial={false}>
				{isExpanded && (
					<motion.div
						key="expanded"
						initial={{ opacity: 0, height: 0, y: -8 }}
						animate={{ opacity: 1, height: "auto", y: 0 }}
						exit={{ opacity: 0, height: 0, y: -8 }}
						transition={{ duration: 0.18, ease: "easeInOut" }}
						className="overflow-hidden mt-4 space-y-4"
					>
						<ParameterSelector
							classificationId={classification.id}
							selectedParams={selectedParams}
							onToggleParameter={onToggleParameter}
						/>
						<div className="pt-2">
							<button
								type="button"
								onClick={(e) => {
									e.stopPropagation()
									onReplace()
								}}
								className="w-full btn btn-primary btn-sm"
								disabled={
									isLoading ||
									!["color", "geometry", "median", "all"].some((p) => selectedParams.has(p))
								}
							>
								{isLoading ? (
									<>
										<AiOutlineLoading3Quarters className="animate-spin mr-2" />
										Расчет корреляции...
									</>
								) : (
									"Перезаписать"
								)}
							</button>
						</div>
					</motion.div>
				)}
			</AnimatePresence>
		</div>
	</div>
)

export default ClassificationCard
