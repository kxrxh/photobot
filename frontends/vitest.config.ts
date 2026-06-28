import { mergeConfig } from "vite";
import viteConfig from "./vite.config";

// https://vitest.dev/config/
export default mergeConfig(viteConfig({ command: "serve", mode: "test" }), {
	test: {
		environment: "jsdom",
		include: ["src/**/*.{test,spec}.{ts,tsx}"],
		setupFiles: ["./src/test/setup.ts"],
		coverage: {
			provider: "v8",
		},
	},
});
