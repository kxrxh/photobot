import type { ThemeParams } from "@tma.js/types"

/** Common messenger user shape (Telegram, MAX) */
export interface MessengerUser {
	id: number
	first_name: string
	last_name: string
	username?: string
	language_code: string
	allows_write_to_pm?: boolean
	photo_url?: string
}

/** Haptic feedback style */
export type HapticImpactStyle = "light" | "medium" | "heavy" | "rigid" | "soft"

/** Haptic notification type */
export type HapticNotificationType = "error" | "success" | "warning"

/** Why messenger-backed location request did not return coordinates */
export type MessengerLocationFailureReason = "denied" | "unavailable" | "timeout" | "unknown"

export type MessengerLocationResult =
	| { ok: true; latitude: number; longitude: number }
	| { ok: false; reason: MessengerLocationFailureReason }

/**
 * Unified interface for messenger platforms (Telegram, MAX, dev mock).
 * All implementations provide the same API surface.
 */
export interface MessengerBridge {
	getUser(): MessengerUser | null
	getPlatform(): string
	getVersion(): string
	getChatInstance(): string | undefined
	getChatType(): string | undefined
	getStartParam(): string | undefined
	isReady(): boolean
	getInitData(): string
	getThemeParams(): ThemeParams | undefined
	isDark(): boolean
	close(): void
	expand(): void
	disableSwipes(): void
	requestFileDownload(url: string, filename: string): void
	impactOccurred(style?: HapticImpactStyle): void
	notificationOccurred(type: HapticNotificationType): void
	selectionChanged(): void
	isLocationSupported(): boolean
	requestLocation(): Promise<MessengerLocationResult>
}
