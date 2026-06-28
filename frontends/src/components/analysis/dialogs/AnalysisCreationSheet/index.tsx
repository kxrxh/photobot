import { useQuery, useQueryClient } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import type React from "react"
import { useCallback, useEffect, useRef, useState } from "react"
import { FaImage, FaList } from "react-icons/fa"

import { createAnalysis, getObjectsByRequestId, getRequests } from "@/api/analysis"
import type { AnalysisRequest } from "@/api/analysis/types"
import { getUserActiveClassification } from "@/api/classification"
import { CACHE } from "@/api/queryConfig"
import { queryKeys } from "@/api/queryKeys"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import { useAuth } from "@/contexts/AuthContext"
import { useAlert } from "@/hooks/useAlert"
import type { FractionType } from "@/hooks/useFractions"
import { useMessenger } from "@/hooks/useMessenger"
import { type RequestUpdatePayload, useAnalysisWebSocket } from "@/hooks/websocket"
import { downloadAnalysisReport } from "@/utils/analysis"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import { resolveChatPlatform } from "@/utils/messengerPlatform"

import RequestsTab from "./tabs/RequestsTab"
import ResultsTab from "./tabs/ResultsTab"
import UploadTab from "./tabs/UploadTab"
import { useOpenRequestDeepLink } from "./useOpenRequestDeepLink"

interface AnalysisCreationSheetProps {
	isOpen: boolean
	currentFraction?: FractionType
	initialTab?: "upload" | "requests"
	initialRequestId?: string
}

type MassInputMode = "mass_1000" | "mass"

/** Coalesce rapid WebSocket-driven invalidations so each event does not trigger its own refetch. */
const REQUESTS_INVALIDATE_DEBOUNCE_MS = 600

/** Non-empty placeholder when `import.meta.env.DEV` and no product chosen (API rejects empty product). */
const DEV_ANALYSIS_PRODUCT_PLACEHOLDER = "dev"

function formatSentBytes(n: number): string {
	if (n < 1024) return `${n} Б`
	if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} КБ`
	return `${(n / (1024 * 1024)).toFixed(1)} МБ`
}

/** Backend only pushes WS on terminal-ish transitions; created→processing has no event — poll until settled. */
function requestsListRefetchIntervalMs(
	data: { requests: AnalysisRequest[] } | undefined
): number | false {
	if (!data?.requests?.length) return false
	const hasInFlight = data.requests.some(
		(r) =>
			r.status === "created" || r.status === "processing" || r.status === "waiting_for_confirmation"
	)
	return hasInFlight ? 6000 : false
}

const AnalysisCreationSheet: React.FC<AnalysisCreationSheetProps> = ({
	isOpen,
	currentFraction,
	initialTab,
	initialRequestId,
}) => {
	const navigate = useNavigate()
	const fileInputRef = useRef<HTMLInputElement>(null)
	const scrollContainerRef = useRef<HTMLDivElement>(null)
	const { showSuccess, showError } = useAlert()
	const { userId } = useAuth()
	const queryClient = useQueryClient()
	const { requestFileDownload, platform } = useMessenger()
	const requestsInvalidateTimerRef = useRef<number | null>(null)
	const backNavigationCooldownRef = useRef<number | null>(null)

	const scheduleRequestsListInvalidate = useCallback(() => {
		if (requestsInvalidateTimerRef.current) {
			clearTimeout(requestsInvalidateTimerRef.current)
		}
		requestsInvalidateTimerRef.current = window.setTimeout(() => {
			requestsInvalidateTimerRef.current = null
			void queryClient.invalidateQueries({ queryKey: queryKeys.requests.all })
		}, REQUESTS_INVALIDATE_DEBOUNCE_MS)
	}, [queryClient])

	useEffect(() => {
		return () => {
			if (requestsInvalidateTimerRef.current) {
				clearTimeout(requestsInvalidateTimerRef.current)
				requestsInvalidateTimerRef.current = null
			}
			if (backNavigationCooldownRef.current) {
				clearTimeout(backNavigationCooldownRef.current)
				backNavigationCooldownRef.current = null
			}
		}
	}, [])

	const [uploadedFiles, setUploadedFiles] = useState<File[]>([])
	const [isUploading, setIsUploading] = useState(false)
	const [uploadProgress, setUploadProgress] = useState<{
		loaded: number
		total?: number
	} | null>(null)
	const [selectedProduct, setSelectedProduct] = useState<string>("")
	const [mass1000, setMass1000] = useState<string>("")
	const [mass, setMass] = useState<string>("")
	const [massInputMode, setMassInputMode] = useState<MassInputMode>("mass_1000")
	const [location, setLocation] = useState<string>("")
	const [year, setYear] = useState<string>(new Date().getFullYear().toString())
	const [massLiter, setMassLiter] = useState<string>("")
	const [activeTab, setActiveTab] = useState<"upload" | "requests" | "results">("upload")
	const [statusFilter, setStatusFilter] = useState<string>("")
	const [selectedRequest, setSelectedRequest] = useState<AnalysisRequest | null>(null)
	const [isNavigatingBack, setIsNavigatingBack] = useState(false)
	const [excludedObjects, setExcludedObjects] = useState<string[]>([])

	const handleMass1000Change = (value: string) => {
		setMass1000(value)
	}

	const handleMassChange = (value: string) => {
		setMass(value)
	}

	const { data: activeClassificationData, isLoading: loadingActiveClassification } = useQuery({
		queryKey: queryKeys.userActiveClassification,
		queryFn: () => getUserActiveClassification(),
		enabled: isOpen && !!userId,
		retry: false,
	})

	const activeClassification = activeClassificationData?.classification

	useEffect(() => {
		if (activeClassification?.product?.name && !selectedProduct && activeTab === "upload") {
			setSelectedProduct(activeClassification.product.name)
		}
	}, [activeClassification, selectedProduct, activeTab])

	const scrollToTop = useCallback(() => {
		if (scrollContainerRef.current) {
			scrollContainerRef.current.scrollTop = 0
		}
	}, [])

	useEffect(() => {
		if (!isOpen) return
		setActiveTab(initialTab ?? "upload")
		const scrollTid = window.setTimeout(() => {
			scrollToTop()
		}, 0)
		return () => clearTimeout(scrollTid)
	}, [isOpen, initialTab, scrollToTop])

	const {
		data: requestsData,
		isLoading: loadingRequests,
		refetch: refetchRequests,
	} = useQuery({
		queryKey: queryKeys.requests.list(statusFilter),
		queryFn: async () => {
			if (!userId) {
				throw new Error("User not authenticated")
			}
			return await getRequests({
				status: statusFilter || undefined,
			})
		},
		enabled: isOpen && !!userId,
		staleTime: CACHE.staleTimeShort,
		refetchOnWindowFocus: false,
		refetchInterval: (q) => requestsListRefetchIntervalMs(q.state.data),
	})

	const openConfirmationFromDeepLink = useCallback((request: AnalysisRequest) => {
		setSelectedRequest(request)
		setActiveTab("results")
	}, [])

	useOpenRequestDeepLink({
		isOpen,
		openRequestId: initialRequestId,
		userId,
		requestsList: requestsData,
		scrollToTop,
		onOpenConfirmation: openConfirmationFromDeepLink,
	})

	const { data: objects, isLoading: loadingAnalysis } = useQuery({
		queryKey: queryKeys.requestObjects(selectedRequest?.id ?? ""),
		queryFn: async () => {
			if (!selectedRequest?.id) {
				throw new Error("No analysis selected")
			}
			return await getObjectsByRequestId(selectedRequest.id)
		},
		enabled: isOpen && activeTab === "results" && !!selectedRequest?.id,
		staleTime: 15_000,
		refetchOnWindowFocus: true,
	})

	const handleRefreshRequests = async () => {
		try {
			await refetchRequests()
			showSuccess("Список запросов обновлен", undefined, "bottom")
		} catch (err) {
			showError(getUserFacingErrorMessage(err), undefined, "bottom")
		}
	}

	const onAnalysisRequestUpdate = useCallback(
		({ requestId, data }: RequestUpdatePayload) => {
			queryClient.setQueriesData({ queryKey: queryKeys.requests.all }, (prev) => {
				if (!prev || typeof prev !== "object" || !("requests" in prev)) return prev
				const p = prev as { requests: AnalysisRequest[]; total: number }
				const idx = p.requests.findIndex((r) => r.id === requestId)
				if (idx === -1) return prev

				const cur = p.requests[idx]
				const status = data.status as AnalysisRequest["status"]
				const temp_id = data.temp_id !== undefined ? data.temp_id : cur.temp_id
				const error_message =
					data.error_message !== undefined ? data.error_message : cur.error_message
				if (
					cur.status === status &&
					cur.temp_id === temp_id &&
					cur.error_message === error_message
				) {
					return prev
				}

				const next = [...p.requests]
				next[idx] = { ...cur, status, temp_id, error_message }
				return { ...p, requests: next }
			})
			scheduleRequestsListInvalidate()
		},
		[queryClient, scheduleRequestsListInvalidate]
	)

	useAnalysisWebSocket({
		userId: userId != null ? userId.toString() : "",
		enabled: isOpen && !!userId,
		onRequestUpdate: onAnalysisRequestUpdate,
	})

	useEffect(() => {
		if (!isOpen) {
			setUploadedFiles([])
			setStatusFilter("")
			setSelectedRequest(null)
			setActiveTab("upload")
			setIsNavigatingBack(false)
			setExcludedObjects([])
			setMass1000("")
			setMass("")
			setMassInputMode("mass_1000")
			setLocation("")
			setYear(new Date().getFullYear().toString())
			setMassLiter("")
		}
	}, [isOpen])

	useEffect(() => {
		if (!isOpen) return

		const onVisible = () => {
			if (document.visibilityState !== "visible") return
			void queryClient.invalidateQueries({ queryKey: queryKeys.requests.all })
		}

		document.addEventListener("visibilitychange", onVisible)
		return () => document.removeEventListener("visibilitychange", onVisible)
	}, [isOpen, queryClient])

	const isAllowedImage = (file: File): boolean => {
		const mime = (file.type || "").toLowerCase()
		const name = (file.name || "").toLowerCase()

		const allowedMimes = new Set([
			"image/jpeg",
			"image/jpg",
			"image/png",
			"image/x-heic",
			"image/x-heif",
			"image/heic",
			"image/heif",
			"image/heic-sequence",
			"image/heif-sequence",
		])

		if (allowedMimes.has(mime)) return true

		// Some browsers/devices (notably Android) may provide empty/unknown MIME for HEIC.
		const allowedExts = [".jpg", ".jpeg", ".png", ".heic", ".heif"]
		return allowedExts.some((ext) => name.endsWith(ext))
	}

	const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
		const files = Array.from(e.target.files || []).filter(isAllowedImage)

		if (files.length === 0) {
			showError("Пожалуйста, выберите только JPG, PNG, HEIC или HEIF файлы", undefined, "bottom")
			e.target.value = ""
			return
		}

		setUploadedFiles((prev) => [...prev, ...files])
		e.target.value = ""
	}

	const removeFile = (index: number) => {
		setUploadedFiles((prev) => prev.filter((_, i) => i !== index))
	}

	const processImages = async () => {
		if (uploadedFiles.length === 0) return

		const trimmedProduct = selectedProduct.trim()
		const productForRequest =
			trimmedProduct || (import.meta.env.DEV ? DEV_ANALYSIS_PRODUCT_PLACEHOLDER : "")
		if (!productForRequest) {
			showError("Пожалуйста, выберите продукт для анализа", undefined, "bottom")
			return
		}

		setIsUploading(true)
		setUploadProgress(null)

		try {
			if (!userId) {
				throw new Error("User not authenticated")
			}

			const sentCount = uploadedFiles.length
			await createAnalysis(productForRequest, uploadedFiles, {
				mass_1000: massInputMode === "mass_1000" ? mass1000 || undefined : undefined,
				mass: massInputMode === "mass" ? mass || undefined : undefined,
				location: location || undefined,
				year: year || undefined,
				mass_liter: massLiter || undefined,
				onUploadProgress: ({ loaded, total }) => {
					setUploadProgress({ loaded, total })
				},
			})

			// Reset upload UI before invalidate/navigate — otherwise 100% sticks until those finish.
			setIsUploading(false)
			setUploadProgress(null)
			setUploadedFiles([])

			showSuccess(`Отправлено ${sentCount} изображений для анализа`, undefined, "bottom")

			await queryClient.invalidateQueries({ queryKey: queryKeys.requests.all })
			navigate({
				to: "/analysis/create",
				search: { tab: "requests", openRequest: undefined },
				replace: true,
			})
		} catch (error) {
			log.error("Failed to send analysis request:", error)
			showError(getUserFacingErrorMessage(error), undefined, "bottom")
		} finally {
			setIsUploading(false)
			setUploadProgress(null)
		}
	}

	const handleViewResults = (request: AnalysisRequest) => {
		setSelectedRequest(request)
		setActiveTab("results")
		scrollToTop()
	}

	const handleDownloadPdf = async (request: AnalysisRequest) => {
		if (!request.temp_id) {
			log.error("Error downloading PDF report: temp_id is missing")
			return
		}

		try {
			const chatPlatform = resolveChatPlatform(platform)
			const { sendToChatOk } = await downloadAnalysisReport(
				request.temp_id,
				requestFileDownload,
				chatPlatform ? { platform: chatPlatform } : undefined
			)
			showSuccess(sendToChatOk ? "Отчёт скачан. Отправлен в чат." : "Отчёт скачан.")
		} catch (error) {
			log.error("Error downloading PDF report:", error)
			showError(getUserFacingErrorMessage(error))
		}
	}

	const handleBackToRequests = () => {
		if (isNavigatingBack) return // Prevent double-clicking

		setIsNavigatingBack(true)
		setActiveTab("requests")
		setExcludedObjects([]) // Reset excluded objects when going back
		setSelectedRequest(null) // Reset selected request when going back
		scrollToTop()

		if (backNavigationCooldownRef.current) {
			clearTimeout(backNavigationCooldownRef.current)
		}
		backNavigationCooldownRef.current = window.setTimeout(() => {
			backNavigationCooldownRef.current = null
			setIsNavigatingBack(false)
		}, 300)
	}

	const handleExcludeObject = (objectId: string) => {
		if (excludedObjects.includes(objectId)) return
		setExcludedObjects((prev) => [...prev, objectId])
	}

	const handleIncludeObject = (objectId: string) => {
		setExcludedObjects((prev) => prev.filter((id) => id !== objectId))
	}

	const handleResetSelection = () => {
		setExcludedObjects([])
	}

	const handleInvertSelection = () => {
		if (!objects) return
		const allObjectIds = objects.map((obj) => obj.id)
		const invertedExclusions = allObjectIds.filter((id) => !excludedObjects.includes(id))
		setExcludedObjects(invertedExclusions)
	}

	if (!isOpen) return null

	const uploadPercent =
		isUploading && uploadProgress && uploadProgress.total !== undefined && uploadProgress.total > 0
			? Math.min(100, Math.round((100 * uploadProgress.loaded) / uploadProgress.total))
			: null

	const showSendDock = activeTab === "upload" && uploadedFiles.length > 0

	const handleHeaderClose = () => {
		if (activeTab === "results") {
			handleBackToRequests()
			return
		}
		void navigate({ to: "/menu" })
	}

	return (
		<div className="flex min-h-0 flex-1 flex-col bg-base-100">
			<header className="sticky top-0 z-10 shrink-0 border-b border-base-200 bg-base-100 pt-[env(safe-area-inset-top)]">
				<div className="flex min-h-12 items-center justify-between gap-3 px-3 py-2 sm:px-4">
					<div className="min-w-0 flex-1">
						<h1 className="truncate text-lg font-bold leading-snug text-base-content sm:text-xl">
							{activeTab === "results" ? "Пересчет анализа" : "Загрузка из фото"}
						</h1>
					</div>
					<SheetHeaderCloseButton
						onClick={handleHeaderClose}
						disabled={activeTab === "results" && isNavigatingBack}
						aria-label={activeTab === "results" ? "Назад к запросам" : "Закрыть"}
						title={activeTab === "results" ? "Назад к запросам" : "В меню"}
					/>
				</div>
			</header>

			<div className="shrink-0 border-b border-base-200 bg-base-100 px-2 py-1.5 sm:px-3">
				<div
					className="mx-auto w-full max-w-2xl"
					role="tablist"
					aria-label="Разделы загрузки анализа"
				>
					<div className="grid grid-cols-2 gap-0.5">
						<button
							type="button"
							role="tab"
							aria-selected={activeTab === "upload"}
							className={`flex min-h-9 cursor-pointer flex-col items-center justify-center gap-0 rounded-lg px-0.5 py-1 text-[10px] transition-colors duration-200 sm:min-h-10 sm:gap-0.5 sm:py-1.5 sm:text-[11px] ${
								activeTab === "upload"
									? "bg-primary/10 font-semibold text-primary"
									: "text-base-content/60 active:bg-base-200/80 sm:hover:bg-base-200/50"
							}`}
							onClick={() => {
								setActiveTab("upload")
								if (activeTab === "results") {
									setSelectedRequest(null)
									setExcludedObjects([])
								}
								scrollToTop()
							}}
						>
							<span className="text-sm leading-none sm:text-base" aria-hidden>
								<FaImage />
							</span>
							<span className="max-w-full truncate px-0.5 text-center font-medium leading-tight">
								<span className="hidden sm:inline">Загрузить фото</span>
								<span className="sm:hidden">Фото</span>
							</span>
						</button>
						<button
							type="button"
							role="tab"
							aria-selected={activeTab === "requests"}
							className={`flex min-h-9 cursor-pointer flex-col items-center justify-center gap-0 rounded-lg px-0.5 py-1 text-[10px] transition-colors duration-200 sm:min-h-10 sm:gap-0.5 sm:py-1.5 sm:text-[11px] ${
								activeTab === "requests"
									? "bg-primary/10 font-semibold text-primary"
									: "text-base-content/60 active:bg-base-200/80 sm:hover:bg-base-200/50"
							}`}
							onClick={() => {
								setActiveTab("requests")
								if (activeTab === "results") {
									setSelectedRequest(null)
									setExcludedObjects([])
								}
								scrollToTop()
								void queryClient.invalidateQueries({ queryKey: queryKeys.requests.all })
							}}
						>
							<span className="text-sm leading-none sm:text-base" aria-hidden>
								<FaList />
							</span>
							<span className="max-w-full truncate px-0.5 text-center font-medium leading-tight">
								Запросы
							</span>
						</button>
					</div>
				</div>

				{activeTab === "requests" && (
					<div className="mx-auto mt-2 max-w-2xl text-center">
						<p className="text-xs text-base-content/60 sm:text-sm">Запросы на анализ изображений</p>
					</div>
				)}
				{activeTab === "results" && selectedRequest && (
					<div className="mx-auto mt-2 max-w-2xl space-y-2 text-center">
						<p className="text-xs text-base-content/60 sm:text-sm">
							Результаты анализа для запроса #{selectedRequest.id}
						</p>
						{objects && objects.length > 0 ? (
							<p className="text-xs text-base-content/65 sm:text-sm">
								{excludedObjects.length > 0 ? (
									<>
										Исключено{" "}
										<span className="tabular-nums text-base-content/90">
											{excludedObjects.length}
										</span>{" "}
										из <span className="tabular-nums text-base-content/90">{objects.length}</span>
									</>
								) : (
									<>
										Найдено{" "}
										<span className="tabular-nums text-base-content/90">{objects.length}</span>{" "}
										объектов
									</>
								)}
							</p>
						) : null}
					</div>
				)}
			</div>

			<div ref={scrollContainerRef} className="min-h-0 flex-1 overflow-y-auto px-2">
				<div
					className={`mx-auto max-w-2xl space-y-4 px-1 pt-2 sm:space-y-6 ${
						showSendDock
							? "pb-[calc(4.25rem+env(safe-area-inset-bottom,0))] sm:pb-[calc(4.5rem+env(safe-area-inset-bottom,0))]"
							: "pb-2 sm:pb-4"
					}`}
				>
					{activeTab === "upload" ? (
						<UploadTab
							fraction={{ current: currentFraction }}
							product={{ selected: selectedProduct, onChange: setSelectedProduct }}
							files={{
								list: uploadedFiles,
								isUploading,
								onSelect: handleFileSelect,
								onRemove: removeFile,
								inputRef: fileInputRef,
							}}
							massFields={{
								mass1000,
								mass,
								massInputMode,
								onMassInputModeChange: setMassInputMode,
								onMass1000Change: handleMass1000Change,
								onMassChange: handleMassChange,
								massLiter,
								onMassLiterChange: setMassLiter,
							}}
							sampleMeta={{
								location,
								onLocationChange: setLocation,
								year,
								onYearChange: setYear,
							}}
							classification={{
								active: activeClassification,
								loading: loadingActiveClassification,
							}}
						/>
					) : activeTab === "requests" ? (
						<RequestsTab
							statusFilter={statusFilter}
							loadingRequests={loadingRequests}
							requestsData={requestsData}
							onStatusFilterChange={setStatusFilter}
							onRefreshRequests={handleRefreshRequests}
							onViewResults={handleViewResults}
							onDownloadPdf={handleDownloadPdf}
						/>
					) : activeTab === "results" ? (
						<ResultsTab
							loadingAnalysis={loadingAnalysis}
							objects={objects}
							selectedRequest={selectedRequest}
							onBackToRequests={handleBackToRequests}
							userId={userId ?? undefined}
							excludedObjects={excludedObjects}
							onExcludeObject={handleExcludeObject}
							onIncludeObject={handleIncludeObject}
							onResetSelection={handleResetSelection}
							onInvertSelection={handleInvertSelection}
						/>
					) : null}
				</div>
			</div>

			{showSendDock ? (
				<footer className="fixed right-0 bottom-0 left-0 z-40 border-t border-base-200 bg-base-100 p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))]">
					<div className="mx-auto w-full max-w-2xl">
						<button
							type="button"
							onClick={processImages}
							className="btn btn-primary relative w-full overflow-hidden"
							disabled={isUploading || uploadedFiles.length === 0}
							aria-busy={isUploading}
							aria-label={
								isUploading ? "Отправка файлов на сервер" : "Отправить изображения на анализ"
							}
						>
							{isUploading ? (
								<>
									{uploadPercent !== null ? (
										<>
											<span className="absolute inset-0 bg-black/25" aria-hidden />
											<span
												role="progressbar"
												aria-valuenow={uploadPercent}
												aria-valuemin={0}
												aria-valuemax={100}
												aria-label={`Загружено ${uploadPercent} процентов`}
												className="absolute inset-y-0 left-0 bg-primary-focus transition-[width] duration-150 ease-out"
												style={{ width: `${uploadPercent}%` }}
											/>
										</>
									) : (
										<span
											className="absolute inset-0 bg-primary-focus/50 motion-safe:animate-pulse"
											aria-hidden
										/>
									)}
									<span className="relative z-10 flex items-center justify-center gap-2 font-medium">
										{uploadPercent !== null ? (
											<span>{uploadPercent}%</span>
										) : (
											<>
												<span className="loading loading-spinner loading-sm" />
												<span>
													Отправка…
													{uploadProgress && uploadProgress.loaded > 0
														? ` ${formatSentBytes(uploadProgress.loaded)}`
														: ""}
												</span>
											</>
										)}
									</span>
								</>
							) : (
								<>
									<FaImage className="mr-2" aria-hidden />
									Отправить
								</>
							)}
						</button>
					</div>
				</footer>
			) : null}
		</div>
	)
}

export default AnalysisCreationSheet
