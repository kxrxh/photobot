/**
 * Central logging: verbose diagnostics in development only; production keeps
 * minimal `error` output for exceptional failures.
 */
export const log = {
	devWarn(...args: unknown[]): void {
		if (import.meta.env.DEV) {
			console.warn(...args)
		}
	},

	devError(...args: unknown[]): void {
		if (import.meta.env.DEV) {
			console.error(...args)
		}
	},

	error(...args: unknown[]): void {
		console.error(...args)
	},
}
