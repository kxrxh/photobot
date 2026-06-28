import { createFileRoute } from "@tanstack/react-router"
import { lazy, Suspense } from "react"
import Loading from "@/components/common/ui/Loading"

const MarkupPage = lazy(() => import("@/routes/_authenticated/markup/-MarkupPage"))

export const Route = createFileRoute("/_authenticated/markup")({
	component: RouteComponent,
})

export type SelectedAnalysis = {
	id: string
}

function RouteComponent() {
	return (
		<Suspense
			fallback={
				<div className="flex items-center justify-center h-screen">
					<Loading />
				</div>
			}
		>
			<MarkupPage />
		</Suspense>
	)
}
