import { isMaxEnvironment } from "@/messenger/wrappers/max"

function hasTelegramLaunchParamsInUrl(): boolean {
	if (typeof window === "undefined") return false
	const blob = `${window.location.href}${window.location.hash}${window.location.search}`
	return (
		blob.includes("tgWebAppData=") ||
		blob.includes("tgWebAppVersion=") ||
		blob.includes("tgWebAppPlatform=")
	)
}

function isTelegramEnvironment(): boolean {
	if (typeof window === "undefined") return false
	if (window.TelegramWebviewProxy) return true

	const webApp = window.Telegram?.WebApp
	if (webApp) {
		if (typeof webApp.initData === "string" && webApp.initData.length > 0) return true
		if (webApp.platform || webApp.version) return true
		if (webApp.initDataUnsafe?.user) return true
	}

	return hasTelegramLaunchParamsInUrl()
}

/** Dev-only: `?mock=messenger`, `?mock=telegram`, or `?mock=max` simulates a mini-app in the browser. */
export function isDevMessengerMock(): boolean {
	if (!import.meta.env.DEV || typeof window === "undefined") return false
	const mock = new URLSearchParams(window.location.search).get("mock")
	return mock === "messenger" || mock === "telegram" || mock === "max"
}

/** True when running inside a messenger WebView (or dev mock), not a plain browser tab. */
export function isMessengerEnvironment(): boolean {
	return isMaxEnvironment() || isTelegramEnvironment() || isDevMessengerMock()
}

/** Password / web-session auth; false when messenger initData auto-login should run. */
export function isWebAuthMode(): boolean {
	return !isMessengerEnvironment()
}

export function isTelegramMessengerEnvironment(): boolean {
	return isTelegramEnvironment()
}
