export type Product = {
	id: string
	name: string
	created_at: string
	updated_at: string
}

export type UpdateProductPayload = {
	id: string
	name: string
}

export type CreateProductPayload = {
	name: string
}
