import type React from "react"
import { useMemo } from "react"
import { IoBarChartOutline } from "react-icons/io5"
import {
	Bar,
	BarChart,
	CartesianGrid,
	Cell,
	ResponsiveContainer,
	Tooltip,
	XAxis,
	YAxis,
} from "recharts"
import type { KalibriObject } from "@/api/analysis/types"
import { useMatchMedia } from "@/hooks/useMatchMedia"
import {
	buildFieldHistogram,
	buildMass1000Histogram,
	type HistogramBin,
} from "@/utils/analysisDistributionData"
import type { FieldStatisticsResult } from "@/utils/analysisStatistics"
import AnalysisSheetSection from "./AnalysisSheetSection"

interface DistributionChartsSectionProps {
	objects: KalibriObject[]
	fieldStats: FieldStatisticsResult
}

const DEFAULT_BAR = "#2563eb"

const ChartBlock: React.FC<{
	title: string
	bins: HistogramBin[]
	emptyText: string
	compact: boolean
}> = ({ title, bins, emptyText, compact }) => {
	if (bins.length === 0) {
		return (
			<div className="rounded-2xl border border-dashed border-base-300 bg-base-200/30 p-6 text-center">
				<p className="text-sm text-base-content/60">{emptyText}</p>
			</div>
		)
	}

	const data = bins.map((b) => ({
		name: b.label,
		count: b.count,
		fill: b.color ?? DEFAULT_BAR,
	}))

	const total = bins.reduce((s, b) => s + b.count, 0)

	const manyBins = data.length > 6
	const xInterval = compact && manyBins ? ("preserveStartEnd" as const) : 0
	const xAngle = compact ? (manyBins ? 0 : -28) : -35
	const xHeight = compact ? (manyBins ? 32 : 40) : 48
	const chartHeight = compact ? "min-h-[220px] h-[220px]" : "min-h-[208px] h-52"
	const yAxisW = compact ? 30 : 36

	return (
		<div className="rounded-2xl border border-base-200/90 bg-base-100 p-3 shadow-sm sm:p-4">
			<h3 className="mb-2 px-0.5 text-sm font-semibold leading-snug text-base-content sm:px-1">
				{title}
			</h3>
			<div className={`w-full ${chartHeight}`}>
				<ResponsiveContainer width="100%" height="100%">
					<BarChart
						data={data}
						margin={{
							top: 8,
							right: compact ? 4 : 8,
							left: compact ? 0 : 0,
							bottom: compact ? 24 : 32,
						}}
					>
						<CartesianGrid strokeDasharray="3 3" className="stroke-base-300" />
						<XAxis
							dataKey="name"
							tick={{ fontSize: compact ? 9 : 10 }}
							interval={xInterval}
							angle={xAngle}
							textAnchor={compact && manyBins ? "middle" : "end"}
							height={xHeight}
						/>
						<YAxis allowDecimals={false} tick={{ fontSize: compact ? 10 : 11 }} width={yAxisW} />
						<Tooltip
							contentStyle={{
								borderRadius: 8,
								border: "1px solid #e5e7eb",
							}}
							formatter={(value) => [value ?? "—", "Количество"]}
						/>
						<Bar dataKey="count" radius={[4, 4, 0, 0]} isAnimationActive={false}>
							{data.map((entry) => (
								<Cell key={entry.name} fill={entry.fill} />
							))}
						</Bar>
					</BarChart>
				</ResponsiveContainer>
			</div>
			<div className="mt-3 overflow-x-auto [-webkit-overflow-scrolling:touch]">
				<table className="table table-sm w-full">
					<caption className="sr-only">
						{title}: таблица частот по интервалам, всего {total} объектов
					</caption>
					<thead>
						<tr className="text-xs text-base-content/70 sm:text-[11px]">
							<th className="text-left">Интервал</th>
							<th className="text-right">Количество</th>
						</tr>
					</thead>
					<tbody>
						{bins.map((b) => (
							<tr key={b.label} className="border-t border-base-200">
								<td className="font-mono text-xs sm:text-[11px]">{b.label}</td>
								<td className="text-right font-mono text-xs sm:text-[11px]">{b.count}</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</div>
	)
}

const DistributionChartsSection: React.FC<DistributionChartsSectionProps> = ({
	objects,
	fieldStats,
}) => {
	const compactCharts = useMatchMedia("(max-width: 639px)")
	const lengthBins = useMemo(() => {
		const s = fieldStats.l
		if (!s || objects.length === 0) return []
		return buildFieldHistogram(objects, "l", s.min, s.max)
	}, [objects, fieldStats.l])

	const widthBins = useMemo(() => {
		const s = fieldStats.w
		if (!s || objects.length === 0) return []
		return buildFieldHistogram(objects, "w", s.min, s.max)
	}, [objects, fieldStats.w])

	const lwBins = useMemo(() => {
		const s = fieldStats.l_w
		if (!s || objects.length === 0) return []
		return buildFieldHistogram(objects, "l_w", s.min, s.max)
	}, [objects, fieldStats.l_w])

	const sqBins = useMemo(() => {
		const s = fieldStats.sq
		if (!s || objects.length === 0) return []
		return buildFieldHistogram(objects, "sq", s.min, s.max)
	}, [objects, fieldStats.sq])

	const massBins = useMemo(() => buildMass1000Histogram(objects), [objects])

	if (objects.length === 0) {
		return null
	}

	return (
		<AnalysisSheetSection
			title="Графики распределения"
			subtitle="Гистограммы по измерениям объектов"
			icon={<IoBarChartOutline size={21} />}
			accent="info"
		>
			<p className="mb-3 text-[11px] leading-snug text-base-content/55 sm:hidden">
				На узком экране подписи оси сокращены; см. таблицу под графиком для точных интервалов.
			</p>
			<div className="grid gap-3 sm:gap-4 md:grid-cols-2">
				<ChartBlock
					title="Длина"
					bins={lengthBins}
					emptyText="График распределения длины отсутствует"
					compact={compactCharts}
				/>
				<ChartBlock
					title="Ширина"
					bins={widthBins}
					emptyText="График распределения ширины отсутствует"
					compact={compactCharts}
				/>
				<ChartBlock
					title="Соотношение L/W"
					bins={lwBins}
					emptyText="График распределения соотношения L/W отсутствует"
					compact={compactCharts}
				/>
				<ChartBlock
					title="Площадь (SQ)"
					bins={sqBins}
					emptyText="График распределения площади отсутствует"
					compact={compactCharts}
				/>
				{massBins.length > 0 ? (
					<ChartBlock
						title="Масса 1000 зёрен"
						bins={massBins}
						emptyText=""
						compact={compactCharts}
					/>
				) : null}
			</div>
		</AnalysisSheetSection>
	)
}

export default DistributionChartsSection
