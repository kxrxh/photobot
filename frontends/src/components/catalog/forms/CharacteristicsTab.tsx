import { type FC, useId, useMemo } from "react"
import { MdOutlineAnalytics } from "react-icons/md"
import type { Analysis, KalibriObject } from "@/api/analysis/types"
import StatTable from "@/components/analysis/dialogs/AnalysisViewSheet/components/StatTable"
import { calculateStatsFromObjects } from "@/utils/stats"

/** Matches `GeneralInfoTab` / `UploadTab` section titles */
const SECTION_TITLE_CLASS =
	"mb-1.5 block text-xs font-semibold uppercase tracking-wide text-base-content/45"

interface CharacteristicsTabProps {
	analyses: Analysis[]
	onSelectAnalysis: () => void
	allObjectsData: Record<string, KalibriObject[]>
	excludedObjects: number[]
	readOnly?: boolean
}

const metricGroups = [
	{
		title: "Геометрические характеристики",
		metrics: [
			{ key: "length", title: "Длина", unit: "мм" },
			{ key: "width", title: "Ширина", unit: "мм" },
			{ key: "sq", title: "Площадь", unit: "мм²" },
			{ key: "lwRatio", title: "Отношение Д/Ш", unit: "" },
		],
	},
	{
		title: "Цветовые характеристики (RGB)",
		metrics: [
			{ key: "r", title: "Красный (R)", unit: "" },
			{ key: "g", title: "Зеленый (G)", unit: "" },
			{ key: "b", title: "Синий (B)", unit: "" },
		],
	},
	{
		title: "Цветовые характеристики (HSV)",
		metrics: [
			{ key: "h", title: "Оттенок (H)", unit: "" },
			{ key: "s", title: "Насыщенность (S)", unit: "" },
			{ key: "v", title: "Яркость (V)", unit: "" },
		],
	},
] as const

function hasAnyMetricData(
	stats: ReturnType<typeof calculateStatsFromObjects>
): stats is NonNullable<typeof stats> {
	if (!stats) return false
	return metricGroups.some((group) =>
		group.metrics.some((metric) => Boolean(stats[metric.key as keyof typeof stats]))
	)
}

const CharacteristicsTab: FC<CharacteristicsTabProps> = ({
	analyses,
	onSelectAnalysis,
	allObjectsData,
	excludedObjects,
	readOnly = false,
}) => {
	const sectionTitleId = useId()
	const stats = useMemo(
		() => calculateStatsFromObjects(allObjectsData, excludedObjects),
		[allObjectsData, excludedObjects]
	)

	const activeObjectsCount = useMemo(() => {
		const allObjects = Object.values(allObjectsData).flat()
		return allObjects.filter((obj) => !excludedObjects.includes(obj.id)).length
	}, [allObjectsData, excludedObjects])

	const hasData = hasAnyMetricData(stats)

	const emptyDescription = useMemo(() => {
		if (analyses.length === 0) {
			return "На вкладке «Анализы» выберите один или несколько анализов — таблицы заполнятся по объектам автоматически."
		}
		return "Сейчас в расчёте нет объектов: откройте «Анализы», дождитесь загрузки или уберите лишние исключения."
	}, [analyses.length])

	return (
		<div className="animate-fadeIn space-y-6 px-4 pb-8">
			<section aria-labelledby={sectionTitleId} className="flex flex-col gap-3">
				<div className="pt-1">
					<h2 id={sectionTitleId} className={SECTION_TITLE_CLASS}>
						Характеристики
					</h2>
					<p className="text-xs leading-relaxed text-base-content/55">
						Сводные минимум, максимум, медиана и среднее по выбранным анализам.
					</p>
				</div>

				{!hasData ? (
					readOnly ? (
						<div className="rounded-2xl border border-base-200 bg-base-200/20 px-4 py-5 text-center">
							<p className="text-sm leading-relaxed text-base-content/70">
								Пока нет данных для таблиц. После выбора анализов характеристики появятся здесь. В
								режиме просмотра заявки изменить состав анализов нельзя.
							</p>
						</div>
					) : (
						<div className="overflow-hidden rounded-2xl border border-primary/20 bg-linear-to-br from-primary/[0.07] via-base-200/30 to-base-200/50 shadow-sm">
							<div className="flex flex-col items-center gap-4 px-4 py-8 text-center sm:px-6">
								<div
									className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/15 text-primary shadow-inner"
									aria-hidden
								>
									<MdOutlineAnalytics className="h-8 w-8" />
								</div>
								<div className="max-w-sm space-y-2">
									<h3 className="text-base font-semibold text-base-content">
										Данные для расчёта не выбраны
									</h3>
									<p className="text-sm leading-relaxed text-base-content/65">{emptyDescription}</p>
								</div>
								<button
									type="button"
									onClick={onSelectAnalysis}
									className="btn btn-primary btn-md mt-1 w-full max-w-xs gap-2 rounded-xl shadow-sm"
								>
									<MdOutlineAnalytics className="h-5 w-5 shrink-0 opacity-90" aria-hidden />
									Перейти к анализам
								</button>
							</div>
						</div>
					)
				) : (
					<>
						<div className="space-y-3 rounded-2xl border border-base-200 bg-base-200/25 p-4">
							<div className="flex flex-wrap items-center gap-2">
								<span className="badge badge-sm border border-base-300/60 bg-base-100/80 font-medium text-base-content/85">
									Анализов: {analyses.length}
								</span>
								<span className="badge badge-sm badge-primary badge-outline font-medium">
									Объектов в расчёте: {activeObjectsCount}
								</span>
							</div>
							<p className="text-xs leading-relaxed text-base-content/55">
								Показатели считаются по объектам выбранных анализов с учётом исключений на вкладке
								«Анализы».
							</p>
							{!readOnly ? (
								<button
									type="button"
									onClick={onSelectAnalysis}
									className="btn btn-ghost btn-sm h-9 min-h-9 gap-1.5 px-2 font-medium text-primary hover:bg-primary/10"
								>
									<MdOutlineAnalytics className="h-4 w-4" aria-hidden />
									Изменить состав анализов
								</button>
							) : null}
						</div>

						<div className="space-y-5">
							{metricGroups.map((group) => {
								const groupHasData = group.metrics.some((metric) =>
									Boolean(stats[metric.key as keyof typeof stats])
								)
								if (!groupHasData) return null

								return (
									<div
										key={group.title}
										className="space-y-3 rounded-2xl border border-base-200 bg-base-200/25 p-4"
									>
										<h3 className="text-xs font-semibold uppercase tracking-wide text-base-content/45">
											{group.title}
										</h3>
										<div className="space-y-5">
											{group.metrics.map((metric) => {
												const statData = stats[metric.key as keyof typeof stats]
												return statData ? (
													<StatTable
														key={metric.key}
														title={metric.title}
														data={statData}
														unit={metric.unit}
													/>
												) : null
											})}
										</div>
									</div>
								)
							})}
						</div>
					</>
				)}
			</section>
		</div>
	)
}

export default CharacteristicsTab
