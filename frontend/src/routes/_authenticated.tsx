import { createFileRoute } from "@tanstack/react-router"
import { AuthenticatedOutlet } from "@/components/common/layout/AuthenticatedOutlet"
import { ErrorBoundary } from "@/components/common/layout/ErrorBoundary"

export const Route = createFileRoute("/_authenticated")({
	component: () => (
		<ErrorBoundary>
			<AuthenticatedOutlet />
		</ErrorBoundary>
	),
})
