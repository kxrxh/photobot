import type { StoredTokens } from "@/storage/types"

const AUTH_TOKEN_KEY = "auth_service_token"
const REFRESH_TOKEN_KEY = "auth_service_refresh_token"

export const getStoredAccessToken = (): string | null => localStorage.getItem(AUTH_TOKEN_KEY)
export const getStoredRefreshToken = (): string | null => localStorage.getItem(REFRESH_TOKEN_KEY)

export const setStoredTokens = (tokens: StoredTokens): void => {
	localStorage.setItem(AUTH_TOKEN_KEY, tokens.access_token)
	localStorage.setItem(REFRESH_TOKEN_KEY, tokens.refresh_token)
}

export const clearStoredAuth = (): void => {
	localStorage.removeItem(AUTH_TOKEN_KEY)
	localStorage.removeItem(REFRESH_TOKEN_KEY)
	localStorage.removeItem("user_data")
}
