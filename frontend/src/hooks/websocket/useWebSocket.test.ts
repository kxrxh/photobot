import { act, renderHook } from "@testing-library/react"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { MockWebSocket } from "@/test/mocks/MockWebSocket"
import { useWebSocket } from "./useWebSocket"

describe("useWebSocket", () => {
	const originalWebSocket = globalThis.WebSocket

	beforeEach(() => {
		vi.useFakeTimers()
		MockWebSocket.instances = []
		globalThis.WebSocket = MockWebSocket as unknown as typeof WebSocket
	})

	afterEach(() => {
		vi.clearAllTimers()
		vi.useRealTimers()
		globalThis.WebSocket = originalWebSocket
	})

	it("reconnects after auth close code 4001 to allow re-auth", () => {
		// Reconnect delay includes random jitter (see nextReconnectDelayMs); fix it so timer advances are deterministic.
		const randomSpy = vi.spyOn(Math, "random").mockReturnValue(0)
		try {
			renderHook(() =>
				useWebSocket({
					url: "ws://localhost:3000/ws",
					enabled: true,
					reconnectInterval: 10,
					reconnectAttempts: 2,
				})
			)

			act(() => {
				vi.advanceTimersByTime(100)
			})

			expect(MockWebSocket.instances).toHaveLength(1)

			act(() => {
				MockWebSocket.instances[0]?.triggerClose(4001, "auth failed")
				vi.advanceTimersByTime(10)
			})

			expect(MockWebSocket.instances).toHaveLength(2)
		} finally {
			randomSpy.mockRestore()
		}
	})
})
