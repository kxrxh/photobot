/**
 * Storage keys (for example, object paths) are not directly renderable in the browser.
 * They must be resolved into presigned URLs by API hydration first.
 */
export function isDirectlyRenderableImageValue(value: string): boolean {
	return value.startsWith("data:") || value.startsWith("http://") || value.startsWith("https://")
}

export const getImageDataUrl = (value: string | string[] | undefined): string | null => {
	if (!value) {
		return null
	}

	if (Array.isArray(value)) {
		for (const item of value) {
			const resolved = getImageDataUrl(item)
			if (resolved) {
				return resolved
			}
		}
		return null
	}

	if (typeof value !== "string" || value.length === 0) {
		return null
	}

	const normalized = value.trim()
	if (!normalized) {
		return null
	}

	return isDirectlyRenderableImageValue(normalized) ? normalized : null
}

export const getAllImageDataUrls = (value: string | string[] | undefined): string[] => {
	if (!value) {
		return []
	}
	if (Array.isArray(value)) {
		return value.map((item) => getImageDataUrl(item)).filter((url): url is string => url !== null)
	}

	const result = getImageDataUrl(value)
	return result ? [result] : []
}

/** Analysis with optional presigned URLs and fallback direct URLs in files_source/files_output. */
interface AnalysisWithImageUrls {
	files_source?: string[]
	files_output?: string[]
	files_source_urls?: string[]
	files_output_urls?: string[]
}

/** Get source image URLs for display. Prefers presigned URLs from API (filtered to renderable URLs only). */
export const getAnalysisSourceUrls = (analysis: AnalysisWithImageUrls): string[] => {
	const fromPresigned = getAllImageDataUrls(analysis.files_source_urls)
	if (fromPresigned.length > 0) {
		return fromPresigned
	}
	return getAllImageDataUrls(analysis.files_source)
}

/** Get output image URLs for display. Prefers presigned URLs from API (filtered to renderable URLs only). */
export const getAnalysisOutputUrls = (analysis: AnalysisWithImageUrls): string[] => {
	const fromPresigned = getAllImageDataUrls(analysis.files_output_urls)
	if (fromPresigned.length > 0) {
		return fromPresigned
	}
	return getAllImageDataUrls(analysis.files_output)
}

/** Object with optional image_url (presigned) or file (direct URL). */
interface ObjectWithImageUrl {
	image_url?: string
	file?: string
}

/** Get object image URL for display. Prefers image_url, otherwise direct renderable file URL. */
export const getObjectImageUrl = (obj: ObjectWithImageUrl): string | null => {
	if (obj.image_url) return obj.image_url
	return getImageDataUrl(obj.file)
}
