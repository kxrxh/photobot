import { handleApiResponse } from "../helpers"
import type { ApiResponse } from "../types"
import { catalogClient as client } from "./catalogClient"

export interface ClassificationsResponse {
	main_groups: Record<string, string>
	main_subgroups: Record<string, string>
	subgroups: Record<string, string>
	hierarchy: Record<string, Record<string, string[]>>
}

export const fetchClassifications = async (): Promise<ClassificationsResponse> => {
	const response = await client.get("classifications").json<ApiResponse<ClassificationsResponse>>()
	return handleApiResponse(response)
}
