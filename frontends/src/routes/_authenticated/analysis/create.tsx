import { createFileRoute } from "@tanstack/react-router"
import AnalysisCreationPage from "@/components/analysis/dialogs/AnalysisCreationSheet"

type CreateAnalysisSearch = {
	openRequest?: string
	tab?: "upload" | "requests"
}

function parseCreateAnalysisSearch(search: Record<string, unknown>): CreateAnalysisSearch {
	const rawOpen = search.openRequest
	const openRequest = typeof rawOpen === "string" && rawOpen.length > 0 ? rawOpen : undefined

	const rawTab = search.tab
	const tab = rawTab === "upload" || rawTab === "requests" ? rawTab : undefined

	return { openRequest, tab }
}

export const Route = createFileRoute("/_authenticated/analysis/create")({
	validateSearch: (search: Record<string, unknown>) => parseCreateAnalysisSearch(search),
	component: RouteComponent,
})

function RouteComponent() {
	const { openRequest, tab } = Route.useSearch()

	const initialTab = tab ?? (openRequest ? "requests" : undefined)

	return (
		<div className="flex min-h-0 flex-1 flex-col overflow-hidden bg-base-100">
			<AnalysisCreationPage isOpen initialTab={initialTab} initialRequestId={openRequest} />
		</div>
	)
}
