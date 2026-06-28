import { isDevMessengerMock, isTelegramMessengerEnvironment } from "@/lib/auth/mode"
import { BrowserBridge } from "./browserBridge"
import { DevBridge } from "./devBridge"
import { MaxBridge } from "./maxBridge"
import { TelegramBridge } from "./telegramBridge"
import type { MessengerBridge } from "./types"
import { isMaxEnvironment } from "./wrappers/max"

type BridgeKind = "max" | "telegram" | "dev" | "browser"

const messengerKinds = new Set<BridgeKind>(["max", "telegram", "dev"])

let bridgeInstance: MessengerBridge | null = null
let bridgeKind: BridgeKind | null = null

function resolveBridgeKind(): BridgeKind {
	if (isMaxEnvironment()) return "max"
	if (isTelegramMessengerEnvironment()) return "telegram"
	if (import.meta.env.DEV && isDevMessengerMock()) return "dev"
	return "browser"
}

function createBridge(kind: BridgeKind): MessengerBridge {
	switch (kind) {
		case "max":
			return new MaxBridge()
		case "telegram":
			return new TelegramBridge()
		case "dev":
			return new DevBridge()
		case "browser":
			return new BrowserBridge()
	}
}

export function createMessengerBridge(): MessengerBridge {
	const kind = resolveBridgeKind()

	if (bridgeInstance && bridgeKind === kind) {
		return bridgeInstance
	}

	if (bridgeInstance && bridgeKind === "browser" && messengerKinds.has(kind)) {
		bridgeInstance = createBridge(kind)
		bridgeKind = kind
		return bridgeInstance
	}

	if (bridgeInstance && bridgeKind !== "browser") {
		return bridgeInstance
	}

	bridgeInstance = createBridge(kind)
	bridgeKind = kind
	return bridgeInstance
}

export function resetMessengerBridge(): void {
	bridgeInstance = null
	bridgeKind = null
}

export type { MessengerBridge, MessengerUser } from "./types"
