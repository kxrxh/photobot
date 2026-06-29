import {
	postEvent,
	request2,
	retrieveLaunchParams,
	retrieveRawInitData,
	TimeoutError,
} from "@tma.js/bridge"
import type { LaunchParams, ThemeParams, User as TMAUser } from "@tma.js/types"
import { log } from "@/utils/log"
import type { MessengerLocationResult } from "../types"

declare global {
	interface Window {
		TelegramWebviewProxy?: {
			postEvent: (eventType: string, data: string) => void
		}
		Telegram?: {
			WebApp?: {
				initData?: string
				initDataUnsafe?: {
					user?: TMAUser
					chat_instance?: string
					chat_type?: string
				}
				platform?: string
				version?: string
				colorScheme?: string
				themeParams?: ThemeParams
				ready?: () => void
				expand?: () => void
				HapticFeedback?: {
					impactOccurred: (style: "light" | "medium" | "heavy" | "rigid" | "soft") => void
					notificationOccurred: (type: "error" | "success" | "warning") => void
					selectionChanged: () => void
				}
				/** @see https://core.telegram.org/bots/webapps#locationmanager */
				LocationManager?: {
					isSupported: boolean
					isInited: boolean
					isLocationAvailable?: boolean
					/** @deprecated Telegram also documents isAccessGranted / isAccessRequested */
					isLocationPermitted?: boolean
					isAccessRequested?: boolean
					isAccessGranted?: boolean
					init: (callback?: () => void) => void
					getLocation: (
						callback: (data: { latitude: number; longitude: number } | null) => void
					) => void
					openSettings?: () => void
				}
			}
		}
	}
}

export class TelegramWrapper {
	private launchParams?: LaunchParams
	private rawInitData?: string
	private webApp?: {
		initData?: string
		initDataUnsafe?: {
			user?: TMAUser
			chat_instance?: string
			chat_type?: string
		}
		platform?: string
		version?: string
		colorScheme?: string
		themeParams?: ThemeParams
		ready?: () => void
		expand?: () => void
		HapticFeedback?: {
			impactOccurred: (style: "light" | "medium" | "heavy" | "rigid" | "soft") => void
			notificationOccurred: (type: "error" | "success" | "warning") => void
			selectionChanged: () => void
		}
	}

	constructor() {
		this.initialize()
	}

	private initialize() {
		this.webApp = window.Telegram?.WebApp
		if (this.webApp) {
			this.webApp.ready?.()
			this.webApp.expand?.()
			if (this.webApp.initData) {
				this.rawInitData = this.webApp.initData
			}
		}

		try {
			this.launchParams = retrieveLaunchParams()
		} catch (error) {
			log.devWarn("Failed to retrieve Telegram launch params:", error)
		}

		if (!this.rawInitData) {
			try {
				this.rawInitData = retrieveRawInitData()
			} catch (error) {
				log.devWarn("Failed to retrieve Telegram raw init data:", error)
			}
		}

		if (!this.rawInitData && this.webApp?.initData) {
			this.rawInitData = this.webApp.initData
		}
	}

	/**
	 * Gets the Telegram user data
	 */
	getUser(): TMAUser | null {
		try {
			const tgWebAppData = this.launchParams?.tgWebAppData

			if (tgWebAppData?.user) {
				const user = tgWebAppData.user
				return {
					id: user.id,
					first_name: user.first_name,
					last_name: user.last_name,
					username: user.username,
					language_code: user.language_code ?? "en",
					allows_write_to_pm: user.allows_write_to_pm,
					photo_url: user.photo_url,
				}
			}

			const webAppUser = this.webApp?.initDataUnsafe?.user
			if (webAppUser) {
				return {
					id: webAppUser.id,
					first_name: webAppUser.first_name,
					last_name: webAppUser.last_name,
					username: webAppUser.username,
					language_code: webAppUser.language_code ?? "en",
					allows_write_to_pm: webAppUser.allows_write_to_pm,
					photo_url: webAppUser.photo_url,
				}
			}

			return null
		} catch (error) {
			log.devWarn("Failed to get user data:", error)
			return null
		}
	}

	/**
	 * Gets the platform (e.g., "ios", "android", "web")
	 */
	getPlatform(): string {
		return this.launchParams?.tgWebAppPlatform || this.webApp?.platform || "unknown"
	}

	/**
	 * Gets the WebApp version
	 */
	getVersion(): string {
		return this.launchParams?.tgWebAppVersion || this.webApp?.version || "unknown"
	}

	/**
	 * Gets the chat instance
	 */
	getChatInstance(): string | undefined {
		const tgWebAppData = this.launchParams?.tgWebAppData
		return tgWebAppData?.chat_instance || this.webApp?.initDataUnsafe?.chat_instance
	}

	/**
	 * Gets the chat type
	 */
	getChatType(): string | undefined {
		const tgWebAppData = this.launchParams?.tgWebAppData
		return tgWebAppData?.chat_type || this.webApp?.initDataUnsafe?.chat_type
	}

	/**
	 * Gets the start parameter
	 */
	getStartParam(): string | undefined {
		return this.launchParams?.tgWebAppStartParam
	}

	/**
	 * Gets the init data string
	 */
	getInitData(): string {
		if (this.rawInitData) {
			return this.rawInitData
		}

		if (this.webApp?.initData) {
			return this.webApp.initData
		}

		const tgWebAppData = this.launchParams?.tgWebAppData
		if (!tgWebAppData) return ""

		const params = new URLSearchParams()
		if (tgWebAppData.user) {
			params.set(
				"user",
				JSON.stringify({
					id: tgWebAppData.user.id,
					first_name: tgWebAppData.user.first_name,
					last_name: tgWebAppData.user.last_name,
					username: tgWebAppData.user.username,
					language_code: tgWebAppData.user.language_code,
					allows_write_to_pm: tgWebAppData.user.allows_write_to_pm,
					photo_url: tgWebAppData.user.photo_url,
				})
			)
		}
		if (tgWebAppData.chat_instance) {
			params.set("chat_instance", tgWebAppData.chat_instance)
		}
		if (tgWebAppData.chat_type) {
			params.set("chat_type", tgWebAppData.chat_type)
		}
		params.set("auth_date", Math.floor(tgWebAppData.auth_date.getTime() / 1000).toString())
		params.set("hash", tgWebAppData.hash)

		return params.toString()
	}

	/**
	 * Gets theme parameters
	 */
	getThemeParams(): ThemeParams | undefined {
		return this.launchParams?.tgWebAppThemeParams || this.webApp?.themeParams
	}

	/**
	 * Checks if dark mode is active
	 */
	isDark(): boolean {
		const colorScheme = this.webApp?.colorScheme
		const bgColor = this.getThemeParams()?.bg_color

		return (
			colorScheme === "dark" ||
			(bgColor?.includes("#000") ?? false) ||
			(this.webApp?.themeParams?.bg_color?.includes("#000") ?? false) ||
			false
		)
	}

	/**
	 * Checks if the WebApp is ready
	 */
	isReady(): boolean {
		return Boolean(this.launchParams || this.rawInitData || this.webApp?.initData)
	}

	/**
	 * Closes the WebApp
	 */
	close(): void {
		try {
			postEvent("web_app_close")
		} catch (error) {
			log.devError("Failed to close WebApp:", error)
		}
	}

	/**
	 * Expands the WebApp to full height
	 */
	expand(): void {
		try {
			postEvent("web_app_expand")
		} catch (error) {
			log.devError("Failed to expand WebApp:", error)
		}
	}

	/**
	 * Disables vertical swipes
	 */
	disableSwipes(): void {
		try {
			postEvent("web_app_setup_swipe_behavior", {
				allow_vertical_swipe: false,
			})
		} catch (error) {
			log.devError("Failed to disable vertical swipes:", error)
		}
	}

	/**
	 * Requests file download
	 */
	requestFileDownload(url: string, filename: string): void {
		try {
			postEvent("web_app_request_file_download", {
				url,
				file_name: filename,
			})
		} catch (error) {
			log.devError("Failed to request file download:", error)
			this.fallbackDownload(url, filename)
		}
	}

	/**
	 * Fallback download method using regular browser download
	 */
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

	/**
	 * Triggers haptic feedback impact
	 */
	impactOccurred(style: "light" | "medium" | "heavy" | "rigid" | "soft" = "light"): void {
		try {
			this.webApp?.HapticFeedback?.impactOccurred(style)
		} catch (error) {
			log.devWarn("Failed to trigger haptic feedback impact:", error)
		}
	}

	/**
	 * Triggers haptic feedback notification
	 */
	notificationOccurred(type: "error" | "success" | "warning"): void {
		try {
			this.webApp?.HapticFeedback?.notificationOccurred(type)
		} catch (error) {
			log.devWarn("Failed to trigger haptic feedback notification:", error)
		}
	}

	/**
	 * Triggers haptic feedback for selection change
	 */
	selectionChanged(): void {
		try {
			this.webApp?.HapticFeedback?.selectionChanged()
		} catch (error) {
			log.devWarn("Failed to trigger haptic feedback selection:", error)
		}
	}

	/**
	 * Checks if location manager is supported
	 */
	isLocationSupported(): boolean {
		const native = window.Telegram?.WebApp?.LocationManager
		if (native?.isSupported) {
			return true
		}
		const version = this.getVersion()
		if (version === "unknown") return false

		const [major] = version.split(".").map(Number)
		return Number.isFinite(major) && major >= 8
	}

	private static readonly LOCATION_TIMEOUT_MS = 45_000

	/**
	 * Requests current location: prefers official {@link https://core.telegram.org/bots/webapps#locationmanager LocationManager}
	 * (init → getLocation). Falls back to the Mini Apps bridge flow recommended by @tma.js:
	 * {@link https://docs.telegram-mini-apps.com/platform/methods#web-app-check-location web_app_check_location}
	 * then {@link https://docs.telegram-mini-apps.com/platform/methods#web-app-request-location web_app_request_location}.
	 */
	async requestLocation(): Promise<MessengerLocationResult> {
		const timeoutMs = TelegramWrapper.LOCATION_TIMEOUT_MS
		const nativeLm = window.Telegram?.WebApp?.LocationManager

		if (nativeLm?.isSupported) {
			return new Promise((resolve) => {
				let settled = false
				const done = (value: MessengerLocationResult) => {
					if (settled) return
					settled = true
					clearTimeout(timer)
					resolve(value)
				}
				const timer = setTimeout(() => done({ ok: false, reason: "timeout" }), timeoutMs)

				const runGetLocation = () => {
					try {
						nativeLm.getLocation((data) => {
							if (!data) {
								done({ ok: false, reason: "denied" })
								return
							}
							done({
								ok: true,
								latitude: data.latitude,
								longitude: data.longitude,
							})
						})
					} catch {
						done({ ok: false, reason: "unknown" })
					}
				}

				try {
					if (nativeLm.isInited) {
						runGetLocation()
					} else {
						nativeLm.init(() => runGetLocation())
					}
				} catch {
					done({ ok: false, reason: "unknown" })
				}
			})
		}

		let checked: Awaited<ReturnType<typeof request2>>
		try {
			checked = await request2("web_app_check_location", "location_checked", {
				timeout: 15_000,
			})
		} catch (error) {
			if (error instanceof TimeoutError) {
				return { ok: false, reason: "timeout" }
			}
			log.devError("Failed to check location via bridge:", error)
			return { ok: false, reason: "unknown" }
		}

		if (!checked.available) {
			return { ok: false, reason: "unavailable" }
		}

		if (checked.access_requested === true && checked.access_granted === false) {
			return { ok: false, reason: "denied" }
		}

		try {
			const payload = await request2("web_app_request_location", "location_requested", {
				timeout: timeoutMs,
			})

			if (!payload.available) {
				return { ok: false, reason: "denied" }
			}
			return {
				ok: true,
				latitude: payload.latitude,
				longitude: payload.longitude,
			}
		} catch (error) {
			if (error instanceof TimeoutError) {
				return { ok: false, reason: "timeout" }
			}
			log.devError("Failed to request location via bridge:", error)
			return { ok: false, reason: "unknown" }
		}
	}
}
