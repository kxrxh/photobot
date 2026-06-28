import { act, render, screen } from "@testing-library/react"
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest"
import { analysisRequestFixture } from "@/test/factories/analysisRequest"
import RequestsTab from "./RequestsTab"

describe("RequestsTab relative time", () => {
	beforeEach(() => {
		vi.useFakeTimers()
		vi.setSystemTime(new Date("2026-03-16T13:00:40.000Z"))
	})

	afterEach(() => {
		vi.clearAllTimers()
		vi.useRealTimers()
	})

	it("updates at minute boundary without manual refresh", () => {
		render(
			<RequestsTab
				statusFilter=""
				loadingRequests={false}
				requestsData={{ requests: [analysisRequestFixture()], total: 1 }}
				onStatusFilterChange={() => {}}
				onRefreshRequests={() => {}}
				onViewResults={() => {}}
				onDownloadPdf={async () => {}}
			/>
		)

		expect(screen.getByText("1 мин назад")).toBeInTheDocument()

		// RequestsTab resyncs `now` on a 30s interval; advance past one tick so relative minutes can update.
		act(() => {
			vi.advanceTimersByTime(30_000)
		})

		expect(screen.getByText("2 мин назад")).toBeInTheDocument()
	})
})
