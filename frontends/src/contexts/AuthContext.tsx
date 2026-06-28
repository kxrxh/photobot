import {
	createContext,
	type ReactNode,
	useCallback,
	useContext,
	useEffect,
	useLayoutEffect,
	useMemo,
	useReducer,
	useRef,
} from "react"
import {
	login as apiLogin,
	loginWithPassword as apiLoginWithPassword,
	refreshTokensSingleFlight,
	registerUser,
} from "@/api/auth"
import { setTokensRefreshedListener } from "@/api/auth/refresh"
import type { AuthServiceTokenResponse } from "@/api/auth/types"
import { useMessenger } from "@/hooks/useMessenger"
import { normalizeAuthError } from "@/lib/auth/errors"
import { getAccessTokenExpiryMs, sessionFromPayload, verifyAccessToken } from "@/lib/auth/jwt"
import { INIT_DATA_UNAVAILABLE_MESSAGE, UNKNOWN_ERROR_MESSAGE } from "@/lib/auth/messages"
import { isWebAuthMode } from "@/lib/auth/mode"
import {
	AUTH_LOADING_TIMEOUT_MS,
	type AuthStatus,
	authReducer,
	initialAuthState,
	LOGIN_REQUEST_TIMEOUT_MS,
} from "@/lib/auth/state"
import {
	clearStoredAuth,
	getStoredAccessToken,
	getStoredRefreshToken,
	setStoredTokens,
} from "@/lib/auth/storage"
import { log } from "@/utils/log"

/** Public auth status union; exported as AuthState for backward compatibility. */
export type AuthState = AuthStatus

interface AuthContextType {
	state: AuthState
	userId: number | null
	telegramId: number | null
	token: string | null
	roles: Set<string>
	login: () => Promise<void>
	loginWithPassword: (login: string, password: string) => Promise<void>
	logout: () => void
	isWebAuth: boolean
	error: string | null
	applyTokenResponse: (response: AuthServiceTokenResponse) => Promise<void>
}

const LOGIN_TIMEOUT_MESSAGE = "Не удалось подключиться. Проверьте интернет."
const AUTH_LOADING_TIMEOUT_MESSAGE =
	"Не удалось восстановить сессию. Проверьте интернет и попробуйте снова."
const PROACTIVE_REFRESH_INTERVAL_MS = 5 * 60 * 1000 // 5 minutes
const PROACTIVE_REFRESH_THRESHOLD_MS = 5 * 60 * 1000 // Refresh when < 5 min left

function withTimeout<T>(promise: Promise<T>, ms: number, message: string): Promise<T> {
	return Promise.race([
		promise,
		new Promise<never>((_, reject) => setTimeout(() => reject(new Error(message)), ms)),
	])
}

const AuthContext = createContext<AuthContextType | null>(null)

export function useAuth() {
	const context = useContext(AuthContext)
	if (!context) {
		throw new Error("useAuth must be used within AuthProvider")
	}
	return context
}

interface AuthProviderProps {
	children: ReactNode
}

export function AuthProvider({ children }: AuthProviderProps) {
	const [model, dispatch] = useReducer(authReducer, initialAuthState)
	const autoLoginAttemptedRef = useRef(false)
	const { initData, isReady } = useMessenger()

	const clearAuthData = useCallback(() => {
		clearStoredAuth()
	}, [])

	const applyAuthResponse = useCallback(async (response: AuthServiceTokenResponse) => {
		if (!response.access_token) {
			throw new Error("Token not received from server")
		}

		const validation = await verifyAccessToken(response.access_token)
		if (!validation || validation.expired) {
			throw new Error("Received token is invalid or already expired")
		}

		const session = sessionFromPayload(validation.payload)
		if (!session) {
			throw new Error("JWT claims missing required fields")
		}

		setStoredTokens(response)

		const messengerId = session.telegram_id ?? session.max_id ?? null

		dispatch({
			type: "AUTHENTICATED",
			token: response.access_token,
			userId: session.id,
			telegramId: messengerId,
			roles: session.roles,
		})
	}, [])

	// Any refresh (including ky 401 / XHR) must update roles in state, not only storage.
	useLayoutEffect(() => {
		setTokensRefreshedListener(async (tokens) => {
			await applyAuthResponse(tokens)
		})
		return () => setTokensRefreshedListener(null)
	}, [applyAuthResponse])

	const tryRefresh = useCallback(async () => {
		const storedRefreshToken = getStoredRefreshToken()
		if (!storedRefreshToken) {
			clearAuthData()
			dispatch({ type: "UNAUTHENTICATED" })
			return
		}

		try {
			await refreshTokensSingleFlight(storedRefreshToken)
		} catch (err) {
			clearAuthData()
			const normalized = await normalizeAuthError(err, { isRefreshError: true })
			if (normalized.kind === "user_not_found") {
				dispatch({ type: "USER_NOT_FOUND" })
			} else {
				const errorMessage =
					isWebAuthMode() &&
					(normalized.kind === "unauthenticated" || normalized.kind === "invalid_data")
						? normalized.message
						: null
				dispatch({ type: "UNAUTHENTICATED", error: errorMessage })
			}
		}
	}, [clearAuthData])

	const initFromStorage = useCallback(async () => {
		dispatch({ type: "LOADING" })

		const storedToken = getStoredAccessToken()
		if (!storedToken) {
			await tryRefresh()
			return
		}

		const validation = await verifyAccessToken(storedToken)
		if (!validation || validation.expired) {
			await tryRefresh()
			return
		}

		const session = sessionFromPayload(validation.payload)
		if (!session) {
			clearAuthData()
			dispatch({ type: "UNAUTHENTICATED" })
			return
		}

		const messengerId = session.telegram_id ?? session.max_id ?? null

		dispatch({
			type: "AUTHENTICATED",
			token: storedToken,
			userId: session.id,
			telegramId: messengerId,
			roles: session.roles,
		})
	}, [tryRefresh, clearAuthData])

	useEffect(() => {
		void initFromStorage()
	}, [initFromStorage])

	useEffect(() => {
		if (model.status !== "loading") return
		const t = setTimeout(() => {
			clearAuthData()
			dispatch({ type: "UNAUTHENTICATED", error: AUTH_LOADING_TIMEOUT_MESSAGE })
		}, AUTH_LOADING_TIMEOUT_MS)
		return () => clearTimeout(t)
	}, [model.status, clearAuthData])

	useEffect(() => {
		if (model.status !== "authenticated" || !model.token) return
		const refreshIfNeeded = async () => {
			const token = getStoredAccessToken()
			const refreshToken = getStoredRefreshToken()
			if (!token || !refreshToken) return
			const expiryMs = getAccessTokenExpiryMs(token)
			if (expiryMs == null) return
			const timeLeftMs = expiryMs - Date.now()
			if (timeLeftMs >= PROACTIVE_REFRESH_THRESHOLD_MS) return
			try {
				await refreshTokensSingleFlight(refreshToken)
			} catch (err) {
				log.devWarn("[Auth] Proactive token refresh failed:", err)
			}
		}
		void refreshIfNeeded()
		const id = setInterval(refreshIfNeeded, PROACTIVE_REFRESH_INTERVAL_MS)
		return () => clearInterval(id)
	}, [model.status, model.token])

	const authenticate = useCallback(
		async (initData: string) => {
			try {
				dispatch({ type: "LOADING" })
				const tokenResponse = await withTimeout(
					apiLogin(initData),
					LOGIN_REQUEST_TIMEOUT_MS,
					LOGIN_TIMEOUT_MESSAGE
				)
				await applyAuthResponse(tokenResponse)
			} catch (err) {
				clearAuthData()
				if (err instanceof Error && err.message === LOGIN_TIMEOUT_MESSAGE) {
					dispatch({ type: "UNAUTHENTICATED", error: LOGIN_TIMEOUT_MESSAGE })
					return
				}
				const normalized = await normalizeAuthError(err)
				if (normalized.kind === "user_not_found") {
					try {
						await withTimeout(
							registerUser(initData, {}),
							LOGIN_REQUEST_TIMEOUT_MS,
							LOGIN_TIMEOUT_MESSAGE
						)
					} catch (registerErr) {
						log.devWarn("[Auth] Auto-register after user_not_found failed:", registerErr)
					}
					try {
						const tokenResponse = await withTimeout(
							apiLogin(initData),
							LOGIN_REQUEST_TIMEOUT_MS,
							LOGIN_TIMEOUT_MESSAGE
						)
						await applyAuthResponse(tokenResponse)
					} catch (retryErr) {
						if (retryErr instanceof Error && retryErr.message === LOGIN_TIMEOUT_MESSAGE) {
							dispatch({ type: "UNAUTHENTICATED", error: LOGIN_TIMEOUT_MESSAGE })
						} else {
							dispatch({ type: "USER_NOT_FOUND" })
						}
					}
					return
				}
				switch (normalized.kind) {
					case "invalid_data":
						dispatch({ type: "INVALID_DATA", error: normalized.message })
						break
					case "unauthenticated":
						dispatch({ type: "UNAUTHENTICATED", error: normalized.message })
						break
					default: {
						const msg =
							"message" in normalized
								? (normalized as { message: string }).message
								: UNKNOWN_ERROR_MESSAGE
						dispatch({ type: "UNAUTHENTICATED", error: msg })
					}
				}
			}
		},
		[applyAuthResponse, clearAuthData]
	)

	const loginWithPassword = useCallback(
		async (loginName: string, password: string) => {
			try {
				dispatch({ type: "LOADING" })
				const tokenResponse = await withTimeout(
					apiLoginWithPassword(loginName, password),
					LOGIN_REQUEST_TIMEOUT_MS,
					LOGIN_TIMEOUT_MESSAGE
				)
				await applyAuthResponse(tokenResponse)
			} catch (err) {
				clearAuthData()
				const normalized = await normalizeAuthError(err)
				const msg =
					normalized.kind === "unauthenticated" || normalized.kind === "invalid_data"
						? normalized.message
						: err instanceof Error
							? err.message
							: "Не удалось войти"
				dispatch({ type: "UNAUTHENTICATED", error: msg })
				throw new Error(msg)
			}
		},
		[applyAuthResponse, clearAuthData]
	)

	useEffect(() => {
		if (isWebAuthMode()) return
		if (!isReady) return
		if (model.status === "loading") return
		if (model.status === "authenticated") return
		if (autoLoginAttemptedRef.current) return

		autoLoginAttemptedRef.current = true

		if (!initData) {
			dispatch({
				type: "INVALID_DATA",
				error: INIT_DATA_UNAVAILABLE_MESSAGE,
			})
			return
		}

		void authenticate(initData)
	}, [isReady, initData, authenticate, model.status])

	const login = useCallback(async () => {
		if (!initData) {
			dispatch({
				type: "INVALID_DATA",
				error: INIT_DATA_UNAVAILABLE_MESSAGE,
			})
			return
		}
		await authenticate(initData)
	}, [initData, authenticate])

	const logout = useCallback(() => {
		autoLoginAttemptedRef.current = false
		clearAuthData()
		dispatch({ type: "LOGOUT" })
	}, [clearAuthData])

	const contextValue = useMemo<AuthContextType>(
		() => ({
			state: model.status,
			userId: model.userId,
			telegramId: model.telegramId,
			token: model.token,
			roles: model.roles,
			login,
			loginWithPassword,
			logout,
			isWebAuth: isWebAuthMode(),
			error: model.error,
			applyTokenResponse: applyAuthResponse,
		}),
		[
			model.status,
			model.userId,
			model.telegramId,
			model.token,
			model.roles,
			model.error,
			login,
			loginWithPassword,
			logout,
			applyAuthResponse,
		]
	)

	return <AuthContext.Provider value={contextValue}>{children}</AuthContext.Provider>
}
