import { useCallback, useEffect, useRef, useState } from "react"
import { FaChevronLeft, FaChevronRight } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import ImageFullscreen from "../dialogs/ImageFullscreen"

interface GalleryProps {
	images: string[]
	altText?: string
	className?: string
}

export default function Gallery({ images, altText = "Изображение", className = "" }: GalleryProps) {
	const [currentIndex, setCurrentIndex] = useState(0)
	const [validImages, setValidImages] = useState<string[]>([])
	const [isLoading, setIsLoading] = useState(true)
	const touchStartX = useRef<number | null>(null)
	const touchEndX = useRef<number | null>(null)

	useEffect(() => {
		let cancelled = false
		const cleanImages = images.filter((img) => img && img.trim() !== "")

		if (cleanImages.length === 0) {
			setValidImages([])
			setIsLoading(false)
			return
		}

		setIsLoading(true)

		const imagePromises = cleanImages.map((src) => {
			return new Promise<string | null>((resolve) => {
				const img = new Image()
				let settled = false
				const timeoutId = window.setTimeout(() => {
					if (settled) return
					settled = true
					resolve(null)
				}, 8000)
				const finish = (value: string | null) => {
					if (settled) return
					settled = true
					window.clearTimeout(timeoutId)
					resolve(value)
				}
				img.onload = () => finish(src)
				img.onerror = () => finish(null)
				img.src = src
			})
		})

		void Promise.all(imagePromises).then((results) => {
			if (cancelled) return
			const loaded = results.filter((s): s is string => s !== null)
			setValidImages(loaded)
			setIsLoading(false)
		})

		return () => {
			cancelled = true
		}
	}, [images])

	const goToPrevious = useCallback(() => {
		setCurrentIndex((prevIndex) => (prevIndex === 0 ? validImages.length - 1 : prevIndex - 1))
	}, [validImages.length])

	const goToNext = useCallback(() => {
		setCurrentIndex((prevIndex) => (prevIndex === validImages.length - 1 ? 0 : prevIndex + 1))
	}, [validImages.length])

	useEffect(() => {
		if (currentIndex >= validImages.length && validImages.length > 0) {
			setCurrentIndex(0)
		}
	}, [validImages.length, currentIndex])

	const goToImage = (index: number) => {
		setCurrentIndex(index)
	}

	const handleTouchStart = (e: React.TouchEvent) => {
		touchStartX.current = e.targetTouches[0].clientX
	}

	const handleTouchMove = (e: React.TouchEvent) => {
		touchEndX.current = e.targetTouches[0].clientX
	}

	const handleTouchEnd = () => {
		if (!touchStartX.current || !touchEndX.current) return

		const distance = touchStartX.current - touchEndX.current
		const isLeftSwipe = distance > 50
		const isRightSwipe = distance < -50

		if (isLeftSwipe && validImages.length > 1) {
			goToNext()
		} else if (isRightSwipe && validImages.length > 1) {
			goToPrevious()
		}

		touchStartX.current = null
		touchEndX.current = null
	}

	useEffect(() => {
		const handleKeyDown = (e: KeyboardEvent) => {
			if (validImages.length <= 1) return

			if (e.key === "ArrowLeft") {
				e.preventDefault()
				goToPrevious()
			} else if (e.key === "ArrowRight") {
				e.preventDefault()
				goToNext()
			}
		}

		window.addEventListener("keydown", handleKeyDown)
		return () => window.removeEventListener("keydown", handleKeyDown)
	}, [validImages.length, goToNext, goToPrevious])

	const chromeBtn =
		"btn btn-circle btn-sm min-h-0 h-9 w-9 border border-base-300/90 bg-base-100/90 p-0 text-base-content shadow-md backdrop-blur-sm transition-colors hover:border-primary/30 hover:bg-base-100"

	if (isLoading) {
		return (
			<div
				className={`flex aspect-4/3 items-center justify-center rounded-2xl border border-base-200 bg-base-200/30 ${className}`}
			>
				<div className="text-center">
					<div className="loading loading-spinner loading-lg text-primary mb-2" />
				</div>
			</div>
		)
	}

	if (validImages.length === 0) {
		return (
			<div
				className={`flex aspect-4/3 flex-col items-center justify-center gap-2 rounded-2xl border border-dashed border-base-300 bg-base-200/25 px-6 text-center ${className}`}
			>
				<div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-base-200/80 text-base-content/50">
					<IoImageOutline className="h-8 w-8" aria-hidden />
				</div>
				<p className="text-sm text-base-content/60">Нет изображения</p>
			</div>
		)
	}

	return (
		<div className={`space-y-3 ${className}`}>
			<div className="group relative">
				<div
					className="aspect-4/3 w-full touch-pan-y select-none overflow-hidden rounded-2xl border border-base-200 bg-base-100 shadow-sm"
					onTouchStart={handleTouchStart}
					onTouchMove={handleTouchMove}
					onTouchEnd={handleTouchEnd}
				>
					<ImageFullscreen
						src={validImages[currentIndex]}
						alt={`${altText} ${currentIndex + 1}`}
						className="h-full w-full object-cover transition-opacity duration-150 pointer-events-none"
					/>
				</div>

				{validImages.length > 1 && (
					<>
						<button
							type="button"
							onClick={goToPrevious}
							className={`absolute left-2 top-1/2 z-10 -translate-y-1/2 sm:opacity-90 ${chromeBtn} max-sm:opacity-100`}
							aria-label="Предыдущее изображение"
						>
							<FaChevronLeft className="h-3.5 w-3.5" />
						</button>
						<button
							type="button"
							onClick={goToNext}
							className={`absolute right-2 top-1/2 z-10 -translate-y-1/2 sm:opacity-90 ${chromeBtn} max-sm:opacity-100`}
							aria-label="Следующее изображение"
						>
							<FaChevronRight className="h-3.5 w-3.5" />
						</button>
					</>
				)}

				{validImages.length > 1 && (
					<div className="absolute bottom-2 right-2 rounded-full border border-base-300/80 bg-base-100/90 px-2 py-0.5 text-xs font-medium tabular-nums text-base-content shadow-sm backdrop-blur-sm">
						{currentIndex + 1} / {validImages.length}
					</div>
				)}
			</div>

			{validImages.length > 1 && (
				<div className="hidden gap-2 overflow-x-auto pb-1 scrollbar-hide sm:flex">
					{validImages.map((image, index) => (
						<button
							key={`thumbnail-${image}`}
							type="button"
							onClick={() => goToImage(index)}
							className={`h-16 w-16 shrink-0 overflow-hidden rounded-xl border-2 bg-base-100 transition-all duration-200 ${
								index === currentIndex
									? "border-primary ring-2 ring-primary/25"
									: "border-base-200 hover:border-base-300"
							}`}
							aria-label={`Показать изображение ${index + 1}`}
						>
							<img
								src={image}
								alt={`${altText} thumbnail ${index + 1}`}
								className="h-full w-full object-cover"
								loading="lazy"
							/>
						</button>
					))}
				</div>
			)}

			{validImages.length > 1 && (
				<div className="sm:hidden">
					{validImages.length <= 4 ? (
						<div className="flex justify-center gap-2">
							{validImages.map((image, index) => (
								<button
									key={`mobile-thumb-${image}`}
									type="button"
									onClick={() => goToImage(index)}
									className={`h-12 w-12 shrink-0 touch-manipulation overflow-hidden rounded-xl border-2 bg-base-100 transition-all duration-200 ${
										index === currentIndex
											? "border-primary ring-2 ring-primary/25"
											: "border-base-200 active:border-base-300"
									}`}
									aria-label={`Показать изображение ${index + 1}`}
								>
									<img
										src={image}
										alt={`${altText} thumbnail ${index + 1}`}
										className="h-full w-full object-cover"
										loading="lazy"
									/>
								</button>
							))}
						</div>
					) : (
						<div className="space-y-2">
							<div className="flex justify-center gap-2.5">
								{validImages.map((image, index) => (
									<button
										key={`dot-${image}`}
										type="button"
										onClick={() => goToImage(index)}
										className={`h-2.5 w-2.5 touch-manipulation rounded-full transition-all duration-200 ${
											index === currentIndex
												? "scale-110 bg-primary"
												: "bg-base-content/25 active:bg-base-content/45"
										}`}
										aria-label={`Перейти к изображению ${index + 1}`}
									/>
								))}
							</div>
							<p className="text-center text-xs text-base-content/55">Свайп для переключения</p>
						</div>
					)}
				</div>
			)}
		</div>
	)
}
