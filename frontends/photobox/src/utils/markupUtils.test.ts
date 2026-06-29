import { describe, expect, it } from "vitest"
import { kalibriObject } from "@/test/factories/kalibriObject"
import { createObjectLookup } from "./markupUtils"

describe("createObjectLookup", () => {
	it("returns empty lookup for empty map and ids", () => {
		expect(createObjectLookup({}, [])).toEqual({})
	})

	it("builds lookup with analysisId", () => {
		const obj = kalibriObject(1)
		const result = createObjectLookup({ a1: [obj] }, ["a1"])
		expect(result[1]).toEqual({ ...obj, analysisId: "a1" })
	})

	it("handles numeric analysisIds", () => {
		const obj = kalibriObject(1)
		const result = createObjectLookup({ "42": [obj] }, [42])
		expect(result[1].analysisId).toBe("42")
	})

	it("overwrites when same object id appears in multiple analyses", () => {
		const obj1 = kalibriObject(1)
		const obj2 = kalibriObject(1, { l: 99 })
		const result = createObjectLookup({ a1: [obj1], a2: [obj2] }, ["a1", "a2"])
		// Later analysis wins (a2 processed after a1)
		expect(result[1].analysisId).toBe("a2")
		expect(result[1].l).toBe(99)
	})

	it("ignores missing analysis ids in map", () => {
		const result = createObjectLookup({}, ["a1", "a2"])
		expect(result).toEqual({})
	})

	it("merges objects from multiple analyses", () => {
		const obj1 = kalibriObject(1)
		const obj2 = kalibriObject(2)
		const result = createObjectLookup({ a1: [obj1], a2: [obj2] }, ["a1", "a2"])
		expect(result[1].analysisId).toBe("a1")
		expect(result[2].analysisId).toBe("a2")
	})
})
