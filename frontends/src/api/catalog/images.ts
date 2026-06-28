import { handleApiResponse } from "../helpers"
import type { ApiResponse } from "../types"
import { catalogClient as client } from "./catalogClient"
import type { WeedImage } from "./types"

export const addCoffeeImage = async (coffeeId: number, image: File): Promise<WeedImage> => {
	const formData = new FormData()
	formData.append("file", image)

	const response = await client
		.post(`weeds/${coffeeId}/images`, {
			body: formData,
		})
		.json<ApiResponse<WeedImage>>()

	return handleApiResponse(response)
}

export const deleteCoffeeImage = async (coffeeId: number, imageId: number): Promise<void> => {
	await client.delete(`weeds/${coffeeId}/images/${imageId}`)
}
