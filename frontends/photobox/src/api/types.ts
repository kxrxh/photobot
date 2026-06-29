export interface ApiResponse<T> {
	success: true
	result: T
}

export interface ApiErrorInfo {
	code: number
	message: string
	details?: unknown
	path: string
}

export interface ApiErrorResponse {
	success: false
	error: ApiErrorInfo
}

export class ApiError extends Error {
	public readonly code: number
	public readonly details?: unknown
	public readonly path: string

	constructor(errorInfo: ApiErrorInfo) {
		super(errorInfo.message)
		this.name = "ApiError"
		this.code = errorInfo.code
		this.details = errorInfo.details
		this.path = errorInfo.path
	}

	static isApiError(error: unknown): error is ApiError {
		return error instanceof ApiError
	}

	isNotFound(): boolean {
		return this.code === 404
	}

	isUnauthorized(): boolean {
		return this.code === 401
	}

	isForbidden(): boolean {
		return this.code === 403
	}

	isBadRequest(): boolean {
		return this.code === 400
	}

	isServerError(): boolean {
		return this.code >= 500
	}
}

export interface PaginationParams {
	limit?: number
	offset?: number
}

export interface PaginatedResponse<T> {
	data: T[]
	total: number
	limit: number
	offset: number
}
