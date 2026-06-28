import { describe, expect, it } from "vitest"
import { kalibriObject } from "@/test/factories/kalibriObject"
import {
	calculateFieldStatisticsResult,
	calculateKernelBrokenPercentage,
} from "@/utils/analysisStatistics"

describe("calculateFieldStatisticsResult", () => {
	it("computes l and l_w from objects", () => {
		const objects = [
			kalibriObject(1, { l: 10, w: 2, sq: 20 }),
			kalibriObject(2, { l: 12, w: 3, sq: 36 }),
		]
		const r = calculateFieldStatisticsResult(objects)
		expect(r.l?.min).toBe(10)
		expect(r.l?.max).toBe(12)
		expect(r.l?.avg).toBe(11)
		expect(r.l_w?.min).toBeCloseTo(12 / 3)
		expect(r.l_w?.max).toBeCloseTo(10 / 2)
	})

	it("returns empty object for empty list", () => {
		expect(calculateFieldStatisticsResult([])).toEqual({})
	})
})

describe("calculateKernelBrokenPercentage", () => {
	it("returns null for non-kernel product", () => {
		const objects = [kalibriObject(1, { l: 4, w: 2, class: "broken" })]
		expect(calculateKernelBrokenPercentage(objects, "seeds")).toBeNull()
	})

	it("returns percentage for kernel product with broken class", () => {
		const objects = [
			kalibriObject(1, { l: 4, w: 2, class: "whole" }),
			kalibriObject(2, { l: 4, w: 2, class: "broken" }),
		]
		const p = calculateKernelBrokenPercentage(objects, "kernels")
		expect(p).not.toBeNull()
		expect(p).toBeGreaterThan(0)
		expect(p).toBeLessThanOrEqual(100)
	})
})
