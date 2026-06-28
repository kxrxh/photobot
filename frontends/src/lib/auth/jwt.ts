import * as jose from "jose"
import { getAuthServiceBaseUrl } from "@/api/auth/helpers"
import type { StoredUserSession } from "@/storage/types"

function getJWKSUrl(): URL {
	const baseUrl = getAuthServiceBaseUrl().replace(/\/$/, "")
	return new URL(`${baseUrl}/auth/.well-known/jwks.json`, window.location.origin)
}

const JWKS = jose.createRemoteJWKSet(getJWKSUrl())

export interface VerifyResult {
	payload: object
	expired: false
}

export interface VerifyExpiredResult {
	payload: object
	expired: true
}

/**
 * Verify access token with JWKS. Returns payload and expired flag;
 * expired tokens still return decoded payload for session extraction.
 */
export async function verifyAccessToken(
	token: string
): Promise<VerifyResult | VerifyExpiredResult | null> {
	try {
		const { payload } = await jose.jwtVerify(token, JWKS)
		return { payload: payload as object, expired: false as const }
	} catch (err) {
		if (err instanceof jose.errors.JWTExpired) {
			return { payload: jose.decodeJwt(token) as object, expired: true as const }
		}
		return null
	}
}

/**
 * Get access token expiry timestamp in ms (for proactive refresh). Returns null if invalid.
 */
export function getAccessTokenExpiryMs(token: string): number | null {
	try {
		const payload = jose.decodeJwt(token)
		const exp = payload.exp
		if (typeof exp !== "number") return null
		return exp * 1000
	} catch {
		return null
	}
}

/**
 * Extract id, telegram_id, max_id, roles from JWT payload (user_id, telegram_id, max_id, roles).
 */
export function sessionFromPayload(payload: object): StoredUserSession | null {
	const p = payload as Record<string, unknown>
	const id = typeof p.user_id === "number" ? p.user_id : null
	const telegramId = typeof p.telegram_id === "number" ? p.telegram_id : null
	const maxId = typeof p.max_id === "number" ? p.max_id : null
	if (id == null) return null

	const roles = Array.isArray(p.roles) ? (p.roles as string[]) : []
	return { id, telegram_id: telegramId, max_id: maxId, roles }
}
