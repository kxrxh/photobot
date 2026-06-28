import { redirect } from "@tanstack/react-router"
import { decodeJwt } from "jose"
import { sessionFromPayload } from "@/lib/auth/jwt"
import { getStoredAccessToken } from "@/lib/auth/storage"

/**
 * Route guard: redirects to /login if unauthenticated, /menu if missing required role.
 * Uses JWT claims (client-side UX only; APIs must enforce authorization).
 */
export function requireRole(requiredRoles: string[]): void {
	const token = getStoredAccessToken()
	if (!token) {
		throw redirect({ to: "/login" })
	}

	let payload: object
	try {
		payload = decodeJwt(token) as object
	} catch {
		throw redirect({ to: "/login" })
	}

	const session = sessionFromPayload(payload)
	if (!session) {
		throw redirect({ to: "/login" })
	}

	const hasRole = requiredRoles.some((role) => session.roles.includes(role))
	if (!hasRole) {
		throw redirect({ to: "/menu" })
	}
}
