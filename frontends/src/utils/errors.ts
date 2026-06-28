import { HTTPError, isTimeoutError } from "ky"
import { extractAuthServiceMessage, isRefreshRequest } from "@/api/auth"
import {
	AUTH_UNAUTHORIZED_MESSAGE,
	FORBIDDEN_MESSAGE,
	GENERIC_REQUEST_ERROR_MESSAGE,
	MESSENGER_RELOAD_MESSAGE,
	NETWORK_ERROR_MESSAGE,
	NOT_FOUND_MESSAGE,
	REFRESH_EXPIRED_MESSAGE,
	REQUEST_TIMEOUT_MESSAGE,
	SERVER_ERROR_MESSAGE,
	toUserFacingAuthMessage,
	UNKNOWN_ERROR_MESSAGE,
} from "@/lib/auth/messages"

const URL_PATTERN = /https?:\/\//i

export const SESSION_EXPIRED_MESSAGE = MESSENGER_RELOAD_MESSAGE

export function isInitDataExpiredResponse(body: unknown): boolean {
	if (body == null || typeof body !== "object") return false
	const o = body as Record<string, unknown>
	const message = [
		o.message,
		o.error,
		(o.error as Record<string, unknown>)?.message,
		(o.error as Record<string, unknown>)?.details,
	]
		.filter(Boolean)
		.join(" ")
		.toLowerCase()
	if (message.includes("refresh token")) return false
	return (
		message.includes("expired") ||
		message.includes("telegram data validation") ||
		message.includes("init data")
	)
}

export async function getAuthErrorMessage(error: unknown): Promise<string> {
	if (error instanceof HTTPError) {
		try {
			const body = (await error.response.json()) as unknown
			if (error.response.status === 401 && isInitDataExpiredResponse(body)) {
				return SESSION_EXPIRED_MESSAGE
			}
			const msg = extractAuthServiceMessage(body)
			if (msg != null) {
				if (isRefreshRequest(error.request.url)) return REFRESH_EXPIRED_MESSAGE
				return toUserFacingAuthMessage(msg)
			}
		} catch {}
	}
	return getUserFacingErrorMessage(error)
}

export function getUserFacingErrorMessage(error: unknown): string {
	if (isTimeoutError(error)) {
		return REQUEST_TIMEOUT_MESSAGE
	}
	if (error instanceof HTTPError) {
		const status = error.response.status
		if (status === 401) return AUTH_UNAUTHORIZED_MESSAGE
		if (status === 403) return FORBIDDEN_MESSAGE
		if (status === 404) return NOT_FOUND_MESSAGE
		if (status >= 500) return SERVER_ERROR_MESSAGE
		return GENERIC_REQUEST_ERROR_MESSAGE
	}
	if (error instanceof TypeError) return NETWORK_ERROR_MESSAGE
	if (error instanceof Error) {
		if (URL_PATTERN.test(error.message)) return UNKNOWN_ERROR_MESSAGE
		return error.message
	}
	return UNKNOWN_ERROR_MESSAGE
}
