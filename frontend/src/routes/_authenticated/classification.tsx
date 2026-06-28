import { createFileRoute } from "@tanstack/react-router"
import { lazy, Suspense } from "react"
import Loading from "@/components/common/ui/Loading"

const ClassificationPage = lazy(
	() => import("@/routes/_authenticated/classification/-ClassificationPage")
)

export const Route = createFileRoute("/_authenticated/classification")({
	component: RouteComponent,
})

export enum PageType {
	LIST = "list",
	CONSTRUCTOR = "constructor",
	COPY = "copy",
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
			<ClassificationPage />
		</Suspense>
	)
}
