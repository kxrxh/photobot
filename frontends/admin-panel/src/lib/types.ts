export type SuccessResponse<T> = {
	success: true
	result: T
}

export type ErrorResponse = {
	success: false
	message: string
}
