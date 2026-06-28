import { afterEach, describe, expect, it, vi } from "vitest"
import { isDevMessengerMock, isMessengerEnvironment, isWebAuthMode } from "./mode"

describe("auth mode", () => {
	afterEach(() => {
		vi.unstubAllEnvs()
		window.history.replaceState({}, "", "/")
		;(window as Window & { Telegram?: unknown }).Telegram = undefined
		;(window as Window & { WebApp?: unknown }).WebApp = undefined
		;(window as Window & { TelegramWebviewProxy?: unknown }).TelegramWebviewProxy = undefined
	})

	it("uses web auth in a plain browser", () => {
		expect(isMessengerEnvironment()).toBe(false)
		expect(isWebAuthMode()).toBe(true)
	})

	it("uses messenger auth when Telegram provides initData", () => {
		window.Telegram = { WebApp: { initData: "user=..." } }
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("uses messenger auth when Telegram WebApp is injected without initData yet", () => {
		window.Telegram = { WebApp: { platform: "ios", version: "8.0" } }
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("uses messenger auth when Telegram launch params are in the URL", () => {
		window.history.replaceState({}, "", "/#tgWebAppData=eyJ1c2VyIjp7ImlkIjoxfX0")
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("uses messenger auth when TelegramWebviewProxy is present", () => {
		;(window as Window & { TelegramWebviewProxy?: unknown }).TelegramWebviewProxy = {
			postEvent: () => {},
		}
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("uses messenger auth when MAX provides initData", () => {
		window.WebApp = { initData: "mock-init-data" }
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("uses messenger auth when MAX host exposes initDataUnsafe before initData string", () => {
		window.WebApp = {
			initDataUnsafe: { user: { id: 1, first_name: "Max", last_name: "User" } },
		}
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})

	it("ignores empty messenger stubs in the browser", () => {
		window.WebApp = {}
		window.Telegram = { WebApp: {} }
		expect(isMessengerEnvironment()).toBe(false)
		expect(isWebAuthMode()).toBe(true)
	})

	it("supports dev messenger mock via query param", () => {
		vi.stubEnv("DEV", true)
		window.history.replaceState({}, "", "/?mock=messenger")
		expect(isDevMessengerMock()).toBe(true)
		expect(isMessengerEnvironment()).toBe(true)
		expect(isWebAuthMode()).toBe(false)
	})
})
