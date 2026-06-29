import apiBase from "@/lib/apiBase"
import type { SuccessResponse } from "@/lib/types"

export async function login(
	loginField: string,
	password: string
): Promise<
	SuccessResponse<{
		access_token: string
		refresh_token: string
		roles: string[]
	}>
> {
	const response = await apiBase
		.post("auth/login", {
			json: { login: loginField, password },
			headers: {
				"X-Grant-Type": "password",
			},
		})
		.json()
	return response as SuccessResponse<{
		access_token: string
		refresh_token: string
		roles: string[]
	}>
}

export async function refreshToken(
	refreshToken: string
): Promise<SuccessResponse<{ access_token: string; refresh_token: string }>> {
	const response = await apiBase
		.post("auth/refresh", {
			json: { refresh_token: refreshToken },
		})
		.json()
	return response as SuccessResponse<{
		access_token: string
		refresh_token: string
	}>
}
