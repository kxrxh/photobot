import { HTTPError, TimeoutError } from "ky"
import { describe, expect, it } from "vitest"
import { REFRESH_EXPIRED_MESSAGE } from "@/lib/auth/messages"
import { emptyNormalizedOptions } from "@/test/fixtures/ky"
import {
	getAuthErrorMessage,
	getUserFacingErrorMessage,
	isInitDataExpiredResponse,
	SESSION_EXPIRED_MESSAGE,
} from "./errors"

describe("isInitDataExpiredResponse", () => {
	it("returns false for null or non-object", () => {
		expect(isInitDataExpiredResponse(null)).toBe(false)
		expect(isInitDataExpiredResponse(undefined)).toBe(false)
		expect(isInitDataExpiredResponse("string")).toBe(false)
	})

	it("returns true when message contains 'expired'", () => {
		expect(isInitDataExpiredResponse({ message: "Init data expired" })).toBe(true)
	})

	it("returns true when error contains 'telegram data validation'", () => {
		expect(
			isInitDataExpiredResponse({ error: { message: "telegram data validation failed" } })
		).toBe(true)
	})

	it("returns true when details contains 'init data'", () => {
		expect(isInitDataExpiredResponse({ error: { details: "init data invalid" } })).toBe(true)
	})

	it("returns false when none of the keywords match", () => {
		expect(isInitDataExpiredResponse({ message: "Something went wrong" })).toBe(false)
	})

	it("returns false when message is about refresh token expiry", () => {
		expect(isInitDataExpiredResponse({ message: "refresh token expired or already used" })).toBe(
			false
		)
		expect(
			isInitDataExpiredResponse({ error: { message: "Refresh token expired or already used" } })
		).toBe(false)
	})
})

describe("getUserFacingErrorMessage", () => {
	it("returns timeout message for ky TimeoutError", () => {
		const err = new TimeoutError(new Request("https://api.test/analyses", { method: "POST" }))
		expect(getUserFacingErrorMessage(err)).toBe(
			"Превышено время ожидания. Проверьте интернет-соединение; при загрузке файлов попробуйте меньше файлов за раз или более стабильную сеть."
		)
	})

	it("returns 401 message for HTTP 401", () => {
		const err = new HTTPError(
			new Response(null, { status: 401 }),
			new Request("https://api.test/"),
			emptyNormalizedOptions
		)
		expect(getUserFacingErrorMessage(err)).toBe("Ошибка авторизации.")
	})

	it("returns 403 message for HTTP 403", () => {
		const err = new HTTPError(
			new Response(null, { status: 403 }),
			new Request("https://api.test/"),
			emptyNormalizedOptions
		)
		expect(getUserFacingErrorMessage(err)).toBe("Доступ запрещён.")
	})

	it("returns 404 message for HTTP 404", () => {
		const err = new HTTPError(
			new Response(null, { status: 404 }),
			new Request("https://api.test/"),
			emptyNormalizedOptions
		)
		expect(getUserFacingErrorMessage(err)).toBe("Не найдено.")
	})

	it("returns server error message for 5xx", () => {
		const err = new HTTPError(
			new Response(null, { status: 500 }),
			new Request("https://api.test/"),
			emptyNormalizedOptions
		)
		expect(getUserFacingErrorMessage(err)).toBe("Временная ошибка сервера. Попробуйте позже.")
	})

	it("returns generic request error for other HTTP status", () => {
		const err = new HTTPError(
			new Response(null, { status: 400 }),
			new Request("https://api.test/"),
			emptyNormalizedOptions
		)
		expect(getUserFacingErrorMessage(err)).toBe("Произошла ошибка при запросе.")
	})

	it("returns network message for TypeError", () => {
		expect(getUserFacingErrorMessage(new TypeError("Failed to fetch"))).toBe(
			"Сервер не отвечает. Пожалуйста, проверьте ваше интернет-соединение или попробуйте позже."
		)
	})

	it("returns error message for plain Error when no URL in message", () => {
		expect(getUserFacingErrorMessage(new Error("Custom error"))).toBe("Custom error")
	})

	it("returns generic message when Error contains URL (to avoid exposure)", () => {
		expect(getUserFacingErrorMessage(new Error("Failed: https://internal.api/secret"))).toBe(
			"Произошла непредвиденная ошибка."
		)
	})

	it("returns generic message for non-Error", () => {
		expect(getUserFacingErrorMessage("unknown")).toBe("Произошла непредвиденная ошибка.")
	})
})

describe("getAuthErrorMessage", () => {
	it("returns SESSION_EXPIRED_MESSAGE for 401 with expired init data", async () => {
		const res = new Response(JSON.stringify({ message: "init data expired" }), {
			status: 401,
		})
		const err = new HTTPError(
			res,
			new Request("https://api.test/auth/login"),
			emptyNormalizedOptions
		)
		expect(await getAuthErrorMessage(err)).toBe(SESSION_EXPIRED_MESSAGE)
	})

	it("falls through to getUserFacingErrorMessage for non-401", async () => {
		const res = new Response(JSON.stringify({}), { status: 500 })
		const err = new HTTPError(res, new Request("https://api.test/"), emptyNormalizedOptions)
		expect(await getAuthErrorMessage(err)).toBe("Временная ошибка сервера. Попробуйте позже.")
	})

	it("returns Russian message for login already taken on register", async () => {
		const res = new Response(
			JSON.stringify({
				success: false,
				error: { code: 409, message: "login already taken", path: "/auth/register-web" },
			}),
			{ status: 409 }
		)
		const err = new HTTPError(
			res,
			new Request("https://api.test/auth/register-web"),
			emptyNormalizedOptions
		)
		expect(await getAuthErrorMessage(err)).toBe("Логин уже занят")
	})

	it("returns Russian message for invalid credentials", async () => {
		const res = new Response(
			JSON.stringify({
				success: false,
				error: { code: 401, message: "invalid credentials", path: "/auth/login" },
			}),
			{ status: 401 }
		)
		const err = new HTTPError(
			res,
			new Request("https://api.test/auth/login"),
			emptyNormalizedOptions
		)
		expect(await getAuthErrorMessage(err)).toBe("Неверный логин или пароль")
	})

	it("returns REFRESH_EXPIRED_MESSAGE for 401 on refresh endpoint", async () => {
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
		expect(await getAuthErrorMessage(err)).toBe(REFRESH_EXPIRED_MESSAGE)
	})

	it("does not expose raw ky HTTPError message with URL", async () => {
		const res = new Response(null, { status: 401 })
		const err = new HTTPError(
			res,
			new Request("https://csort-superset.ru/auth/api/v1/auth/login"),
			emptyNormalizedOptions
		)
		expect(await getAuthErrorMessage(err)).toBe("Ошибка авторизации.")
	})
})
