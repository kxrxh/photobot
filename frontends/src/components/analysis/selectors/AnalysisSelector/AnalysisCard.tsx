import { memo, useEffect, useMemo, useState } from "react"
import { FaCheckCircle, FaClock, FaDownload, FaEye, FaPlus, FaTrash } from "react-icons/fa"
import { IoImageOutline } from "react-icons/io5"
import type { Analysis } from "@/api/analysis/types"
import { getAnalysisSourceUrls } from "@/utils/image"
import { log } from "@/utils/log"

export const formatAnalysisDate = (dateStr: string) => {
	const date = new Date(dateStr)
	return new Intl.DateTimeFormat("ru-RU", {
		day: "2-digit",
		month: "2-digit",
		year: "numeric",
		hour: "2-digit",
		minute: "2-digit",
	}).format(date)
}

export interface AnalysisCardProps {
	analysis: Analysis
	imageUrls?: string[]
	isSelected: boolean
	isDownloading: boolean
	mode: "selection" | "view-only"
	selectionMode: "single" | "multiple"
	onSelect: (analysis: Analysis) => void
	onView: (analysis: Analysis) => void
	onDownload: (analysis: Analysis) => Promise<void>
}

export const AnalysisCard = memo<AnalysisCardProps>(
	({
		analysis,
		imageUrls: hydratedImageUrls,
		isSelected,
		isDownloading,
		mode,
		selectionMode,
		onSelect,
		onView,
		onDownload,
	}) => {
		const imageUrls = useMemo(
			() => hydratedImageUrls ?? getAnalysisSourceUrls(analysis),
			[analysis, hydratedImageUrls]
		)
		const urlsKey = useMemo(() => imageUrls.join("\0"), [imageUrls])
		const [previewIndex, setPreviewIndex] = useState(0)
		const [imageError, setImageError] = useState(false)
		const [isImageLoading, setIsImageLoading] = useState(false)

		// biome-ignore lint/correctness/useExhaustiveDependencies: urlsKey intentionally drives reset
		useEffect(() => {
			setPreviewIndex(0)
			setImageError(false)
		}, [urlsKey])

		const previewImageUrl = imageUrls[previewIndex]

		useEffect(() => {
			setIsImageLoading(Boolean(previewImageUrl))
		}, [previewImageUrl])

		const showImageFallback = !previewImageUrl || imageError

		return (
			<div
				className={`group relative flex flex-col gap-4 rounded-3xl border p-4 shadow-sm transition-all duration-300 ease-out backdrop-blur-md sm:p-5 ${
					isSelected
						? "border-primary bg-primary/10 shadow-lg"
						: "border-base-200 bg-base-100 hover:border-primary hover:bg-primary/5 hover:shadow-lg hover:-translate-y-0.5"
				}`}
			>
				{isSelected && (
					<span className="absolute right-4 top-4 flex items-center gap-1 rounded-full bg-primary px-3 py-1 text-xs font-semibold uppercase tracking-wide text-primary-content shadow">
						<FaCheckCircle size={12} />
						Выбрано
					</span>
				)}
				<div className="flex flex-col gap-4 sm:flex-row sm:items-center">
					<div className="relative w-full shrink-0 overflow-hidden rounded-2xl border border-base-200 bg-base-200 aspect-4/3 sm:w-28 sm:aspect-square">
						<div className="flex h-full w-full items-center justify-center bg-base-200">
							{isImageLoading && !showImageFallback && (
								<div className="absolute inset-0 z-10 flex items-center justify-center bg-base-200/90 animate-pulse">
									<div className="h-7 w-7 rounded-full border-2 border-primary border-t-transparent animate-spin" />
								</div>
							)}

							{showImageFallback ? (
								<div className="absolute inset-0 flex items-center justify-center text-base-content opacity-40">
									<IoImageOutline size={32} />
								</div>
							) : (
								<img
									src={previewImageUrl}
									alt="Превью анализа"
									loading="lazy"
									decoding="async"
									className={`h-full w-full object-cover transition-all duration-500 ${
										isImageLoading
											? "scale-105 opacity-0"
											: "scale-100 opacity-100 group-hover:scale-105"
									}`}
									onError={() => {
										const next = previewIndex + 1
										if (next < imageUrls.length) {
											setPreviewIndex(next)
											setIsImageLoading(true)
										} else {
											setImageError(true)
											setIsImageLoading(false)
										}
									}}
									onLoad={() => {
										setImageError(false)
										setIsImageLoading(false)
									}}
								/>
							)}
						</div>
						{imageUrls.length > 1 && (
							<span className="absolute right-3 top-3 rounded-full bg-base-100 px-2 py-1 text-[10px] font-semibold uppercase tracking-wide text-base-content shadow">
								{imageUrls.length} фото
							</span>
						)}
					</div>
					<div className="flex flex-1 flex-col gap-3">
						<div className="flex flex-wrap items-center gap-2">
							<h3 className="text-lg font-semibold leading-tight text-base-content">
								Анализ №{analysis.id}
							</h3>
						</div>
						<div className="flex flex-wrap items-center gap-2 text-sm text-base-content opacity-80">
							<FaClock className="text-primary" size={14} />
							<span>{formatAnalysisDate(analysis.date_time)}</span>
						</div>
					</div>
				</div>
				<div className="flex flex-wrap items-center gap-2 sm:justify-end sm:gap-3">
					<button
						type="button"
						onClick={async (e) => {
							e.stopPropagation()
							if (isDownloading) return
							try {
								await onDownload(analysis)
							} catch (error) {
								log.devError("Failed to download report:", error)
							}
						}}
						disabled={isDownloading}
						className={`btn btn-sm btn-soft btn-neutral gap-2 rounded-full px-4 ${
							isDownloading ? "btn-disabled" : ""
						}`}
						title={isDownloading ? "Генерация отчета..." : "Скачать отчет"}
						aria-label={`Скачать отчет для анализа ${analysis.id}`}
					>
						{isDownloading ? (
							<div className="loading loading-spinner loading-xs" />
						) : (
							<FaDownload size={16} />
						)}
						<span className="text-xs font-semibold">{isDownloading ? "Генерация..." : "PDF"}</span>
					</button>
					<button
						type="button"
						onClick={(e) => {
							e.stopPropagation()
							onView(analysis)
						}}
						className="btn btn-sm btn-soft btn-accent gap-2 rounded-full px-4"
						title="Просмотр анализа"
						aria-label={`Просмотр анализа ${analysis.id}`}
					>
						<FaEye size={16} />
						<span className="text-xs font-semibold">Открыть</span>
					</button>

					{mode === "selection" && (
						<button
							type="button"
							onClick={(e) => {
								e.stopPropagation()
								onSelect(analysis)
							}}
							className={`btn btn-sm btn-block gap-2 rounded-full px-4 ${
								isSelected ? "btn-soft btn-error" : "btn-soft btn-primary"
							}`}
							title={
								isSelected
									? selectionMode === "single"
										? "Отменить выбор"
										: "Удалить анализ"
									: "Добавить анализ"
							}
							aria-label={
								isSelected
									? selectionMode === "single"
										? `Отменить выбор анализа ${analysis.id}`
										: `Удалить анализ ${analysis.id}`
									: `Добавить анализ ${analysis.id}`
							}
						>
							{isSelected ? <FaTrash size={16} /> : <FaPlus size={16} />}
							<span className="text-xs font-semibold">
								{isSelected ? (selectionMode === "single" ? "Убрать" : "Удалить") : "Добавить"}
							</span>
						</button>
					)}
				</div>
			</div>
		)
	}
)

AnalysisCard.displayName = "AnalysisCard"
