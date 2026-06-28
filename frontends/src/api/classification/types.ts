export type Classification = {
	created_at: string
	updated_at: string
	id: string
	created_by: number
	product: Product
	name: string
	is_public: boolean
}

export type CompleteClassification = {
	classification: Classification
	fractions: Fraction[]
}

export type Fraction = {
	id: string
	conditions: Condition[]
	name: string
	order_index: number
}

export type Condition = {
	id: string
	name: string
	params: Param[]
	operator: string
	connection: string
	order_index: number
}

export type Param = {
	id: string
	name: string
	operator: string
	value: number | null
}

export type Product = {
	id: string
	name: string
	created_at?: string
	updated_at?: string
}

export type SaveCompleteClassificationRequest = {
	product: Product
	fractions: SaveFractionRequest[]
	name: string
	is_public: boolean
}

export type SaveFractionRequest = {
	conditions: SaveConditionRequest[]
	name: string
	order_index: number
}

export type SaveConditionRequest = {
	params: SaveParamRequest[]
	name: string
	operator: string
	connection: string
	order_index: number
}

export type SaveParamRequest = {
	name: string
	operator: "<" | "<=" | "==" | ">=" | ">" | "!="
	value: number
}

export type ClassificationFilters = {
	product_id?: string
	name?: string
}
