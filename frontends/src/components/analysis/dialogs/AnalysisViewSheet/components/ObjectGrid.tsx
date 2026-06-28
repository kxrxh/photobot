import type React from "react"
import { useEffect, useState } from "react"
import { IoImageOutline } from "react-icons/io5"
import { getImageDataUrl } from "@/utils/image"

export interface GridObject {
	id: number
	image?: string
}

interface ObjectGridProps {
	objects: GridObject[]
}

const ObjectImage: React.FC<{ image?: string; alt: string }> = ({ image, alt }) => {
	const imageUrl = getImageDataUrl(image)
	const [hasError, setHasError] = useState(!imageUrl)

	useEffect(() => {
		setHasError(!imageUrl)
	}, [imageUrl])

	if (hasError) {
		return (
			<div className="flex justify-center items-center w-full h-full bg-base-200">
				<IoImageOutline className="w-1/2 h-1/2 text-base-content/30" />
			</div>
		)
	}

	return (
		<img
			src={imageUrl || undefined}
			alt={alt}
			className="object-contain w-full h-full"
			loading="lazy"
			onError={() => setHasError(true)}
		/>
	)
}

const ObjectGrid: React.FC<ObjectGridProps> = ({ objects }) => {
	return (
		<div className="grid grid-cols-4 gap-2 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10">
			{objects.map((object) => (
				<div
					key={object.id}
					className="overflow-hidden relative rounded-2xl border aspect-square border-base-200 bg-white"
				>
					<ObjectImage image={object.image} alt={`Object ${object.id}`} />
				</div>
			))}
		</div>
	)
}

export default ObjectGrid
