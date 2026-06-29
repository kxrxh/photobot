import { createBrowserHistory, createRouter, RouterProvider } from "@tanstack/react-router"
import { useEffect } from "react"
import ReactDOM from "react-dom/client"

import * as TanStackQueryProvider from "./integrations/tanstack-query/root-provider.tsx"

// Import the generated route tree
import { routeTree } from "./routeTree.gen"

import "./styles.css"
import { AuthProvider, useAuth } from "./features/auth"
import reportWebVitals from "./reportWebVitals.ts"

// Create a new router instance
const router = createRouter({
	routeTree,
	context: {
		...TanStackQueryProvider.getContext(),
		auth: null,
	},
	defaultPreload: "intent",
	scrollRestoration: true,
	defaultStructuralSharing: true,
	basepath: "/admin-panel/",
	history: createBrowserHistory(),
})

// Register the router instance for type safety
declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router
	}
}

function App() {
	const auth = useAuth()

	// Invalidate the router when the auth state changes
	useEffect(() => {
		router.invalidate()
	}, [])

	if (auth.isValidating) {
		return (
			<div className="min-h-screen flex items-center justify-center bg-linear-to-br from-base-100 via-primary/5 to-secondary/5">
				<span className="loading loading-spinner loading-lg text-primary" />
			</div>
		)
	}

	return (
		<RouterProvider
			router={router}
			context={{
				...TanStackQueryProvider.getContext(),
				auth,
			}}
		/>
	)
}

// Render the app
const rootElement = document.getElementById("app")
if (rootElement && !rootElement.innerHTML) {
	const root = ReactDOM.createRoot(rootElement)
	root.render(
		<TanStackQueryProvider.Provider>
			<AuthProvider>
				<App />
			</AuthProvider>
		</TanStackQueryProvider.Provider>
	)
}

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals()
