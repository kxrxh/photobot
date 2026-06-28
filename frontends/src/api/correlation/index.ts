import type { CorrelationRequest, CorrelationWithTest } from "@/api/correlation/types"
import type { ApiResponse } from "@/api/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"
import { handleApiResponse } from "../helpers"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.classification,
	timeout: 10000,
})

export const calculateCorrelation = async (
	data: CorrelationRequest
): Promise<CorrelationWithTest[]> => {
	const response = await client
		.post("correlation", { json: data })
		.json<ApiResponse<CorrelationWithTest[]>>()
	return handleApiResponse(response)
}
