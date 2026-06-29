import api from "@/lib/api"
import type { SuccessResponse } from "@/lib/types"
import type { Service } from "@/types/service"

export async function getServices(): Promise<SuccessResponse<Service[]>> {
	return api.get("services").json<SuccessResponse<Service[]>>()
}

export async function createService(
	service_id: string,
	service_secret: string
): Promise<SuccessResponse<{ message: string }>> {
	return api
		.post("services", { json: { service_id, service_secret } })
		.json<SuccessResponse<{ message: string }>>()
}

export async function deleteService(
	service_id: string
): Promise<SuccessResponse<{ message: string }>> {
	return api.delete(`services/${service_id}`).json<SuccessResponse<{ message: string }>>()
}
