import type React from "react"
import { FaChevronDown, FaChevronUp, FaEdit } from "react-icons/fa"
import type { Markup } from "@/api/markup/types"

const MarkupCard: React.FC<{
	markup: Markup
	isSelected: boolean
	isExpanded: boolean
	isEditing: boolean
	editName: string
	onToggleExpansion: () => void
	onStartEdit: () => void
	onCancelEdit: () => void
	onNameChange: (name: string) => void
	onReplace: () => void
}> = ({
	markup,
	isSelected,
	isExpanded,
	isEditing,
	editName,
	onToggleExpansion,
	onStartEdit,
	onCancelEdit,
	onNameChange,
	onReplace,
}) => (
	<div
		className={`card card-border transition-colors ${
			isSelected ? "ring-2 ring-primary bg-primary/5" : "hover:border-primary/50"
		}`}
	>
		<div className="p-4">
			<button
				type="button"
				className="flex flex-col gap-3 p-0 w-full text-left bg-transparent border-none cursor-pointer"
				onClick={onToggleExpansion}
				aria-expanded={isExpanded}
			>
				<div className="flex gap-4 justify-between items-center w-full">
					<div className="flex overflow-hidden grow gap-2 items-center">
						<h3 className="text-base font-semibold truncate">{markup.name}</h3>
					</div>

					<div className="flex gap-2 items-center shrink-0">
						{isExpanded ? (
							<FaChevronUp className="text-sm" />
						) : (
							<FaChevronDown className="text-sm" />
						)}
					</div>
				</div>

				<div className="flex flex-wrap gap-y-1 gap-x-4 items-center text-sm text-base-content/70">
					<div className="flex gap-1 items-center">
						Обновлено:
						<span className="badge badge-outline badge-sm">
							{new Date(markup.updated_at).toLocaleDateString("ru-RU", {
								day: "2-digit",
								month: "2-digit",
								year: "numeric",
								hour: "2-digit",
								minute: "2-digit",
							})}
						</span>
					</div>
				</div>
			</button>

			{isExpanded && (
				<div className="overflow-hidden mt-4 space-y-4">
					{isEditing ? (
						<div className="space-y-3">
							<h4 className="text-sm font-medium text-base-content/80">Новое название:</h4>
							<div className="px-1">
								<input
									type="text"
									value={editName}
									onChange={(e) => onNameChange(e.target.value)}
									className="input input-bordered input-sm"
									placeholder="Введите новое название..."
									ref={(input) => input?.focus()}
								/>
							</div>
							<div className="flex gap-2">
								<button
									type="button"
									onClick={(e) => {
										e.stopPropagation()
										onReplace()
									}}
									className="flex-1 btn btn-primary btn-sm"
									disabled={!editName.trim()}
								>
									Перезаписать
								</button>
								<button
									type="button"
									onClick={(e) => {
										e.stopPropagation()
										onCancelEdit()
									}}
									className="flex-1 btn btn-sm"
								>
									Отмена
								</button>
							</div>
						</div>
					) : (
						<div className="pt-2">
							<button
								type="button"
								onClick={(e) => {
									e.stopPropagation()
									onStartEdit()
								}}
								className="flex gap-2 justify-center items-center w-full btn btn-primary btn-sm"
							>
								<FaEdit className="w-3 h-3" />
								Перезаписать
							</button>
						</div>
					)}
				</div>
			)}
		</div>
	</div>
)

export default MarkupCard
