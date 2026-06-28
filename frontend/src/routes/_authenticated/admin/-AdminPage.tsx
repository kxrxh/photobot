import ParamsTab from "@/components/admin/tabs/ParamsTab"
import ProductsTab from "@/components/admin/tabs/ProductsTab"
import { Tabs } from "@/components/admin/tabs/Tabs"

const tabItems = [
	{
		label: "Продукты",
		content: <ProductsTab />,
	},
	{
		label: "Параметры",
		content: <ParamsTab />,
	},
]

export default function AdminPage() {
	return (
		<>
			<header className="flex sticky top-0 z-50 flex-col gap-2 px-2 py-2 w-full bg-base-100">
				<div className="flex justify-between items-center align-center">
					<h1 className="text-2xl font-bold">Панель администратора</h1>
				</div>
			</header>
			<div>
				<Tabs items={tabItems} />
			</div>
		</>
	)
}
