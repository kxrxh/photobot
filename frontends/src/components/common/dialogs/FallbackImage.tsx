import type React from "react"
import { useState } from "react"
import coffeeBeanIcon from "@/assets/coffee-bean.svg"

interface FallbackImageProps {
	src?: string
	fallbackSrc?: string
	className?: string
	isFullscreen?: boolean
	alt?: string
}

const FallbackImage: React.FC<FallbackImageProps> = ({
	src,
	fallbackSrc = coffeeBeanIcon,
	className = "",
	isFullscreen = false,
	alt = "Изображение",
}) => {
	const [hasError, setHasError] = useState(false)

	const handleError = () => {
		setHasError(true)
	}

	const showFallback = !src || hasError

	return (
		<div
			className={`bg-neutral-content flex items-center justify-center ${
				isFullscreen ? "w-96 h-96 rounded-lg shadow-2xl" : className
			}`}
			style={!isFullscreen ? { aspectRatio: "4/3" } : undefined}
		>
			{showFallback ? (
				<img src={fallbackSrc} alt={alt} className="text-neutral-content w-full h-full" />
			) : (
				<img
					src={src}
					alt={alt}
					onError={handleError}
					className="object-cont w-full h-full"
					loading="lazy"
					decoding="async"
				/>
			)}
		</div>
	)
}

export default FallbackImage
