export type BotPlatform = "telegram" | "max"

export interface Bot {
	id: string
	name: string
	platform?: BotPlatform
	token?: string
	created_at: string
	updated_at: string
}
