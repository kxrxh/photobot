/** Successful refresh response shape returned by `client.post(...).json()` in tests. */
export const authRefreshSuccessEnvelope = {
	success: true as const,
	result: {
		access_token: "access",
		refresh_token: "refresh-new",
	},
} as const
