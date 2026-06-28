import { describe, expect, it } from "vitest"
import { extractAuthServiceMessage, isRefreshRequest } from "./helpers"

describe("extractAuthServiceMessage", () => {
	it("returns null for null or non-object", () => {
		expect(extractAuthServiceMessage(null)).toBe(null)
		expect(extractAuthServiceMessage(undefined)).toBe(null)
		expect(extractAuthServiceMessage("string")).toBe(null)
	})

	it("returns null when error object is missing", () => {
		expect(extractAuthServiceMessage({})).toBe(null)
		expect(extractAuthServiceMessage({ success: false })).toBe(null)
	})

	it("returns message from error.message", () => {
		expect(extractAuthServiceMessage({ error: { message: "Invalid credentials" } })).toBe(
			"Invalid credentials"
		)
	})

	it("returns null for empty message", () => {
		expect(extractAuthServiceMessage({ error: { message: "" } })).toBe(null)
	})

	it("returns null when message contains URL", () => {
		expect(extractAuthServiceMessage({ error: { message: "See https://example.com" } })).toBe(null)
	})

	it("returns null when message exceeds 500 chars", () => {
		expect(extractAuthServiceMessage({ error: { message: "x".repeat(501) } })).toBe(null)
	})

	it("returns null when error is not an object", () => {
		expect(extractAuthServiceMessage({ error: "string" })).toBe(null)
	})
})

describe("isRefreshRequest", () => {
	it("returns true for URL ending with /auth/refresh", () => {
		expect(isRefreshRequest("https://auth.example.com/auth/refresh")).toBe(true)
		expect(isRefreshRequest("http://localhost:3000/api/auth/refresh")).toBe(true)
	})

	it("returns false for URL without auth/refresh", () => {
		expect(isRefreshRequest("https://auth.example.com/auth/login")).toBe(false)
		expect(isRefreshRequest("https://example.com/")).toBe(false)
	})

	it("handles invalid URL via fallback", () => {
		expect(isRefreshRequest("not-a-url-but-has-auth/refresh")).toBe(true)
	})

	it("returns false for string without auth/refresh", () => {
		expect(isRefreshRequest("random/path")).toBe(false)
	})
})
