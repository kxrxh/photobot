import type { MessengerBridge, MessengerUser } from "./types"
import { TelegramWrapper } from "./wrappers/telegram"

export class TelegramBridge implements MessengerBridge {
	private readonly wrapper: TelegramWrapper

	constructor() {
		this.wrapper = new TelegramWrapper()
	}

	getUser(): MessengerUser | null {
		return this.wrapper.getUser() as MessengerUser | null
	}

	getPlatform(): string {
		return this.wrapper.getPlatform()
	}

	getVersion(): string {
		return this.wrapper.getVersion()
	}

	getChatInstance(): string | undefined {
		return this.wrapper.getChatInstance()
	}

	getChatType(): string | undefined {
		return this.wrapper.getChatType()
	}

	getStartParam(): string | undefined {
		return this.wrapper.getStartParam()
	}

	isReady(): boolean {
		return this.wrapper.isReady()
	}

	getInitData(): string {
		return this.wrapper.getInitData()
	}

	getThemeParams() {
		return this.wrapper.getThemeParams()
	}

	isDark(): boolean {
		return this.wrapper.isDark()
	}

	close(): void {
		this.wrapper.close()
	}

	expand(): void {
		this.wrapper.expand()
	}

	disableSwipes(): void {
		this.wrapper.disableSwipes()
	}

	requestFileDownload(url: string, filename: string): void {
		this.wrapper.requestFileDownload(url, filename)
	}

	impactOccurred(style?: "light" | "medium" | "heavy" | "rigid" | "soft"): void {
		this.wrapper.impactOccurred(style)
	}

	notificationOccurred(type: "error" | "success" | "warning"): void {
		this.wrapper.notificationOccurred(type)
	}

	selectionChanged(): void {
		this.wrapper.selectionChanged()
	}

	isLocationSupported(): boolean {
		return this.wrapper.isLocationSupported()
	}

	requestLocation() {
		return this.wrapper.requestLocation()
	}
}
