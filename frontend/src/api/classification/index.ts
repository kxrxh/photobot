import type {
	Classification,
	ClassificationFilters,
	CompleteClassification,
	SaveCompleteClassificationRequest,
} from "@/api/classification/types"
import type { ApiResponse } from "@/api/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"
import { handleApiResponse } from "../helpers"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.classification,
})

export const getClassification = async (id: string): Promise<CompleteClassification> => {
	const response = await client
		.get(`classifications/${id}`)
		.json<ApiResponse<CompleteClassification>>()
	return handleApiResponse(response)
}

export const getClassifications = async (
	filters?: ClassificationFilters
): Promise<{
	classifications: Classification[]
	active_classification: Classification | null
}> => {
	const searchParams = new URLSearchParams()
	if (filters?.name) searchParams.append("name", filters.name)
	if (filters?.product_id) searchParams.append("product_id", filters.product_id)

	const response = await client.get(`classifications?${searchParams.toString()}`).json<
		ApiResponse<{
			classifications: Classification[]
			active_classification: Classification | null
		}>
	>()
	return handleApiResponse(response)
}

export const createClassification = async (
	data: SaveCompleteClassificationRequest
): Promise<Classification> => {
	const response = await client
		.post("classifications", { json: data })
		.json<ApiResponse<Classification>>()
	return handleApiResponse(response)
}

export const updateClassification = async (
	id: string,
	data: SaveCompleteClassificationRequest
): Promise<Classification> => {
	const response = await client
		.put(`classifications/${id}`, { json: data })
		.json<ApiResponse<Classification>>()
	return handleApiResponse(response)
}

export const deleteClassification = async (id: string): Promise<{ message: string }> => {
	const response = await client
		.delete(`classifications/${id}`)
		.json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const makeClassificationPublic = async (id: string): Promise<{ message: string }> => {
	const response = await client
		.put(`classifications/${id}/public`)
		.json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const makeClassificationPrivate = async (id: string): Promise<{ message: string }> => {
	const response = await client
		.put(`classifications/${id}/private`)
		.json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const setUserActiveClassification = async (
	classificationId: string
): Promise<{ message: string }> => {
	const response = await client
		.post("user-active-classifications", {
			json: { classification_id: classificationId },
		})
		.json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const deleteUserActiveClassification = async (): Promise<{ message: string }> => {
	const response = await client
		.delete("user-active-classifications")
		.json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const getUserActiveClassification = async (): Promise<CompleteClassification> => {
	const response = await client
		.get(`user-active-classifications`)
		.json<ApiResponse<CompleteClassification>>()
	return handleApiResponse(response)
}
