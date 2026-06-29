import { createRootRouteWithContext, Outlet } from "@tanstack/react-router"
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import type { AuthContextType } from "@/features/auth"

interface MyRouterContext {
	auth: AuthContextType | null
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
	component: () => (
		<div className="min-h-screen bg-linear-to-br from-base-100 via-primary/5 to-secondary/5 relative overflow-hidden">
			<div className="absolute inset-0 overflow-hidden pointer-events-none">
				<div className="absolute -top-40 -right-40 w-80 h-80 bg-primary/10 rounded-full blur-3xl"></div>
				<div className="absolute -bottom-40 -left-40 w-80 h-80 bg-secondary/10 rounded-full blur-3xl"></div>
				<div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-accent/5 rounded-full blur-3xl"></div>
			</div>

			<div className="relative z-10">
				<Outlet />
			</div>
			<TanStackRouterDevtools />
		</div>
	),
})
