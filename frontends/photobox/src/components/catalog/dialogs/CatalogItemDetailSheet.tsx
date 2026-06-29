import { useQuery } from "@tanstack/react-query"
import { useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { FaRegFileAlt, FaRegStickyNote } from "react-icons/fa"
import { IoClose, IoSettingsOutline } from "react-icons/io5"
import { MdOutlineAnalytics } from "react-icons/md"
import type { KalibriObject } from "@/api/analysis/types"
import { fetchWeedDetails } from "@/api/catalog"
import type { WeedDetails } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import AnalysisTabView from "@/components/catalog/views/AnalysisTabView"
import CharacteristicsTabView from "@/components/catalog/views/CharacteristicsTabView"
import GeneralInfoTabView from "@/components/catalog/views/GeneralInfoTabView"
import NotesTabView from "@/components/catalog/views/NotesTabView"
import ErrorPage from "@/components/common/layout/ErrorPage"
import Loading from "@/components/common/ui/Loading"
import { useAnalysesWithObjects } from "@/hooks/useAnalysesWithObjects"

enum Tab {
	GENERAL = "Основное",
	CHARACTERISTICS = "Характеристики",
	ANALYTICS = "Анализы",
	NOTES = "Заметки",
}

interface CatalogItemDetailSheetProps {
	isOpen: boolean
	onClose: () => void
	catalogItemId: number | null
}

export default function CatalogItemDetailSheet({
	isOpen,
	onClose,
	catalogItemId,
}: CatalogItemDetailSheetProps) {
	const [activeTab, setActiveTab] = useState<Tab>(Tab.GENERAL)
	const mainRef = useRef<HTMLElement>(null)
	const dialogRef = useRef<HTMLDialogElement>(null)

	const {
		isPending,
		error,
		data: item,
	} = useQuery<WeedDetails>({
		queryKey: queryKeys.catalog.weedDetails(String(catalogItemId ?? "")),
		queryFn: () => fetchWeedDetails(Number(catalogItemId)),
		enabled: !!catalogItemId && isOpen,
	})

	const { analyses: analysesData, objectsData: fetchedObjectsData } = useAnalysesWithObjects(
		String(catalogItemId ?? ""),
		item?.analyses ?? [],
		{
			enabled: isOpen && !!catalogItemId && !!item && (item.analyses?.length ?? 0) > 0,
		}
	)

	const [allObjectsData, setAllObjectsData] = useState<Record<string, KalibriObject[]>>({})

	useEffect(() => {
		if (dialogRef.current) {
			if (isOpen) {
				dialogRef.current.showModal()
				document.body.style.overflow = "hidden"
			} else {
				dialogRef.current.close()
				document.body.style.overflow = "unset"
			}
		}

		return () => {
			document.body.style.overflow = "unset"
		}
	}, [isOpen])

	useEffect(() => {
		if (Object.keys(fetchedObjectsData).length > 0) {
			setAllObjectsData(fetchedObjectsData)
		}
	}, [fetchedObjectsData])

	useEffect(() => {
		if (isOpen) {
			setActiveTab(Tab.GENERAL)
		} else {
			setAllObjectsData({})
		}
	}, [isOpen])

	const handleObjectsDataChange = (newData: Record<string, KalibriObject[]>) => {
		setAllObjectsData((prev) => ({ ...prev, ...newData }))
	}

	const renderTabContent = () => {
		if (!item) return null

		switch (activeTab) {
			case Tab.GENERAL:
				return (
					<GeneralInfoTabView
						images={item.images}
						description={item.description}
						mainGroup={item.main_group}
						mainSubgroup={item.main_subgroup}
						subgroup={item.subgroup}
						harmfulness={item.harmfulness}
						isQuarantine={item.is_quarantine}
					/>
				)
			case Tab.CHARACTERISTICS:
				return <CharacteristicsTabView item={item} />
			case Tab.ANALYTICS:
				return (
					<AnalysisTabView
						selectedAnalyses={analysesData}
						allObjectsData={allObjectsData}
						onObjectsDataChange={handleObjectsDataChange}
						excludedObjects={item.statistics?.excluded_objects || []}
					/>
				)
			case Tab.NOTES:
				return <NotesTabView catalogItemId={item.id.toString()} />
			default:
				return null
		}
	}

	const tabs = [
		{ id: Tab.GENERAL, icon: <FaRegFileAlt /> },
		{ id: Tab.CHARACTERISTICS, icon: <IoSettingsOutline /> },
		{ id: Tab.ANALYTICS, icon: <MdOutlineAnalytics /> },
		{ id: Tab.NOTES, icon: <FaRegStickyNote /> },
	]

	return createPortal(
		<dialog ref={dialogRef} className="modal">
			<div className="flex flex-col w-full h-full bg-base-100">
				<header className="flex sticky top-0 z-20 justify-between items-center p-2 backdrop-blur-sm bg-base-100/95">
					<h1 className="text-xl font-bold truncate">
						{item ? item.name : isPending ? "Загрузка..." : "Ошибка"}
					</h1>
					<button
						type="button"
						onClick={onClose}
						className="btn btn-ghost btn-circle btn-sm"
						aria-label="Закрыть"
					>
						<IoClose size={24} />
					</button>
				</header>

				<main ref={mainRef} className="overflow-y-auto flex-1 mx-auto w-full max-w-md pb-18">
					{isPending && <Loading />}
					{error && <ErrorPage error={error} />}
					{item && renderTabContent()}
				</main>
				{item && (
					<footer className="sticky bottom-0 z-10 w-full border-t shadow-2xl backdrop-blur-sm bg-base-100/95 border-base-300">
						<div className="mx-auto max-w-md">
							<div className="p-2 space-y-2">
								<div className="flex justify-between">
									{tabs.map((tab) => (
										<button
											key={tab.id}
											role="tab"
											type="button"
											className={`flex flex-col items-center justify-center gap-1 text-xs transition-all duration-200 p-2 rounded-lg ${
												activeTab === tab.id ? "text-primary font-bold" : "text-base-content/60"
											}`}
											onClick={() => {
												setActiveTab(tab.id)
											}}
										>
											<span className="text-base leading-none">{tab.icon}</span>
											<span className="font-medium whitespace-nowrap">{tab.id}</span>
										</button>
									))}
								</div>
							</div>
						</div>
					</footer>
				)}
			</div>
		</dialog>,
		document.body
	)
}
