import ky from "ky"

export const BOT_NAME = "photobot"

// Only retry safe, read-only requests. Retrying POST/PUT/PATCH after a successful server
// commit (e.g. client timeout) can duplicate mutations (e.g. proposal request-changes).
const DEFAULT_RETRY = {
	limit: 2,
	methods: ["get"] as "get"[],
	statusCodes: [408, 413, 429, 500, 502, 503, 504],
}

export const baseClient = ky.create({
	timeout: 10000,
	retry: DEFAULT_RETRY,
	throwHttpErrors: false,
})

export { DEFAULT_RETRY }
