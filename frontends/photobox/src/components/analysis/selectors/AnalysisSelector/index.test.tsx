import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react"
import type { ComponentProps } from "react"
import { useEffect, useState } from "react"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import type { Analysis } from "@/api/analysis/types"
import {
	analysisListPage,
	analysisListRow,
	analysisWithObjects,
} from "@/test/factories/analysisSelector"
import { getAnalysisSourceUrls } from "@/utils/image"
import AnalysisSelectorDialog from "./index"

const mockFetchAnalysisById = vi.fn()

const defaultInfiniteQueryResult = {
	data: {
		pages: [
			analysisListPage([
				analysisListRow({
					id: "analysis-1",
					files_source: ["analysis/analysis-1/source_0.jpg"],
					files_output: [],
				}),
			]),
		],
	},
	isLoading: false,
	isFetching: false,
	error: null,
	fetchNextPage: vi.fn(),
	hasNextPage: false,
	isFetchingNextPage: false,
	refetch: vi.fn(),
}

const mockUseInfiniteQuery = vi.fn(() => defaultInfiniteQueryResult)

vi.mock("@tanstack/react-query", async (importOriginal) => {
	const actual = await importOriginal<typeof import("@tanstack/react-query")>()
	return {
		...actual,
		useInfiniteQuery: () => mockUseInfiniteQuery(),
		useQueryClient: () => ({
			fetchQuery: async ({ queryFn }: { queryFn: () => Promise<unknown> }) => queryFn(),
		}),
		useQueries: () => [],
	}
})

vi.mock("./useAnalysisSourceHydration", () => ({
	useAnalysisSourceHydration: (analyses: Analysis[], isOpen: boolean) => {
		const [map, setMap] = useState<Record<string, string[]>>({})
		useEffect(() => {
			if (!isOpen) {
				setMap({})
				return
			}
			let cancelled = false
			void (async () => {
				const next: Record<string, string[]> = {}
				for (const analysis of analyses) {
					const existing = getAnalysisSourceUrls(analysis)
					if (existing.length > 0) {
						next[analysis.id] = existing
						continue
					}
					try {
						const detail = await mockFetchAnalysisById(analysis.id)
						const urls = getAnalysisSourceUrls(detail)
						if (urls.length > 0) next[analysis.id] = urls
					} catch {
						// ignore
					}
				}
				if (!cancelled) setMap(next)
			})()
			return () => {
				cancelled = true
			}
		}, [analyses, isOpen])
		return map
	},
}))

vi.mock("react-intersection-observer", () => ({
	useInView: vi.fn(() => ({ ref: vi.fn(), inView: false })),
}))

vi.mock("@/api/analysis", () => ({
	fetchAnalyses: vi.fn(),
	fetchAnalysisById: (...args: unknown[]) => mockFetchAnalysisById(...args),
}))

vi.mock("@/contexts/AuthContext", () => ({
	useAuth: () => ({ userId: 1 }),
}))

vi.mock("@/hooks/useAlert", () => ({
	useAlert: () => ({
		showError: vi.fn(),
		showSuccess: vi.fn(),
	}),
}))

vi.mock("@/hooks/useMessenger", () => ({
	useMessenger: () => ({
		requestFileDownload: vi.fn(),
		hapticFeedback: { impactOccurred: vi.fn() },
		platform: "web",
	}),
}))

vi.mock("@/components/analysis/dialogs/AnalysisViewSheet", () => ({
	default: () => null,
}))

vi.mock("./AnalysisFilterSheet", () => ({
	default: () => null,
}))

vi.mock("@/components/common/ui/Loading", () => ({
	default: () => null,
}))

beforeEach(() => {
	cleanup()
	mockFetchAnalysisById.mockReset()
	mockUseInfiniteQuery.mockReset()
	mockUseInfiniteQuery.mockImplementation(() => defaultInfiniteQueryResult)
})

afterEach(() => {
	cleanup()
})

const dialogProps: ComponentProps<typeof AnalysisSelectorDialog> = {
	isOpen: true,
	onClose: vi.fn(),
	selectedAnalysisIds: [],
	onAddAnalysis: vi.fn(),
	onRemoveAnalysis: vi.fn(),
	onRemoveAllAnalyses: vi.fn(),
	hasAddedAnalyses: false,
}

const renderDialog = () => render(<AnalysisSelectorDialog {...dialogProps} />)

/** Preview img lives in a card without a dedicated test id; scope from the analysis heading. */
function previewSrcForAnalysisId(id: string): string | null {
	const heading = screen.getByRole("heading", { name: `Анализ №${id}` })
	const card = heading.closest("div.group")
	const img = card?.querySelector('img[alt="Превью анализа"]')
	return img?.getAttribute("src") ?? null
}

describe("AnalysisSelectorDialog image hydration", () => {
	it("hydrates preview image URL when list item has non-renderable storage key", async () => {
		mockFetchAnalysisById.mockResolvedValue(
			analysisWithObjects({
				id: "analysis-1",
				files_source: ["analysis/analysis-1/source_0.jpg"],
				files_output: [],
				files_source_urls: ["https://cdn.example.com/analysis-1-source-0.jpg"],
			})
		)

		renderDialog()

		expect(screen.queryByAltText("Превью анализа")).not.toBeInTheDocument()

		await waitFor(() => {
			expect(screen.getByAltText("Превью анализа")).toHaveAttribute(
				"src",
				"https://cdn.example.com/analysis-1-source-0.jpg"
			)
		})

		expect(mockFetchAnalysisById).toHaveBeenCalledWith("analysis-1")
		expect(mockFetchAnalysisById).toHaveBeenCalledTimes(1)
	})

	it("hydrates when files_source_urls has no renderable URLs after filtering", async () => {
		mockUseInfiniteQuery.mockImplementation(() => ({
			...defaultInfiniteQueryResult,
			data: {
				pages: [
					analysisListPage([
						analysisListRow({
							id: "analysis-2",
							files_source: ["analysis/analysis-2/source_0.jpg"],
							files_output: [],
							files_source_urls: ["analysis/analysis-2/bad-presigned-key.jpg"],
						}),
					]),
				],
			},
		}))

		mockFetchAnalysisById.mockResolvedValue(
			analysisWithObjects({
				id: "analysis-2",
				files_source: ["analysis/analysis-2/source_0.jpg"],
				files_output: [],
				files_source_urls: ["https://cdn.example.com/analysis-2-hydrated.jpg"],
			})
		)

		renderDialog()

		expect(screen.queryByAltText("Превью анализа")).not.toBeInTheDocument()

		await waitFor(() => {
			expect(screen.getByAltText("Превью анализа")).toHaveAttribute(
				"src",
				"https://cdn.example.com/analysis-2-hydrated.jpg"
			)
		})

		expect(mockFetchAnalysisById).toHaveBeenCalledWith("analysis-2")
	})

	it("hydrates every card when the first fetch completes and re-renders before a slower fetch finishes", async () => {
		mockUseInfiniteQuery.mockImplementation(() => ({
			...defaultInfiniteQueryResult,
			data: {
				pages: [
					analysisListPage(
						[
							analysisListRow({
								id: "id-fast",
								files_source: ["analysis/id-fast/source_0.jpg"],
								files_output: [],
							}),
							analysisListRow({
								id: "id-slow",
								files_source: ["analysis/id-slow/source_0.jpg"],
								files_output: [],
							}),
						],
						{ total: 2 }
					),
				],
			},
		}))

		mockFetchAnalysisById.mockImplementation(async (id: string) => {
			if (id === "id-fast") {
				await Promise.resolve()
				return analysisWithObjects({
					id: "id-fast",
					files_source: ["analysis/id-fast/source_0.jpg"],
					files_output: [],
					files_source_urls: ["https://cdn.example.com/fast.jpg"],
				})
			}
			await new Promise((r) => setTimeout(r, 30))
			return analysisWithObjects({
				id: "id-slow",
				files_source: ["analysis/id-slow/source_0.jpg"],
				files_output: [],
				files_source_urls: ["https://cdn.example.com/slow.jpg"],
			})
		})

		renderDialog()

		await waitFor(() => {
			expect(previewSrcForAnalysisId("id-fast")).toBe("https://cdn.example.com/fast.jpg")
		})

		await waitFor(
			() => {
				expect(previewSrcForAnalysisId("id-slow")).toBe("https://cdn.example.com/slow.jpg")
			},
			{ timeout: 2000 }
		)

		expect(mockFetchAnalysisById).toHaveBeenCalledWith("id-fast")
		expect(mockFetchAnalysisById).toHaveBeenCalledWith("id-slow")
	})
})

describe("AnalysisSelectorDialog preview fallback", () => {
	it("advances to next source URL when first image fails to load", async () => {
		mockUseInfiniteQuery.mockImplementation(() => ({
			...defaultInfiniteQueryResult,
			data: {
				pages: [
					analysisListPage([
						analysisListRow({
							id: "analysis-3",
							files_source: [],
							files_output: [],
							files_source_urls: [
								"https://bad.example.com/first.jpg",
								"https://good.example.com/second.jpg",
							],
						}),
					]),
				],
			},
		}))

		renderDialog()

		const img = await waitFor(() => screen.getByAltText("Превью анализа"))
		expect(img).toHaveAttribute("src", "https://bad.example.com/first.jpg")

		fireEvent.error(img)

		await waitFor(() => {
			expect(img).toHaveAttribute("src", "https://good.example.com/second.jpg")
		})
	})
})

describe("AnalysisSelectorDialog list keys", () => {
	it("keeps each analysis preview URL correct when list order changes", async () => {
		const itemA = analysisListRow({
			id: "id-a",
			files_source: [],
			files_output: [],
			files_source_urls: ["https://cdn.example.com/a-only.jpg"],
		})
		const itemB = analysisListRow({
			id: "id-b",
			files_source: [],
			files_output: [],
			files_source_urls: ["https://cdn.example.com/b-only.jpg"],
		})

		const listOrder = [itemA, itemB]
		mockUseInfiniteQuery.mockImplementation(() => ({
			...defaultInfiniteQueryResult,
			data: {
				pages: [analysisListPage(listOrder, { total: 2 })],
			},
		}))

		const view = render(<AnalysisSelectorDialog {...dialogProps} />)

		await waitFor(() => {
			expect(screen.getAllByAltText("Превью анализа")).toHaveLength(2)
		})
		expect(previewSrcForAnalysisId("id-a")).toBe("https://cdn.example.com/a-only.jpg")
		expect(previewSrcForAnalysisId("id-b")).toBe("https://cdn.example.com/b-only.jpg")

		listOrder.length = 0
		listOrder.push(itemB, itemA)
		view.rerender(<AnalysisSelectorDialog {...dialogProps} />)

		await waitFor(() => {
			expect(previewSrcForAnalysisId("id-a")).toBe("https://cdn.example.com/a-only.jpg")
			expect(previewSrcForAnalysisId("id-b")).toBe("https://cdn.example.com/b-only.jpg")
		})
	})
})
