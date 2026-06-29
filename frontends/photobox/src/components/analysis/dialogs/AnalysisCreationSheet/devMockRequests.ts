import type { AnalysisRequest } from "@/api/analysis/types"

/** True in `vite dev`; false in production and in Vitest (`MODE === "test"`). */
export function shouldUseDevMockRequestCards(): boolean {
	return import.meta.env.DEV && import.meta.env.MODE !== "test"
}

function isoMinutesAgo(mins: number): string {
	return new Date(Date.now() - mins * 60_000).toISOString()
}

const MOCK_REQUESTS: AnalysisRequest[] = [
	{
		id: "aaaaaaaa-bbbb-4ccc-dddd-000000000001",
		user_id: "0",
		product: "Пшеница",
		status: "created",
		created_at: isoMinutesAgo(2),
		updated_at: isoMinutesAgo(2),
	},
	{
		id: "aaaaaaaa-bbbb-4ccc-dddd-000000000002",
		user_id: "0",
		product: "Подсолнечник",
		status: "processing",
		created_at: isoMinutesAgo(45),
		updated_at: isoMinutesAgo(5),
	},
	{
		id: "aaaaaaaa-bbbb-4ccc-dddd-000000000003",
		user_id: "0",
		product: "Рапс",
		status: "waiting_for_confirmation",
		created_at: isoMinutesAgo(120),
		updated_at: isoMinutesAgo(15),
	},
	{
		id: "aaaaaaaa-bbbb-4ccc-dddd-000000000004",
		user_id: "0",
		product: "Ячмень",
		status: "completed",
		temp_id: "mock-analysis-complete",
		created_at: isoMinutesAgo(2880),
		updated_at: isoMinutesAgo(60),
	},
	{
		id: "aaaaaaaa-bbbb-4ccc-dddd-000000000005",
		user_id: "0",
		product: "Кукуруза",
		status: "failed",
		error_message: "Пример текста ошибки для отладки UI (mock).",
		created_at: isoMinutesAgo(200),
		updated_at: isoMinutesAgo(200),
	},
]

export function getDevMockRequestsPayload(statusFilter: string): {
	requests: AnalysisRequest[]
	total: number
} {
	const requests = statusFilter
		? MOCK_REQUESTS.filter((r) => r.status === statusFilter)
		: MOCK_REQUESTS
	return { requests, total: requests.length }
}
