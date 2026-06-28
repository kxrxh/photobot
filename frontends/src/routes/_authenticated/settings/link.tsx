import { createFileRoute, redirect } from "@tanstack/react-router"

export const Route = createFileRoute("/_authenticated/settings/link")({
	beforeLoad: () => {
		throw redirect({ to: "/profile", replace: true })
	},
})
