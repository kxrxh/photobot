import { createFileRoute } from "@tanstack/react-router"
import { lazy, Suspense } from "react"
import Loading from "@/components/common/ui/Loading"
import { requireRole } from "@/lib/auth/requireRole"

const AdminPage = lazy(() => import("@/routes/_authenticated/admin/-AdminPage"))

export const Route = createFileRoute("/_authenticated/admin")({
	beforeLoad: () => {
		requireRole(["admin"])
	},
	component: RouteComponent,
})

function RouteComponent() {
	return (
		<Suspense
			fallback={
				<div className="flex items-center justify-center h-screen">
					<Loading />
				</div>
			}
		>
			<AdminPage />
		</Suspense>
	)
}
