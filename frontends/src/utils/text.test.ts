import { describe, expect, it } from "vitest"
import { truncate, truncateText } from "./text"

describe("truncateText", () => {
	it("returns empty string for empty input", () => {
		expect(truncateText("")).toBe("")
	})

	it("returns text unchanged when shorter than maxLength", () => {
		expect(truncateText("short")).toBe("short")
		expect(truncateText("exactly12chars!", 20)).toBe("exactly12chars!")
	})

	it("returns text unchanged when exactly at maxLength", () => {
		expect(truncateText("exactly17charss!", 17)).toBe("exactly17charss!")
	})

	it("truncates with default suffix when exceeding maxLength", () => {
		expect(truncateText("this is a very long string", 20)).toBe("this is a very lo...")
	})

	it("uses custom maxLength and suffix", () => {
		expect(truncateText("hello world", 8, "…")).toBe("hello w…")
	})

	it("handles truncation when maxLength barely fits suffix", () => {
		// maxLength 5 with suffix "..." (3 chars) leaves 2 for text
		expect(truncateText("hello world", 5, "...")).toBe("he...")
	})
})

describe("truncate", () => {
	it("short truncates to 15 chars", () => {
		// 15 chars total: 12 chars + "..." (3)
		expect(truncate.short("a".repeat(20))).toBe("aaaaaaaaaaaa...")
	})

	it("medium truncates to 30 chars", () => {
		expect(truncate.medium("a".repeat(40))).toBe(`${"a".repeat(27)}...`)
	})

	it("long truncates to 50 chars", () => {
		expect(truncate.long("a".repeat(60))).toBe(`${"a".repeat(47)}...`)
	})

	it("custom truncates to specified length", () => {
		expect(truncate.custom("hello world", 8)).toBe("hello...")
	})
})
