import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { createRouter, RouterProvider } from "@tanstack/react-router"
import ReactDOM from "react-dom/client"
import { routeTree } from "./routeTree.gen"
import "./index.css"
import { ThemeManager } from "@/components/common/layout/ThemeManager"
import { BASE_PATH } from "@/constants"
import { AuthProvider } from "@/contexts/AuthContext"

const CHUNK_RELOAD_KEY = "photobox:chunk-reload"

/**
 * After a deploy, cached index.html may reference removed chunks. Client-side navigation
 * then fetches index.html for missing /assets/*.js (SPA fallback) → MIME type error.
 * Reload once to pick up the new asset manifest.
 */
function installChunkLoadRecovery(): void {
	window.addEventListener("vite:preloadError", (event) => {
		event.preventDefault()
		if (sessionStorage.getItem(CHUNK_RELOAD_KEY)) {
			sessionStorage.removeItem(CHUNK_RELOAD_KEY)
			return
		}
		sessionStorage.setItem(CHUNK_RELOAD_KEY, "1")
		window.location.reload()
	})
}

export const queryClient = new QueryClient({
	defaultOptions: {
		queries: {
			retry: false,
			staleTime: 60_000,
			gcTime: 10 * 60_000,
			refetchOnWindowFocus: false,
		},
	},
})

const router = createRouter({
	routeTree,
	context: { queryClient },
	basepath: BASE_PATH,
	defaultPreload: "intent",
})

declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router
	}
}

installChunkLoadRecovery()

const rootElement = document.getElementById("root")
if (rootElement && !rootElement.innerHTML) {
	sessionStorage.removeItem(CHUNK_RELOAD_KEY)
	const root = ReactDOM.createRoot(rootElement)
	root.render(
		<ThemeManager>
			<QueryClientProvider client={queryClient}>
				<AuthProvider>
					<RouterProvider router={router} />
				</AuthProvider>
			</QueryClientProvider>
		</ThemeManager>
	)
}
