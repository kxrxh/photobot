import { ApiError, type ApiErrorResponse, type ApiResponse } from "./types"

export const handleApiResponse = <T>(response: ApiResponse<T> | ApiErrorResponse): T => {
	if (response.success) {
		return response.result
	} else {
		throw new ApiError(response.error)
	}
}
