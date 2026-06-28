import { sendReportToChat } from "@/api/analysis"
import { fetchReportDownloadURL } from "@/api/reports"
import { log } from "@/utils/log"

const SEND_TO_CHAT_WAIT_MS = 1500

export const downloadAnalysisReport = async (
	analysisId: string | number,
	requestFileDownload: (url: string, filename: string) => void,
	options?: { platform?: "telegram" | "max" }
): Promise<{ sendToChatOk: boolean }> => {
	if (!analysisId) {
		log.error("Error downloading PDF report: analysisId is missing")
		throw new Error("analysisId is required")
	}

	const { url } = await fetchReportDownloadURL(analysisId, "pdf")

	const platform = options?.platform
	let sendToChatOk = false

	if (platform === "telegram" || platform === "max") {
		const sendPromise = sendReportToChat(analysisId, platform).then(
			() => true,
			(err) => {
				log.devError("Failed to send report to chat:", err)
				return false
			}
		)
		requestFileDownload(url, `Отчёт_по_анализу_${analysisId}.pdf`)
		sendToChatOk = await Promise.race<boolean>([
			sendPromise,
			new Promise((resolve) =>
				setTimeout(() => {
					resolve(false)
				}, SEND_TO_CHAT_WAIT_MS)
			),
		])
	} else {
		requestFileDownload(url, `Отчёт_по_анализу_${analysisId}.pdf`)
	}

	return { sendToChatOk }
}
