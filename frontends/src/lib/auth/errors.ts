import { HTTPError } from "ky"
import { extractAuthServiceMessage, isRefreshRequest } from "@/api/auth"
import { ApiError } from "@/api/types"
import {
	MESSENGER_RELOAD_MESSAGE,
	NETWORK_ERROR_MESSAGE,
	REFRESH_EXPIRED_MESSAGE,
	SERVER_ERROR_MESSAGE,
	TELEGRAM_VALIDATION_ERROR_MESSAGE,
	toUserFacingAuthMessage,
} from "@/lib/auth/messages"
import { isInitDataExpiredResponse } from "@/utils/errors"

type AuthErrorResult =
	| { kind: "user_not_found" }
	| { kind: "invalid_data"; message: string }
	| { kind: "unauthenticated"; message: string }

export { MESSENGER_RELOAD_MESSAGE, REFRESH_EXPIRED_MESSAGE, toUserFacingAuthMessage }

type NormalizeAuthErrorOptions = { isRefreshError?: boolean }

export async function normalizeAuthError(
	err: unknown,
	options?: NormalizeAuthErrorOptions
): Promise<AuthErrorResult> {
	if (err instanceof HTTPError) {
		const status = err.response.status
		if (status === 404) return { kind: "user_not_found" }
		if (status === 401) {
			const isRefresh =
				options?.isRefreshError ??
				(typeof err.request?.url === "string" && isRefreshRequest(err.request.url))
			try {
				const body = (await err.response.json()) as unknown
				const apiMsg = extractAuthServiceMessage(body)
				if (isRefresh) {
					return { kind: "unauthenticated", message: REFRESH_EXPIRED_MESSAGE }
				}
				if (isInitDataExpiredResponse(body)) {
					return { kind: "invalid_data", message: MESSENGER_RELOAD_MESSAGE }
				}
				if (apiMsg) {
					return {
						kind: "unauthenticated",
						message: toUserFacingAuthMessage(apiMsg),
					}
				}
			} catch {}
			return {
				kind: isRefresh ? "unauthenticated" : "invalid_data",
				message: isRefresh ? REFRESH_EXPIRED_MESSAGE : TELEGRAM_VALIDATION_ERROR_MESSAGE,
			}
		}
		if (status === 500) {
			try {
				const body = (await err.response.json()) as {
					error?: { message?: string; details?: string }
				}
				const text = [body?.error?.message, body?.error?.details].filter(Boolean).join(" ")
				if (text.includes("user") && text.includes("not found") && text.includes("Telegram")) {
					return { kind: "user_not_found" }
				}
			} catch {}
			return { kind: "unauthenticated", message: SERVER_ERROR_MESSAGE }
		}
		return { kind: "unauthenticated", message: SERVER_ERROR_MESSAGE }
	}

	if (err instanceof ApiError) {
		if (err.isNotFound()) return { kind: "user_not_found" }
		if (err.isUnauthorized()) {
			return {
				kind: "unauthenticated",
				message: toUserFacingAuthMessage(err.message),
			}
		}
	}

	return { kind: "unauthenticated", message: NETWORK_ERROR_MESSAGE }
}
