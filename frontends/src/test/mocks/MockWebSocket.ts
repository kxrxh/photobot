import { vi } from "vitest"

/** Minimal WebSocket stub for hook tests; tracks instances for reconnect assertions. */
export class MockWebSocket {
	static readonly CONNECTING = 0
	static readonly OPEN = 1
	static readonly CLOSING = 2
	static readonly CLOSED = 3

	static instances: MockWebSocket[] = []

	readonly url: string
	readyState = MockWebSocket.CONNECTING
	onopen: ((event: Event) => void) | null = null
	onmessage: ((event: MessageEvent) => void) | null = null
	onclose: ((event: CloseEvent) => void) | null = null
	onerror: ((event: Event) => void) | null = null
	send = vi.fn()
	close = vi.fn(() => {
		this.readyState = MockWebSocket.CLOSED
	})

	constructor(url: string) {
		this.url = url
		MockWebSocket.instances.push(this)
	}

	triggerOpen() {
		this.readyState = MockWebSocket.OPEN
		this.onopen?.(new Event("open"))
	}

	triggerClose(code: number, reason = "closed") {
		this.readyState = MockWebSocket.CLOSED
		this.onclose?.({ code, reason } as CloseEvent)
	}
}
