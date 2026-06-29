import type { IconType } from "react-icons"
import { FaHome, FaRobot, FaServer, FaSignOutAlt, FaUsers, FaUserTag } from "react-icons/fa"

export interface NavItem {
	to: string
	label: string
	icon: IconType
	linkClass: string
}

export const NAV_ITEMS: NavItem[] = [
	{
		to: "/",
		label: "Главная",
		icon: FaHome,
		linkClass: "hover:bg-primary/10 hover:text-primary",
	},
	{
		to: "/users",
		label: "Пользователи",
		icon: FaUsers,
		linkClass: "hover:bg-secondary/10 hover:text-secondary",
	},
	{
		to: "/roles",
		label: "Роли",
		icon: FaUserTag,
		linkClass: "hover:bg-secondary/10 hover:text-secondary",
	},
	{
		to: "/bots",
		label: "Боты",
		icon: FaRobot,
		linkClass: "hover:bg-accent/10 hover:text-accent",
	},
	{
		to: "/services",
		label: "Сервисы",
		icon: FaServer,
		linkClass: "hover:bg-info/10 hover:text-info",
	},
]

export const LOGOUT_ICON = FaSignOutAlt
export const LOGOUT_LINK_CLASS = "hover:bg-error/10 hover:text-error"
