import type { ReactNode } from "react"
import { SheetHeaderBackButton } from "@/components/common/ui/SheetHeaderActions"

export type CatalogFormTabItem = {
	id: string
	label: string
	icon: ReactNode
}

function catalogTabNavColumnClass(tabCount: number): string {
	if (tabCount <= 1) return "grid-cols-1"
	if (tabCount === 2) return "grid-cols-2"
	if (tabCount === 3) return "grid-cols-3"
	return "grid-cols-4"
}

export type CatalogItemFormLayoutProps = {
	title: string
	subtitle?: string
	onBack: () => void
	backButtonTitle?: string
	tabs: CatalogFormTabItem[]
	activeTabId: string
	onTabChange: (tabId: string) => void
	banner?: ReactNode
	children: ReactNode
	footerActions?: ReactNode
}

export function CatalogItemFormLayout({
	title,
	subtitle,
	onBack,
	backButtonTitle = "Назад в каталог",
	tabs,
	activeTabId,
	onTabChange,
	banner,
	children,
	footerActions,
}: CatalogItemFormLayoutProps) {
	return (
		<div className="flex min-h-0 flex-1 flex-col bg-base-100">
			<header className="flex shrink-0 items-center justify-between gap-2 border-b border-base-300 bg-base-100 px-3 py-3 backdrop-blur-sm">
				<div className="min-w-0 flex-1 pr-1">
					<h1 className="text-lg font-bold leading-snug tracking-tight text-base-content">
						{title}
					</h1>
					{subtitle ? (
						<p className="mt-0.5 text-xs leading-snug text-base-content/60">{subtitle}</p>
					) : null}
				</div>
				<SheetHeaderBackButton
					onClick={onBack}
					aria-label={backButtonTitle}
					title={backButtonTitle}
				/>
			</header>

			<main className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden">
				{banner}
				<div className="mx-auto w-full max-w-lg pb-4 pt-1">{children}</div>
			</main>

			<footer className="shrink-0 border-t border-base-300 bg-base-100/95 shadow-[0_-8px_32px_-12px_color-mix(in_oklch,var(--color-neutral),transparent_86%)] backdrop-blur-md">
				<div className="mx-auto w-full max-w-lg px-2 pt-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] sm:px-4">
					<nav
						className={`grid w-full touch-manipulation gap-1 ${catalogTabNavColumnClass(tabs.length)} ${footerActions ? "mb-2" : ""}`}
						aria-label="Разделы карточки"
					>
						{tabs.map((tab) => {
							const selected = activeTabId === tab.id
							return (
								<button
									key={tab.id}
									type="button"
									role="tab"
									aria-selected={selected}
									className={`flex min-h-11 min-w-0 cursor-pointer flex-col items-center justify-center gap-0.5 rounded-lg px-0.5 py-1 text-center transition-colors active:opacity-90 ${
										selected
											? "bg-primary text-primary-content"
											: "text-base-content/75 hover:bg-base-200/70 hover:text-base-content"
									}`}
									onClick={() => onTabChange(tab.id)}
								>
									<span
										className={`text-[0.95rem] leading-none sm:text-[1.05rem] ${selected ? "" : "opacity-90"}`}
										aria-hidden
									>
										{tab.icon}
									</span>
									<span
										className="w-full max-w-full truncate px-0.5 text-[9px] font-medium leading-none tracking-tighter sm:text-[10px]"
										title={tab.label}
									>
										{tab.label}
									</span>
								</button>
							)
						})}
					</nav>
					{footerActions}
				</div>
			</footer>
		</div>
	)
}
