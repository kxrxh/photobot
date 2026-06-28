import type { WeedStatistics } from "@/api/catalog/types"

interface CharacteristicsTabViewProps {
	item: {
		length: number
		width: number
		statistics?: WeedStatistics
	}
}

interface StatDisplayProps {
	title: string
	stats: {
		avg: number
		median: number
		min: number
		max: number
	}
	unit?: string
}

const StatDisplay = ({ title, stats, unit = "" }: StatDisplayProps) => {
	const formatValue = (value: number) => {
		return `${value.toFixed(2)}${unit}`
	}

	return (
		<div className="mb-6">
			<h3 className="mb-3 font-semibold text-base text-base-content/90 flex items-center gap-2">
				<span className="w-1 h-4 bg-primary rounded-full" />
				{title}
			</h3>
			<div className="border border-base-300 rounded-lg overflow-hidden bg-base-100 shadow-sm">
				<table className="table w-full">
					<tbody>
						<tr className="bg-base-200/50 border-b border-base-300">
							<td className="font-medium text-sm text-base-content/70 px-4 py-2 border-r border-base-300">
								Минимум
							</td>
							<td className="font-medium text-sm text-base-content/70 px-4 py-2">Максимум</td>
						</tr>
						<tr className="border-b border-base-300/50">
							<td className="px-4 py-2 font-mono text-sm border-r border-base-300">
								{formatValue(stats.min)}
							</td>
							<td className="px-4 py-2 font-mono text-sm">{formatValue(stats.max)}</td>
						</tr>
						<tr className="bg-base-200/50 border-b border-base-300">
							<td className="font-medium text-sm text-base-content/70 px-4 py-2 border-r border-base-300">
								Медиана
							</td>
							<td className="font-medium text-sm text-base-content/70 px-4 py-2">Среднее</td>
						</tr>
						<tr>
							<td className="px-4 py-2 font-mono text-sm border-r border-base-300">
								{formatValue(stats.median)}
							</td>
							<td className="px-4 py-2 font-mono text-sm">{formatValue(stats.avg)}</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	)
}

export default function CharacteristicsTabView({ item }: CharacteristicsTabViewProps) {
	const hasStatistics = item.statistics

	if (!hasStatistics) {
		return (
			<div className="p-4 text-center text-base-content/60">Нет характеристик для отображения.</div>
		)
	}

	return (
		<div className="p-4 space-y-6">
			{hasStatistics && item.statistics && (
				<div>
					<h2 className="mb-4 text-lg font-semibold">Статистические характеристики</h2>

					<StatDisplay
						title="Длина (мм)"
						stats={{
							avg: item.statistics.l_avg,
							median: item.statistics.l_median,
							min: item.statistics.l_min,
							max: item.statistics.l_max,
						}}
						unit=" мм"
					/>

					<StatDisplay
						title="Ширина (мм)"
						stats={{
							avg: item.statistics.w_avg,
							median: item.statistics.w_median,
							min: item.statistics.w_min,
							max: item.statistics.w_max,
						}}
						unit=" мм"
					/>

					<StatDisplay
						title="Площадь (мм²)"
						stats={{
							avg: item.statistics.sq_avg,
							median: item.statistics.sq_median,
							min: item.statistics.sq_min,
							max: item.statistics.sq_max,
						}}
						unit=" мм²"
					/>

					<StatDisplay
						title="Отношение длины к ширине"
						stats={{
							avg: item.statistics.lw_avg,
							median: item.statistics.lw_median,
							min: item.statistics.lw_min,
							max: item.statistics.lw_max,
						}}
					/>

					<StatDisplay
						title="Красный (R)"
						stats={{
							avg: item.statistics.r_avg,
							median: item.statistics.r_median,
							min: item.statistics.r_min,
							max: item.statistics.r_max,
						}}
					/>

					<StatDisplay
						title="Зеленый (G)"
						stats={{
							avg: item.statistics.g_avg,
							median: item.statistics.g_median,
							min: item.statistics.g_min,
							max: item.statistics.g_max,
						}}
					/>

					<StatDisplay
						title="Синий (B)"
						stats={{
							avg: item.statistics.b_avg,
							median: item.statistics.b_median,
							min: item.statistics.b_min,
							max: item.statistics.b_max,
						}}
					/>

					<StatDisplay
						title="Оттенок (H)"
						stats={{
							avg: item.statistics.h_avg,
							median: item.statistics.h_median,
							min: item.statistics.h_min,
							max: item.statistics.h_max,
						}}
						unit="°"
					/>

					<StatDisplay
						title="Насыщенность (S)"
						stats={{
							avg: item.statistics.s_avg,
							median: item.statistics.s_median,
							min: item.statistics.s_min,
							max: item.statistics.s_max,
						}}
						unit="%"
					/>

					<StatDisplay
						title="Яркость (V)"
						stats={{
							avg: item.statistics.v_avg,
							median: item.statistics.v_median,
							min: item.statistics.v_min,
							max: item.statistics.v_max,
						}}
						unit="%"
					/>

					<StatDisplay
						title="Яркость"
						stats={{
							avg: item.statistics.brt_avg,
							median: item.statistics.brt_median,
							min: item.statistics.brt_min,
							max: item.statistics.brt_max,
						}}
					/>

					<StatDisplay
						title="Плотность"
						stats={{
							avg: item.statistics.solid_avg,
							median: item.statistics.solid_median,
							min: item.statistics.solid_min,
							max: item.statistics.solid_max,
						}}
					/>

					<StatDisplay
						title="Отношение площади к окружности"
						stats={{
							avg: item.statistics.sq_sqcrl_avg,
							median: item.statistics.sq_sqcrl_median,
							min: item.statistics.sq_sqcrl_min,
							max: item.statistics.sq_sqcrl_max,
						}}
					/>
				</div>
			)}
		</div>
	)
}
