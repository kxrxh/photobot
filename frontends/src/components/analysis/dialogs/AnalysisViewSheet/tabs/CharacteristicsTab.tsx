import type React from "react"
import { useMemo } from "react"
import { FaTable } from "react-icons/fa"
import { IoIosStats } from "react-icons/io"
import { buildPdfLikeKpis } from "@/utils/analysisKpis"
import {
	calculateClassStatisticsResult,
	calculateFieldStatisticsResult,
} from "@/utils/analysisStatistics"
import AnalysisKpiStrip from "../components/AnalysisKpiStrip"
import AnalysisSheetSection from "../components/AnalysisSheetSection"
import ClassificationSection from "../components/ClassificationSection"
import DistributionChartsSection from "../components/DistributionChartsSection"
import ObjectAggregateTables from "../components/ObjectAggregateTables"
import type { TabComponentProps } from "../types"

const CharacteristicsTab: React.FC<TabComponentProps> = ({
	analysis,
	objects,
	objectsLoading,
	objectsError,
}) => {
	const list = objects ?? []

	const fieldStats = useMemo(() => calculateFieldStatisticsResult(list), [list])

	const classStats = useMemo(() => calculateClassStatisticsResult(list), [list])

	const kpis = useMemo(
		() => buildPdfLikeKpis(analysis, objectsLoading ? undefined : objects, fieldStats),
		[analysis, objects, objectsLoading, fieldStats]
	)

	const showObjectBlocks = !objectsLoading && !objectsError && objects !== undefined

	return (
		<div className="space-y-5 sm:space-y-6">
			{objectsLoading ? (
				<p className="flex items-center gap-2 text-sm text-base-content/55">
					<span className="loading loading-spinner loading-sm text-primary" />
					Загрузка объектов для таблиц и графиков…
				</p>
			) : null}

			{objectsError ? (
				<div
					className="rounded-2xl border border-error/30 bg-error/10 px-4 py-3 text-sm text-error"
					role="alert"
				>
					Не удалось загрузить объекты для статистики и графиков. Ниже — параметры только из ответа
					анализа (если есть).
				</div>
			) : null}

			<AnalysisSheetSection
				title="Основные показатели"
				subtitle="Ключевые величины по данным анализа"
				icon={<IoIosStats size={20} />}
				accent="primary"
			>
				{kpis.cards.length > 0 ? (
					<AnalysisKpiStrip kpis={kpis} />
				) : (
					<p className="rounded-2xl border border-dashed border-base-300/80 bg-base-100/80 px-4 py-6 text-center text-sm text-base-content/60">
						Нет доступных показателей для отображения.
					</p>
				)}
			</AnalysisSheetSection>

			{showObjectBlocks ? (
				<AnalysisSheetSection
					title="Статистика по объектам"
					subtitle="Минимум, максимум, среднее и медиана по распознанным объектам"
					icon={<FaTable size={17} />}
					accent="primary"
				>
					<ObjectAggregateTables fieldStats={fieldStats} />
				</AnalysisSheetSection>
			) : null}

			{showObjectBlocks && classStats ? <ClassificationSection classStats={classStats} /> : null}

			{showObjectBlocks && list.length > 0 ? (
				<DistributionChartsSection objects={list} fieldStats={fieldStats} />
			) : null}
		</div>
	)
}

export default CharacteristicsTab
