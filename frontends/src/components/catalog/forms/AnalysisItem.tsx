import { AnimatePresence, motion } from "framer-motion"
import { useEffect, useState } from "react"
import { FaBan, FaChevronDown, FaChevronUp, FaTrash } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import { getAnalysisSourceUrls, getObjectImageUrl } from "@/utils/image"

const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString()

interface AnalysisItemProps {
	analysis: Analysis
	isExpanded: boolean
	onToggleExpand: (id: string) => void
	onRemove?: (id: string) => void
	excludedObjects?: number[]
	onToggleExclude?: (id: number) => void
}

const AnalysisItem: React.FC<AnalysisItemProps> = ({
	analysis,
	isExpanded,
	onToggleExpand,
	onRemove,
	excludedObjects = [],
	onToggleExclude,
}) => {
	const [objects, setObjects] = useState<KalibriObject[]>([])
	const [isLoadingObjects, setIsLoadingObjects] = useState(false)
	const thumbnailUrl = getAnalysisSourceUrls(analysis)[0]

	useEffect(() => {
		if (isExpanded && analysis.id) {
			setIsLoadingObjects(true)
			fetchAnalysisObjects(analysis.id)
				.then((data) => {
					setObjects(data)
				})
				.catch(() => {
					setObjects([])
				})
				.finally(() => {
					setIsLoadingObjects(false)
				})
		}
	}, [isExpanded, analysis.id])

	return (
		<div className="transition-colors border rounded-lg shadow-sm border-base-300 bg-base-100 hover:border-primary-focus">
			<button
				type="button"
				className="flex items-center justify-between w-full p-3 text-left cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
				onClick={() => onToggleExpand(analysis.id)}
				aria-expanded={isExpanded}
			>
				<div className="flex items-center grow space-x-3">
					{thumbnailUrl ? (
						<div className="shrink-0 w-16 h-16 overflow-hidden rounded">
							<img
								src={thumbnailUrl}
								className="object-cover w-full h-full"
								alt={`Анализ №${analysis.id}`}
								loading="lazy"
								decoding="async"
							/>
						</div>
					) : (
						<div className="flex items-center justify-center shrink-0 w-16 h-16 overflow-hidden rounded">
							<IoImageOutline size={24} className="text-gray-400" />
						</div>
					)}
					<div className="flex-1 min-w-0">
						<h3 className="font-semibold text-base-content text-md">Анализ №{analysis.id}</h3>
						<div className="space-y-1">
							<p className="text-sm text-gray-600">{formatDate(analysis.date_time)}</p>
						</div>
					</div>
				</div>
				<div className="flex items-center shrink-0 ml-4">
					{onRemove && (
						<button
							type="button"
							className="btn btn-circle btn-ghost text-primary"
							onClick={(e: React.MouseEvent) => {
								e.stopPropagation()
								onRemove(analysis.id)
							}}
							title="Удалить анализ из выбранных"
						>
							<FaTrash size={16} />
						</button>
					)}
					{isExpanded ? (
						<FaChevronUp size={16} className="ml-2 text-primary" />
					) : (
						<FaChevronDown size={16} className="ml-2 text-primary" />
					)}
				</div>
			</button>

			<AnimatePresence initial={false}>
				{isExpanded && (
					<motion.div
						key="content"
						initial="collapsed"
						animate="open"
						exit="collapsed"
						variants={{
							open: { opacity: 1, height: "auto" },
							collapsed: { opacity: 0, height: 0 },
						}}
						transition={{
							duration: 0.3,
							ease: [0.04, 0.62, 0.23, 0.98],
						}}
						className="overflow-hidden"
					>
						<div className="px-3 pt-2 pb-3 border-t border-base-300">
							{isLoadingObjects ? (
								<div className="flex items-center justify-center py-2 text-sm text-gray-500">
									Загрузка объектов...
								</div>
							) : objects.length > 0 ? (
								<div>
									<h4 className="mb-2 text-sm font-medium text-base-content">
										Объекты ({objects.length})
									</h4>
									<div className="grid grid-cols-5 gap-2 sm:grid-cols-7">
										{objects.map((obj, index) => {
											const isExcluded = excludedObjects.includes(obj.id)
											const objImageUrl = getObjectImageUrl(obj)
											return (
												<button
													key={obj.id || `obj-${analysis.id}-${index}`}
													type="button"
													className={
														"relative shrink-0 w-16 h-16 overflow-hidden rounded border transition-colors border-base-300 focus:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2"
													}
													onClick={() => onToggleExclude?.(obj.id)}
													title={isExcluded ? "Включить объект" : "Исключить объект"}
												>
													{objImageUrl ? (
														<img
															src={objImageUrl}
															className={`object-fit w-full h-full transition-all ${isExcluded ? "filter grayscale brightness-30" : ""}`}
															alt={`Объект ${obj.id}`}
															loading="lazy"
															decoding="async"
														/>
													) : (
														<div
															className={`h-16 w-16 transition-all bg-gray-200 flex items-center justify-center text-gray-400 ${isExcluded ? "filter grayscale brightness-30" : ""}`}
														>
															<IoImageOutline size={24} />
														</div>
													)}
													{isExcluded && (
														<div className="absolute inset-0 flex items-center justify-center">
															<FaBan size={24} className="text-white drop-shadow-md" />
														</div>
													)}
												</button>
											)
										})}
									</div>
								</div>
							) : (
								<p className="py-2 text-sm text-center text-gray-500">
									Объекты не найдены для этого анализа.
								</p>
							)}
						</div>
					</motion.div>
				)}
			</AnimatePresence>
		</div>
	)
}

export default AnalysisItem
