import type React from "react"
import { useMemo } from "react"
import { FaList } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import { getObjectImageUrl } from "@/utils/image"
import AnalysisSheetSection from "../components/AnalysisSheetSection"
import ObjectGrid, { type GridObject as GridAnalysisObject } from "../components/ObjectGrid"
import type { TabComponentProps } from "../types"

const ObjectsTab: React.FC<TabComponentProps> = ({ objects, objectsLoading, objectsError }) => {
	const gridObjects: GridAnalysisObject[] = useMemo(() => {
		if (!objects) return []
		return objects.map((obj) => ({
			id: obj.id,
			image: getObjectImageUrl(obj) ?? undefined,
		}))
	}, [objects])

	let content: React.ReactNode
	if (objectsLoading) {
		content = (
			<div className="flex h-48 items-center justify-center text-primary">
				<span className="loading loading-spinner loading-md" />
			</div>
		)
	} else if (objectsError || objects === undefined) {
		content = (
			<div className="flex flex-col items-center justify-center py-12 text-center">
				<div className="w-16 h-16 rounded-full bg-base-200 flex items-center justify-center mb-3">
					<IoImageOutline size={32} className="text-base-content/40" />
				</div>
				<p className="text-sm font-medium text-base-content/70">Ошибка загрузки объектов</p>
				<p className="text-xs text-base-content/50 mt-1">
					Не удалось получить данные об объектах для этого анализа.
				</p>
			</div>
		)
	} else if (objects.length === 0) {
		content = (
			<div className="flex flex-col items-center justify-center py-12 text-center">
				<div className="w-16 h-16 rounded-full bg-base-200 flex items-center justify-center mb-3">
					<IoImageOutline size={32} className="text-base-content/40" />
				</div>
				<p className="text-sm font-medium text-base-content/70">Объекты не найдены</p>
				<p className="text-xs text-base-content/50 mt-1">
					Для данного анализа не найдено ни одного объекта.
				</p>
			</div>
		)
	} else {
		content = <ObjectGrid objects={gridObjects} />
	}

	return (
		<AnalysisSheetSection
			title="Список объектов"
			subtitle="Превью распознанных объектов анализа"
			icon={<FaList size={17} />}
			accent="primary"
			headerExtra={
				!objectsLoading && objects && objects.length > 0 ? (
					<span className="badge badge-primary badge-sm">
						{objects.length}{" "}
						{objects.length === 1 ? "объект" : objects.length < 5 ? "объекта" : "объектов"}
					</span>
				) : undefined
			}
		>
			{content}
		</AnalysisSheetSection>
	)
}

export default ObjectsTab
