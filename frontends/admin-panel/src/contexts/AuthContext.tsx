import { createContext, useCallback, useContext, useEffect, useState } from "react"
import { refreshToken as apiRefreshToken } from "@/features/auth/api"
import { LOGIN_PATH } from "@/lib/constants"
import type { SuccessResponse } from "@/lib/types"

export interface AuthContextType {
	isAuthenticated: boolean
	accessToken: string | null
	refreshToken: string | null
	login: (accessToken: string, refreshToken: string) => void
	logout: () => void
	isValidating: boolean
}

const AuthContext = createContext<AuthContextType | null>(null)

export function AuthProvider({ children }: { children: React.ReactNode }) {
	const [accessToken, setAccessToken] = useState<string | null>(localStorage.getItem("accessToken"))
	const [refreshToken, setRefreshToken] = useState<string | null>(
		localStorage.getItem("refreshToken")
	)
	const [isValidating, setIsValidating] = useState(true)

	const login = useCallback((newAccessToken: string, newRefreshToken: string) => {
		setAccessToken(newAccessToken)
		setRefreshToken(newRefreshToken)
		localStorage.setItem("accessToken", newAccessToken)
		localStorage.setItem("refreshToken", newRefreshToken)
	}, [])

	const logout = useCallback(() => {
		setAccessToken(null)
		setRefreshToken(null)
		localStorage.removeItem("accessToken")
		localStorage.removeItem("refreshToken")
	}, [])

	useEffect(() => {
		const storedRefreshToken = localStorage.getItem("refreshToken")
		if (!storedRefreshToken) {
			setIsValidating(false)
			return
		}
		apiRefreshToken(storedRefreshToken)
			.then((res: SuccessResponse<{ access_token: string; refresh_token: string }>) => {
				const { access_token, refresh_token } = res.result
				login(access_token, refresh_token)
			})
			.catch(() => {
				logout()
				window.location.href = LOGIN_PATH
			})
			.finally(() => {
				setIsValidating(false)
			})
	}, [login, logout])

	const isAuthenticated = !!accessToken

	return (
		<AuthContext.Provider
			value={{
				accessToken,
				refreshToken,
				login,
				logout,
				isAuthenticated,
				isValidating,
			}}
		>
			{children}
		</AuthContext.Provider>
	)
}

export function useAuth() {
	const context = useContext(AuthContext)
	if (!context) {
		throw new Error("useAuth must be used within an AuthProvider")
	}
	return context
}
