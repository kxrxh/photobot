import type {
	Analysis,
	AnalysisRequest,
	AnalysisWithObjects,
	ConfirmRequestRequest,
	CreateAnalysisResponse,
	GetAnalysesParams,
	KalibriObject,
	SimpleObject,
} from "@/api/analysis/types"
import { handleApiResponse } from "@/api/helpers"
import type { ApiResponse, PaginatedResponse } from "@/api/types"
import { log } from "@/utils/log"
import { BOT_NAME } from "../client"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"
import {
	postCreateAnalysisWithUploadProgress,
	type UploadProgressInfo,
} from "./createAnalysisUpload"

export type { UploadProgressInfo }

export type CreateAnalysisOptions = {
	mass_1000?: string
	mass?: string
	location?: string
	year?: string
	mass_liter?: string
	signal?: AbortSignal
	onUploadProgress?: (info: UploadProgressInfo) => void
}

const ANALYSIS_LONG_OPERATION_TIMEOUT_MS = 30 * 60 * 1000

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.analysis,
	timeout: 20 * 1000,
})

export const fetchAnalyses = async (
	params?: GetAnalysesParams
): Promise<PaginatedResponse<Analysis>> => {
	const searchParams = new URLSearchParams()

	if (params?.limit) searchParams.set("limit", params.limit.toString())
	if (params?.offset !== undefined) searchParams.set("offset", params.offset.toString())
	if (params?.product) searchParams.set("product", params.product)
	if (params?.id) searchParams.set("id", params.id)
	if (params?.sort_by) searchParams.set("sort_by", params.sort_by)
	if (params?.sort_order) searchParams.set("sort_order", params.sort_order)

	const response = await client
		.get(`analyses?${searchParams.toString()}`)
		.json<ApiResponse<PaginatedResponse<Analysis>>>()

	return handleApiResponse(response)
}

export const fetchAnalysisObjects = async (analysisId: string): Promise<KalibriObject[]> => {
	const response = await client
		.get(`analyses/${analysisId}/objects`)
		.json<ApiResponse<KalibriObject[]>>()
	return handleApiResponse(response)
}

export const createAnalysis = async (
	product: string,
	files: File[],
	options?: CreateAnalysisOptions
): Promise<CreateAnalysisResponse> => {
	const createFormData = (): FormData => {
		const formData = new FormData()
		formData.append("product", product)
		formData.append("bot", BOT_NAME)

		if (options?.mass_1000) formData.append("mass_1000", options.mass_1000)
		if (options?.mass) formData.append("mass", options.mass)
		if (options?.location) formData.append("location", options.location)
		if (options?.year) formData.append("year", options.year)
		if (options?.mass_liter) formData.append("mass_liter", options.mass_liter)

		files.forEach((file) => {
			formData.append("files", file)
		})

		return formData
	}

	const totalBytes = files.reduce((sum, f) => sum + f.size, 0)
	const sizeMb = totalBytes / (1024 * 1024)
	const uploadTimeoutMs = Math.min(
		20 * 60 * 1000,
		Math.max(3 * 60 * 1000, 3 * 60 * 1000 + sizeMb * 30 * 1000)
	)

	try {
		return await postCreateAnalysisWithUploadProgress(createFormData, {
			timeoutMs: uploadTimeoutMs,
			signal: options?.signal,
			onUploadProgress: options?.onUploadProgress,
		})
	} catch (error) {
		log.error("Analysis creation failed:", error)
		throw error
	}
}

export const fetchAnalysisById = async (id: string): Promise<AnalysisWithObjects> => {
	const response = await client.get(`analyses/${id}`).json<ApiResponse<AnalysisWithObjects>>()

	return handleApiResponse(response)
}

export const mergeAnalyses = async (analyses: string[]): Promise<{ message: string }> => {
	const response = await client
		.post("analyses/merge", {
			json: {
				analyses,
			},
			timeout: ANALYSIS_LONG_OPERATION_TIMEOUT_MS,
			retry: 0,
		})
		.json<ApiResponse<{ message: string }>>()

	return handleApiResponse(response)
}

export const sendReportToChat = async (
	analysisId: string | number,
	platform: "telegram" | "max",
	signal?: AbortSignal
): Promise<{ message: string }> => {
	const response = await client
		.post(`analyses/${analysisId}/report/send-to-chat`, {
			headers: { "X-Messenger-Platform": platform, "X-Bot-Name": BOT_NAME },
			signal,
		})
		.json<ApiResponse<{ message: string }>>()

	return handleApiResponse(response)
}

export const getObjectsByRequestId = async (requestId: string): Promise<SimpleObject[]> => {
	const response = await client
		.get(`requests/objects/${requestId}`)
		.json<ApiResponse<SimpleObject[]>>()

	return handleApiResponse(response)
}

export const getRequests = async (params?: {
	status?: string
	limit?: number
	offset?: number
}): Promise<{ requests: AnalysisRequest[]; total: number }> => {
	const searchParams = new URLSearchParams()

	if (params?.status) searchParams.set("status", params.status)
	if (params?.limit) searchParams.set("limit", params.limit.toString())
	if (params?.offset !== undefined) searchParams.set("offset", params.offset.toString())

	const response = await client
		.get(`requests?${searchParams.toString()}`)
		.json<ApiResponse<{ requests: AnalysisRequest[]; total: number }>>()

	return handleApiResponse(response)
}

export const confirmRequest = async (
	request: ConfirmRequestRequest,
	signal?: AbortSignal
): Promise<{ message: string }> => {
	try {
		const response = await client
			.post("requests/confirm", {
				json: request,
				timeout: ANALYSIS_LONG_OPERATION_TIMEOUT_MS,
				retry: 0,
				signal,
			})
			.json<ApiResponse<{ message: string }>>()

		return handleApiResponse(response)
	} catch (error) {
		log.error("Confirm request failed:", error)
		throw error
	}
}
