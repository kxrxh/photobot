import path from "node:path";
import tailwindcss from "@tailwindcss/vite";
import { tanstackRouter } from "@tanstack/router-plugin/vite";
import react from "@vitejs/plugin-react-swc";
import { defineConfig, loadEnv } from "vite";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
	const env = loadEnv(mode, process.cwd(), "");
	const rawBase = env.VITE_BASE_PATH || "/";
	const base = rawBase.endsWith("/") ? rawBase : `${rawBase}/`;
	return {
		base,
		plugins: [
			tanstackRouter({ target: "react", autoCodeSplitting: true }),
			react(),
			tailwindcss(),
		],
		build: {
			minify: "esbuild",
			rollupOptions: {
				output: {
					manualChunks(id) {
						if (!id.includes("node_modules")) return
						if (id.includes("recharts")) return "vendor-recharts"
						if (id.includes("leaflet") || id.includes("react-leaflet")) {
							return "vendor-leaflet"
						}
						if (id.includes("framer-motion")) return "vendor-framer-motion"
					},
				},
			},
		},
		server: {
			host: env.VITE_HOST || "0.0.0.0",
		},
		resolve: {
			alias: {
				"@": path.resolve(__dirname, "./src"),
				"@/components": path.resolve(__dirname, "./src/components"),
				"@/api": path.resolve(__dirname, "./src/api"),
				"@/hooks": path.resolve(__dirname, "./src/hooks"),
				"@/utils": path.resolve(__dirname, "./src/utils"),
				"@/types": path.resolve(__dirname, "./src/types"),
				"@/constants": path.resolve(__dirname, "./src/constants"),
				"@/contexts": path.resolve(__dirname, "./src/contexts"),
				"@/routes": path.resolve(__dirname, "./src/routes"),
				"@/assets": path.resolve(__dirname, "./src/assets"),
			},
		},
	};
});
