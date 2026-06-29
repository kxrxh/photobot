/**
 * MAX Bridge wrapper for mini-apps running inside MAX.
 * @see https://dev.max.ru/docs/webapps/bridge
 * @see https://dev.max.ru/docs/webapps/introduction
 */

import type { ThemeParams } from "@tma.js/types"
import { log } from "@/utils/log"
import type { MessengerLocationResult } from "../types"

interface MaxWebAppUser {
	id: number
	first_name: string
	last_name: string
	username?: string
	language_code?: string
	photo_url?: string
}

interface MaxInitDataUnsafe {
	query_id?: string
	auth_date?: number
	hash?: string
	start_param?: string | { [key: string]: unknown }
	user?: MaxWebAppUser
	chat?: {
		id?: number
		type?: string
	}
}

declare global {
	interface Window {
		WebApp?: {
			initData?: string
			InitData?: string
			initDataUnsafe?: MaxInitDataUnsafe
			platform?: string
			version?: string
			ready?: () => void
			close?: () => void
			downloadFile?: (url: string, file_name: string) => void
			HapticFeedback?: {
				impactOccurred: (
					impactStyle?: "soft" | "light" | "medium" | "heavy" | "rigid",
					disableVibrationFallback?: boolean
				) => void
				notificationOccurred: (
					notificationType?: "error" | "success" | "warning",
					disableVibrationFallback?: boolean
				) => void
				selectionChanged?: (disableVibrationFallback?: boolean) => void
			}
		}
	}
}

/** User type compatible with Telegram User (has allows_write_to_pm for compatibility) */
export interface MaxWrapperUser {
	id: number
	first_name: string
	last_name: string
	username?: string
	language_code: string
	allows_write_to_pm?: boolean
	photo_url?: string
}

export class MaxWrapper {
	constructor() {
		this.initialize()
	}

	private getWebApp(): Window["WebApp"] {
		return typeof window !== "undefined" ? window.WebApp : undefined
	}

	private initialize() {
		if (!this.getWebApp()) return
		// Defer so the host WebView can attach its message transport before WebAppReady is sent.
		queueMicrotask(() => {
			this.getWebApp()?.ready?.()
		})
	}

	private getInitDataString(): string {
		const webApp = this.getWebApp()
		return webApp?.initData ?? webApp?.InitData ?? ""
	}

	getUser(): MaxWrapperUser | null {
		const user = this.getWebApp()?.initDataUnsafe?.user
		if (!user) return null
		return {
			id: user.id,
			first_name: user.first_name,
			last_name: user.last_name,
			username: user.username,
			language_code: user.language_code ?? "en",
			photo_url: user.photo_url,
		}
	}

	getPlatform(): string {
		// MAX reports device platform (ios/android/desktop/web), but backend expects messenger platform.
		return this.getWebApp() ? "max" : "unknown"
	}

	getVersion(): string {
		return this.getWebApp()?.version ?? "unknown"
	}

	getChatInstance(): string | undefined {
		const chat = this.getWebApp()?.initDataUnsafe?.chat
		return chat?.id?.toString()
	}

	getChatType(): string | undefined {
		return this.getWebApp()?.initDataUnsafe?.chat?.type
	}

	getStartParam(): string | undefined {
		const raw = this.getWebApp()?.initDataUnsafe?.start_param
		if (typeof raw === "string") return raw
		if (raw && typeof raw === "object") {
			return JSON.stringify(raw)
		}
		return undefined
	}

	getInitData(): string {
		return this.getInitDataString()
	}

	getThemeParams(): ThemeParams | undefined {
		return undefined
	}

	isDark(): boolean {
		if (typeof window !== "undefined" && window.matchMedia) {
			return window.matchMedia("(prefers-color-scheme: dark)").matches
		}
		return false
	}

	isReady(): boolean {
		return this.getInitDataString().length > 0
	}

	close(): void {
		try {
			this.getWebApp()?.close?.()
		} catch (error) {
			log.devError("Failed to close MAX WebApp:", error)
		}
	}

	expand(): void {}

	disableSwipes(): void {}

	requestFileDownload(url: string, filename: string): void {
		try {
			const webApp = this.getWebApp()
			if (webApp?.downloadFile && url.startsWith("https://")) {
				webApp.downloadFile(url, filename)
			} else {
				this.fallbackDownload(url, filename)
			}
		} catch (error) {
			log.devWarn("MAX downloadFile failed, using fallback:", error)
			this.fallbackDownload(url, filename)
		}
	}

	private fallbackDownload(url: string, filename: string): void {
		const link = document.createElement("a")
		link.href = url
		link.download = filename
		link.target = "_blank"
		link.rel = "noopener noreferrer"
		document.body.appendChild(link)
		link.click()
		document.body.removeChild(link)
	}

	impactOccurred(style: "light" | "medium" | "heavy" | "rigid" | "soft" = "light"): void {
		try {
			this.getWebApp()?.HapticFeedback?.impactOccurred?.(style)
		} catch (error) {
			log.devWarn("Failed to trigger MAX haptic feedback impact:", error)
		}
	}

	notificationOccurred(type: "error" | "success" | "warning"): void {
		try {
			this.getWebApp()?.HapticFeedback?.notificationOccurred?.(type)
		} catch (error) {
			log.devWarn("Failed to trigger MAX haptic feedback notification:", error)
		}
	}

	selectionChanged(): void {
		try {
			this.getWebApp()?.HapticFeedback?.selectionChanged?.()
		} catch (error) {
			log.devWarn("Failed to trigger MAX haptic feedback selection:", error)
		}
	}

	isLocationSupported(): boolean {
		return false
	}

	async requestLocation(): Promise<MessengerLocationResult> {
		return { ok: false, reason: "unavailable" }
	}
}

/**
 * Detect MAX host WebView. The MAX script defines `window.WebApp` in any browser, but only the
 * host sets initData / initDataUnsafe — use those to avoid false positives in a plain tab.
 */
export function isMaxEnvironment(): boolean {
	if (typeof window === "undefined") return false
	const wa = window.WebApp
	if (!wa) return false

	const data = wa.initData ?? wa.InitData
	if (typeof data === "string" && data.length > 0) return true

	// Host may expose user in initDataUnsafe before the signed initData string is ready.
	if (wa.initDataUnsafe?.user) return true

	return false
}
