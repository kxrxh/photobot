import type { QueryClient } from "@tanstack/react-query"
import { createRootRouteWithContext, Outlet } from "@tanstack/react-router"
import ErrorPage from "@/components/common/layout/ErrorPage"
import GlobalAlerts from "@/components/common/layout/GlobalAlerts"
import NotFoundPage from "@/components/common/layout/NotFoundPage"

interface MyRouterContext {
	queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
	component: () => (
		<div className="w-full h-full min-h-screen bg-base-100">
			<div className="h-full">
				<Outlet />
			</div>
			<GlobalAlerts />
		</div>
	),
	notFoundComponent: () => <NotFoundPage />,
	errorComponent: ({ error }) => <ErrorPage error={error} />,
})
