import { Link, useLocation } from "@tanstack/react-router"
import type { ReactNode } from "react"

interface AuthPageLayoutProps {
	title: string
	subtitle?: string
	children: ReactNode
	footer?: ReactNode
}

export function AuthPageLayout({ title, subtitle, children, footer }: AuthPageLayoutProps) {
	const { pathname } = useLocation()
	const showLoginLink = pathname !== "/login"

	return (
		<div className="flex min-h-screen flex-col items-center justify-center gap-6 p-4">
			<div className="w-full max-w-sm flex flex-col gap-4">
				<div className="text-center">
					<h1 className="text-2xl font-semibold">{title}</h1>
					{subtitle ? <p className="mt-1 text-sm text-base-content/70">{subtitle}</p> : null}
				</div>
				{children}
				{footer}
			</div>
			{showLoginLink ? (
				<Link to="/login" className="text-sm link link-primary">
					На страницу входа
				</Link>
			) : null}
		</div>
	)
}
