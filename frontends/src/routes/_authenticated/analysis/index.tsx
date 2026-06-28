import { createFileRoute, redirect } from "@tanstack/react-router"

function truthyParam(v: unknown): boolean {
	return v === true || v === "true" || v === "1"
}

/**
 * Legacy query contract (still supported via redirect):
 * - ?create=true → /analysis/create
 * - ?list=true → /analysis/list
 * - ?openRequest=<id> → /analysis/create?openRequest=<id>
 *
 * Plain /analysis (no legacy flags) → /menu
 *
 * No validateSearch here: avoid polluting the URL with create=false&list=false.
 */
export const Route = createFileRoute("/_authenticated/analysis/")({
	beforeLoad: ({ search }) => {
		const s = search as Record<string, unknown>
		const list = truthyParam(s.list)
		const create = truthyParam(s.create)
		const rawOpen = s.openRequest
		const openRequest = typeof rawOpen === "string" && rawOpen.length > 0 ? rawOpen : undefined

		if (list) {
			throw redirect({ to: "/analysis/list", replace: true })
		}
		if (create || openRequest) {
			throw redirect({
				to: "/analysis/create",
				search: { openRequest },
				replace: true,
			})
		}
		throw redirect({ to: "/menu", replace: true })
	},
	component: () => null,
})
