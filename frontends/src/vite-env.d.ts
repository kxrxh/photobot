/// <reference types="vite/client" />

interface ImportMetaEnv {
	readonly VITE_HOST: string
	readonly VITE_BASE_PATH?: string
	readonly VITE_AUTH_API_URL: string
	readonly VITE_WEED_API_URL: string
	readonly VITE_CLASSIFICATION_API_URL: string
	readonly VITE_ANALYSIS_API_URL: string
	readonly VITE_REPORTS_API_URL: string
}

/** Merges with vite/client; ties custom env keys to `import.meta.env`. */
interface ImportMeta {
	readonly env: ImportMetaEnv
}
