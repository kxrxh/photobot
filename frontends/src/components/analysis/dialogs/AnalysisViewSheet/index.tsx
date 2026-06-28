import { useQuery } from "@tanstack/react-query"
import { AnimatePresence, motion, useReducedMotion } from "framer-motion"
import type React from "react"
import { useEffect, useState } from "react"
import { createPortal } from "react-dom"
import { FaInfo, FaList } from "react-icons/fa"
import { IoIosStats } from "react-icons/io"
import { fetchAnalysisObjects } from "@/api/analysis"
import type { Analysis } from "@/api/analysis/types"
import { queryKeys } from "@/api/queryKeys"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import CharacteristicsTab from "./tabs/CharacteristicsTab"
import MainTab from "./tabs/MainTab"
import ObjectsTab from "./tabs/ObjectsTab"
import type { TabDefinition, TabType } from "./types"

interface AnalysisViewSheetProps {
	isOpen: boolean
	onClose: () => void
	analysis: Analysis | null
}

const tabs: TabDefinition[] = [
	{ id: "main", label: "Основное", icon: FaInfo },
	{ id: "characteristics", label: "Характеристики", icon: IoIosStats },
	{ id: "objects", label: "Объекты", icon: FaList },
]

const AnalysisViewSheet: React.FC<AnalysisViewSheetProps> = ({ isOpen, onClose, analysis }) => {
	const [activeTab, setActiveTab] = useState<TabType>("main")
	const reduceMotion = useReducedMotion()
	const analysisId = analysis?.id

	const {
		data: objects,
		isLoading: objectsLoading,
		isError: objectsError,
	} = useQuery({
		queryKey: queryKeys.analyses.objects(analysisId ?? ""),
		queryFn: () => {
			if (analysisId === undefined) {
				return Promise.reject(new Error("fetchAnalysisObjects: missing analysis id"))
			}
			return fetchAnalysisObjects(analysisId)
		},
		enabled: isOpen && analysisId !== undefined,
		staleTime: 15 * 60 * 1000,
		refetchOnWindowFocus: false,
		refetchOnReconnect: false,
	})

	// analysisId: same sheet open, different analysis → return to «Основное»
	// biome-ignore lint/correctness/useExhaustiveDependencies: analysisId is required in addition to isOpen
	useEffect(() => {
		if (!isOpen) return
		setActiveTab("main")
	}, [isOpen, analysisId])

	const tabProps = analysis
		? {
				analysis,
				objects,
				objectsLoading,
				objectsError,
			}
		: null

	const renderTabContent = () => {
		if (!analysis || !tabProps) return null

		switch (activeTab) {
			case "main":
				return <MainTab {...tabProps} />
			case "characteristics":
				return <CharacteristicsTab {...tabProps} />
			case "objects":
				return <ObjectsTab {...tabProps} />
			default:
				return <MainTab {...tabProps} />
		}
	}

	if (!analysis) return null

	return createPortal(
		<AnimatePresence>
			{isOpen && (
				<>
					<motion.div
						initial={{ opacity: 0 }}
						animate={{ opacity: 1 }}
						exit={{ opacity: 0 }}
						transition={{ duration: reduceMotion ? 0 : 0.2 }}
						className="fixed inset-0 z-40 bg-black/50"
						onClick={onClose}
					/>

					<motion.div
						initial={{ opacity: 0 }}
						animate={{ opacity: 1 }}
						exit={{ opacity: 0 }}
						transition={{ duration: reduceMotion ? 0 : 0.2 }}
						className="fixed inset-0 z-50 flex flex-col bg-base-100"
					>
						<div className="sticky top-0 z-10 border-b border-base-200 bg-base-100 pt-[env(safe-area-inset-top)]">
							<div className="flex min-h-12 items-center justify-between gap-3 px-3 py-2 sm:px-4">
								<div className="min-w-0 flex-1">
									<h1 className="text-lg font-bold leading-snug text-base-content sm:text-xl">
										Информация об анализе
									</h1>
								</div>
								<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть просмотр анализа" />
							</div>
						</div>

						<main className="mx-auto w-full max-w-3xl flex-1 overflow-y-auto overscroll-y-contain px-3 pb-8 pt-3 sm:px-4 sm:pb-10 sm:pt-4">
							{renderTabContent()}
						</main>

						<footer className="border-t border-base-200 bg-base-100 pb-[max(0.5rem,env(safe-area-inset-bottom))]">
							<div className="grid grid-cols-3 gap-0.5 p-1.5 sm:gap-1 sm:p-2">
								{tabs.map((tab) => {
									const Icon = tab.icon
									return (
										<button
											key={tab.id}
											type="button"
											onClick={() => setActiveTab(tab.id)}
											className={`flex min-h-12 cursor-pointer flex-col items-center justify-center gap-0.5 rounded-xl px-1 py-2 text-[11px] transition-colors duration-200 sm:min-h-0 sm:gap-1 sm:py-2 sm:text-xs ${
												activeTab === tab.id
													? "bg-primary/10 font-semibold text-primary"
													: "text-base-content/60 active:bg-base-200/80 sm:hover:bg-base-200/50"
											}`}
											aria-label={tab.label}
											aria-pressed={activeTab === tab.id}
										>
											<span className="text-lg sm:text-base" aria-hidden>
												<Icon />
											</span>
											<span className="max-w-full truncate px-0.5 text-center font-medium leading-tight">
												{tab.label}
											</span>
										</button>
									)
								})}
							</div>
						</footer>
					</motion.div>
				</>
			)}
		</AnimatePresence>,
		document.body
	)
}

export default AnalysisViewSheet
