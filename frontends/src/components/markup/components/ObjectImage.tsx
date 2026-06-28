import { memo, useState } from "react"
import { FaCheck, FaTimes } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import { getImageDataUrl } from "@/utils/image"

/** Analysis `SimpleObject` or markup fraction rows (`KalibriObjectWithSource` shape). */
export type ObjectImageObject = {
	id: string | number
	file?: string
	presigned_url?: string
	image_url?: string
}

function objectImageSource(o: ObjectImageObject): string | undefined {
	return o.presigned_url ?? o.file ?? o.image_url
}

interface ObjectImageProps {
	object: ObjectImageObject
	isSelected: boolean
	isExcluded?: boolean
	isControlModeActive: boolean
	mode?: "select" | "exclude"
	onClick: () => void
	size?: "sm" | "md" | "lg" | "xl"
}

const sizeClasses = {
	sm: "w-12 h-12",
	md: "w-16 h-16",
	lg: "w-20 h-20",
	xl: "w-24 h-24",
} as const

const iconSizeClasses = {
	sm: "w-4 h-4",
	md: "w-6 h-6",
	lg: "w-8 h-8",
	xl: "w-10 h-10",
} as const

type ObjectImageTileProps = ObjectImageProps & { imageUrl: string | null }

/** Keyed by `imageUrl` in the parent so load error state resets when the URL changes (no sync effect). */
const ObjectImageTile = memo(
	({
		object,
		isSelected,
		isExcluded = false,
		isControlModeActive,
		mode = "select",
		onClick,
		size = "md",
		imageUrl,
	}: ObjectImageTileProps) => {
		const [loadFailed, setLoadFailed] = useState(false)
		const showPlaceholder = !imageUrl || loadFailed

		const isExcludeMode = mode === "exclude"
		const tileBorder =
			isExcludeMode && isExcluded
				? "border-2 border-dashed border-base-content/25"
				: "border-2 border-base-300"
		const tileRing =
			isControlModeActive && isExcludeMode && !isExcluded
				? "hover:border-primary/35 focus-visible:border-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/25"
				: ""

		return (
			<button
				type="button"
				className={`relative flex justify-center items-center ${sizeClasses[size]} rounded-lg bg-base-100 transition-colors duration-200 ${tileBorder} ${tileRing} ${
					isControlModeActive ? "cursor-pointer" : ""
				} ${isExcludeMode && isExcluded ? "opacity-80" : ""}`}
				onClick={onClick}
			>
				{showPlaceholder ? (
					<IoImageOutline className={`${iconSizeClasses[size]} text-base-content/30`} />
				) : (
					<img
						src={imageUrl}
						alt={String(object.id)}
						className={`w-full h-full object-contain rounded-md ${isExcludeMode && isExcluded ? "grayscale-[0.35]" : ""}`}
						loading="lazy"
						onError={() => setLoadFailed(true)}
					/>
				)}
				{isSelected && mode === "select" && (
					<div className="absolute inset-0 flex items-center justify-center rounded-md bg-primary/50">
						<FaCheck className={`${iconSizeClasses[size]} text-primary-content`} />
					</div>
				)}
				{isExcluded && mode === "exclude" && (
					<div
						className="absolute inset-0 flex items-center justify-center rounded-md bg-base-content/45"
						aria-hidden
					>
						<FaTimes className={`${iconSizeClasses[size]} text-base-100`} />
					</div>
				)}
			</button>
		)
	}
)

export const ObjectImage = memo((props: ObjectImageProps) => {
	const imageUrl = getImageDataUrl(objectImageSource(props.object))
	return (
		<ObjectImageTile key={`${props.object.id}-${imageUrl ?? ""}`} {...props} imageUrl={imageUrl} />
	)
})
