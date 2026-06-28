import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"

export const catalogClient = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.catalog,
})
