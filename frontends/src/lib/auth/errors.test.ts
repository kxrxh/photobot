import { HTTPError } from "ky"
import { describe, expect, it } from "vitest"
import {
	GENERIC_AUTH_ERROR_MESSAGE,
	REFRESH_EXPIRED_MESSAGE,
	toUserFacingAuthMessage,
} from "@/lib/auth/messages"
import { emptyNormalizedOptions } from "@/test/fixtures/ky"
import { normalizeAuthError } from "./errors"

describe("toUserFacingAuthMessage", () => {
	it("maps invalid refresh token to a friendly message", () => {
		expect(toUserFacingAuthMessage("invalid refresh token")).toBe(REFRESH_EXPIRED_MESSAGE)
	})

	it("returns generic Russian message for unknown API errors", () => {
		expect(toUserFacingAuthMessage("some unknown api error")).toBe(GENERIC_AUTH_ERROR_MESSAGE)
	})
})

describe("normalizeAuthError", () => {
	it("maps refresh 401 to REFRESH_EXPIRED_MESSAGE regardless of API text", async () => {
		const res = new Response(
			JSON.stringify({
				success: false,
				error: { code: 401, message: "invalid refresh token", path: "/auth/refresh" },
			}),
			{ status: 401 }
		)
		const err = new HTTPError(
			res,
			new Request("https://api.test/auth/refresh"),
			emptyNormalizedOptions
		)
		await expect(normalizeAuthError(err, { isRefreshError: true })).resolves.toEqual({
			kind: "unauthenticated",
			message: REFRESH_EXPIRED_MESSAGE,
		})
	})
})
