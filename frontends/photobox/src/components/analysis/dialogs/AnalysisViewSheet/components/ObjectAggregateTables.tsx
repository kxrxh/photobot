import type React from "react"
import type { FieldStatistics, FieldStatisticsResult } from "@/utils/analysisStatistics"

interface ObjectAggregateTablesProps {
	fieldStats: FieldStatisticsResult
}

function fmt(n: number): string {
	return n.toFixed(2)
}

function SizeRow({ label, s, withAsym }: { label: string; s: FieldStatistics; withAsym: boolean }) {
	return (
		<tr className="border-b border-base-200">
			<td className="sticky left-0 z-1 border-r border-base-200/80 bg-base-100 px-3 py-2.5 text-sm font-medium text-base-content sm:py-2">
				{label}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.min)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.max)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.avg)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.med)}
			</td>
			{withAsym ? (
				<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
					{fmt(s.skew)}
				</td>
			) : (
				<td className="px-3 py-2.5 text-base-content/40 sm:py-2">—</td>
			)}
		</tr>
	)
}

function ColorRow({ label, s }: { label: string; s: FieldStatistics }) {
	return (
		<tr className="border-b border-base-200">
			<td className="sticky left-0 z-1 border-r border-base-200/80 bg-base-100 px-3 py-2.5 text-sm font-medium text-base-content sm:py-2">
				{label}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.min)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.max)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.avg)}
			</td>
			<td className="px-3 py-2.5 font-mono text-xs tabular-nums sm:py-2 sm:text-xs">
				{fmt(s.med)}
			</td>
		</tr>
	)
}

const ObjectAggregateTables: React.FC<ObjectAggregateTablesProps> = ({ fieldStats }) => {
	const hasSize = fieldStats.l || fieldStats.w || fieldStats.sq || fieldStats.l_w
	const hasColor = !!(
		fieldStats.r ||
		fieldStats.g ||
		fieldStats.b ||
		fieldStats.h ||
		fieldStats.s ||
		fieldStats.v
	)

	const vStats = fieldStats.v
	const vValid =
		vStats &&
		Number.isFinite(vStats.min) &&
		Number.isFinite(vStats.max) &&
		Number.isFinite(vStats.avg) &&
		Number.isFinite(vStats.med)

	if (!hasSize && !hasColor) {
		return (
			<p className="rounded-2xl border border-dashed border-base-300/80 bg-base-100/80 px-4 py-6 text-center text-sm text-base-content/60">
				Статистические данные по объектам недоступны (нет измерений).
			</p>
		)
	}

	return (
		<div className="space-y-6">
			{hasSize ? (
				<div>
					<p className="mb-2 text-[11px] leading-snug text-base-content/55 sm:hidden">
						Прокрутите таблицу вправо, чтобы увидеть все столбцы. Первый столбец закреплён.
					</p>
					<h3 className="mb-2 px-0.5 text-base font-semibold leading-snug text-base-content sm:px-1">
						Статистика по размерам
					</h3>
					<div className="overflow-x-auto rounded-2xl border border-base-200 [-webkit-overflow-scrolling:touch]">
						<table className="table w-full min-w-[520px]">
							<caption className="sr-only">Статистика по размерам</caption>
							<thead>
								<tr className="bg-base-200/80 text-[11px] uppercase tracking-wide text-base-content/70 sm:text-xs">
									<th className="sticky left-0 z-2 border-r border-base-200/90 bg-base-200/95 px-3 py-2.5 text-left backdrop-blur-sm sm:py-2">
										Параметр
									</th>
									<th className="px-3 py-2.5 sm:py-2">Мин.</th>
									<th className="px-3 py-2.5 sm:py-2">Макс.</th>
									<th className="px-3 py-2.5 sm:py-2">Среднее</th>
									<th className="px-3 py-2.5 sm:py-2">Медиана</th>
									<th className="px-3 py-2.5 sm:py-2">Асимметрия</th>
								</tr>
							</thead>
							<tbody>
								{fieldStats.l ? <SizeRow label="L (Длина), мм" s={fieldStats.l} withAsym /> : null}
								{fieldStats.w ? <SizeRow label="W (Ширина), мм" s={fieldStats.w} withAsym /> : null}
								{fieldStats.sq ? (
									<SizeRow label="SQ (Площадь), мм²" s={fieldStats.sq} withAsym />
								) : null}
								{fieldStats.l_w ? (
									<SizeRow
										label="L/W (Соотношение длины к ширине)"
										s={fieldStats.l_w}
										withAsym={false}
									/>
								) : null}
							</tbody>
						</table>
					</div>
				</div>
			) : null}

			{hasColor ? (
				<div>
					<p className="mb-2 text-[11px] leading-snug text-base-content/55 sm:hidden">
						Прокрутите таблицу вправо. Первый столбец закреплён.
					</p>
					<h3 className="mb-2 px-0.5 text-base font-semibold leading-snug text-base-content sm:px-1">
						Статистика по цветовым параметрам
					</h3>
					<div className="overflow-x-auto rounded-2xl border border-base-200 [-webkit-overflow-scrolling:touch]">
						<table className="table w-full min-w-[480px]">
							<caption className="sr-only">Статистика по цветовым параметрам</caption>
							<thead>
								<tr className="bg-base-200/80 text-[11px] uppercase tracking-wide text-base-content/70 sm:text-xs">
									<th className="sticky left-0 z-2 border-r border-base-200/90 bg-base-200/95 px-3 py-2.5 text-left backdrop-blur-sm sm:py-2">
										Параметр
									</th>
									<th className="px-3 py-2.5 sm:py-2">Мин.</th>
									<th className="px-3 py-2.5 sm:py-2">Макс.</th>
									<th className="px-3 py-2.5 sm:py-2">Среднее</th>
									<th className="px-3 py-2.5 sm:py-2">Медиана</th>
								</tr>
							</thead>
							<tbody>
								{fieldStats.r ? <ColorRow label="R (Красный)" s={fieldStats.r} /> : null}
								{fieldStats.g ? <ColorRow label="G (Зеленый)" s={fieldStats.g} /> : null}
								{fieldStats.b ? <ColorRow label="B (Синий)" s={fieldStats.b} /> : null}
								{fieldStats.h ? <ColorRow label="H (Оттенок)" s={fieldStats.h} /> : null}
								{fieldStats.s ? <ColorRow label="S (Насыщенность)" s={fieldStats.s} /> : null}
								{vValid && vStats ? <ColorRow label="V (Яркость)" s={vStats} /> : null}
							</tbody>
						</table>
					</div>
				</div>
			) : null}
		</div>
	)
}

export default ObjectAggregateTables
