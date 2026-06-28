import type { Markup, MarkupFilters, SaveMarkup } from "@/api/markup/types"
import type { ApiResponse } from "@/api/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"
import { handleApiResponse } from "../helpers"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.classification,
	timeout: 20000,
})

export async function createMarkup(data: SaveMarkup): Promise<Markup> {
	const response = await client.post("markup", { json: data }).json<ApiResponse<Markup>>()
	return handleApiResponse(response)
}

export async function listMarkups(filters?: MarkupFilters): Promise<Markup[]> {
	const searchParams = new URLSearchParams()
	if (filters?.name) searchParams.set("name", filters.name)
	const response = await client
		.get(`markup?${searchParams.toString()}`)
		.json<ApiResponse<Markup[]>>()
	return handleApiResponse(response)
}

export async function getMarkup(id: string): Promise<Markup> {
	const response = await client.get(`markup/${id}`).json<ApiResponse<Markup>>()
	return handleApiResponse(response)
}

export async function updateMarkup(id: string, data: SaveMarkup): Promise<Markup> {
	const response = await client.put(`markup/${id}`, { json: data }).json<ApiResponse<Markup>>()
	return handleApiResponse(response)
}
