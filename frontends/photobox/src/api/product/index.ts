import { handleApiResponse } from "@/api/helpers"
import type { CreateProductPayload, Product, UpdateProductPayload } from "@/api/product/types"
import type { ApiResponse } from "@/api/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.classification,
	timeout: 10000,
})

export const getAllProducts = async (): Promise<Product[]> => {
	const response = await client.get("products").json<ApiResponse<Product[]>>()
	return handleApiResponse(response)
}

export const updateProduct = async (product: UpdateProductPayload): Promise<Product> => {
	const response = await client
		.put(`products/${product.id}`, { json: product })
		.json<ApiResponse<Product>>()
	return handleApiResponse(response)
}

export const deleteProduct = async (id: string): Promise<Product> => {
	const response = await client.delete(`products/${id}`).json<ApiResponse<Product>>()
	return handleApiResponse(response)
}

export const createProduct = async (product: CreateProductPayload): Promise<Product> => {
	const response = await client.post("products", { json: product }).json<ApiResponse<Product>>()
	return handleApiResponse(response)
}
