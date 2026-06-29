import { describe, expect, it } from "vitest"
import { kalibriObject } from "@/test/factories/kalibriObject"
import {
	buildSmartRangesFromObjects,
	calculateMetricStats,
	calculateStatsFromObjects,
	convertToBackendStatistics,
} from "./stats"

describe("calculateMetricStats", () => {
	it("returns zeros for empty array", () => {
		expect(calculateMetricStats([])).toEqual({
			min: 0,
			max: 0,
			avg: 0,
			median: 0,
		})
	})

	it("returns correct stats for single value", () => {
		expect(calculateMetricStats([42])).toEqual({
			min: 42,
			max: 42,
			avg: 42,
			median: 42,
		})
	})

	it("returns correct stats for multiple values", () => {
		const result = calculateMetricStats([1, 2, 3, 4, 5])
		expect(result.min).toBe(1)
		expect(result.max).toBe(5)
		expect(result.avg).toBe(3)
		expect(result.median).toBe(3)
	})

	it("returns correct median for even count", () => {
		const result = calculateMetricStats([1, 2, 3, 4])
		expect(result.median).toBe(2.5)
	})
})

describe("calculateStatsFromObjects", () => {
	it("returns null for empty objects", () => {
		expect(calculateStatsFromObjects({}, [])).toBe(null)
	})

	it("returns null when all objects excluded", () => {
		const obj = kalibriObject(1)
		expect(calculateStatsFromObjects({ a1: [obj] }, [1])).toBe(null)
	})

	it("returns stats for included objects", () => {
		const obj1 = kalibriObject(1, { l: 10, w: 5 })
		const obj2 = kalibriObject(2, { l: 20, w: 10 })
		const result = calculateStatsFromObjects({ a1: [obj1], a2: [obj2] }, [])
		expect(result).not.toBe(null)
		expect(result?.length).toBeDefined()
		expect(result?.width).toBeDefined()
		expect(result?.lwRatio).toBeDefined()
	})

	it("excludes objects by id", () => {
		const obj1 = kalibriObject(1, { l: 10 })
		const obj2 = kalibriObject(2, { l: 100 })
		const result = calculateStatsFromObjects({ a1: [obj1, obj2] }, [2])
		expect(result?.length?.min).toBe(10)
		expect(result?.length?.max).toBe(10)
	})
})

describe("convertToBackendStatistics", () => {
	it("returns null for empty objects", () => {
		expect(convertToBackendStatistics({}, [])).toBe(null)
	})

	it("returns backend format with lw_avg, lw_median, etc.", () => {
		const obj = kalibriObject(1, { l: 10, w: 5 })
		const result = convertToBackendStatistics({ a1: [obj] }, [])
		expect(result).not.toBe(null)
		expect(result?.lw_avg).toBeDefined()
		expect(result?.lw_median).toBeDefined()
		expect(result?.l_avg).toBeDefined()
		expect(result?.w_avg).toBeDefined()
	})
})

describe("buildSmartRangesFromObjects", () => {
	it("returns empty object for empty array", () => {
		expect(buildSmartRangesFromObjects([])).toEqual({})
	})

	it("returns ranges for objects with numeric values", () => {
		const obj = kalibriObject(1, { l: 10, w: 5, h: 3 })
		const result = buildSmartRangesFromObjects([obj], 2)
		expect(result.l_min).toBeDefined()
		expect(result.l_max).toBeDefined()
		expect(result.w_min).toBeDefined()
		expect(result.w_max).toBeDefined()
		expect(result.lw_min).toBeDefined()
		expect(result.lw_max).toBeDefined()
	})
})
