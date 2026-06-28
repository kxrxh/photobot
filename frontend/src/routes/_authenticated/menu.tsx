import { createFileRoute } from "@tanstack/react-router"
import { FaBook, FaCalculator, FaCog, FaImage, FaListUl, FaUser } from "react-icons/fa"
import { FiSettings } from "react-icons/fi"
import { Button } from "@/components/common/ui/Button"
import { useAuth } from "@/contexts/AuthContext"

const MENU_ITEMS: {
	key: string
	to: string
	search?: Record<string, boolean | undefined>
	label: string
	icon: React.ElementType
	roles?: string[]
}[] = [
	{
		key: "analysis-create",
		to: "/analysis/create",
		label: "Создать анализ",
		icon: FaCalculator,
	},
	{
		key: "analysis-list",
		to: "/analysis/list",
		label: "Список анализов",
		icon: FaListUl,
	},
	{ key: "classification", to: "/classification", label: "Классификация", icon: FaCog },
	{ key: "markup", to: "/markup", label: "Разметка", icon: FaImage },
	{ key: "catalog", to: "/catalog", label: "Каталог", icon: FaBook },
	{ key: "profile", to: "/profile", label: "Профиль", icon: FaUser },
	{ key: "admin", to: "/admin", label: "Настройки", icon: FiSettings, roles: ["admin"] },
]

export const Route = createFileRoute("/_authenticated/menu")({
	component: RouteComponent,
})

function RouteComponent() {
	const navigate = Route.useNavigate()
	const { roles } = useAuth()

	const handleNavigation = (item: (typeof MENU_ITEMS)[number]) => () => {
		navigate({ to: item.to, search: item.search })
	}

	const filteredMenuItems = MENU_ITEMS.filter((item) => {
		if (item.roles && roles) {
			return item.roles.some((role) => roles.has(role))
		}
		return true
	})

	return (
		<div className="w-full max-w-md p-2 mx-auto">
			<div className="flex flex-col gap-4">
				{filteredMenuItems.map((item) => (
					<Button
						type="button"
						key={item.key}
						onClick={handleNavigation(item)}
						variant="primary"
						className="btn-lg"
						startIcon={<item.icon />}
					>
						{item.label}
					</Button>
				))}
			</div>
		</div>
	)
}
