import { Link, useLocation } from "@tanstack/react-router"
import { FaBars } from "react-icons/fa"
import { LOGOUT_ICON, LOGOUT_LINK_CLASS, NAV_ITEMS } from "./nav-config"

function navLinkClass(base: string, active: boolean) {
	return active ? `${base} btn-active` : base
}

const MOBILE_BASE = "gap-2 rounded-lg"

export function MobileMenu({
	open,
	onToggle,
	onClose,
	onLogout,
}: {
	open: boolean
	onToggle: () => void
	onClose: () => void
	onLogout: () => void
}) {
	const { pathname } = useLocation()

	const handleLogout = () => {
		onClose()
		onLogout()
	}

	return (
		<div className="navbar-end lg:hidden">
			<div className="dropdown dropdown-end">
				<button
					type="button"
					className="btn btn-ghost btn-square"
					onClick={onToggle}
					aria-label="Меню"
					aria-expanded={open}
					tabIndex={0}
				>
					<FaBars className="h-6 w-6" />
				</button>
				{open && (
					<>
						<button
							type="button"
							className="fixed inset-0 z-40 cursor-default"
							aria-label="Закрыть меню"
							onClick={onClose}
						/>
						<ul
							className="menu dropdown-content bg-base-100 rounded-box z-50 mt-2 w-56 gap-1 border border-base-300 p-2 shadow-xl"
							onClick={onClose}
							onKeyDown={(e) => {
								if (e.key === "Escape") onClose()
							}}
						>
							{NAV_ITEMS.map((item) => {
								const isActive = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to)
								return (
									<li key={item.to}>
										<Link
											to={item.to}
											className={navLinkClass(`${MOBILE_BASE} ${item.linkClass}`, isActive)}
										>
											<item.icon className="text-lg" />
											{item.label}
										</Link>
									</li>
								)
							})}
							<li>
								<button
									type="button"
									onClick={handleLogout}
									className={`${MOBILE_BASE} ${LOGOUT_LINK_CLASS}`}
								>
									<LOGOUT_ICON className="text-lg" />
									Выйти
								</button>
							</li>
						</ul>
					</>
				)}
			</div>
		</div>
	)
}
