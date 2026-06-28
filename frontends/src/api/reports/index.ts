import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"

const REPORTS_DOWNLOAD_TIMEOUT_MS = 30 * 60 * 1000

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.reports,
	timeout: REPORTS_DOWNLOAD_TIMEOUT_MS,
})

export type ReportDownloadURLResponse = {
	url: string
	expiresInSeconds: number
}

export const fetchReportDownloadURL = async (
	analysisId: string | number,
	format: "pdf" | "csv" = "pdf",
	signal?: AbortSignal
): Promise<ReportDownloadURLResponse> => {
	const base = API_ENDPOINTS.reports.trim()
	if (!base) {
		throw new Error("VITE_REPORTS_API_URL is not configured")
	}

	const response = await client.get(`${analysisId}/download-url`, {
		searchParams: { format },
		signal,
	})

	if (!response.ok) {
		let message = response.statusText
		try {
			const body = (await response.json()) as { error?: string }
			if (body.error) message = body.error
		} catch {
			/* ignore */
		}
		throw new Error(message || `Reports API error ${response.status}`)
	}

	return response.json<ReportDownloadURLResponse>()
}
