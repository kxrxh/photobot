import React, { useEffect, useId, useRef } from "react"
import {
	FaCheck,
	FaChevronDown,
	FaChevronUp,
	FaClock,
	FaDownload,
	FaExclamationTriangle,
	FaHashtag,
	FaInbox,
	FaPlus,
	FaSpinner,
	FaSync,
} from "react-icons/fa"
import type { AnalysisRequest } from "@/api/analysis/types"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { parseAnalysisErrorMessage } from "@/utils/analysisRequestErrors"
import { log } from "@/utils/log"

import { getDevMockRequestsPayload, shouldUseDevMockRequestCards } from "../devMockRequests"

const REQUEST_STATUS_OPTIONS = [
	{ value: "created", label: "Создан" },
	{ value: "processing", label: "Обработка" },
	{ value: "waiting_for_confirmation", label: "На подтверждении" },
	{ value: "completed", label: "Завершено" },
	{ value: "failed", label: "Ошибка" },
] as const

interface RequestsTabProps {
	statusFilter: string
	loadingRequests: boolean
	requestsData: { requests: AnalysisRequest[]; total: number } | undefined
	onStatusFilterChange: (value: string) => void
	onRefreshRequests: () => void
	onViewResults: (request: AnalysisRequest) => void
	onDownloadPdf: (request: AnalysisRequest) => Promise<void>
}

function formatRequestErrorBody(text: string) {
	const rows = parseAnalysisErrorMessage(text)
	const structured = rows.length > 1 || Boolean(rows[0]?.fileLabel)
	if (!structured) {
		return <div className="text-xs text-error/90 whitespace-pre-line leading-snug">{text}</div>
	}
	return (
		<ul className="list-none space-y-2 text-xs text-error/90 leading-snug">
			{rows.map((row) => (
				<li key={row.fileLabel ? `${row.fileLabel}:${row.message}` : row.message} className="pl-0">
					{row.fileLabel ? (
						<>
							<span className="font-semibold text-error">{row.fileLabel}</span>
							<span className="text-base-content/80"> — {row.message}</span>
						</>
					) : (
						row.message
					)}
				</li>
			))}
		</ul>
	)
}

const RequestsTab: React.FC<RequestsTabProps> = ({
	statusFilter,
	loadingRequests,
	requestsData,
	onStatusFilterChange,
	onRefreshRequests,
	onViewResults,
	onDownloadPdf,
}) => {
	const statusFilterId = useId()
	const displayRequestsData = React.useMemo(() => {
		if (loadingRequests) {
			return requestsData
		}
		if (shouldUseDevMockRequestCards() && (requestsData?.requests?.length ?? 0) === 0) {
			return getDevMockRequestsPayload(statusFilter)
		}
		return requestsData
	}, [loadingRequests, requestsData, statusFilter])

	const [expandedErrors, setExpandedErrors] = React.useState<Set<string>>(new Set())
	const [downloadingRequestIds, setDownloadingRequestIds] = React.useState<Set<string>>(new Set())
	const [copiedAnalysisIds, setCopiedAnalysisIds] = React.useState<Set<string>>(new Set())
	const [copiedRequestIds, setCopiedRequestIds] = React.useState<Set<string>>(new Set())
	const [nowMs, setNowMs] = React.useState(() => Date.now())

	const analysisCopyTimeoutsRef = useRef<Map<string, number>>(new Map())
	const requestCopyTimeoutsRef = useRef<Map<string, number>>(new Map())

	useEffect(() => {
		return () => {
			for (const t of analysisCopyTimeoutsRef.current.values()) {
				clearTimeout(t)
			}
			analysisCopyTimeoutsRef.current.clear()
			for (const t of requestCopyTimeoutsRef.current.values()) {
				clearTimeout(t)
			}
			requestCopyTimeoutsRef.current.clear()
		}
	}, [])

	// Relative labels track wall clock; resync when the tab becomes visible (timers are throttled in background).
	React.useEffect(() => {
		const syncNow = () => setNowMs(Date.now())

		const onVisible = () => {
			if (document.visibilityState === "visible") {
				syncNow()
			}
		}

		document.addEventListener("visibilitychange", onVisible)
		const intervalId = window.setInterval(syncNow, 30_000)

		return () => {
			document.removeEventListener("visibilitychange", onVisible)
			window.clearInterval(intervalId)
		}
	}, [])

	const toggleErrorExpansion = (requestId: string) => {
		setExpandedErrors((prev) => {
			const newSet = new Set(prev)
			if (newSet.has(requestId)) {
				newSet.delete(requestId)
			} else {
				newSet.add(requestId)
			}
			return newSet
		})
	}

	const handleDownloadPdf = async (request: AnalysisRequest) => {
		if (!request.temp_id || downloadingRequestIds.has(request.id)) {
			return
		}

		setDownloadingRequestIds((prev) => new Set(prev).add(request.id))

		try {
			await onDownloadPdf(request)
		} catch (error) {
			log.devError("Error downloading PDF:", error)
		} finally {
			setDownloadingRequestIds((prev) => {
				const newSet = new Set(prev)
				newSet.delete(request.id)
				return newSet
			})
		}
	}

	const handleCopyAnalysisId = async (analysisId: string, event: React.MouseEvent) => {
		event.stopPropagation()

		try {
			await navigator.clipboard.writeText(analysisId)
			setCopiedAnalysisIds((prev) => new Set(prev).add(analysisId))
			const prevT = analysisCopyTimeoutsRef.current.get(analysisId)
			if (prevT !== undefined) clearTimeout(prevT)
			const tid = window.setTimeout(() => {
				analysisCopyTimeoutsRef.current.delete(analysisId)
				setCopiedAnalysisIds((prev) => {
					const newSet = new Set(prev)
					newSet.delete(analysisId)
					return newSet
				})
			}, 2000)
			analysisCopyTimeoutsRef.current.set(analysisId, tid)
		} catch (error) {
			log.devError("Failed to copy analysis ID:", error)
		}
	}

	const handleCopyRequestId = async (id: string, event: React.MouseEvent) => {
		event.stopPropagation()

		try {
			await navigator.clipboard.writeText(id)
			setCopiedRequestIds((prev) => new Set(prev).add(id))
			const prevT = requestCopyTimeoutsRef.current.get(id)
			if (prevT !== undefined) clearTimeout(prevT)
			const tid = window.setTimeout(() => {
				requestCopyTimeoutsRef.current.delete(id)
				setCopiedRequestIds((prev) => {
					const next = new Set(prev)
					next.delete(id)
					return next
				})
			}, 2000)
			requestCopyTimeoutsRef.current.set(id, tid)
		} catch (error) {
			log.devError("Failed to copy request ID:", error)
		}
	}

	const statusFilterLabel = (() => {
		switch (statusFilter) {
			case "created":
				return "«Создан»"
			case "processing":
				return "«Обработка»"
			case "waiting_for_confirmation":
				return "«На подтверждении»"
			case "completed":
				return "«Завершено»"
			case "failed":
				return "«Ошибка»"
			default:
				return ""
		}
	})()

	const getStatusConfig = (status: string) => {
		switch (status) {
			case "created":
				return {
					icon: <FaPlus className="h-4 w-4" />,
					label: "Создан",
					iconWrap: "bg-base-200 text-base-content/80",
					badge: "badge badge-sm badge-ghost",
				}
			case "processing":
				return {
					icon: <FaSpinner className="h-4 w-4 animate-spin" />,
					label: "Обработка",
					iconWrap: "bg-warning/10 text-warning",
					badge: "badge badge-sm badge-warning badge-outline",
				}
			case "waiting_for_confirmation":
				return {
					icon: <FaClock className="h-4 w-4" />,
					label: "На подтверждении",
					iconWrap: "bg-info/10 text-info",
					badge: "badge badge-sm badge-info badge-outline",
				}
			case "completed":
				return {
					icon: <FaCheck className="h-4 w-4" />,
					label: "Завершено",
					iconWrap: "bg-success/10 text-success",
					badge: "badge badge-sm badge-success badge-outline",
				}
			case "failed":
				return {
					icon: <FaExclamationTriangle className="h-4 w-4" />,
					label: "Ошибка",
					iconWrap: "bg-error/10 text-error",
					badge: "badge badge-sm badge-error badge-outline",
				}
			default:
				return {
					icon: <FaClock className="h-4 w-4" />,
					label: "Неизвестно",
					iconWrap: "bg-base-200 text-base-content/70",
					badge: "badge badge-sm badge-ghost",
				}
		}
	}

	return (
		<div className="space-y-4 sm:space-y-6">
			<div className="flex gap-3 items-center">
				<div className="flex-1 min-w-0">
					<ModalSelect
						id={statusFilterId}
						title="Фильтр по статусу"
						placeholder="Все запросы"
						options={REQUEST_STATUS_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
						value={statusFilter}
						onChange={onStatusFilterChange}
						size="sm"
					/>
				</div>

				<button
					type="button"
					onClick={onRefreshRequests}
					className={`btn btn-outline btn-primary btn-xs gap-1.5 min-h-8 px-3 whitespace-nowrap transition-all duration-200 touch-manipulation active:scale-95 shrink-0 ${
						loadingRequests ? "btn-disabled animate-pulse" : "hover:shadow-md hover:scale-105"
					}`}
					disabled={loadingRequests}
					title="Обновить список запросов"
				>
					<FaSync
						className={`w-3.5 h-3.5 transition-transform duration-200 ${
							loadingRequests ? "animate-spin" : "group-hover:rotate-180"
						}`}
					/>
					<span className="text-xs font-medium">
						{loadingRequests ? "Обновление..." : "Обновить"}
					</span>
				</button>
			</div>

			<div className="space-y-3">
				{loadingRequests ? (
					<div className="space-y-4">
						{["skeleton-1", "skeleton-2", "skeleton-3"].map((skeletonId) => (
							<div key={skeletonId} className="card border border-base-300 bg-base-100">
								<div className="card-body gap-3 p-3 sm:p-4">
									<div className="flex items-start justify-between gap-3">
										<div className="flex items-center gap-3">
											<div className="skeleton h-10 w-10 shrink-0 rounded-full" />
											<div className="space-y-2">
												<div className="skeleton h-4 w-40" />
												<div className="skeleton h-3 w-24" />
											</div>
										</div>
										<div className="skeleton h-6 w-16 rounded-field" />
									</div>
									<div className="skeleton h-4 w-32" />
									<div className="flex flex-col gap-2 border-t border-base-300/50 pt-3 sm:flex-row sm:gap-4">
										<div className="skeleton h-3 w-44" />
										<div className="skeleton h-3 w-36" />
									</div>
								</div>
							</div>
						))}
					</div>
				) : displayRequestsData?.requests?.length ? (
					displayRequestsData.requests.map((request) => {
						const statusCfg = getStatusConfig(request.status)
						const createdDate = new Date(request.created_at)
						const updatedDate = new Date(request.updated_at)
						const isUpdated = updatedDate.getTime() !== createdDate.getTime()

						const referenceDate =
							updatedDate.getTime() >= createdDate.getTime() ? updatedDate : createdDate
						const diffMs = nowMs - referenceDate.getTime()
						const diffMins = Math.floor(diffMs / (1000 * 60))
						const diffHours = Math.floor(diffMins / 60)
						const diffDays = Math.floor(diffHours / 24)

						let timeAgo: string
						if (diffMins < 1) timeAgo = "только что"
						else if (diffMins < 60) timeAgo = `${diffMins} мин назад`
						else if (diffHours < 24) timeAgo = `${diffHours} ч назад`
						else timeAgo = `${diffDays} д назад`

						const requestId = request.id

						return (
							<div key={request.id} className="card border border-base-300 bg-base-100">
								<div className="card-body gap-3 p-3 sm:p-4">
									<div className="flex items-start justify-between gap-3">
										<div className="flex min-w-0 items-center gap-3">
											<div
												className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${statusCfg.iconWrap}`}
											>
												{statusCfg.icon}
											</div>
											<div className="min-w-0">
												<h3 className="text-base font-semibold leading-tight">Запрос</h3>
												<p className="text-sm text-base-content/70">{timeAgo}</p>
											</div>
										</div>
										<span className={`shrink-0 ${statusCfg.badge}`}>{statusCfg.label}</span>
									</div>

									<div className="space-y-2">
										<div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
											<span className="text-sm text-base-content/70">Продукт:</span>
											<span className="text-sm font-medium text-base-content">
												{request.product}
											</span>
										</div>

										{request.status === "failed" && (
											<div className="pt-2">
												<button
													type="button"
													onClick={() => toggleErrorExpansion(request.id)}
													className="flex items-center gap-2 text-xs text-error/80 hover:text-error transition-colors duration-200 p-2 rounded-md hover:bg-error/5 w-full text-left"
													title="Показать/скрыть информацию об ошибке"
												>
													<FaExclamationTriangle className="w-3 h-3 shrink-0" />
													<span className="font-medium">Информация об ошибке</span>
													{expandedErrors.has(request.id) ? (
														<FaChevronUp className="w-3 h-3 ml-auto" />
													) : (
														<FaChevronDown className="w-3 h-3 ml-auto" />
													)}
												</button>
												{expandedErrors.has(request.id) && request.error_message && (
													<div className="mt-1 ml-5 p-2 bg-error/5 border border-error/10 rounded-md">
														{formatRequestErrorBody(request.error_message)}
													</div>
												)}
											</div>
										)}

										<div className="flex flex-col gap-2 border-t border-base-300/50 pt-3 sm:flex-row sm:flex-wrap sm:items-start sm:gap-x-4 sm:gap-y-2">
											<div className="flex min-w-0 items-center gap-1 text-xs text-base-content/70">
												<FaClock className="w-3 h-3 shrink-0" />
												<span className="min-w-0">
													{isUpdated ? "Обновлено" : "Создано"}:{" "}
													{updatedDate.toLocaleString("ru-RU")}
												</span>
											</div>
											<button
												type="button"
												onClick={(e) => void handleCopyRequestId(requestId, e)}
												className="flex w-full min-w-0 items-start gap-1 text-xs opacity-60 transition-opacity duration-200 hover:opacity-100 cursor-pointer bg-transparent border-0 p-0 text-left sm:w-auto"
												title="Нажмите, чтобы скопировать ID запроса"
												aria-label="Скопировать ID запроса"
											>
												{copiedRequestIds.has(requestId) ? (
													<FaCheck className="mt-0.5 h-3 w-3 shrink-0 text-success" />
												) : (
													<FaHashtag className="mt-0.5 h-3 w-3 shrink-0" />
												)}
												<span className="min-w-0 break-all">
													ID запроса: {requestId}
													{copiedRequestIds.has(requestId) && (
														<span className="text-success"> (скопировано)</span>
													)}
												</span>
											</button>
											{request.status === "completed" && request.temp_id && (
												<button
													type="button"
													onClick={(e) => {
														if (request.temp_id) {
															handleCopyAnalysisId(request.temp_id, e)
														}
													}}
													className="flex w-full min-w-0 items-start gap-1 text-xs opacity-60 transition-opacity duration-200 hover:opacity-100 cursor-pointer bg-transparent border-0 p-0 text-left sm:w-auto"
													title="Нажмите, чтобы скопировать ID анализа"
												>
													{copiedAnalysisIds.has(request.temp_id) ? (
														<FaCheck className="mt-0.5 h-3 w-3 shrink-0 text-success" />
													) : (
														<FaHashtag className="mt-0.5 h-3 w-3 shrink-0" />
													)}
													<span className="min-w-0 break-all">
														ID Анализа: {request.temp_id}
														{copiedAnalysisIds.has(request.temp_id) && (
															<span className="text-success"> (скопировано)</span>
														)}
													</span>
												</button>
											)}
										</div>
									</div>

									{request.status === "waiting_for_confirmation" && (
										<div className="border-t border-base-300/50 pt-3">
											<button
												type="button"
												onClick={() => onViewResults(request)}
												className="btn btn-primary w-full gap-2 transition-all duration-200 touch-manipulation active:scale-[0.98]"
												title="Открыть результаты анализа"
											>
												<FaCheck className="w-4 h-4" />
												<span className="text-sm font-medium">Открыть результаты анализа</span>
											</button>
										</div>
									)}

									{request.status === "completed" && (
										<div className="border-t border-base-300/50 pt-3">
											<button
												type="button"
												onClick={() => handleDownloadPdf(request)}
												disabled={!request.temp_id || downloadingRequestIds.has(requestId)}
												className={`btn btn-primary w-full gap-2 transition-all duration-200 touch-manipulation active:scale-[0.98] ${
													!request.temp_id || downloadingRequestIds.has(requestId)
														? "btn-disabled"
														: ""
												}`}
												title={
													downloadingRequestIds.has(requestId)
														? "Генерация отчета..."
														: request.temp_id
															? "Скачать PDF отчет"
															: "PDF отчет недоступен"
												}
											>
												{downloadingRequestIds.has(requestId) ? (
													<div className="loading loading-spinner loading-sm"></div>
												) : (
													<FaDownload className="w-4 h-4" />
												)}
												<span className="text-sm font-medium">
													{downloadingRequestIds.has(requestId)
														? "Генерация отчета..."
														: "Скачать PDF отчет"}
												</span>
											</button>
										</div>
									)}
								</div>
							</div>
						)
					})
				) : (
					<div className="card bg-base-100 border border-base-300">
						<div className="card-body p-6 text-center">
							<FaInbox className="w-12 h-12 mx-auto mb-4 text-base-content/30" />
							{statusFilter ? (
								<>
									<h4 className="text-lg font-semibold mb-2">Нет запросов с выбранным статусом</h4>
									<p className="text-sm text-base-content/70 leading-relaxed">
										Здесь собраны ваши{" "}
										<strong className="font-medium text-base-content">
											запросы на анализ фотографий
										</strong>
										{statusFilterLabel ? ` (фильтр ${statusFilterLabel})` : ""}. С таким статусом
										сейчас ничего нет — выберите «Все запросы» выше или зайдите позже, когда статус
										обновится.
									</p>
								</>
							) : (
								<>
									<h4 className="text-lg font-semibold mb-2">Пока нет запросов на анализ</h4>
									<p className="text-sm text-base-content/70 leading-relaxed">
										Это список{" "}
										<strong className="font-medium text-base-content">
											заявок на анализ изображений
										</strong>
										: от отправки снимков до готового отчёта. Чтобы появился первый запрос, откройте
										вкладку «Фото», загрузите фотографии и отправьте их на обработку.
									</p>
								</>
							)}
						</div>
					</div>
				)}
			</div>
		</div>
	)
}

export default RequestsTab
