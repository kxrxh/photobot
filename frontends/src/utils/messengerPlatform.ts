const telegramPlatforms = new Set([
	"telegram",
	"android",
	"ios",
	"macos",
	"tdesktop",
	"weba",
	"webk",
	"web",
	"unigram",
])

const isMaxRuntimeDetected = (): boolean => {
	if (typeof window === "undefined") {
		return false
	}

	const webApp = (window as Window & { WebApp?: { initData?: string; InitData?: string } }).WebApp
	return Boolean(webApp && (webApp.initData || webApp.InitData))
}

export const resolveChatPlatform = (
	platform: string | null | undefined
): "telegram" | "max" | undefined => {
	const normalizedPlatform = platform?.trim().toLowerCase()

	if (!normalizedPlatform || normalizedPlatform === "unknown") {
		return undefined
	}

	if (normalizedPlatform === "max") {
		return "max"
	}

	// In MAX mini-app runtime, bridge platform can be device-like (android/ios/web).
	if (isMaxRuntimeDetected()) {
		return "max"
	}

	if (telegramPlatforms.has(normalizedPlatform)) {
		return "telegram"
	}

	return undefined
}
