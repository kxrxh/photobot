import type React from "react"
import { FaPlus } from "react-icons/fa"

interface SearchAndCreateBarProps {
	searchTerm: string
	onSearchChange: (value: string) => void
	onCreateNew: () => void
	searchPlaceholder?: string
	createLabel?: string
}

const SearchAndCreateBar: React.FC<SearchAndCreateBarProps> = ({
	searchTerm,
	onSearchChange,
	onCreateNew,
	searchPlaceholder = "Поиск...",
	createLabel = "Создать",
}) => (
	<div className="p-2">
		<div className="flex gap-3 items-center">
			<input
				type="text"
				placeholder={searchPlaceholder}
				className="flex-1 input input-bordered input-sm"
				value={searchTerm}
				onChange={(e) => onSearchChange(e.target.value)}
			/>
			<button
				type="button"
				className="flex gap-2 items-center btn btn-primary btn-sm shrink-0"
				onClick={onCreateNew}
			>
				<FaPlus className="w-3 h-3" />
				{createLabel}
			</button>
		</div>
	</div>
)

export default SearchAndCreateBar
