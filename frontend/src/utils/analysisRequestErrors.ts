export type ParsedAnalysisRequestError = {
	fileLabel?: string
	message: string
}

export function parseAnalysisErrorMessage(
	raw: string | undefined | null
): ParsedAnalysisRequestError[] {
	if (raw == null) return []
	const head = raw.trim()
	if (!head) return []

	const sep = ": "
	const out: ParsedAnalysisRequestError[] = []
	for (const rawLine of head.split("\n")) {
		const line = rawLine.trim()
		if (!line) continue
		const i = line.indexOf(sep)
		if (i > 0) {
			const fileLabel = line.slice(0, i).trim()
			const message = line.slice(i + sep.length).trim()
			if (fileLabel && message) {
				out.push({ fileLabel, message })
				continue
			}
		}
		out.push({ message: line })
	}
	return out
}
