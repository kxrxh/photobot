import type { ThemeParams } from "@tma.js/types"
import { useEffect, useMemo, useState } from "react"
import { createMessengerBridge } from "@/messenger"
import type { MessengerBridge, MessengerLocationResult } from "@/messenger/types"
import type { UseMessengerReturnType } from "@/types/telegram"
import { log } from "@/utils/log"

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

const unavailableLocation: Promise<MessengerLocationResult> = Promise.resolve({
	ok: false,
	reason: "unavailable",
})

function buildFallbackReturn(): UseMessengerReturnType {
	return {
		user: null,
		platform: "unknown",
		version: "unknown",
		chatInstance: undefined,
		chatType: undefined,
		startParam: undefined,
		isReady: false,
		initData: "",
		themeParams: undefined,
		isDark: false,
		close: () => {},
		expand: () => {},
		disableSwipes: () => {},
		requestFileDownload: baseDownload,
		hapticFeedback: {
			impactOccurred: () => {},
			notificationOccurred: () => {},
			selectionChanged: () => {},
		},
		location: {
			isSupported: () => false,
			requestLocation: async (): Promise<MessengerLocationResult> => ({
				ok: false,
				reason: "unavailable",
			}),
		},
	}
}

/**
 * Hook that provides messenger platform API (Telegram, MAX, or dev mock).
 * Chooses implementation via createMessengerBridge().
 * Return value is referentially stable across renders when underlying snapshot fields are unchanged.
 */
function loadMessengerBridge(): MessengerBridge | null {
	try {
		return createMessengerBridge()
	} catch (error) {
		log.devWarn("Failed to initialize messenger bridge:", error)
		return null
	}
}

export const useMessenger = (): UseMessengerReturnType => {
	const [bridge, setBridge] = useState<MessengerBridge | null>(loadMessengerBridge)

	useEffect(() => {
		const refresh = () => {
			const next = loadMessengerBridge()
			setBridge((prev) => (prev === next ? prev : next))
			return next
		}

		const current = refresh()
		if (current?.isReady()) return

		let attempts = 0
		const id = window.setInterval(() => {
			attempts += 1
			const next = refresh()
			if (next?.isReady() || attempts >= 40) {
				clearInterval(id)
			}
		}, 50)

		return () => clearInterval(id)
	}, [])

	const hapticFeedback = useMemo(
		() => ({
			impactOccurred: (style?: Parameters<MessengerBridge["impactOccurred"]>[0]) =>
				bridge?.impactOccurred(style),
			notificationOccurred: (type: Parameters<MessengerBridge["notificationOccurred"]>[0]) =>
				bridge?.notificationOccurred(type),
			selectionChanged: () => bridge?.selectionChanged(),
		}),
		[bridge]
	)

	const location = useMemo(
		() => ({
			isSupported: () => bridge?.isLocationSupported() ?? false,
			requestLocation: () => bridge?.requestLocation() ?? unavailableLocation,
		}),
		[bridge]
	)

	const isReady = bridge?.isReady() ?? false
	const initData = bridge?.getInitData() ?? ""
	const platform = bridge?.getPlatform()
	const version = bridge?.getVersion()
	const chatInstance = bridge?.getChatInstance()
	const chatType = bridge?.getChatType()
	const startParam = bridge?.getStartParam()
	const isDark = bridge?.isDark() ?? false

	return useMemo(() => {
		if (!bridge) {
			return buildFallbackReturn()
		}

		return {
			user: bridge.getUser() as UseMessengerReturnType["user"],
			platform,
			version,
			chatInstance,
			chatType,
			startParam,
			isReady,
			initData,
			themeParams: bridge.getThemeParams() as ThemeParams | undefined,
			isDark,
			close: () => bridge.close(),
			expand: () => bridge.expand(),
			disableSwipes: () => bridge.disableSwipes(),
			requestFileDownload: (url: string, filename: string) =>
				bridge.requestFileDownload(url, filename),
			hapticFeedback,
			location,
		}
	}, [
		bridge,
		isReady,
		initData,
		platform,
		version,
		chatInstance,
		chatType,
		startParam,
		isDark,
		hapticFeedback,
		location,
	])
}
