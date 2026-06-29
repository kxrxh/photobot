/**
 * Auth state slice: status, error, token, user ids, roles.
 * Single source of truth for state shape and transitions.
 */

export type AuthStatus =
	| "loading"
	| "authenticated"
	| "unauthenticated"
	| "user_not_found"
	| "invalid_data"

export interface AuthState {
	status: AuthStatus
	error: string | null
	token: string | null
	userId: number | null
	telegramId: number | null
	roles: Set<string>
}

export type AuthAction =
	| { type: "LOADING" }
	| {
			type: "AUTHENTICATED"
			token: string
			userId: number
			telegramId: number | null
			roles: string[]
	  }
	| { type: "UNAUTHENTICATED"; error?: string | null }
	| { type: "USER_NOT_FOUND" }
	| { type: "INVALID_DATA"; error: string }
	| { type: "LOGOUT" }

export const initialAuthState: AuthState = {
	status: "loading",
	error: null,
	token: null,
	userId: null,
	telegramId: null,
	roles: new Set(),
}

/**
 * If stuck in loading (e.g. JWKS/refresh hang), we eventually show register so user can proceed.
 * Configurable here for tests or environment overrides later.
 */
export const AUTH_LOADING_TIMEOUT_MS = 10_000

/**
 * Max time to wait for login/register request (avoids infinite loading in MAX/Telegram when network hangs).
 */
export const LOGIN_REQUEST_TIMEOUT_MS = 15_000

export function authReducer(state: AuthState, action: AuthAction): AuthState {
	switch (action.type) {
		case "LOADING":
			return { ...state, status: "loading", error: null }
		case "AUTHENTICATED":
			return {
				status: "authenticated",
				error: null,
				token: action.token,
				userId: action.userId,
				telegramId: action.telegramId,
				roles: new Set(action.roles),
			}
		case "UNAUTHENTICATED":
			return {
				status: "unauthenticated",
				error: action.error ?? null,
				token: null,
				userId: null,
				telegramId: null,
				roles: new Set(),
			}
		case "USER_NOT_FOUND":
			return {
				status: "user_not_found",
				error: null,
				token: null,
				userId: null,
				telegramId: null,
				roles: new Set(),
			}
		case "INVALID_DATA":
			return {
				status: "invalid_data",
				error: action.error,
				token: null,
				userId: null,
				telegramId: null,
				roles: new Set(),
			}
		case "LOGOUT":
			return {
				status: "unauthenticated",
				error: null,
				token: null,
				userId: null,
				telegramId: null,
				roles: new Set(),
			}
		default:
			return state
	}
}
