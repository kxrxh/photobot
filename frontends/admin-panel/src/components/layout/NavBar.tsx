import { Link, useLocation } from "@tanstack/react-router"
import { LOGOUT_ICON, LOGOUT_LINK_CLASS, NAV_ITEMS } from "./nav-config"

function navLinkClass(base: string, active: boolean) {
	return active ? `${base} btn-active` : base
}

const DESKTOP_BASE = "btn btn-ghost btn-sm gap-2 transition-all duration-200 rounded-lg"

export function NavBar({ onLogout }: { onLogout: () => void }) {
	const { pathname } = useLocation()

	return (
		<div className="navbar-end hidden lg:flex">
			<ul className="menu menu-horizontal gap-1">
				{NAV_ITEMS.map((item) => {
					const isActive = item.to === "/" ? pathname === "/" : pathname.startsWith(item.to)
					return (
						<li key={item.to}>
							<Link
								to={item.to}
								className={navLinkClass(`${DESKTOP_BASE} ${item.linkClass}`, isActive)}
							>
								<item.icon className="text-lg" />
								<span className="hidden sm:inline">{item.label}</span>
							</Link>
						</li>
					)
				})}
				<li>
					<button
						type="button"
						onClick={onLogout}
						className={navLinkClass(`${DESKTOP_BASE} ${LOGOUT_LINK_CLASS}`, false)}
					>
						<LOGOUT_ICON className="text-lg" />
						<span className="hidden sm:inline">Выйти</span>
					</button>
				</li>
			</ul>
		</div>
	)
}
