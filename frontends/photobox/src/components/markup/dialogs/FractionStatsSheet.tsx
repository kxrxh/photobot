import type React from "react"
import { useMemo } from "react"
import type { KalibriObject } from "@/api/analysis/types"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import type { FractionType } from "@/hooks/useFractions"
import { calculateMetricStats } from "@/utils/stats"
import StatTable from "../../analysis/dialogs/AnalysisViewSheet/components/StatTable"

interface FractionStatsSheetProps {
	isOpen: boolean
	onClose: () => void
	fraction: FractionType | null
}

interface MetricConfig {
	key: string
	title: string
	unit: string
}

interface MetricGroup {
	title: string
	metrics: MetricConfig[]
}

// Type for object properties that can be extracted as metrics
type ExtractableProperty = "l" | "w" | "sq" | "r" | "g" | "b" | "h" | "s" | "v"

const METRIC_GROUPS: MetricGroup[] = [
	{
		title: "Геометрические характеристики",
		metrics: [
			{ key: "length", title: "Длина", unit: "мм" },
			{ key: "width", title: "Ширина", unit: "мм" },
			{ key: "area", title: "Площадь", unit: "мм²" },
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
]

/**
 * Extracts valid numeric values from objects for a given property
 */
const extractValidValues = (objects: KalibriObject[], property: ExtractableProperty): number[] => {
	return objects
		.map((obj) => obj[property])
		.filter((val): val is number => typeof val === "number" && !Number.isNaN(val))
}

/**
 * Calculates length-to-width ratio for objects
 */
const calculateLWRatio = (objects: KalibriObject[]): number[] => {
	return objects
		.map((obj) => {
			const { l: length, w: width } = obj
			return typeof length === "number" && typeof width === "number" && width !== 0
				? length / width
				: Number.NaN
		})
		.filter((val): val is number => typeof val === "number" && !Number.isNaN(val))
}

/**
 * Custom hook for calculating fraction statistics
 */
const useFractionStats = (fraction: FractionType | null) => {
	return useMemo(() => {
		if (!fraction?.objects?.length) {
			return null
		}

		const { objects } = fraction

		// Extract values for each metric
		const lengthValues = extractValidValues(objects, "l")
		const widthValues = extractValidValues(objects, "w")
		const areaValues = extractValidValues(objects, "sq")
		const lwRatioValues = calculateLWRatio(objects)
		const rValues = extractValidValues(objects, "r")
		const gValues = extractValidValues(objects, "g")
		const bValues = extractValidValues(objects, "b")
		const hValues = extractValidValues(objects, "h")
		const sValues = extractValidValues(objects, "s")
		const vValues = extractValidValues(objects, "v")

		// Calculate statistics for each metric that has data
		const calculateStats = (values: number[]) =>
			values.length > 0 ? calculateMetricStats(values) : undefined

		return {
			length: calculateStats(lengthValues),
			width: calculateStats(widthValues),
			area: calculateStats(areaValues),
			lwRatio: calculateStats(lwRatioValues),
			r: calculateStats(rValues),
			g: calculateStats(gValues),
			b: calculateStats(bValues),
			h: calculateStats(hValues),
			s: calculateStats(sValues),
			v: calculateStats(vValues),
		}
	}, [fraction])
}

/**
 * Component for rendering empty state
 */
const EmptyState: React.FC<{ activeObjectsCount: number }> = ({ activeObjectsCount }) => (
	<div className="p-6 text-center rounded-lg border shadow-sm bg-base-100 border-base-300">
		<p className="text-base-content/70">
			{activeObjectsCount === 0
				? "В этой фракции нет активных объектов."
				: "Статистические данные для этой фракции отсутствуют."}
		</p>
	</div>
)

/**
 * Component for rendering modal header
 */
const ModalHeader: React.FC<{
	fractionName: string
	activeObjectsCount: number
	onClose: () => void
}> = ({ fractionName, activeObjectsCount, onClose }) => (
	<div className="flex justify-between items-center mb-6">
		<div>
			<h3 className="text-lg font-bold">Статистика фракции</h3>
			<p className="text-sm text-base-content/70">
				{fractionName} • {activeObjectsCount} объектов
			</p>
		</div>
		<SheetHeaderCloseButton onClick={onClose} />
	</div>
)

/**
 * Component for rendering metric group
 */
const MetricGroupSection: React.FC<{
	group: MetricGroup
	stats: Partial<Record<string, ReturnType<typeof calculateMetricStats>>>
}> = ({ group, stats }) => {
	const hasData = group.metrics.some((metric) => stats[metric.key])

	if (!hasData) {
		return null
	}

	return (
		<div>
			<h2 className="mb-4 text-lg font-semibold text-base-content">{group.title}</h2>
			<div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
				{group.metrics.map((metric) => {
					const metricData = stats[metric.key]
					return metricData ? (
						<StatTable key={metric.key} title={metric.title} data={metricData} unit={metric.unit} />
					) : null
				})}
			</div>
		</div>
	)
}

/**
 * Main modal component for displaying fraction statistics
 */
const FractionStatsSheet: React.FC<FractionStatsSheetProps> = ({ isOpen, onClose, fraction }) => {
	const stats = useFractionStats(fraction)

	if (!isOpen || !fraction) {
		return null
	}

	const activeObjectsCount = fraction.objects.length

	return (
		<div className="modal modal-open">
			<div className="modal-box w-11/12 max-w-5xl">
				<ModalHeader
					fractionName={fraction.name}
					activeObjectsCount={activeObjectsCount}
					onClose={onClose}
				/>

				{!stats || activeObjectsCount === 0 ? (
					<EmptyState activeObjectsCount={activeObjectsCount} />
				) : (
					<div className="space-y-6 max-h-96 overflow-y-auto">
						{METRIC_GROUPS.map((group) => (
							<MetricGroupSection key={group.title} group={group} stats={stats} />
						))}
					</div>
				)}

				<div className="modal-action">
					<button type="button" className="btn btn-block" onClick={onClose}>
						Закрыть
					</button>
				</div>
			</div>
		</div>
	)
}

export default FractionStatsSheet
