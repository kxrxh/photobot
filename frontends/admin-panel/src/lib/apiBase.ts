import ky from "ky"

/**
 * Base API client without auth hooks. Used by auth endpoints (login, refresh)
 * to avoid circular dependency with the main api client.
 */
const apiBase = ky.create({
	prefixUrl: import.meta.env.VITE_AUTH_API_URL,
})

export default apiBase
