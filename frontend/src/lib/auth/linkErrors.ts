import { ApiError } from "@/api/types"
import { toUserFacingAuthMessage, UNKNOWN_ERROR_MESSAGE } from "@/lib/auth/messages"
import { getUserFacingErrorMessage } from "@/utils/errors"

export function getLinkErrorMessage(
	error: unknown,
	fallback = "Не удалось привязать аккаунт"
): string {
	if (ApiError.isApiError(error)) {
		const rawMessage = (error.message || "").trim()
		if (!rawMessage) {
			return fallback
		}
		return toUserFacingAuthMessage(rawMessage)
	}
	const message = getUserFacingErrorMessage(error)
	return message === UNKNOWN_ERROR_MESSAGE ? fallback : message
}
