export interface StoredUserSession {
	id: number
	telegram_id: number | null
	max_id: number | null
	roles: string[]
}

export interface StoredTokens {
	access_token: string
	refresh_token: string
}
