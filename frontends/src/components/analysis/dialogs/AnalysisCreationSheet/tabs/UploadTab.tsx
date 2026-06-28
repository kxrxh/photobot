import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import {
	type ChangeEvent,
	type ComponentProps,
	lazy,
	type ReactNode,
	type RefObject,
	Suspense,
	useEffect,
	useId,
	useMemo,
	useState,
} from "react"
import {
	FaCalendarAlt,
	FaCheckCircle,
	FaChevronDown,
	FaCloudUploadAlt,
	FaEdit,
	FaExclamationTriangle,
	FaInfoCircle,
	FaLayerGroup,
	FaLocationArrow,
	FaMap,
	FaMapMarkerAlt,
	FaPrescriptionBottle,
	FaTrash,
	FaWeightHanging,
} from "react-icons/fa"
import { deleteUserActiveClassification, setUserActiveClassification } from "@/api/classification"
import type { Classification } from "@/api/classification/types"
import { getAllProducts } from "@/api/product"
import { queryKeys } from "@/api/queryKeys"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { useAlert } from "@/hooks/useAlert"
import type { FractionType } from "@/hooks/useFractions"
import { useMessenger } from "@/hooks/useMessenger"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import ClassificationSelectionAlert from "../components/modal/ClassificationSelectionAlert"
import DeactivateClassificationAlert from "../components/modal/DeactivateClassificationAlert"

const LocationMapSheet = lazy(() => import("../components/modal/LocationMapSheet"))

const SECTION_LABEL_CLASS =
	"mb-1.5 block text-xs font-semibold uppercase tracking-wide text-base-content/45"

/** Positive numeric mass fields: allow empty or finite number > 0. */
function sanitizePositiveNumberInput(value: string, onChange: (value: string) => void): void {
	if (value === "" || (Number(value) > 0 && Number.isFinite(Number(value)))) {
		onChange(value)
	}
}

const UPLOAD_ZONE_BASE =
	"flex w-full flex-col items-center justify-center gap-2 rounded-2xl border-2 border-dashed px-4 py-10 transition-all sm:py-12"
const UPLOAD_ZONE_FOCUS =
	"focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
const UPLOAD_ZONE_DISABLED = "cursor-not-allowed border-base-300/60 bg-base-200/20 opacity-60"
const UPLOAD_ZONE_ENABLED =
	"cursor-pointer border-primary/35 bg-primary/5 hover:border-primary/50 hover:bg-primary/10 active:scale-[0.99]"
const UPLOAD_ZONE_PULSE = "animate-pulse"

const UPLOAD_ICON_WRAP_BASE = "flex h-14 w-14 items-center justify-center rounded-2xl"
const UPLOAD_ICON_WRAP_DISABLED = "bg-base-300/40 text-base-content/35"
const UPLOAD_ICON_WRAP_ENABLED = "bg-primary/15 text-primary"

const UPLOAD_TITLE_BASE = "text-sm font-semibold"
const UPLOAD_TITLE_DISABLED = "text-base-content/45"
const UPLOAD_TITLE_ENABLED = "text-base-content"

type MassInputMode = "mass_1000" | "mass"

interface UploadTabFractionContext {
	current?: FractionType
}

interface UploadTabProductSelection {
	selected: string
	onChange: (product: string) => void
}

interface UploadTabFileUpload {
	list: File[]
	isUploading: boolean
	onSelect: (e: ChangeEvent<HTMLInputElement>) => void
	onRemove: (index: number) => void
	inputRef: RefObject<HTMLInputElement | null>
}

interface UploadTabMassFields {
	mass1000: string
	mass: string
	massInputMode: MassInputMode
	onMassInputModeChange: (mode: MassInputMode) => void
	onMass1000Change: (value: string) => void
	onMassChange: (value: string) => void
	massLiter: string
	onMassLiterChange: (value: string) => void
}

interface UploadTabSampleMeta {
	location: string
	onLocationChange: (value: string) => void
	year: string
	onYearChange: (value: string) => void
}

interface UploadTabClassificationState {
	active?: Classification | null
	loading?: boolean
}

interface UploadTabProps {
	fraction?: UploadTabFractionContext
	product: UploadTabProductSelection
	files: UploadTabFileUpload
	massFields: UploadTabMassFields
	sampleMeta: UploadTabSampleMeta
	classification?: UploadTabClassificationState
}

type SectionLabelProps = { children: ReactNode } & (
	| (Pick<ComponentProps<"label">, "htmlFor"> & { htmlFor: string })
	| { htmlFor?: undefined }
)

function SectionLabel(props: SectionLabelProps) {
	if ("htmlFor" in props && props.htmlFor) {
		const { children, htmlFor } = props
		return (
			<label className={SECTION_LABEL_CLASS} htmlFor={htmlFor}>
				{children}
			</label>
		)
	}
	return <div className={SECTION_LABEL_CLASS}>{props.children}</div>
}

export default function UploadTab({
	fraction,
	product,
	files,
	massFields,
	sampleMeta,
	classification,
}: UploadTabProps) {
	const currentFraction = fraction?.current
	const selectedProduct = product.selected
	const onProductChange = product.onChange
	const uploadedFiles = files.list
	const isUploading = files.isUploading
	const onFileSelect = files.onSelect
	const onRemoveFile = files.onRemove
	const fileInputRef = files.inputRef
	const {
		mass1000,
		mass,
		massInputMode,
		onMassInputModeChange,
		onMass1000Change,
		onMassChange,
		massLiter,
		onMassLiterChange,
	} = massFields
	const { location, onLocationChange, year, onYearChange } = sampleMeta
	const activeClassification = classification?.active
	const loadingActiveClassification = classification?.loading
	const selectId = useId()
	const massId = useId()
	const massSampleId = useId()
	const locationId = useId()
	const yearId = useId()
	const massLiterId = useId()
	const { showSuccess, showError } = useAlert()
	const { location: tgLocation } = useMessenger()
	const queryClient = useQueryClient()
	const [isClassificationModalOpen, setIsClassificationModalOpen] = useState(false)
	const [isDeactivateModalOpen, setIsDeactivateModalOpen] = useState(false)
	const [isMapOpen, setIsMapOpen] = useState(false)
	const [isMetadataExpanded, setIsMetadataExpanded] = useState(false)
	const isDev = import.meta.env.DEV
	const uploadBlockedWithoutProduct = !isDev && !selectedProduct

	const currentCalendarYear = new Date().getFullYear()
	const yearSelectOptions = useMemo(() => {
		return Array.from({ length: 21 }, (_, i) => {
			const y = currentCalendarYear - 20 + i
			return y
		})
	}, [currentCalendarYear])

	const previewEntries = useMemo(
		() => uploadedFiles.map((file) => ({ file, url: URL.createObjectURL(file) })),
		[uploadedFiles]
	)

	useEffect(() => {
		return () => {
			for (const { url } of previewEntries) {
				URL.revokeObjectURL(url)
			}
		}
	}, [previewEntries])

	const {
		data: productsData,
		isLoading: loadingProducts,
		error: productsError,
	} = useQuery({
		queryKey: queryKeys.products,
		queryFn: getAllProducts,
		staleTime: 10 * 60_000,
	})

	const productOptions = useMemo(
		() =>
			(productsData ?? [])
				.map((productItem) => ({ value: productItem.name, label: productItem.name }))
				.sort((left, right) => left.label.localeCompare(right.label, "ru")),
		[productsData]
	)

	const setActiveClassificationMutation = useMutation({
		mutationFn: (classificationId: string) => setUserActiveClassification(classificationId),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: queryKeys.userActiveClassification,
			})
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Активная классификация изменена")
			setIsClassificationModalOpen(false)
		},
		onError: () => {
			showError("Ошибка при изменении активной классификации")
		},
	})

	const removeActiveClassificationMutation = useMutation({
		mutationFn: () => deleteUserActiveClassification(),
		onSuccess: () => {
			queryClient.invalidateQueries({
				queryKey: queryKeys.userActiveClassification,
			})
			queryClient.invalidateQueries({ queryKey: queryKeys.classifications.all })
			showSuccess("Активная классификация деактивирована")
			setIsDeactivateModalOpen(false)
		},
		onError: () => {
			showError("Ошибка при деактивации активной классификации")
		},
	})

	const handleOpenModal = () => {
		setIsClassificationModalOpen(true)
	}

	const handleCloseModal = () => {
		setIsClassificationModalOpen(false)
	}

	const handleSetActiveClassification = (classificationId: string) => {
		setActiveClassificationMutation.mutate(classificationId)
	}

	const handleRemoveActiveClassification = () => {
		setIsDeactivateModalOpen(true)
	}

	const handleConfirmDeactivate = () => {
		removeActiveClassificationMutation.mutate()
	}

	const handleCancelDeactivate = () => {
		setIsDeactivateModalOpen(false)
	}

	const handleGetTelegramLocation = async () => {
		try {
			if (!tgLocation.isSupported()) {
				showError("Получение геолокации не поддерживается в этой версии Telegram")
				return
			}

			const result = await tgLocation.requestLocation()
			if (result.ok) {
				const locationString = `${result.latitude.toFixed(6)}, ${result.longitude.toFixed(6)}`
				onLocationChange(locationString)
				showSuccess("Местоположение получено")
			} else {
				let msg: string
				switch (result.reason) {
					case "denied":
						msg =
							"Доступ к геолокации отклонён. Разрешите доступ в настройках или укажите точку на карте."
						break
					case "timeout":
						msg = "Не удалось определить местоположение: время ожидания истекло."
						break
					case "unavailable":
						msg = "Геолокация недоступна в этой среде."
						break
					default:
						msg = "Не удалось получить местоположение."
				}
				showError(msg)
			}
		} catch (err) {
			log.devError("Error getting location:", err)
			showError(getUserFacingErrorMessage(err))
		}
	}

	const productMismatch =
		activeClassification?.product?.name &&
		selectedProduct &&
		activeClassification.product.name !== selectedProduct

	const hasMetadataSummary = Boolean(mass1000 || mass || location || massLiter || year)

	return (
		<>
			<div className="mx-auto flex max-w-2xl flex-col gap-8 pb-2">
				<section aria-label="Классификация">
					<SectionLabel>Классификация</SectionLabel>
					{loadingActiveClassification ? (
						<div className="skeleton h-20 w-full rounded-2xl" />
					) : (
						<div className="rounded-2xl border border-base-200 bg-base-200/25 p-4">
							{activeClassification ? (
								<div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
									<div className="flex min-w-0 gap-3">
										<span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-success/15 text-success">
											<FaCheckCircle className="h-5 w-5" aria-hidden />
										</span>
										<div className="min-w-0">
											<p className="text-sm font-semibold text-base-content">
												{activeClassification.name}
											</p>
											<p className="mt-0.5 text-xs text-base-content/55">
												Продукт классификации:{" "}
												<span className="font-medium text-base-content/80">
													{activeClassification.product?.name || "—"}
												</span>
											</p>
										</div>
									</div>
									<div className="flex shrink-0 gap-2 sm:flex-col sm:items-stretch">
										<Button
											type="button"
											onClick={handleOpenModal}
											variant="primary"
											size="sm"
											disabled={isUploading}
											startIcon={<FaEdit className="h-3.5 w-3.5" />}
											className="flex-1 sm:flex-none"
										>
											Сменить
										</Button>
										<Button
											type="button"
											onClick={handleRemoveActiveClassification}
											variant="ghost"
											size="sm"
											disabled={isUploading || removeActiveClassificationMutation.isPending}
											startIcon={<FaTrash className="h-3.5 w-3.5" />}
											className="flex-1 text-error sm:flex-none"
										>
											Отключить
										</Button>
									</div>
								</div>
							) : (
								<div className="flex flex-col gap-2">
									<div className="flex items-center gap-2.5">
										<span
											className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border border-dashed border-base-300/70 bg-base-100/50 text-base-content/40"
											aria-hidden
										>
											<FaLayerGroup className="h-3.5 w-3.5" />
										</span>
										<p className="min-w-0 flex-1 text-xs font-semibold text-base-content">
											Активная классификация не выбрана
										</p>
									</div>
									<button
										type="button"
										onClick={handleOpenModal}
										className="btn btn-primary w-full"
										disabled={isUploading}
									>
										Выбрать классификацию
									</button>
								</div>
							)}

							{currentFraction ? (
								<div className="mt-3 flex items-center gap-2 border-t border-base-200/80 pt-3">
									<span className="h-1.5 w-1.5 shrink-0 animate-pulse rounded-full bg-primary" />
									<p className="text-xs text-base-content/70">
										Разметка: объекты попадут во фракцию{" "}
										<strong className="text-base-content">{currentFraction.name}</strong>
									</p>
								</div>
							) : null}

							{productMismatch ? (
								<div className="mt-3 flex gap-2 rounded-xl border border-warning/30 bg-warning/10 px-3 py-2 text-xs text-base-content/85">
									<FaExclamationTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0 text-warning" />
									<span>
										<strong className="font-semibold">Разные продукты:</strong> в классификации — «
										{activeClassification?.product?.name}», в поле ниже — «{selectedProduct}».
									</span>
								</div>
							) : null}
						</div>
					)}
				</section>

				<section aria-label="Продукт">
					<SectionLabel htmlFor={selectId}>Продукт для анализа</SectionLabel>
					{loadingProducts ? (
						<div className="skeleton h-12 w-full rounded-xl" />
					) : productsError ? (
						<div
							role="alert"
							className="flex gap-3 rounded-xl border border-error/20 bg-error/5 px-3 py-3"
						>
							<FaExclamationTriangle className="mt-0.5 h-4 w-4 shrink-0 text-error" />
							<div className="min-w-0 text-sm">
								<p className="font-medium text-base-content">Не удалось загрузить список</p>
								<p className="mt-1 text-xs text-base-content/70">
									{getUserFacingErrorMessage(productsError)}
								</p>
							</div>
						</div>
					) : (
						<ModalSelect
							id={selectId}
							title="Продукт для анализа"
							placeholder="Выберите продукт"
							options={productOptions}
							value={selectedProduct}
							onChange={onProductChange}
							disabled={isUploading}
							clearable={false}
						/>
					)}
				</section>

				<section aria-label="Изображения">
					<SectionLabel>Фотографии</SectionLabel>
					<button
						type="button"
						disabled={uploadBlockedWithoutProduct || isUploading}
						onClick={() => fileInputRef.current?.click()}
						className={[
							UPLOAD_ZONE_BASE,
							UPLOAD_ZONE_FOCUS,
							uploadBlockedWithoutProduct || isUploading
								? UPLOAD_ZONE_DISABLED
								: UPLOAD_ZONE_ENABLED,
							isUploading ? UPLOAD_ZONE_PULSE : "",
						]
							.filter(Boolean)
							.join(" ")}
					>
						<input
							ref={fileInputRef}
							type="file"
							accept="image/*,.jpg,.jpeg,.png,.heic,.heif"
							multiple
							onChange={onFileSelect}
							className="hidden"
						/>
						<span
							className={[
								UPLOAD_ICON_WRAP_BASE,
								uploadBlockedWithoutProduct ? UPLOAD_ICON_WRAP_DISABLED : UPLOAD_ICON_WRAP_ENABLED,
							].join(" ")}
						>
							<FaCloudUploadAlt className="h-7 w-7" aria-hidden />
						</span>
						<div className="text-center">
							<p
								className={[
									UPLOAD_TITLE_BASE,
									uploadBlockedWithoutProduct ? UPLOAD_TITLE_DISABLED : UPLOAD_TITLE_ENABLED,
								].join(" ")}
							>
								{uploadBlockedWithoutProduct
									? "Сначала выберите продукт"
									: "Нажмите, чтобы добавить файлы"}
							</p>
							<p className="mt-1 max-w-xs text-xs text-base-content/55">
								{uploadBlockedWithoutProduct
									? "После выбора продукта можно загрузить снимки."
									: "JPG, PNG, HEIC/HEIF · несколько файлов · до 10 МБ каждый"}
							</p>
						</div>
					</button>

					{previewEntries.length > 0 ? (
						<ul className="mt-4 grid grid-cols-2 gap-2 sm:grid-cols-3 sm:gap-3">
							{previewEntries.map(({ file, url }, index) => (
								<li
									key={`${file.name}-${file.size}-${file.lastModified}`}
									className="group relative overflow-hidden rounded-xl border border-base-200 bg-base-100 shadow-sm"
								>
									<div className="aspect-square bg-base-200">
										<img
											src={url}
											alt={file.name}
											className="h-full w-full object-cover"
											loading="lazy"
											decoding="async"
										/>
									</div>
									<div className="border-t border-base-200 bg-base-100/95 p-2 backdrop-blur-sm">
										<p className="truncate text-xs font-medium" title={file.name}>
											{file.name}
										</p>
										<p className="text-[10px] text-base-content/50">
											{(file.size / 1024 / 1024).toFixed(1)} МБ
										</p>
									</div>
									<button
										type="button"
										onClick={() => onRemoveFile(index)}
										disabled={isUploading}
										className="absolute right-1.5 top-1.5 flex h-8 w-8 items-center justify-center rounded-full bg-base-100/90 text-error shadow-md backdrop-blur-sm transition-opacity hover:bg-error/10"
										aria-label={`Удалить ${file.name}`}
									>
										<FaTrash className="h-3.5 w-3.5" />
									</button>
								</li>
							))}
						</ul>
					) : null}
				</section>

				<section aria-label="Дополнительные данные пробы">
					<div className="overflow-hidden rounded-2xl border border-base-200 bg-base-200/20">
						<button
							type="button"
							onClick={() => setIsMetadataExpanded(!isMetadataExpanded)}
							className="flex w-full items-center justify-between gap-3 px-4 py-3 text-left transition-colors hover:bg-base-200/40"
						>
							<div className="flex min-w-0 items-center gap-2">
								<FaInfoCircle className="h-4 w-4 shrink-0 text-base-content/40" aria-hidden />
								<span className="text-sm font-semibold text-base-content/80">
									Параметры пробы{" "}
									<span className="font-normal text-base-content/45">(необязательно)</span>
								</span>
							</div>
							<FaChevronDown
								className={`h-4 w-4 shrink-0 text-base-content/40 transition-transform duration-200 ${
									isMetadataExpanded ? "rotate-180" : ""
								}`}
								aria-hidden
							/>
						</button>
						{!isMetadataExpanded && hasMetadataSummary ? (
							<div className="flex flex-wrap gap-1.5 border-t border-base-200/60 px-4 py-2">
								{year ? (
									<span className="badge badge-sm badge-ghost gap-1 border border-base-300/60">
										<FaCalendarAlt className="h-3 w-3" />
										{year}
									</span>
								) : null}
								{mass1000 ? (
									<span className="badge badge-sm badge-ghost gap-1 border border-base-300/60">
										<FaWeightHanging className="h-3 w-3" />
										1000: {mass1000} г
									</span>
								) : null}
								{mass ? (
									<span className="badge badge-sm badge-ghost gap-1 border border-base-300/60">
										<FaWeightHanging className="h-3 w-3" />
										обр.: {mass} г
									</span>
								) : null}
								{massLiter ? (
									<span className="badge badge-sm badge-ghost gap-1 border border-base-300/60">
										<FaPrescriptionBottle className="h-3 w-3" />1 л: {massLiter} г
									</span>
								) : null}
								{location ? (
									<span className="badge badge-sm badge-ghost max-w-[140px] justify-start gap-1 truncate border border-base-300/60">
										<FaMapMarkerAlt className="h-3 w-3 shrink-0" />
										<span className="truncate">{location}</span>
									</span>
								) : null}
							</div>
						) : null}
						<div
							className={`grid transition-all duration-200 ease-out ${
								isMetadataExpanded ? "grid-rows-[1fr] opacity-100" : "grid-rows-[0fr] opacity-0"
							}`}
						>
							<div className="overflow-hidden">
								<div className="space-y-4 border-t border-base-200/60 px-4 pb-4 pt-3">
									<div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
										<div className="space-y-2">
											<div className="flex items-center justify-between gap-2">
												<label
													className="text-xs font-medium text-base-content/70"
													htmlFor={massInputMode === "mass_1000" ? massId : massSampleId}
												>
													{massInputMode === "mass_1000"
														? "Масса 1000 семян (г)"
														: "Масса образца (г)"}
												</label>
												<div className="join rounded-lg shadow-sm">
													<button
														type="button"
														className={`btn btn-xs join-item px-2 ${
															massInputMode === "mass_1000" ? "btn-primary" : "btn-ghost"
														}`}
														onClick={() => onMassInputModeChange("mass_1000")}
														disabled={isUploading}
													>
														1000 с.
													</button>
													<button
														type="button"
														className={`btn btn-xs join-item px-2 ${
															massInputMode === "mass" ? "btn-primary" : "btn-ghost"
														}`}
														onClick={() => onMassInputModeChange("mass")}
														disabled={isUploading}
													>
														Образец
													</button>
												</div>
											</div>
											<Input
												id={massInputMode === "mass_1000" ? massId : massSampleId}
												type="number"
												step="0.000001"
												min="0.000001"
												value={massInputMode === "mass_1000" ? mass1000 : mass}
												onChange={(e) => {
													const val = e.target.value
													sanitizePositiveNumberInput(
														val,
														massInputMode === "mass_1000" ? onMass1000Change : onMassChange
													)
												}}
												placeholder="Например: 40.5"
												disabled={isUploading}
												size="sm"
											/>
										</div>
										<div className="space-y-2">
											<label className="text-xs font-medium text-base-content/70" htmlFor={yearId}>
												Год
											</label>
											<ModalSelect
												id={yearId}
												title="Год"
												placeholder="Год"
												options={yearSelectOptions.map((y) => ({
													value: String(y),
													label: String(y),
												}))}
												value={year}
												onChange={onYearChange}
												disabled={isUploading}
												clearable={false}
												size="sm"
											/>
										</div>
										<div className="space-y-2 sm:col-span-2">
											<label
												className="text-xs font-medium text-base-content/70"
												htmlFor={locationId}
											>
												Местоположение
											</label>
											<div className="flex flex-wrap gap-2">
												<div className="relative min-w-48 flex-1">
													<div
														id={locationId}
														className={`input input-bordered flex h-10 min-h-10 w-full items-center px-3 text-sm ${
															!location ? "text-base-content/40" : ""
														} ${isUploading ? "opacity-50" : ""}`}
													>
														<span className="truncate">{location || "Не задано"}</span>
													</div>
													{location ? (
														<button
															type="button"
															onClick={(e) => {
																e.stopPropagation()
																onLocationChange("")
															}}
															className="absolute inset-y-0 right-2 flex items-center text-base-content/40 hover:text-error"
															title="Очистить"
															disabled={isUploading}
														>
															<FaTrash className="h-3.5 w-3.5" />
														</button>
													) : null}
												</div>
												<Button
													type="button"
													onClick={() => setIsMapOpen(true)}
													variant="primary"
													size="sm"
													className="btn-circle h-10 w-10 shrink-0 p-0"
													disabled={isUploading}
													aria-label="Выбрать местоположение на карте"
													title="Карта"
												>
													<FaMap className="h-4 w-4" aria-hidden />
												</Button>
												{tgLocation.isSupported() ? (
													<Button
														type="button"
														onClick={handleGetTelegramLocation}
														variant="primary"
														size="sm"
														className="h-10 shrink-0 rounded-xl px-3"
														disabled={isUploading}
													>
														<FaLocationArrow className="mr-1.5 h-3.5 w-3.5" />
														GPS
													</Button>
												) : null}
											</div>
										</div>
										<div className="space-y-2">
											<label
												className="text-xs font-medium text-base-content/70"
												htmlFor={massLiterId}
											>
												Масса 1 л (г)
											</label>
											<Input
												id={massLiterId}
												type="text"
												value={massLiter}
												onChange={(e) => onMassLiterChange(e.target.value)}
												placeholder="Например: 40.5"
												disabled={isUploading}
												size="sm"
											/>
										</div>
									</div>
								</div>
							</div>
						</div>
					</div>
				</section>
			</div>

			<ClassificationSelectionAlert
				isOpen={isClassificationModalOpen}
				onClose={handleCloseModal}
				onConfirm={handleSetActiveClassification}
				activeClassification={activeClassification}
				isPending={setActiveClassificationMutation.isPending}
			/>

			<DeactivateClassificationAlert
				isOpen={isDeactivateModalOpen}
				onClose={handleCancelDeactivate}
				onConfirm={handleConfirmDeactivate}
				isPending={removeActiveClassificationMutation.isPending}
			/>

			{isMapOpen ? (
				<Suspense
					fallback={
						<div className="fixed inset-0 z-100 flex items-center justify-center bg-black/50 backdrop-blur-sm">
							<span className="loading loading-spinner loading-lg text-primary" />
						</div>
					}
				>
					<LocationMapSheet
						isOpen={isMapOpen}
						onClose={() => setIsMapOpen(false)}
						onConfirm={(coords) => onLocationChange(coords)}
						initialLocation={location}
					/>
				</Suspense>
			) : null}
		</>
	)
}
