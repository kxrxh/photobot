import { useState } from "react"
import { FaInfoCircle, FaPlus, FaTrash } from "react-icons/fa"
import type { WeedListItem } from "@/api/catalog/types"
import coffeeBeanIcon from "@/assets/coffee-bean.svg"

interface CatalogSelectorCardProps {
	item: WeedListItem
	onSelect: (item: WeedListItem) => void
	onViewDetails?: (item: WeedListItem) => void
	isSelected?: boolean
}

const CatalogSelectorCard = ({
	item,
	onSelect,
	onViewDetails,
	isSelected,
}: CatalogSelectorCardProps) => {
	const [imageError, setImageError] = useState(false)
	const [isImageLoading, setIsImageLoading] = useState(!!item.primary_image_url)

	const handleImageError = () => {
		setImageError(true)
		setIsImageLoading(false)
	}

	const handleImageLoad = () => {
		setIsImageLoading(false)
	}

	const showFallback = !item.primary_image_url || imageError

	const handleViewDetails = () => {
		if (onViewDetails) {
			onViewDetails(item)
		}
	}

	return (
		<div
			className={`card card-border bg-base-100 border-base-content/10 shadow-xl transition-all duration-300 ${isSelected ? "ring-2 ring-primary bg-primary/5" : ""}`}
		>
			<figure className="relative aspect-4/3 overflow-hidden">
				<div className="absolute top-2 right-2 z-10">
					{onViewDetails && (
						<button
							type="button"
							className="p-2 text-gray-700 bg-white/90 rounded-full shadow-md cursor-pointer"
							onClick={handleViewDetails}
							aria-label={`View details for ${item.name}`}
							title="Подробнее"
						>
							<FaInfoCircle className="w-4 h-4" />
						</button>
					)}
				</div>

				<div className="flex justify-center items-center w-full h-full bg-neutral-content">
					{isImageLoading && (
						<div className="flex absolute inset-0 justify-center items-center animate-pulse bg-neutral-content">
							<div className="w-8 h-8 rounded-full border-2 animate-spin border-primary border-t-transparent" />
						</div>
					)}

					{showFallback ? (
						<img src={coffeeBeanIcon} alt="Нет фото" className="w-16 h-16 opacity-50" />
					) : (
						<img
							src={item.primary_image_url || ""}
							alt={item.name}
							className="object-cover w-full h-full"
							loading="lazy"
							decoding="async"
							onError={handleImageError}
							onLoad={handleImageLoad}
						/>
					)}
				</div>
			</figure>

			<div className="p-4 card-body">
				<h2 className="font-semibold text-md card-title text-primary line-clamp-2">{item.name}</h2>

				<div className="space-y-1 text-sm">
					<div className="flex gap-1 items-center">
						<span className="font-medium">Длина:</span>
						{item.length > 0 ? `${item.length} мм` : "-"}
					</div>
					<div className="flex gap-1 items-center">
						<span className="font-medium">Ширина:</span>
						{item.width > 0 ? `${item.width} мм` : "-"}
					</div>
					<div className="flex gap-1 items-center">
						<span className="font-medium">Д/Ш:</span>
						{item.length && item.width ? (item.length / item.width).toFixed(2) : "-"}
					</div>
				</div>
			</div>
			<div className="card-actions justify-center w-full p-2">
				<button
					type="button"
					className={`btn btn-sm w-full ${!isSelected ? "btn-primary" : ""}`}
					onClick={() => onSelect(item)}
				>
					{isSelected ? <FaTrash /> : <FaPlus />}
					{isSelected ? "Убрать" : "Добавить"}
				</button>
			</div>
		</div>
	)
}

export default CatalogSelectorCard
