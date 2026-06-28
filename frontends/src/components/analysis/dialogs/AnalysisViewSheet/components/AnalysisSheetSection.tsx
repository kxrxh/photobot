import type React from "react"

export type AnalysisSheetSectionAccent = "primary" | "secondary" | "info" | "success"

const accentChip: Record<AnalysisSheetSectionAccent, string> = {
	primary: "bg-primary/10 text-primary",
	secondary: "bg-secondary/10 text-secondary",
	info: "bg-info/10 text-info",
	success: "bg-success/10 text-success",
}

export interface AnalysisSheetSectionProps {
	title: string
	subtitle?: string
	icon: React.ReactNode
	accent?: AnalysisSheetSectionAccent
	headerExtra?: React.ReactNode
	children: React.ReactNode
	className?: string
}

const AnalysisSheetSection: React.FC<AnalysisSheetSectionProps> = ({
	title,
	subtitle,
	icon,
	accent = "primary",
	headerExtra,
	children,
	className = "",
}) => {
	const chip = accentChip[accent]
	return (
		<section
			className={`rounded-3xl border border-base-200/90 bg-base-200/30 p-3 shadow-sm sm:p-4 dark:bg-base-200/20 ${className}`.trim()}
		>
			<header className="mb-3 sm:mb-4">
				<div className="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
					<h2 className="flex min-w-0 items-center gap-2.5 text-base font-semibold leading-snug text-base-content">
						<span
							className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ${chip}`}
							aria-hidden
						>
							{icon}
						</span>
						<span className="min-w-0">
							{title}
							{subtitle ? (
								<span className="mt-0.5 block text-xs font-normal leading-snug text-base-content/55">
									{subtitle}
								</span>
							) : null}
						</span>
					</h2>
					{headerExtra ? <div className="shrink-0 sm:pt-0.5">{headerExtra}</div> : null}
				</div>
			</header>
			{children}
		</section>
	)
}

export default AnalysisSheetSection
