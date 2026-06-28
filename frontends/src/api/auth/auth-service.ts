import { handleApiResponse } from "../helpers"
import type { ApiResponse } from "../types"
import { BOT_NAME, client, detectMessengerPlatform } from "./client"
import { authHooks } from "./hooks"
import type {
	AuthServiceRegisterRequest,
	AuthServiceTokenResponse,
	AuthServiceUser,
	AuthServiceWebRegisterRequest,
	AuthServiceWebRegisterResponse,
} from "./types"

const authServiceClient = client.extend({
	hooks: authHooks,
})

export const registerUser = async (
	initData: string,
	params: Partial<AuthServiceRegisterRequest> = {}
): Promise<unknown> => {
	const platform = detectMessengerPlatform(initData)

	const response = await client
		.post("auth/register", {
			json: params,
			headers: {
				"X-Init-Data": initData,
				"X-Messenger-Platform": platform,
				"X-Bot-Name": BOT_NAME,
			},
		})
		.json<ApiResponse<unknown>>()

	return handleApiResponse(response)
}

export const loginWithPassword = async (
	loginName: string,
	password: string
): Promise<AuthServiceTokenResponse> => {
	const response = await client
		.post("auth/login", {
			json: { login: loginName, password },
			headers: { "X-Grant-Type": "user_password" },
		})
		.json<ApiResponse<AuthServiceTokenResponse>>()

	return handleApiResponse(response)
}

export const registerWeb = async (
	params: AuthServiceWebRegisterRequest
): Promise<AuthServiceWebRegisterResponse> => {
	const response = await client
		.post("auth/register-web", { json: params })
		.json<ApiResponse<AuthServiceWebRegisterResponse>>()

	return handleApiResponse(response)
}

export const forgotPassword = async (login: string): Promise<void> => {
	const response = await client
		.post("auth/forgot-password", { json: { login } })
		.json<ApiResponse<{ message: string }>>()
	handleApiResponse(response)
}

export const resetPassword = async (
	login: string,
	otp: string,
	newPassword: string
): Promise<void> => {
	const response = await client
		.post("auth/reset-password", {
			json: { login, otp, new_password: newPassword },
		})
		.json<ApiResponse<{ message: string }>>()
	handleApiResponse(response)
}

export const resetPasswordRecovery = async (
	login: string,
	recoveryCode: string,
	newPassword: string
): Promise<void> => {
	const response = await client
		.post("auth/reset-password-recovery", {
			json: { login, recovery_code: recoveryCode, new_password: newPassword },
		})
		.json<ApiResponse<{ message: string }>>()
	handleApiResponse(response)
}

export const setupWebAccess = async (
	login: string,
	password: string
): Promise<{ recovery_codes: string[] }> => {
	const response = await authServiceClient
		.post("auth/setup-web-access", { json: { login, password } })
		.json<ApiResponse<{ recovery_codes: string[] }>>()

	return handleApiResponse(response)
}

export const login = async (initData: string): Promise<AuthServiceTokenResponse> => {
	const platform = detectMessengerPlatform(initData)

	const response = await client
		.post("auth/login", {
			json: {},
			headers: {
				"X-Grant-Type": "initdata",
				"X-Init-Data": initData,
				"X-Messenger-Platform": platform,
				"X-Bot-Name": BOT_NAME,
			},
		})
		.json<ApiResponse<AuthServiceTokenResponse>>()

	return handleApiResponse(response)
}

export const getMe = async (): Promise<AuthServiceUser> => {
	const response = await authServiceClient.get("users/me").json<ApiResponse<AuthServiceUser>>()
	return handleApiResponse(response)
}

export const updateMe = async (params: Partial<AuthServiceUser>): Promise<AuthServiceUser> => {
	const response = await authServiceClient
		.put("users/me", { json: params })
		.json<ApiResponse<AuthServiceUser>>()

	return handleApiResponse(response)
}

interface LinkCodeResult {
	code: string
	expires_in_seconds: number
}

/** JWT-only: returns a short-lived code for cross-messenger account linking. */
export const requestLinkCode = async (): Promise<LinkCodeResult> => {
	const response = await authServiceClient
		.post("auth/link-code")
		.json<ApiResponse<LinkCodeResult> | LinkCodeResult>()

	if (response && typeof response === "object" && "code" in response) {
		return response as LinkCodeResult
	}
	return handleApiResponse(response as ApiResponse<LinkCodeResult>)
}

interface LinkWithCodeResult {
	message: string
	access_token?: string
	refresh_token?: string
	roles?: string[]
}

/** JWT + init data: completes linking using a code from the other messenger. */
export const linkWithCode = async (
	code: string,
	initData: string,
	platform: "telegram" | "max"
): Promise<LinkWithCodeResult> => {
	const response = await authServiceClient
		.post("auth/link-with-code", {
			json: { code },
			headers: {
				"X-Init-Data": initData,
				"X-Bot-Name": BOT_NAME,
				"X-Messenger-Platform": platform,
			},
		})
		.json<ApiResponse<LinkWithCodeResult> | LinkWithCodeResult>()

	if (response && typeof response === "object" && "message" in response) {
		return response as LinkWithCodeResult
	}
	return handleApiResponse(response as ApiResponse<LinkWithCodeResult>)
}

/** JWT-only: completes linking using a code from a messenger mini-app. */
export const linkWithCodeFromWeb = async (code: string): Promise<LinkWithCodeResult> => {
	const response = await authServiceClient
		.post("auth/link-with-code-from-web", { json: { code } })
		.json<ApiResponse<LinkWithCodeResult> | LinkWithCodeResult>()

	if (response && typeof response === "object" && "message" in response) {
		return response as LinkWithCodeResult
	}
	return handleApiResponse(response as ApiResponse<LinkWithCodeResult>)
}
