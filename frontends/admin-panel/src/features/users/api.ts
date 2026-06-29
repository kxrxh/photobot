import api from "@/lib/api"
import type { SuccessResponse } from "@/lib/types"
import type { User } from "@/types/user"

export async function getUsers(): Promise<SuccessResponse<User[]>> {
	return api.get("users").json<SuccessResponse<User[]>>()
}
