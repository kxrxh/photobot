import ky, { type Hooks } from "ky"
import { BASE_PATH } from "@/constants"
import { clearStoredAuth, getStoredAccessToken, getStoredRefreshToken } from "@/lib/auth/storage"
import {
	extractAuthServiceMessage,
	isLinkAuthRequest,
	isNonRetryableAuth401,
	isRefreshRequest,
} from "./helpers"
import { refreshTokensSingleFlight } from "./refresh"

export const authHooks: Hooks = {
	beforeRequest: [
		(request) => {
			const token = getStoredAccessToken()
			if (token) {
				request.headers.set("Authorization", `Bearer ${token}`)
			}
		},
	],

	afterResponse: [
		async (request, options, response) => {
			if (response.status !== 401) return response
			if (isRefreshRequest(request.url)) return response
			if (options.context?.retried) return response

			if (isLinkAuthRequest(request.url)) {
				const body = await response
					.clone()
					.json()
					.catch(() => null)
				if (isNonRetryableAuth401(extractAuthServiceMessage(body))) {
					return response
				}
			}

			if (!getStoredRefreshToken()) {
				clearStoredAuth()
				window.location.href = BASE_PATH
				return response
			}

			try {
				await refreshTokensSingleFlight()

				const token = getStoredAccessToken()
				const newRequest = request.clone()
				if (token) {
					newRequest.headers.set("Authorization", `Bearer ${token}`)
				}

				return ky(newRequest, {
					...options,
					context: { ...options.context, retried: true },
					retry: 0,
				})
			} catch {
				clearStoredAuth()
				window.location.href = BASE_PATH
				return response
			}
		},
	],
}
