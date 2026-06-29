import api from "@/lib/api"
import type { SuccessResponse } from "@/lib/types"
import type { Bot } from "@/types/bot"

export async function getBots(): Promise<SuccessResponse<Bot[]>> {
	return api.get("bots").json<SuccessResponse<Bot[]>>()
}

export async function createBot(
	name: string,
	token: string,
	platform: "telegram" | "max" = "telegram"
): Promise<SuccessResponse<Bot>> {
	return api.post("bots", { json: { name, token, platform } }).json<SuccessResponse<Bot>>()
}

export async function updateBot(
	id: string,
	name?: string,
	token?: string
): Promise<SuccessResponse<Bot>> {
	const json: { name?: string; token?: string } = {}
	if (name !== undefined) {
		json.name = name
	}
	if (token !== undefined) {
		json.token = token
	}

	return api.put(`bots/${id}`, { json }).json<SuccessResponse<Bot>>()
}

export async function deleteBot(id: string): Promise<SuccessResponse<null>> {
	return api.delete(`bots/${id}`).json<SuccessResponse<null>>()
}
