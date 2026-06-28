import { useQuery, useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { FaRegFileAlt, FaRegStickyNote } from "react-icons/fa"
import { IoSettingsOutline } from "react-icons/io5"
import { MdOutlineAnalytics } from "react-icons/md"
import type { KalibriObject } from "@/api/analysis/types"
import { fetchWeedDetails } from "@/api/catalog"
import { fetchClassifications } from "@/api/catalog/classifications"
import { queryKeys } from "@/api/queryKeys"
import { canEditCatalogItem } from "@/components/catalog/components/CatalogProposalsShared"
import { CatalogItemFormLayout } from "@/components/catalog/layout/CatalogItemFormLayout"
import AnalysisTabView from "@/components/catalog/views/AnalysisTabView"
import CharacteristicsTabView from "@/components/catalog/views/CharacteristicsTabView"
import GeneralInfoTabView from "@/components/catalog/views/GeneralInfoTabView"
import NotesTabView from "@/components/catalog/views/NotesTabView"
import ErrorPage from "@/components/common/layout/ErrorPage"
import { Button } from "@/components/common/ui/Button"
import Loading from "@/components/common/ui/Loading"
import { useAuth } from "@/contexts/AuthContext"
import { useAnalysesWithObjects } from "@/hooks/useAnalysesWithObjects"
import { resetScroll } from "@/utils/scroll"

enum Tab {
	GENERAL = "Основное",
	CHARACTERISTICS = "Характеристики",
	ANALYTICS = "Анализы",
	NOTES = "Заметки",
}

export const Route = createFileRoute("/_authenticated/catalog/$catalogItemId/")({
	component: CatalogItemView,
})

function CatalogItemView() {
	const navigate = useNavigate()
	const queryClient = useQueryClient()
	const { catalogItemId } = Route.useParams()
	const { roles } = useAuth()
	const showEditButton = canEditCatalogItem(roles)

	const [activeTab, setActiveTab] = useState<Tab>(Tab.GENERAL)
	const [isOpeningEdit, setIsOpeningEdit] = useState(false)

	const [allObjectsData, setAllObjectsData] = useState<Record<string, KalibriObject[]>>({})

	const {
		isPending,
		error,
		data: item,
	} = useQuery({
		queryKey: queryKeys.catalog.weedDetails(catalogItemId),
		queryFn: () => fetchWeedDetails(Number(catalogItemId)),
		enabled: !!catalogItemId,
	})

	const { analyses: analysesData, objectsData: fetchedObjectsData } = useAnalysesWithObjects(
		catalogItemId,
		item?.analyses ?? [],
		{
			enabled: !!item && (item.analyses?.length ?? 0) > 0,
		}
	)

	useEffect(() => {
		if (Object.keys(fetchedObjectsData).length > 0) {
			setAllObjectsData(fetchedObjectsData)
		} else if (!item?.analyses?.length) {
			setAllObjectsData({})
		}
	}, [fetchedObjectsData, item?.analyses])

	const handleObjectsDataChange = (newData: Record<string, KalibriObject[]>) => {
		setAllObjectsData((prev) => ({ ...prev, ...newData }))
	}

	const handleBack = () => {
		navigate({ to: "/catalog" })
	}

	const handleEditOpen = async () => {
		setIsOpeningEdit(true)
		try {
			await Promise.allSettled([
				queryClient.prefetchQuery({
					queryKey: queryKeys.catalog.weedDetails(catalogItemId),
					queryFn: () => fetchWeedDetails(Number(catalogItemId)),
				}),
				queryClient.prefetchQuery({
					queryKey: queryKeys.classifications.all,
					queryFn: fetchClassifications,
				}),
			])
		} finally {
			navigate({
				to: "/catalog/$catalogItemId/edit",
				params: { catalogItemId },
			})
		}
	}

	const renderTabContent = () => {
		if (!item) return null

		switch (activeTab) {
			case Tab.GENERAL: {
				return (
					<GeneralInfoTabView
						images={item.images}
						description={item.description}
						mainGroup={item.main_group ?? null}
						mainSubgroup={item.main_subgroup ?? null}
						subgroup={item.subgroup ?? null}
						isQuarantine={item.is_quarantine ?? false}
						harmfulness={item.harmfulness ?? undefined}
					/>
				)
			}
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
		{ id: Tab.GENERAL, label: Tab.GENERAL, icon: <FaRegFileAlt /> },
		{ id: Tab.CHARACTERISTICS, label: Tab.CHARACTERISTICS, icon: <IoSettingsOutline /> },
		{ id: Tab.ANALYTICS, label: Tab.ANALYTICS, icon: <MdOutlineAnalytics /> },
		{ id: Tab.NOTES, label: Tab.NOTES, icon: <FaRegStickyNote /> },
	]

	if (isPending) {
		return (
			<div className="flex min-h-0 flex-1 items-center justify-center bg-base-100">
				<Loading />
			</div>
		)
	}
	if (error) return <ErrorPage error={error} />
	if (!item) return <ErrorPage error={new Error("Элемент не найден")} />

	return (
		<CatalogItemFormLayout
			title={item.name}
			subtitle="Карточка каталога"
			onBack={handleBack}
			tabs={tabs}
			activeTabId={activeTab}
			onTabChange={(id) => {
				setActiveTab(id as Tab)
				resetScroll()
			}}
			footerActions={
				showEditButton ? (
					<Button
						type="button"
						fullWidth
						loading={isOpeningEdit}
						disabled={isOpeningEdit}
						onClick={handleEditOpen}
					>
						Изменить
					</Button>
				) : undefined
			}
		>
			{renderTabContent()}
		</CatalogItemFormLayout>
	)
}
