import type React from "react"
import type { PdfLikeKpis } from "@/utils/analysisKpis"

interface AnalysisKpiStripProps {
	kpis: PdfLikeKpis
}

const AnalysisKpiStrip: React.FC<AnalysisKpiStripProps> = ({ kpis }) => {
	const { cards, rgbSwatch } = kpis

	if (cards.length === 0) {
		return null
	}

	return (
		<dl className="grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-base-200/90 bg-base-200/70 text-left shadow-sm sm:grid-cols-3 dark:border-base-300/50 dark:bg-base-300/40">
			{cards.map((card) => {
				const isRhs = card.id === "rhs"
				const swatchRgb = isRhs && rgbSwatch ? rgbSwatch : null
				const swatchCss = swatchRgb
					? `rgb(${Math.max(0, Math.min(255, swatchRgb.r))}, ${Math.max(0, Math.min(255, swatchRgb.g))}, ${Math.max(0, Math.min(255, swatchRgb.b))})`
					: null

				return (
					<div
						key={card.id}
						className={`flex min-h-21 flex-col justify-center gap-1 p-3.5 sm:min-h-22 sm:p-4 ${
							isRhs ? "col-span-2 sm:col-span-3" : ""
						} ${card.id === "broken" ? "bg-warning/[0.07]" : "bg-base-100"}`}
					>
						<dt className="text-[11px] font-medium leading-snug text-base-content/60 sm:text-xs">
							{card.title}
						</dt>
						<dd className="m-0 flex min-w-0 items-start justify-between gap-3">
							<span className="wrap-break-word text-base font-bold tabular-nums text-base-content sm:text-lg">
								{card.value}
							</span>
							{swatchCss && swatchRgb ? (
								<span
									className="mt-0.5 h-10 w-10 shrink-0 rounded-lg border border-base-200 shadow-inner sm:h-11 sm:w-11"
									style={{ backgroundColor: swatchCss }}
									title={`RGB: ${swatchRgb.r}, ${swatchRgb.g}, ${swatchRgb.b}`}
								/>
							) : null}
						</dd>
						{card.detail ? (
							<p className="m-0 font-mono text-[11px] tabular-nums text-base-content/50 sm:text-xs">
								{card.detail}
							</p>
						) : null}
					</div>
				)
			})}
		</dl>
	)
}

export default AnalysisKpiStrip
