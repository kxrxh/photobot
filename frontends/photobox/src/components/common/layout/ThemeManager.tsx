import { type ReactNode, useEffect } from "react"

export function ThemeManager({ children }: { children: ReactNode }) {
	useEffect(() => {
		document.documentElement.setAttribute("data-theme", "csort")
	}, [])

	return <>{children}</>
}
