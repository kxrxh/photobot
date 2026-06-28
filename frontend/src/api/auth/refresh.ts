import { clearStoredAuth, getStoredRefreshToken, setStoredTokens } from "@/lib/auth/storage"
import { handleApiResponse } from "../helpers"
import type { ApiResponse } from "../types"
import { client } from "./client"
import type { AuthServiceTokenResponse } from "./types"

let inFlight: Promise<AuthServiceTokenResponse> | null = null

/** Keeps React auth state (e.g. roles from JWT) in sync when refresh runs outside AuthContext (401 retry, XHR). */
type TokensRefreshedListener = (tokens: AuthServiceTokenResponse) => void | Promise<void>
let tokensRefreshedListener: TokensRefreshedListener | null = null

export function setTokensRefreshedListener(listener: TokensRefreshedListener | null): void {
	tokensRefreshedListener = listener
}

async function performRefresh(refreshToken: string): Promise<AuthServiceTokenResponse> {
	const refreshEnvelope = await client
		.post("auth/refresh", { json: { refresh_token: refreshToken } })
		.json<ApiResponse<AuthServiceTokenResponse>>()

	const tokens = handleApiResponse(refreshEnvelope)
	setStoredTokens(tokens)
	if (tokensRefreshedListener) {
		await tokensRefreshedListener(tokens)
	}
	return tokens
}

/** Single-flight refresh: AuthService invalidates the refresh token on each use. */
export function refreshTokensSingleFlight(
	refreshToken?: string
): Promise<AuthServiceTokenResponse> {
	if (inFlight) {
		return inFlight
	}

	const token = refreshToken ?? getStoredRefreshToken()
	if (!token) {
		clearStoredAuth()
		return Promise.reject(new Error("Missing refresh token"))
	}

	inFlight = performRefresh(token).finally(() => {
		inFlight = null
	})
	return inFlight
}

export const refresh = (refreshToken?: string): Promise<AuthServiceTokenResponse> =>
	refreshTokensSingleFlight(refreshToken)
