import { createFileRoute, Outlet, redirect, useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { MobileMenu, NavBar } from "@/components/layout"
import { ThemeSwitch } from "@/components/ui"
import { useAuth } from "@/features/auth"

export const Route = createFileRoute("/_protected")({
	beforeLoad: ({ context, location }) => {
		if (!context.auth || !context.auth.isAuthenticated) {
			throw redirect({
				to: "/login",
				search: {
					redirect: location.href,
				},
			})
		}
	},
	component: ProtectedLayout,
})

function ProtectedLayout() {
	const { logout } = useAuth()
	const navigate = useNavigate()
	const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

	const handleLogout = () => {
		logout()
		navigate({ to: "/login" })
	}

	const closeMobileMenu = () => setMobileMenuOpen(false)

	return (
		<div className="min-h-screen bg-linear-to-br from-base-100 to-base-200">
			<div className="navbar bg-base-100/80 backdrop-blur-md border-b border-base-300 shadow-lg sticky top-0 z-50">
				<div className="navbar-start">
					<div className="flex items-center gap-3">
						<ThemeSwitch />
						<span className="text-xl font-bold bg-linear-to-r from-primary to-secondary bg-clip-text text-transparent">
							Панель управления
						</span>
					</div>
				</div>

				<NavBar onLogout={handleLogout} />

				<MobileMenu
					open={mobileMenuOpen}
					onToggle={() => setMobileMenuOpen(!mobileMenuOpen)}
					onClose={closeMobileMenu}
					onLogout={handleLogout}
				/>
			</div>

			<div className="container mx-auto p-6">
				<Outlet />
			</div>
		</div>
	)
}
