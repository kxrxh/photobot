import { refreshTokensSingleFlight } from "@/api/auth/refresh"
import { handleApiResponse } from "@/api/helpers"
import type { ApiErrorResponse, ApiResponse } from "@/api/types"
import { BASE_PATH } from "@/constants"
import { clearStoredAuth, getStoredAccessToken, getStoredRefreshToken } from "@/lib/auth/storage"
import { API_ENDPOINTS } from "../config"
import type { CreateAnalysisResponse } from "./types"

export type UploadProgressInfo = { loaded: number; total?: number }

function analysesPostUrl(): string {
	const base = API_ENDPOINTS.analysis
	const normalized = base.endsWith("/") ? base : `${base}/`
	return new URL("analyses", normalized).href
}

function postFormDataOnce(
	url: string,
	formData: FormData,
	options: {
		timeoutMs: number
		signal?: AbortSignal
		onUploadProgress?: (info: UploadProgressInfo) => void
		token: string | null
	}
): Promise<{ status: number; responseText: string }> {
	return new Promise((resolve, reject) => {
		const xhr = new XMLHttpRequest()
		xhr.open("POST", url)
		xhr.responseType = "text"
		xhr.timeout = options.timeoutMs

		if (options.token) {
			xhr.setRequestHeader("Authorization", `Bearer ${options.token}`)
		}

		let rafId = 0
		const cancelPendingProgress = () => {
			if (rafId) {
				cancelAnimationFrame(rafId)
				rafId = 0
			}
		}
		const reportProgress = (loaded: number, total?: number) => {
			if (!options.onUploadProgress) return
			if (rafId) cancelAnimationFrame(rafId)
			rafId = requestAnimationFrame(() => {
				rafId = 0
				options.onUploadProgress?.({ loaded, total })
			})
		}

		xhr.upload.addEventListener("progress", (e) => {
			if (e.lengthComputable) {
				reportProgress(e.loaded, e.total)
			} else {
				reportProgress(e.loaded, undefined)
			}
		})

		const onAbort = () => xhr.abort()

		if (options.signal) {
			if (options.signal.aborted) {
				reject(new DOMException("The operation was aborted.", "AbortError"))
				return
			}
			options.signal.addEventListener("abort", onAbort, { once: true })
		}

		xhr.onload = () => {
			cancelPendingProgress()
			if (options.signal) {
				options.signal.removeEventListener("abort", onAbort)
			}
			resolve({ status: xhr.status, responseText: xhr.responseText })
		}

		xhr.onerror = () => {
			cancelPendingProgress()
			if (options.signal) {
				options.signal.removeEventListener("abort", onAbort)
			}
			reject(new Error("Network error during analysis upload"))
		}

		xhr.ontimeout = () => {
			cancelPendingProgress()
			if (options.signal) {
				options.signal.removeEventListener("abort", onAbort)
			}
			reject(new Error("Analysis upload timed out"))
		}

		xhr.onabort = () => {
			cancelPendingProgress()
			if (options.signal) {
				options.signal.removeEventListener("abort", onAbort)
			}
			reject(new DOMException("The operation was aborted.", "AbortError"))
		}

		xhr.send(formData)
	})
}

/** XHR multipart POST for upload progress; mirrors ky auth (Bearer, one 401 refresh + retry). */
export async function postCreateAnalysisWithUploadProgress(
	createFormData: () => FormData,
	options: {
		timeoutMs: number
		signal?: AbortSignal
		onUploadProgress?: (info: UploadProgressInfo) => void
	}
): Promise<CreateAnalysisResponse> {
	const url = analysesPostUrl()

	const parseAndReturn = (responseText: string): CreateAnalysisResponse => {
		let parsed: ApiResponse<CreateAnalysisResponse> | ApiErrorResponse
		try {
			parsed = JSON.parse(responseText) as ApiResponse<CreateAnalysisResponse> | ApiErrorResponse
		} catch {
			throw new Error("Invalid response from analysis service")
		}
		return handleApiResponse(parsed)
	}

	let token = getStoredAccessToken()
	let { status, responseText } = await postFormDataOnce(url, createFormData(), {
		...options,
		token,
	})

	if (status === 401) {
		if (!getStoredRefreshToken()) {
			clearStoredAuth()
			window.location.href = BASE_PATH
			throw new Error("Unauthorized")
		}

		try {
			await refreshTokensSingleFlight()
		} catch {
			clearStoredAuth()
			window.location.href = BASE_PATH
			throw new Error("Unauthorized")
		}

		token = getStoredAccessToken()
		;({ status, responseText } = await postFormDataOnce(url, createFormData(), {
			...options,
			token,
		}))

		if (status === 401) {
			clearStoredAuth()
			window.location.href = BASE_PATH
			throw new Error("Unauthorized")
		}
	}

	if (status < 200 || status >= 300) {
		try {
			const parsed = JSON.parse(responseText) as ApiResponse<CreateAnalysisResponse>
			return handleApiResponse(parsed)
		} catch {
			throw new Error(`Analysis upload failed (${status})`)
		}
	}

	return parseAndReturn(responseText)
}
