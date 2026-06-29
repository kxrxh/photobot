import { useState } from "react"
import { FaEdit, FaInfoCircle } from "react-icons/fa"
import type { WeedListItem } from "@/api/catalog/types"
import coffeeBeanIcon from "@/assets/coffee-bean.svg"

interface CatalogCardProps {
	item: WeedListItem
	onViewDetails: (item: WeedListItem) => void
	onEditItem?: (item: WeedListItem) => void
	showEditButton?: boolean
}

const CatalogCard = ({ item, onViewDetails, onEditItem, showEditButton }: CatalogCardProps) => {
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

	return (
		<div className="card card-border bg-base-100 border-base-content/10 shadow-xl transition-all duration-300">
			<figure className="relative aspect-4/3 overflow-hidden">
				<div className="absolute top-2 right-2 z-10 flex gap-2">
					{showEditButton && onEditItem && (
						<button
							type="button"
							className="p-2 text-gray-700 bg-white bg-opacity-70 rounded-full shadow-md transition-all hover:bg-opacity-100 cursor-pointer"
							onClick={() => onEditItem(item)}
							aria-label={`Edit ${item.name}`}
							title="Редактировать"
						>
							<FaEdit className="w-4 h-4" />
						</button>
					)}
					{onViewDetails && (
						<button
							type="button"
							className="p-2 text-gray-700 bg-white bg-opacity-70 rounded-full shadow-md transition-all hover:bg-opacity-100 cursor-pointer"
							onClick={() => onViewDetails(item)}
							aria-label={`View details for ${item.name}`}
							title="Подробнее"
						>
							<FaInfoCircle className="w-4 h-4" />
						</button>
					)}
				</div>

				<div className="w-full h-full bg-neutral-content flex items-center justify-center">
					{isImageLoading && (
						<div className="absolute inset-0 bg-neutral-content animate-pulse flex items-center justify-center">
							<div className="w-8 h-8 border-2 border-primary border-t-transparent rounded-full animate-spin" />
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
		</div>
	)
}

export default CatalogCard
