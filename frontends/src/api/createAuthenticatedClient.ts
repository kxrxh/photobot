import { authHooks } from "./auth/hooks"
import { baseClient, DEFAULT_RETRY } from "./client"

/** Ky instance with Bearer + 401 refresh (use for all protected service bases). */
export function createAuthenticatedClient(config: {
	prefixUrl: string
	timeout?: number
	retry?: number
}) {
	return baseClient.extend({
		prefixUrl: config.prefixUrl,
		timeout: config.timeout ?? 10000,
		retry: config.retry ?? DEFAULT_RETRY,
		hooks: authHooks,
	})
}
