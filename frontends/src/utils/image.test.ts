import { describe, expect, it } from "vitest"
import {
	getAllImageDataUrls,
	getAnalysisOutputUrls,
	getAnalysisSourceUrls,
	getImageDataUrl,
	getObjectImageUrl,
	isDirectlyRenderableImageValue,
} from "./image"

describe("getImageDataUrl", () => {
	it("returns null for undefined or empty", () => {
		expect(getImageDataUrl(undefined)).toBe(null)
		expect(getImageDataUrl("")).toBe(null)
	})

	it("returns data URL as-is when already a data URL", () => {
		const dataUrl = "data:image/png;base64,iVBORw0KGgo="
		expect(getImageDataUrl(dataUrl)).toBe(dataUrl)
	})

	it("returns http/https URL as-is", () => {
		expect(getImageDataUrl("https://example.com/image.jpg")).toBe("https://example.com/image.jpg")
		expect(getImageDataUrl("http://example.com/image.jpg")).toBe("http://example.com/image.jpg")
	})

	it("returns null for storage key values", () => {
		expect(getImageDataUrl("analysis/123/source_0.jpg")).toBe(null)
	})

	it("handles arrays and returns first renderable URL", () => {
		expect(getImageDataUrl(["analysis/123/source_0.jpg", "https://example.com/image.jpg"])).toBe(
			"https://example.com/image.jpg"
		)
	})

	it("returns null for array with no renderable values", () => {
		expect(getImageDataUrl(["notbase64", "alsonot"])).toBe(null)
	})
})

describe("isDirectlyRenderableImageValue", () => {
	it("returns false for non-renderable storage keys", () => {
		expect(isDirectlyRenderableImageValue("analysis/123/source_0.jpg")).toBe(false)
	})

	it("returns true for data URL and http URL", () => {
		expect(isDirectlyRenderableImageValue("data:image/jpeg;base64,/9j/4AAQ")).toBe(true)
		expect(isDirectlyRenderableImageValue("https://example.com/image.jpg")).toBe(true)
	})
})

describe("getAllImageDataUrls", () => {
	it("returns empty array for undefined", () => {
		expect(getAllImageDataUrls(undefined)).toEqual([])
	})

	it("returns single URL for one valid renderable URL", () => {
		const url = "https://example.com/image.jpg"
		expect(getAllImageDataUrls(url)).toEqual([url])
	})

	it("filters non-renderable values from arrays", () => {
		expect(
			getAllImageDataUrls([
				"analysis/123/source_0.jpg",
				"https://example.com/image-1.jpg",
				"data:image/png;base64,iVBORw0KGgo=",
			])
		).toEqual(["https://example.com/image-1.jpg", "data:image/png;base64,iVBORw0KGgo="])
	})
})

describe("getAnalysisSourceUrls", () => {
	it("prefers files_source_urls when present", () => {
		const urls = ["https://a.com/1.jpg", "https://a.com/2.jpg"]
		expect(
			getAnalysisSourceUrls({
				files_source_urls: urls,
				files_source: ["analysis/old"],
			})
		).toEqual(urls)
	})

	it("falls back to files_source direct URLs", () => {
		expect(getAnalysisSourceUrls({ files_source: ["https://a.com/1.jpg"] })).toEqual([
			"https://a.com/1.jpg",
		])
	})

	it("filters invalid entries from files_source_urls and keeps only renderable URLs", () => {
		expect(
			getAnalysisSourceUrls({
				files_source_urls: ["analysis/bad/key.jpg", "https://a.com/ok.jpg", ""],
				files_source: ["analysis/ignored"],
			})
		).toEqual(["https://a.com/ok.jpg"])
	})

	it("falls back to files_source when files_source_urls has no renderable URLs", () => {
		expect(
			getAnalysisSourceUrls({
				files_source_urls: ["analysis/only-keys.jpg", ""],
				files_source: ["https://fallback.com/1.jpg"],
			})
		).toEqual(["https://fallback.com/1.jpg"])
	})
})

describe("getAnalysisOutputUrls", () => {
	it("prefers files_output_urls when present", () => {
		const urls = ["https://a.com/out.jpg"]
		expect(
			getAnalysisOutputUrls({
				files_output_urls: urls,
				files_output: ["analysis/old"],
			})
		).toEqual(urls)
	})

	it("filters invalid entries from files_output_urls and falls back to files_output", () => {
		expect(
			getAnalysisOutputUrls({
				files_output_urls: ["analysis/bad.jpg"],
				files_output: ["https://fallback.com/out.jpg"],
			})
		).toEqual(["https://fallback.com/out.jpg"])
	})
})

describe("getObjectImageUrl", () => {
	it("returns image_url when present", () => {
		const url = "https://example.com/obj.jpg"
		expect(getObjectImageUrl({ image_url: url })).toBe(url)
	})

	it("falls back to file when it is a direct URL", () => {
		const url = "https://legacy.com/pic.jpg"
		expect(getObjectImageUrl({ file: url })).toBe(url)
	})

	it("returns null for non-renderable object file values", () => {
		expect(getObjectImageUrl({ file: "analysis/123/object.jpg" })).toBe(null)
	})
})
