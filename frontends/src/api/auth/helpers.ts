import { API_ENDPOINTS } from "../config"

const URL_PATTERN = /https?:\/\//i

export function extractAuthServiceMessage(body: unknown): string | null {
	if (body == null || typeof body !== "object") return null
	const o = body as Record<string, unknown>
	const err = o.error as Record<string, unknown> | undefined
	if (err == null || typeof err !== "object") return null
	const msg = err.message
	if (typeof msg !== "string" || msg.length === 0 || msg.length > 500) return null
	if (URL_PATTERN.test(msg)) return null
	return msg
}

export const getAuthServiceBaseUrl = () => {
	return API_ENDPOINTS.auth.replace(/\/+$/, "")
}

export const isRefreshRequest = (requestUrl: string) => {
	try {
		const { pathname } = new URL(requestUrl)
		return pathname.endsWith("/auth/refresh")
	} catch {
		return requestUrl.includes("auth/refresh")
	}
}

export const isLinkAuthRequest = (requestUrl: string) => {
	try {
		const { pathname } = new URL(requestUrl)
		return (
			pathname.endsWith("/auth/link-with-code") ||
			pathname.endsWith("/auth/link-with-code-from-web")
		)
	} catch {
		return (
			requestUrl.includes("auth/link-with-code-from-web") ||
			requestUrl.includes("auth/link-with-code")
		)
	}
}

const NON_RETRYABLE_AUTH_401 = new Set([
	"link code not found or expired",
	"cannot link account to itself",
	"code user not found",
	"telegram data validation failed",
	"max data validation failed",
	"init data does not match the authenticated user",
	"invalid reset request",
	"reset code not found or expired",
	"invalid reset code",
	"invalid recovery code",
])

export const isNonRetryableAuth401 = (message: string | null | undefined) => {
	if (!message) return false
	return NON_RETRYABLE_AUTH_401.has(message.trim().toLowerCase())
}
