import React from "react"

type TabItem = {
	label: string
	content: React.ReactNode
}

type TabsProps = {
	items: TabItem[]
}

export function Tabs({ items }: TabsProps) {
	const [activeTab, setActiveTab] = React.useState(0)

	const onTabClick = (index: number) => {
		setActiveTab(index)
		window.scrollTo({ top: 0, behavior: "smooth" })
	}

	return (
		<div>
			<div
				role="tablist"
				className="tabs tabs-border w-full justify-center sticky top-0 z-50 bg-base-100"
			>
				{items.map((item, index) => (
					<button
						type="button"
						key={item.label}
						className={`tab ${activeTab === index ? "tab-active" : ""} tab-bordered`}
						onClick={() => onTabClick(index)}
					>
						{item.label}
					</button>
				))}
			</div>
			<div>{items[activeTab].content}</div>
		</div>
	)
}
