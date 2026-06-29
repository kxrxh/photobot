import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import tanstackRouter from "@tanstack/router-plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { loadEnv } from "vite";
import { defineConfig } from "vite";

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
	const env = loadEnv(mode, process.cwd(), "");
	return {
		base: "/admin-panel/",
		plugins: [
			tanstackRouter({ target: "react", autoCodeSplitting: true }),
			viteReact(),
			tailwindcss(),
		],
		resolve: {
			alias: {
				"@": fileURLToPath(new URL("./src", import.meta.url)),
			},
		},
		esbuild: {
			drop: command === "build" ? ["console", "debugger"] : [],
		},
		build: {
			minify: "esbuild",
		},
		server: {
			host: env.VITE_HOST || "0.0.0.0",
		},
	};
});
