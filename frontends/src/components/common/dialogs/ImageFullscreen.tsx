import { type ReactNode, useCallback, useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { IoClose } from "react-icons/io5"
import FallbackImage from "./FallbackImage"

interface ImageFullscreenProps {
	src: string
	alt: string
	className?: string
	loadingDelay?: number // in milliseconds
	children?: ReactNode // Optional custom thumbnail content
	isClickable?: boolean // New prop
}

export default function ImageFullscreen({
	src,
	alt,
	className = "",
	loadingDelay = 0, // no delay by default
	children,
	isClickable = true, // Default to true
}: ImageFullscreenProps) {
	const [isOpen, setIsOpen] = useState(false)
	const [isMounted, setIsMounted] = useState(false)
	const [error, setError] = useState(false)
	const [isLoading, setIsLoading] = useState(true)
	const loadingTimerRef = useRef<number | null>(null)
	const fallbackTimerRef = useRef<number | null>(null)

	useEffect(() => {
		setIsMounted(true)
	}, [])

	useEffect(() => {
		if (loadingTimerRef.current !== null) {
			window.clearTimeout(loadingTimerRef.current)
			loadingTimerRef.current = null
		}
		if (fallbackTimerRef.current !== null) {
			window.clearTimeout(fallbackTimerRef.current)
			fallbackTimerRef.current = null
		}

		if (!src) {
			setError(true)
			setIsLoading(false)
			return
		}

		setError(false)
		setIsLoading(true)

		// Safety net: never leave skeleton hanging forever if browser misses events.
		fallbackTimerRef.current = window.setTimeout(() => {
			setIsLoading(false)
			fallbackTimerRef.current = null
		}, 8000)

		return () => {
			if (fallbackTimerRef.current !== null) {
				window.clearTimeout(fallbackTimerRef.current)
				fallbackTimerRef.current = null
			}
		}
	}, [src])

	useEffect(() => {
		return () => {
			if (loadingTimerRef.current !== null) {
				window.clearTimeout(loadingTimerRef.current)
				loadingTimerRef.current = null
			}
			if (fallbackTimerRef.current !== null) {
				window.clearTimeout(fallbackTimerRef.current)
				fallbackTimerRef.current = null
			}
		}
	}, [])

	const handleOpen = () => setIsOpen(true)
	const handleClose = useCallback(() => setIsOpen(false), [])
	const handleImageLoad = useCallback(() => {
		if (loadingTimerRef.current !== null) {
			window.clearTimeout(loadingTimerRef.current)
			loadingTimerRef.current = null
		}
		if (fallbackTimerRef.current !== null) {
			window.clearTimeout(fallbackTimerRef.current)
			fallbackTimerRef.current = null
		}
		if (loadingDelay > 0) {
			loadingTimerRef.current = window.setTimeout(() => {
				setIsLoading(false)
				loadingTimerRef.current = null
			}, loadingDelay)
			return
		}
		setIsLoading(false)
	}, [loadingDelay])

	const handleImageElement = useCallback(
		(node: HTMLImageElement | null) => {
			if (!node) return
			if (node.complete && node.naturalWidth > 0) {
				handleImageLoad()
			}
		},
		[handleImageLoad]
	)

	useEffect(() => {
		if (!isOpen) return
		const handleKeyDown = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				handleClose()
			}
		}
		window.addEventListener("keydown", handleKeyDown)
		return () => {
			window.removeEventListener("keydown", handleKeyDown)
		}
	}, [isOpen, handleClose])

	const handleBackdropKeyDown = (e: React.KeyboardEvent) => {
		if (e.key === "Enter" || e.key === " ") {
			if (e.target === e.currentTarget) {
				handleClose()
			}
		}
	}

	const renderImage = (isFullscreen = false) => {
		if (!src || error) {
			return <FallbackImage className={isFullscreen ? "" : className} isFullscreen={isFullscreen} />
		}

		return (
			<div className={isFullscreen ? "" : "relative"}>
				{isLoading && !isFullscreen && (
					<div
						className={`skeleton absolute inset-0 rounded-none ${className}`}
						style={{ aspectRatio: "4/3" }}
					/>
				)}
				<img
					ref={handleImageElement}
					src={src}
					alt={alt}
					className={
						isFullscreen
							? "max-h-[90vh] max-w-[90vw] object-contain rounded-lg shadow-2xl"
							: `${className} ${isLoading ? "opacity-0" : "opacity-100"} transition-opacity duration-200`
					}
					onError={() => {
						setError(true)
						setIsLoading(false)
					}}
					onLoad={handleImageLoad}
				/>
			</div>
		)
	}

	const thumbnailContent = children || renderImage(false)

	return (
		<>
			{isClickable ? (
				<button type="button" onClick={handleOpen} className="contents" disabled={false}>
					{thumbnailContent}
				</button>
			) : (
				<div>{thumbnailContent}</div>
			)}
			{isMounted && isOpen
				? createPortal(
						<div
							className="fixed inset-0 z-10000 flex items-center justify-center p-4 bg-black/80 backdrop-blur-sm"
							onClick={handleClose}
							onKeyDown={handleBackdropKeyDown}
							role="dialog"
							aria-modal="true"
							aria-label="Просмотр изображения"
						>
							<button
								type="button"
								className="btn btn-sm btn-circle btn-ghost absolute right-4 top-4 z-50 text-white hover:bg-white/20 border-white/20"
								onClick={(e) => {
									e.stopPropagation()
									handleClose()
								}}
								aria-label="Закрыть изображение"
							>
								<IoClose size={24} />
							</button>
							{/* biome-ignore lint/a11y/noStaticElementInteractions: Inner wrapper only stops propagation, not interactive */}
							<div
								className="relative max-h-full max-w-full flex items-center justify-center"
								onClick={(e) => e.stopPropagation()}
								onKeyDown={(e) => e.stopPropagation()}
							>
								{renderImage(true)}
							</div>
						</div>,
						document.body
					)
				: null}
		</>
	)
}
