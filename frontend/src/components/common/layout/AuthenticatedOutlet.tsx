import { Outlet, useLocation, useNavigate } from "@tanstack/react-router"
import { useEffect, useRef, useState } from "react"
import { MessengerReloadPrompt } from "@/components/auth/MessengerReloadPrompt"
import Loading from "@/components/common/ui/Loading"
import { useAuth } from "@/contexts/AuthContext"
import { useMessenger } from "@/hooks/useMessenger"
import { isWebAuthMode } from "@/lib/auth/mode"
import { BackButtonVisibilityContext } from "../context/BackButtonVisibilityContext"
import { BackButton } from "./BackButton"
import ErrorPage from "./ErrorPage"

/**
 * Routes where the back button should not be displayed
 */
const EXCLUDED_BACK_BUTTON_PATHS = ["/menu", "/profile", "/login"] as const

/**
 * Path prefix for catalog routes where back button is hidden
 */
const CATALOG_PATH_PREFIX = "/catalog/" as const

/**
 * Authenticated outlet component that handles authentication state and back button visibility
 * Provides a layout with conditional back button footer for authenticated users
 */
export function AuthenticatedOutlet() {
	const location = useLocation()
	const currentPath = location.pathname
	const [isBackButtonHidden, setIsBackButtonHidden] = useState(false)
	const { state, error } = useAuth()
	const navigate = useNavigate()
	const { startParam } = useMessenger()
	const hasHandledStartParamRef = useRef(false)

	// After successful authentication, if Telegram start param exists, deep-link to analysis
	useEffect(() => {
		if (state === "authenticated" && startParam && !hasHandledStartParamRef.current) {
			hasHandledStartParamRef.current = true
			navigate({
				to: "/analysis/create",
				search: { openRequest: startParam },
				replace: true,
			} as const)
		}
	}, [state, startParam, navigate])

	const webMode = isWebAuthMode()
	const shouldRedirectToLogin =
		webMode && (state === "user_not_found" || state === "unauthenticated")

	useEffect(() => {
		if (!shouldRedirectToLogin) return
		void navigate({ to: "/login", replace: true })
	}, [shouldRedirectToLogin, navigate])

	// Handle authentication states
	switch (state) {
		case "loading":
			return (
				<div className="flex items-center justify-center h-screen">
					<Loading />
				</div>
			)

		case "user_not_found":
		case "unauthenticated":
			if (!webMode) {
				return <MessengerReloadPrompt />
			}
			return (
				<div className="flex items-center justify-center h-screen">
					<Loading />
				</div>
			)

		case "invalid_data":
			return <ErrorPage error={new Error(error || "Ошибка валидации данных")} />

		case "authenticated":
			// Continue to render authenticated content
			break

		default:
			return (
				<div className="flex items-center justify-center h-screen">
					<Loading />
				</div>
			)
	}

	const shouldShowBackButton = shouldShowBackButtonForPath(currentPath)
	const showBottomBackFooter = shouldShowBackButton && !isBackButtonHidden

	// Reserve space only when the fixed footer is actually shown
	const mainPaddingClass = showBottomBackFooter ? "pb-16" : "pb-0"

	return (
		<div className="flex h-dvh max-h-dvh min-h-0 flex-col overflow-hidden">
			<main
				className={`flex min-h-0 flex-1 flex-col overflow-y-auto overflow-x-hidden ${mainPaddingClass}`}
			>
				<BackButtonVisibilityContext.Provider value={{ isBackButtonHidden, setIsBackButtonHidden }}>
					<Outlet />
				</BackButtonVisibilityContext.Provider>
			</main>
			{showBottomBackFooter && (
				<footer className="fixed right-0 bottom-0 left-0 z-50 p-2 bg-base-100">
					<BackButton />
				</footer>
			)}
		</div>
	)
}

function shouldShowBackButtonForPath(currentPath: string): boolean {
	const isExcludedPath = EXCLUDED_BACK_BUTTON_PATHS.some((excludedPath) =>
		currentPath.startsWith(excludedPath)
	)

	const isCatalogPath = currentPath.startsWith(CATALOG_PATH_PREFIX)

	// Create-analysis uses in-screen app bar + bottom action dock (no double footer).
	if (currentPath.startsWith("/analysis/create")) {
		return false
	}

	return !isExcludedPath && !isCatalogPath
}
