import { afterEach, beforeEach, describe, expect, it } from "vitest"
import { resolveChatPlatform } from "./messengerPlatform"

describe("resolveChatPlatform", () => {
	it("maps Telegram WebApp platforms to telegram", () => {
		expect(resolveChatPlatform("android")).toBe("telegram")
		expect(resolveChatPlatform("ios")).toBe("telegram")
		expect(resolveChatPlatform("tdesktop")).toBe("telegram")
		expect(resolveChatPlatform("telegram")).toBe("telegram")
	})

	it("keeps max as max", () => {
		expect(resolveChatPlatform("max")).toBe("max")
	})

	it("returns undefined for unknown or empty platform", () => {
		expect(resolveChatPlatform("unknown")).toBeUndefined()
		expect(resolveChatPlatform("mock_platform")).toBeUndefined()
		expect(resolveChatPlatform("")).toBeUndefined()
	})

	describe("when MAX WebApp runtime is detected", () => {
		const win = window as Window & {
			WebApp?: { initData?: string; InitData?: string }
		}
		let originalWebApp: (typeof win)["WebApp"]

		beforeEach(() => {
			originalWebApp = win.WebApp
			win.WebApp = { initData: "mock-init-data" }
		})

		afterEach(() => {
			win.WebApp = originalWebApp
		})

		it("prefers max for telegram-style platform strings", () => {
			expect(resolveChatPlatform("android")).toBe("max")
			expect(resolveChatPlatform("ios")).toBe("max")
		})
	})
})
