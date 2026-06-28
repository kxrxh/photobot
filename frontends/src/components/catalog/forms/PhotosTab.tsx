import type React from "react"
import { useEffect, useId, useMemo, useRef } from "react"
import { FaCloudUploadAlt, FaTrash } from "react-icons/fa"

/** Matches `UploadTab` section labels */
const SECTION_LABEL_CLASS =
	"mb-1.5 block text-xs font-semibold uppercase tracking-wide text-base-content/45"

/** Matches `UploadTab` photo upload zone */
const UPLOAD_ZONE_BASE =
	"flex w-full flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed px-4 py-10 transition-all sm:py-12"
const UPLOAD_ZONE_FOCUS =
	"focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
const UPLOAD_ZONE_DISABLED = "cursor-not-allowed border-base-300/60 bg-base-200/20 opacity-60"
const UPLOAD_ZONE_ENABLED =
	"cursor-pointer border-primary/35 bg-primary/5 hover:border-primary/50 hover:bg-primary/10 active:scale-[0.99]"

const UPLOAD_ICON_WRAP_BASE = "flex h-14 w-14 items-center justify-center rounded-2xl"
const UPLOAD_ICON_WRAP_DISABLED = "bg-base-300/40 text-base-content/35"
const UPLOAD_ICON_WRAP_ENABLED = "bg-primary/15 text-primary"

const UPLOAD_TITLE_BASE = "text-sm font-semibold"
const UPLOAD_TITLE_DISABLED = "text-base-content/45"
const UPLOAD_TITLE_ENABLED = "text-base-content"

/** Compact dashed slot for extra photos (UploadTab-style, smaller) */
const SLOT_ZONE_BASE =
	"flex aspect-square w-full flex-col items-center justify-center gap-1.5 rounded-xl border-2 border-dashed px-2 py-3 transition-all"
const SLOT_ZONE_ENABLED =
	"cursor-pointer border-primary/35 bg-primary/5 hover:border-primary/50 hover:bg-primary/10 active:scale-[0.99]"
const SLOT_ZONE_DISABLED = "cursor-not-allowed border-base-300/60 bg-base-200/20 opacity-60"

const SLOT_ICON_WRAP =
	"flex h-10 w-10 items-center justify-center rounded-xl bg-primary/15 text-primary"
const SLOT_ICON_WRAP_DISABLED =
	"flex h-10 w-10 items-center justify-center rounded-xl bg-base-300/40 text-base-content/35"

const ADDITIONAL_PHOTO_SLOT_KEYS = Array.from({ length: 6 }, () => crypto.randomUUID())

interface PhotosTabProps {
	photos: (File | string | null)[]
	onPhotosChange: (photos: (File | string | null)[]) => void
	readOnly?: boolean
}

type BlobPreviewEntry = { index: number; file: File; url: string }

function photoDisplaySrc(
	photo: File | string | null,
	index: number,
	blobEntries: BlobPreviewEntry[]
): string | null {
	if (!photo) return null
	if (typeof photo === "string") return photo
	const match = blobEntries.find((e) => e.index === index && e.file === photo)
	return match?.url ?? null
}

function photoCaption(photo: File | string | null, index: number): string {
	if (!photo) return ""
	if (typeof photo === "string") return `Фото ${index + 1}`
	return photo.name
}

function photoSizeLine(photo: File | string | null): string | null {
	if (photo && typeof photo !== "string") {
		return `${(photo.size / 1024 / 1024).toFixed(1)} МБ`
	}
	return null
}

const PhotosTab: React.FC<PhotosTabProps> = ({ photos, onPhotosChange, readOnly = false }) => {
	const mainSectionId = useId()
	const additionalSectionId = useId()
	const mainPhoto: File | string | null = photos[0] ?? null
	const additionalPhotos: (File | string | null)[] = [
		...photos.slice(1),
		...Array(Math.max(0, 6 - Math.max(0, photos.length - 1))).fill(null),
	].slice(0, 6)
	const mainPhotoInputRef = useRef<HTMLInputElement>(null)
	const additionalPhotoInputRefs = useRef<(HTMLInputElement | null)[]>(Array(6).fill(null))

	const blobPreviewEntries = useMemo((): BlobPreviewEntry[] => {
		const entries: BlobPreviewEntry[] = []
		photos.forEach((p, index) => {
			if (p && typeof p !== "string") {
				entries.push({ index, file: p, url: URL.createObjectURL(p) })
			}
		})
		return entries
	}, [photos])

	useEffect(() => {
		return () => {
			for (const { url } of blobPreviewEntries) {
				URL.revokeObjectURL(url)
			}
		}
	}, [blobPreviewEntries])

	const handleFileSelect = (file: File, onLoad: (file: File) => void) => {
		if (file?.type.startsWith("image/")) {
			onLoad(file)
		}
	}

	const handleMainPhotoSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		if (readOnly) return
		const file = e.target.files?.[0]
		if (file) {
			handleFileSelect(file, (selectedFile) => {
				const newPhotos = [selectedFile, ...additionalPhotos]
				onPhotosChange(newPhotos)
			})
		}
	}

	const handleAdditionalPhotoSelect = (e: React.ChangeEvent<HTMLInputElement>, index: number) => {
		if (readOnly) return
		const file = e.target.files?.[0]
		if (file) {
			handleFileSelect(file, (selectedFile) => {
				const newAdditional: (File | string | null)[] = [...additionalPhotos]
				newAdditional[index] = selectedFile
				const newPhotos: (File | string | null)[] = [mainPhoto, ...newAdditional]
				onPhotosChange(newPhotos)
			})
		}
	}

	const removeMainPhoto = () => {
		if (readOnly) return
		const newPhotos = [null, ...additionalPhotos]
		onPhotosChange(newPhotos)
		if (mainPhotoInputRef.current) {
			mainPhotoInputRef.current.value = ""
		}
	}

	const removeAdditionalPhoto = (index: number) => {
		if (readOnly) return
		const newAdditional: (File | string | null)[] = [...additionalPhotos]
		newAdditional[index] = null
		const filteredPhotos = newAdditional.filter((photo) => photo !== null)
		const compactedPhotos = filteredPhotos.concat(Array(6 - filteredPhotos.length).fill(null))
		const newPhotosList: (File | string | null)[] = [mainPhoto, ...compactedPhotos]
		onPhotosChange(newPhotosList)
		if (additionalPhotoInputRefs.current[index]) {
			additionalPhotoInputRefs.current[index].value = ""
		}
	}

	const triggerMainPhotoSelect = () => {
		if (readOnly) return
		mainPhotoInputRef.current?.click()
	}

	const triggerAdditionalPhotoSelect = (index: number) => {
		if (readOnly) return
		additionalPhotoInputRefs.current[index]?.click()
	}

	const mainSrc = photoDisplaySrc(mainPhoto, 0, blobPreviewEntries)

	return (
		<div className="animate-fadeIn mx-auto flex max-w-2xl flex-col gap-8 px-4 pb-8">
			<section aria-labelledby={mainSectionId}>
				<div id={mainSectionId} className={SECTION_LABEL_CLASS}>
					Основное фото
				</div>
				<input
					ref={mainPhotoInputRef}
					type="file"
					accept="image/*,.jpg,.jpeg,.png,.heic,.heif"
					onChange={handleMainPhotoSelect}
					className="hidden"
				/>
				{mainPhoto ? (
					<div className="group relative overflow-hidden rounded-xl border border-base-200 bg-base-100 shadow-sm">
						<div className="aspect-video bg-base-200">
							<img
								src={mainSrc ?? ""}
								alt="Основное фото товара"
								className="h-full w-full object-cover"
							/>
						</div>
						<div className="border-t border-base-200 bg-base-100/95 p-2 backdrop-blur-sm">
							<p className="truncate text-xs font-medium" title={photoCaption(mainPhoto, 0)}>
								{photoCaption(mainPhoto, 0)}
							</p>
							{photoSizeLine(mainPhoto) ? (
								<p className="text-[10px] text-base-content/50">{photoSizeLine(mainPhoto)}</p>
							) : (
								<p className="text-[10px] text-base-content/50">Главное изображение карточки</p>
							)}
						</div>
						{!readOnly && (
							<button
								type="button"
								onClick={removeMainPhoto}
								className="absolute right-1.5 top-1.5 flex h-8 w-8 items-center justify-center rounded-full bg-base-100/90 text-error shadow-md backdrop-blur-sm transition-opacity hover:bg-error/10"
								aria-label="Удалить основное фото"
								title="Удалить"
							>
								<FaTrash className="h-3.5 w-3.5" />
							</button>
						)}
					</div>
				) : (
					<button
						type="button"
						disabled={readOnly}
						onClick={triggerMainPhotoSelect}
						className={[
							UPLOAD_ZONE_BASE,
							UPLOAD_ZONE_FOCUS,
							readOnly ? UPLOAD_ZONE_DISABLED : UPLOAD_ZONE_ENABLED,
						].join(" ")}
					>
						<span
							className={[
								UPLOAD_ICON_WRAP_BASE,
								readOnly ? UPLOAD_ICON_WRAP_DISABLED : UPLOAD_ICON_WRAP_ENABLED,
							].join(" ")}
						>
							<FaCloudUploadAlt className="h-7 w-7" aria-hidden />
						</span>
						<div className="text-center">
							<p
								className={[
									UPLOAD_TITLE_BASE,
									readOnly ? UPLOAD_TITLE_DISABLED : UPLOAD_TITLE_ENABLED,
								].join(" ")}
							>
								{readOnly ? "Нет основного фото" : "Нажмите, чтобы добавить основное фото"}
							</p>
							<p className="mt-1 max-w-xs text-xs text-base-content/55">
								{readOnly
									? "Изображение не загружено."
									: "JPG, PNG, HEIC/HEIF · одно изображение для обложки карточки"}
							</p>
						</div>
					</button>
				)}
			</section>

			<section aria-labelledby={additionalSectionId}>
				<div id={additionalSectionId} className={SECTION_LABEL_CLASS}>
					Дополнительные фото
				</div>
				<p className="mb-3 text-xs text-base-content/55">До шести снимков · по желанию</p>
				<ul className="grid list-none grid-cols-2 gap-2 p-0 sm:grid-cols-3 sm:gap-3">
					{additionalPhotos.map((photo, index) => {
						const slotSrc = photoDisplaySrc(photo, index + 1, blobPreviewEntries)
						return (
							<li key={ADDITIONAL_PHOTO_SLOT_KEYS[index]} className="relative">
								<input
									ref={(el) => {
										if (el) additionalPhotoInputRefs.current[index] = el
									}}
									type="file"
									accept="image/*,.jpg,.jpeg,.png,.heic,.heif"
									onChange={(e) => handleAdditionalPhotoSelect(e, index)}
									className="hidden"
								/>
								{photo ? (
									<div className="group relative overflow-hidden rounded-xl border border-base-200 bg-base-100 shadow-sm">
										<div className="aspect-square bg-base-200">
											<img
												src={slotSrc ?? ""}
												alt={`Дополнительное фото ${index + 1}`}
												className="h-full w-full object-cover"
											/>
										</div>
										<div className="border-t border-base-200 bg-base-100/95 p-2 backdrop-blur-sm">
											<p
												className="truncate text-xs font-medium"
												title={photoCaption(photo, index + 1)}
											>
												{photoCaption(photo, index + 1)}
											</p>
											{photoSizeLine(photo) ? (
												<p className="text-[10px] text-base-content/50">{photoSizeLine(photo)}</p>
											) : null}
										</div>
										{!readOnly && (
											<button
												type="button"
												onClick={() => removeAdditionalPhoto(index)}
												className="absolute right-1.5 top-1.5 flex h-8 w-8 items-center justify-center rounded-full bg-base-100/90 text-error shadow-md backdrop-blur-sm transition-opacity hover:bg-error/10"
												aria-label={`Удалить дополнительное фото ${index + 1}`}
												title="Удалить"
											>
												<FaTrash className="h-3.5 w-3.5" />
											</button>
										)}
									</div>
								) : (
									<button
										type="button"
										disabled={readOnly}
										onClick={() => triggerAdditionalPhotoSelect(index)}
										className={[
											SLOT_ZONE_BASE,
											UPLOAD_ZONE_FOCUS,
											readOnly ? SLOT_ZONE_DISABLED : SLOT_ZONE_ENABLED,
										].join(" ")}
									>
										<span className={readOnly ? SLOT_ICON_WRAP_DISABLED : SLOT_ICON_WRAP}>
											<FaCloudUploadAlt className="h-5 w-5" aria-hidden />
										</span>
										<span
											className={[
												"text-center text-[11px] font-semibold",
												readOnly ? UPLOAD_TITLE_DISABLED : UPLOAD_TITLE_ENABLED,
											].join(" ")}
										>
											Добавить
										</span>
									</button>
								)}
							</li>
						)
					})}
				</ul>
			</section>
		</div>
	)
}

export default PhotosTab
