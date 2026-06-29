import type { User as TelegramUser, ThemeParams } from "@tma.js/types"
import type { MessengerLocationResult } from "@/messenger/types"

export type UseMessengerReturnType = {
	user: TelegramUser | null
	platform: string | undefined
	version: string | undefined
	chatInstance: string | undefined
	chatType: string | undefined
	startParam?: string | undefined
	isReady: boolean
	initData: string | undefined
	themeParams: ThemeParams | undefined
	close: () => void
	expand: () => void
	disableSwipes: () => void
	requestFileDownload: (url: string, filename: string) => void
	isDark: boolean
	hapticFeedback: {
		impactOccurred: (style?: "light" | "medium" | "heavy" | "rigid" | "soft") => void
		notificationOccurred: (type: "error" | "success" | "warning") => void
		selectionChanged: () => void
	}
	location: {
		isSupported: () => boolean
		requestLocation: () => Promise<MessengerLocationResult>
	}
}
