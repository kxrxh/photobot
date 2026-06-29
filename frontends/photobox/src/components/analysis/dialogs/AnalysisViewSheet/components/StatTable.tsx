import type React from "react"

interface StatData {
	min: number
	max: number
	avg: number
	median: number
}

interface StatTableProps {
	title: string
	data: StatData
	unit?: string
}

const StatTable: React.FC<StatTableProps> = ({ title, data, unit = "" }) => {
	if (!data) {
		return null
	}

	const formatValue = (value: number) => {
		return `${value.toFixed(2)}${unit ? ` ${unit}` : ""}`
	}

	return (
		<div>
			<h3 className="mb-3 text-sm font-semibold text-base-content flex items-center gap-2">
				<span className="w-1 h-3 bg-primary rounded-full" />
				{title}
			</h3>
			<div className="border border-base-200 rounded-2xl overflow-hidden bg-base-100 shadow-sm">
				<table className="table w-full">
					<tbody>
						<tr className="bg-base-200/50 border-b border-base-200">
							<td className="font-medium text-xs text-base-content/70 px-3 py-2 border-r border-base-200">
								Минимум
							</td>
							<td className="font-medium text-xs text-base-content/70 px-3 py-2">Максимум</td>
						</tr>
						<tr className="border-b border-base-200/50">
							<td className="px-3 py-2 font-mono text-xs border-r border-base-200">
								{formatValue(data.min)}
							</td>
							<td className="px-3 py-2 font-mono text-xs">{formatValue(data.max)}</td>
						</tr>
						<tr className="bg-base-200/50 border-b border-base-200">
							<td className="font-medium text-xs text-base-content/70 px-3 py-2 border-r border-base-200">
								Медиана
							</td>
							<td className="font-medium text-xs text-base-content/70 px-3 py-2">Среднее</td>
						</tr>
						<tr>
							<td className="px-3 py-2 font-mono text-xs border-r border-base-200">
								{formatValue(data.median)}
							</td>
							<td className="px-3 py-2 font-mono text-xs">{formatValue(data.avg)}</td>
						</tr>
					</tbody>
				</table>
			</div>
		</div>
	)
}

export default StatTable
