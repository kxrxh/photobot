import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { authRefreshSuccessEnvelope } from "@/test/factories/authRefresh"

const postSpy = vi.hoisted(() => vi.fn())

vi.mock("./client", () => ({
	client: {
		post: (...args: unknown[]) => postSpy(...args),
	},
}))

vi.mock("@/lib/auth/storage", () => ({
	clearStoredAuth: vi.fn(),
	getStoredRefreshToken: vi.fn(() => "stored-rt"),
	setStoredTokens: vi.fn(),
}))

import { refreshTokensSingleFlight } from "./refresh"

describe("refreshTokensSingleFlight", () => {
	beforeEach(() => {
		postSpy.mockReset()
	})

	afterEach(() => {
		vi.clearAllMocks()
	})

	it("deduplicates concurrent refresh calls into one POST", async () => {
		let release!: () => void
		const gate = new Promise<void>((resolve) => {
			release = resolve
		})

		postSpy.mockReturnValue({
			json: async () => {
				await gate
				return authRefreshSuccessEnvelope
			},
		})

		const p1 = refreshTokensSingleFlight()
		const p2 = refreshTokensSingleFlight()

		await vi.waitFor(() => expect(postSpy).toHaveBeenCalledTimes(1))
		release()
		const [a, b] = await Promise.all([p1, p2])
		expect(a).toEqual(authRefreshSuccessEnvelope.result)
		expect(b).toEqual(authRefreshSuccessEnvelope.result)
		expect(postSpy).toHaveBeenCalledTimes(1)
	})

	it("runs a second refresh after the first completes", async () => {
		postSpy.mockReturnValue({
			json: () => Promise.resolve(authRefreshSuccessEnvelope),
		})

		await refreshTokensSingleFlight()
		await refreshTokensSingleFlight()

		expect(postSpy).toHaveBeenCalledTimes(2)
	})
})
