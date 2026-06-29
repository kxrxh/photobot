import { HTTPError } from "ky"
import api from "@/lib/api"
import type { SuccessResponse } from "@/lib/types"
import type { Role } from "@/types/role"

function unwrap<T>(data: SuccessResponse<T> | T): T {
	if (data && typeof data === "object" && "result" in data && "success" in data) {
		return (data as SuccessResponse<T>).result
	}
	return data as T
}

export const getRoles = async (): Promise<Role[]> => {
	const response = await api.get("roles").json<SuccessResponse<Role[]> | Role[]>()
	return Array.isArray(response) ? response : unwrap(response)
}

export const getUserRoles = async (userId: number): Promise<Role[]> => {
	const response = await api.get(`users/${userId}/roles`).json<SuccessResponse<Role[]> | Role[]>()
	return Array.isArray(response) ? response : unwrap(response)
}

export const addUserRole = async (userId: number, roleId: number): Promise<void> => {
	await api.post("users/roles", {
		json: { user_id: userId, role_id: roleId },
	})
}

export const removeUserRole = async (userId: number, roleId: number): Promise<void> => {
	await api.delete("users/roles", {
		json: { user_id: userId, role_id: roleId },
	})
}

export const createRole = async (role: { name: string }): Promise<Role> => {
	try {
		const response = await api.post("roles", { json: role }).json<SuccessResponse<Role> | Role>()
		return unwrap(response)
	} catch (error: unknown) {
		if (error instanceof HTTPError) {
			if (error.response.status === 409) {
				throw new Error("Role with this name already exists")
			}
		}
		throw error
	}
}

export const updateRole = async (roleId: number, role: { name: string }): Promise<void> => {
	await api.put(`roles/${roleId}`, { json: role })
}

export const deleteRole = async (roleId: number): Promise<void> => {
	await api.delete(`roles/${roleId}`)
}
