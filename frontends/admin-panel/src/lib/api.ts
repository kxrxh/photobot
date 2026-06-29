import ky from "ky"
import { refreshToken as apiRefreshToken } from "@/features/auth/api"
import { LOGIN_PATH } from "./constants"

let refreshPromise: Promise<unknown> | null = null

const api = ky.create({
	prefixUrl: import.meta.env.VITE_AUTH_API_URL,
	hooks: {
		beforeRequest: [
			(request) => {
				const token = localStorage.getItem("accessToken")
				if (token) {
					request.headers.set("Authorization", `Bearer ${token}`)
				}
			},
		],
		afterResponse: [
			async (request, _options, response) => {
				if (response.status !== 401) {
					return response
				}

				if (request.url.includes("auth/refresh") || request.url.includes("auth/login")) {
					return response
				}

				const refreshToken = localStorage.getItem("refreshToken")
				if (!refreshToken) {
					localStorage.removeItem("accessToken")
					localStorage.removeItem("refreshToken")
					window.location.href = LOGIN_PATH
					return response
				}

				if (!refreshPromise) {
					refreshPromise = apiRefreshToken(refreshToken)
						.then((res) => {
							const { access_token, refresh_token } = res.result
							localStorage.setItem("accessToken", access_token)
							localStorage.setItem("refreshToken", refresh_token)
							return res
						})
						.finally(() => {
							refreshPromise = null
						})
				}

				try {
					await refreshPromise
					const newToken = localStorage.getItem("accessToken")
					const newRequest = request.clone()
					if (newToken) {
						newRequest.headers.set("Authorization", `Bearer ${newToken}`)
					}
					return ky(newRequest)
				} catch (error) {
					localStorage.removeItem("accessToken")
					localStorage.removeItem("refreshToken")
					window.location.href = LOGIN_PATH
					throw error
				}
			},
		],
	},
})

export default api
