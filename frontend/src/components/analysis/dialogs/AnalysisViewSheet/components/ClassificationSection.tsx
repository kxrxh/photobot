import type React from "react"
import { FaTags } from "react-icons/fa"
import type { ClassStatistics, ClassStatisticsResult } from "@/utils/analysisStatistics"
import { translateClassName } from "@/utils/reportTranslations"
import AnalysisSheetSection from "./AnalysisSheetSection"

interface ClassificationSectionProps {
	classStats: ClassStatisticsResult
}

function fmt(n: number): string {
	return n.toFixed(2)
}

function SubTable({
	title,
	stats,
	mode,
}: {
	title: string
	stats: ClassStatistics
	mode: "size" | "color"
}) {
	const sizeFields: Array<{ key: keyof ClassStatistics; label: string; asym: boolean }> = [
		{ key: "l", label: "L (Длина), мм", asym: true },
		{ key: "w", label: "W (Ширина), мм", asym: true },
		{ key: "sq", label: "SQ (Площадь), мм²", asym: true },
		{ key: "l_w", label: "L/W", asym: false },
	]

	const colorFields: Array<{ key: keyof ClassStatistics; label: string }> = [
		{ key: "r", label: "R" },
		{ key: "g", label: "G" },
		{ key: "b", label: "B" },
		{ key: "h", label: "H" },
		{ key: "s", label: "S" },
		{ key: "v", label: "V" },
	]

	if (mode === "size") {
		const rows = sizeFields
			.map((f) => {
				const s = stats[f.key]
				if (!s || typeof s === "number") return null
				return { ...f, s }
			})
			.filter((row): row is NonNullable<typeof row> => row !== null)

		if (rows.length === 0) return null

		return (
			<div className="mt-3">
				<h5 className="mb-2 text-sm font-semibold text-base-content">{title}</h5>
				<p className="mb-2 text-[11px] leading-snug text-base-content/55 sm:hidden">
					Прокрутите вправо для остальных столбцов.
				</p>
				<div className="overflow-x-auto rounded-xl border border-base-200 [-webkit-overflow-scrolling:touch]">
					<table className="table table-sm w-full min-w-[400px]">
						<thead>
							<tr className="bg-base-200/60 text-[11px] sm:text-xs">
								<th className="sticky left-0 z-2 border-r border-base-200/90 bg-base-200/95 py-2.5 pl-2 pr-2 text-left backdrop-blur-sm sm:py-2">
									Параметр
								</th>
								<th className="py-2.5 sm:py-2">Мин.</th>
								<th className="py-2.5 sm:py-2">Макс.</th>
								<th className="py-2.5 sm:py-2">Среднее</th>
								<th className="py-2.5 sm:py-2">Медиана</th>
							</tr>
						</thead>
						<tbody>
							{rows.map((row) => (
								<tr key={String(row.key)} className="border-b border-base-200/80">
									<td className="sticky left-0 z-1 border-r border-base-200/80 bg-base-100 py-2.5 pl-2 pr-2 text-xs font-medium sm:py-2">
										{row.label}
									</td>
									<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">
										{fmt(row.s.min)}
									</td>
									<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">
										{fmt(row.s.max)}
									</td>
									<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">
										{fmt(row.s.avg)}
									</td>
									<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">
										{fmt(row.s.med)}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</div>
		)
	}

	const rows = colorFields
		.map((f) => {
			const s = stats[f.key]
			if (!s || typeof s === "number") return null
			return { ...f, s }
		})
		.filter((row): row is NonNullable<typeof row> => row !== null)

	if (rows.length === 0) return null

	return (
		<div className="mt-3">
			<h5 className="mb-2 text-sm font-semibold text-base-content">{title}</h5>
			<p className="mb-2 text-[11px] leading-snug text-base-content/55 sm:hidden">
				Прокрутите вправо для остальных столбцов.
			</p>
			<div className="overflow-x-auto rounded-xl border border-base-200 [-webkit-overflow-scrolling:touch]">
				<table className="table table-sm w-full min-w-[400px]">
					<thead>
						<tr className="bg-base-200/60 text-[11px] sm:text-xs">
							<th className="sticky left-0 z-2 border-r border-base-200/90 bg-base-200/95 py-2.5 pl-2 pr-2 text-left backdrop-blur-sm sm:py-2">
								Параметр
							</th>
							<th className="py-2.5 sm:py-2">Мин.</th>
							<th className="py-2.5 sm:py-2">Макс.</th>
							<th className="py-2.5 sm:py-2">Среднее</th>
							<th className="py-2.5 sm:py-2">Медиана</th>
						</tr>
					</thead>
					<tbody>
						{rows.map((row) => (
							<tr key={String(row.key)} className="border-b border-base-200/80">
								<td className="sticky left-0 z-1 border-r border-base-200/80 bg-base-100 py-2.5 pl-2 pr-2 text-xs font-medium sm:py-2">
									{row.label}
								</td>
								<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">{fmt(row.s.min)}</td>
								<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">{fmt(row.s.max)}</td>
								<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">{fmt(row.s.avg)}</td>
								<td className="py-2.5 font-mono text-xs tabular-nums sm:py-2">{fmt(row.s.med)}</td>
							</tr>
						))}
					</tbody>
				</table>
			</div>
		</div>
	)
}

const ClassificationSection: React.FC<ClassificationSectionProps> = ({ classStats }) => {
	const entries = Object.entries(classStats).filter(([name]) => name !== "objects_img")

	if (entries.length === 0) {
		return null
	}

	return (
		<AnalysisSheetSection
			title="Классификация"
			subtitle="Показатели по классам внутри выборки объектов"
			icon={<FaTags size={17} />}
			accent="secondary"
		>
			<p className="mb-4 text-xs leading-relaxed text-base-content/55">* На основе расчетов</p>
			<div className="space-y-5 sm:space-y-6">
				{entries.map(([className, data]) => (
					<div
						key={className}
						className="rounded-2xl border border-base-200/90 bg-base-100 p-3 shadow-sm sm:p-4"
					>
						<h4 className="text-sm font-semibold leading-snug text-base-content">
							Класс: {translateClassName(className)} (Объектов: {data.count})
						</h4>
						<SubTable title="Статистика по размерам" stats={data} mode="size" />
						<SubTable title="Статистика по цветовым параметрам" stats={data} mode="color" />
					</div>
				))}
			</div>
		</AnalysisSheetSection>
	)
}

export default ClassificationSection
