export interface MarkupFraction {
	object_ids: number[]
	name: string
	id: string
	created_at: string
	updated_at: string
}

export interface Markup {
	id: string
	name: string
	created_by: number
	created_at: string
	updated_at: string
	analyses_ids: string[]
	fractions: MarkupFraction[]
}

export interface SaveMarkup {
	name: string
	fractions: SaveMarkupFraction[]
	analyses_ids: string[]
}

export interface SaveMarkupFraction {
	object_ids: number[]
	name: string
}

export interface MarkupFilters {
	name?: string
}
