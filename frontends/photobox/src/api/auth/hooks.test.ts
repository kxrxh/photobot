import type { KyRequest, KyResponse, NormalizedOptions } from "ky"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { emptyNormalizedOptions } from "@/test/fixtures/ky"

const refreshMock = vi.hoisted(() => vi.fn())
const kyMock = vi.hoisted(() => vi.fn())
const getStoredRefreshTokenMock = vi.hoisted(() => vi.fn())
const getStoredAccessTokenMock = vi.hoisted(() => vi.fn())
const clearStoredAuthMock = vi.hoisted(() => vi.fn())

vi.mock("./refresh", () => ({
	refreshTokensSingleFlight: refreshMock,
}))

vi.mock("@/lib/auth/storage", () => ({
	clearStoredAuth: clearStoredAuthMock,
	getStoredAccessToken: getStoredAccessTokenMock,
	getStoredRefreshToken: getStoredRefreshTokenMock,
	setStoredTokens: vi.fn(),
}))

vi.mock("ky", () => ({
	default: kyMock,
}))

import { authHooks } from "./hooks"

function getAfterResponseHook() {
	const hook = authHooks.afterResponse?.[0]
	if (hook === undefined) {
		throw new Error("expected authHooks.afterResponse[0]")
	}
	return hook
}

const afterResponseHook = getAfterResponseHook()

const retryState = { retryCount: 0 }

async function invokeAfterResponse(
	request: Request,
	options: NormalizedOptions,
	response: Response
) {
	return afterResponseHook(request as KyRequest, options, response as KyResponse, retryState)
}

describe("authHooks.afterResponse", () => {
	let hrefValue: string

	beforeEach(() => {
		hrefValue = ""
		vi.stubGlobal(
			"location",
			Object.defineProperty({} as Location, "href", {
				configurable: true,
				enumerable: true,
				get: () => hrefValue,
				set: (v: string) => {
					hrefValue = v
				},
			}) as Location
		)
		refreshMock.mockReset()
		kyMock.mockReset()
		getStoredRefreshTokenMock.mockReset()
		getStoredAccessTokenMock.mockReset()
		clearStoredAuthMock.mockReset()
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it("returns the response unchanged when status is not 401", async () => {
		const request = new Request("https://api.example.com/data")
		const response = new Response(null, { status: 200 })
		const out = await invokeAfterResponse(request, emptyNormalizedOptions, response)
		expect(out).toBe(response)
		expect(refreshMock).not.toHaveBeenCalled()
	})

	it("does not retry when the request is the auth refresh endpoint", async () => {
		const request = new Request("https://auth.example.com/auth/refresh")
		const response = new Response(null, { status: 401 })
		const out = await invokeAfterResponse(request, emptyNormalizedOptions, response)
		expect(out).toBe(response)
		expect(refreshMock).not.toHaveBeenCalled()
	})

	it("does not retry when context.retried is set", async () => {
		const request = new Request("https://api.example.com/data")
		const response = new Response(null, { status: 401 })
		const out = await invokeAfterResponse(
			request,
			{ context: { retried: true } } as unknown as NormalizedOptions,
			response
		)
		expect(out).toBe(response)
		expect(refreshMock).not.toHaveBeenCalled()
	})

	it("clears auth and redirects when there is no refresh token", async () => {
		getStoredRefreshTokenMock.mockReturnValue(null)
		const request = new Request("https://api.example.com/data")
		const response = new Response(null, { status: 401 })
		const out = await invokeAfterResponse(request, emptyNormalizedOptions, response)
		expect(out).toBe(response)
		expect(clearStoredAuthMock).toHaveBeenCalledTimes(1)
		expect(hrefValue).toBe("/")
		expect(refreshMock).not.toHaveBeenCalled()
	})

	it("refreshes, retries with ky, and returns the retry response on success", async () => {
		getStoredRefreshTokenMock.mockReturnValue("refresh-token")
		refreshMock.mockResolvedValue(undefined)
		getStoredAccessTokenMock.mockReturnValue("new-access")
		const retryResponse = new Response(null, { status: 200 })
		kyMock.mockResolvedValue(retryResponse)

		const request = new Request("https://api.example.com/data", {
			headers: { Authorization: "Bearer old" },
		})
		const response401 = new Response(null, { status: 401 })
		const options = { retry: { limit: 2 } } as NormalizedOptions

		const out = await invokeAfterResponse(request, options, response401)

		expect(refreshMock).toHaveBeenCalledTimes(1)
		expect(kyMock).toHaveBeenCalledTimes(1)
		const firstCall = kyMock.mock.calls[0]
		expect(firstCall).toBeDefined()
		const kyArg0 = firstCall[0] as Request
		expect(kyArg0.headers.get("Authorization")).toBe("Bearer new-access")
		const kyArg1 = firstCall[1] as { context: { retried: boolean }; retry: number }
		expect(kyArg1.context.retried).toBe(true)
		expect(kyArg1.retry).toBe(0)
		expect(out).toBe(retryResponse)
	})

	it("clears auth and redirects when refresh fails", async () => {
		getStoredRefreshTokenMock.mockReturnValue("refresh-token")
		refreshMock.mockRejectedValue(new Error("refresh failed"))

		const request = new Request("https://api.example.com/data")
		const response401 = new Response(null, { status: 401 })

		const out = await invokeAfterResponse(request, emptyNormalizedOptions, response401)

		expect(refreshMock).toHaveBeenCalledTimes(1)
		expect(kyMock).not.toHaveBeenCalled()
		expect(clearStoredAuthMock).toHaveBeenCalledTimes(1)
		expect(hrefValue).toBe("/")
		expect(out).toBe(response401)
	})
})
