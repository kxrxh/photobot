import type { ThemeParams } from "@tma.js/types"
import type { MessengerBridge, MessengerLocationResult } from "./types"

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

export class BrowserBridge implements MessengerBridge {
	getUser() {
		return null
	}

	getPlatform() {
		return "browser"
	}

	getVersion() {
		return "0"
	}

	getChatInstance() {
		return undefined
	}

	getChatType() {
		return undefined
	}

	getStartParam() {
		return undefined
	}

	isReady() {
		return true
	}

	getInitData() {
		return ""
	}

	getThemeParams(): ThemeParams | undefined {
		return undefined
	}

	isDark() {
		return false
	}

	close() {}

	expand() {}

	disableSwipes() {}

	requestFileDownload(url: string, filename: string) {
		baseDownload(url, filename)
	}

	impactOccurred() {}

	notificationOccurred() {}

	selectionChanged() {}

	isLocationSupported() {
		return false
	}

	async requestLocation(): Promise<MessengerLocationResult> {
		return { ok: false, reason: "unavailable" }
	}
}
