export interface AuthServiceTokenResponse {
	access_token: string
	refresh_token: string
	roles: string[]
}

export interface AuthServiceRegisterRequest {
	organization_name?: string
	inn?: string
	full_name?: string
	phone_number?: string
}

export interface AuthServiceWebRegisterRequest {
	login: string
	password: string
	organization_name?: string
	inn?: string
	full_name?: string
	phone_number?: string
}

export interface AuthServiceWebRegisterResponse extends AuthServiceTokenResponse {
	recovery_codes: string[]
}

export interface AuthServiceUser {
	id: number
	login?: string | null
	telegram_id: number | null
	max_id: number | null
	organization_name: string | null
	inn: string | null
	full_name: string | null
	phone_number: string | null
	created_at: string
	updated_at: string
	roles: string[]
}
