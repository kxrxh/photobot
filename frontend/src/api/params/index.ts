import { handleApiResponse } from "@/api/helpers"
import type { ClassificationParam } from "@/api/params/types"
import type { ApiResponse } from "@/api/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.classification,
	timeout: 10000,
})

export const createParam = async (name: string) => {
	const response = await client
		.post("params", { json: { name } })
		.json<ApiResponse<ClassificationParam>>()
	return handleApiResponse(response)
}

export const deleteParam = async (id: string) => {
	const response = await client.delete(`params/${id}`).json<ApiResponse<{ message: string }>>()
	return handleApiResponse(response)
}

export const getParams = async () => {
	const response = await client.get("params").json<ApiResponse<ClassificationParam[]>>()
	return handleApiResponse(response)
}
