import type { ThemeParams } from "@tma.js/types"
import type { MessengerBridge, MessengerLocationResult, MessengerUser } from "./types"

const baseDownload = (url: string, filename: string) => {
	const link = document.createElement("a")
	link.href = url
	link.download = filename
	link.target = "_blank"
	link.rel = "noopener noreferrer"
	document.body.appendChild(link)
	link.click()
	document.body.removeChild(link)
}

const devTelegramId = Number(import.meta.env.VITE_DEV_TELEGRAM_ID) || 919216442

const mockUser: MessengerUser = {
	id: devTelegramId,
	first_name: "test_name",
	last_name: "test_lastname",
	username: "testusername",
	language_code: "en",
	allows_write_to_pm: true,
	photo_url: `https://t.me/i/userpic/320/${devTelegramId}.jpg`,
}

const mockInitData = `user=${encodeURIComponent(
	JSON.stringify({
		id: mockUser.id,
		first_name: mockUser.first_name,
		last_name: mockUser.last_name,
		username: mockUser.username,
		language_code: mockUser.language_code,
	})
)}&auth_date=${Math.floor(Date.now() / 1000)}&hash=mock_hash`

/** Mock implementation for development (outside messenger WebView) */
export class DevBridge implements MessengerBridge {
	getUser(): MessengerUser | null {
		return mockUser
	}

	getPlatform(): string {
		return "mock_platform"
	}

	getVersion(): string {
		return "7.0"
	}

	getChatInstance(): string | undefined {
		return "mock_instance"
	}

	getChatType(): string | undefined {
		return "mock_type"
	}

	getStartParam(): string | undefined {
		return undefined
	}

	isReady(): boolean {
		return true
	}

	getInitData(): string {
		return mockInitData
	}

	getThemeParams(): ThemeParams | undefined {
		return undefined
	}

	isDark(): boolean {
		return false
	}

	close(): void {
		console.log("Mock close")
	}

	expand(): void {
		console.log("Mock expand")
	}

	disableSwipes(): void {
		console.log("Mock disable swipes")
	}

	requestFileDownload(url: string, filename: string): void {
		baseDownload(url, filename)
	}

	impactOccurred(style?: string): void {
		console.log(`Mock haptic feedback impact: ${style || "light"}`)
	}

	notificationOccurred(type: string): void {
		console.log(`Mock haptic feedback notification: ${type}`)
	}

	selectionChanged(): void {
		console.log("Mock haptic feedback selection changed")
	}

	isLocationSupported(): boolean {
		return false
	}

	async requestLocation(): Promise<MessengerLocationResult> {
		return { ok: false, reason: "unavailable" }
	}
}
