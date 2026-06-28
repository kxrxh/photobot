import { baseClient } from "../client"
import { getAuthServiceBaseUrl } from "./helpers"

export const BOT_NAME = "photobot"

type MessengerPlatform = "telegram" | "max"

/** Prefer `window.WebApp` (MAX) over `Telegram.WebApp` when both exist. @see https://dev.max.ru/docs/webapps/bridge */
export function detectMessengerPlatform(_initData: string): MessengerPlatform {
	if (typeof window !== "undefined") {
		const wa = (window as Window & { WebApp?: { initData?: string; InitData?: string } }).WebApp
		if (wa && (wa.initData || wa.InitData)) {
			return "max"
		}
		if ((window as Window & { Telegram?: { WebApp?: unknown } }).Telegram?.WebApp) {
			return "telegram"
		}
	}
	return "telegram"
}

/** No Bearer; login, register, refresh only. */
export const client = baseClient.extend({
	prefixUrl: getAuthServiceBaseUrl(),
	throwHttpErrors: true,
})
