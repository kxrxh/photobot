import { useInfiniteQuery } from "@tanstack/react-query"
import { useNavigate } from "@tanstack/react-router"
import { useEffect, useId, useMemo, useRef, useState } from "react"
import { FaClock, FaInbox, FaStickyNote, FaSync } from "react-icons/fa"
import { fetchProposals } from "@/api/catalog"
import type { ProposalStatus } from "@/api/catalog/types"
import { queryKeys } from "@/api/queryKeys"
import ErrorPage from "@/components/common/layout/ErrorPage"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import {
	formatDateTime,
	getStatusConfig,
	type ProposalItem,
	STATUS_OPTIONS,
} from "./CatalogProposalsShared"

type Props = {
	userId?: number
}

const PAGE_SIZE = 25

const CatalogProposalsUser = ({ userId }: Props) => {
	const navigate = useNavigate()
	const statusFilterId = useId()
	const [statusFilter, setStatusFilter] = useState<ProposalStatus | "">("")
	const [uiLoading, setUiLoading] = useState(false)
	const loadingStartedAt = useRef<number | null>(null)

	const handleBack = () => {
		navigate({ to: "/catalog", replace: true })
	}

	const {
		data: infiniteData,
		isLoading,
		isFetching,
		isFetchingNextPage,
		error,
		refetch,
		fetchNextPage,
		hasNextPage,
	} = useInfiniteQuery({
		queryKey: queryKeys.catalog.proposalsUser(userId ?? 0, statusFilter),
		queryFn: async ({ pageParam }) => {
			const offset = pageParam as number
			return fetchProposals({
				status: statusFilter || undefined,
				sort_order: "desc",
				limit: PAGE_SIZE,
				offset,
			})
		},
		initialPageParam: 0,
		getNextPageParam: (lastPage) => {
			const nextOffset = lastPage.offset + lastPage.limit
			return nextOffset < lastPage.total ? nextOffset : undefined
		},
		enabled: !!userId,
		staleTime: 30 * 1000,
		refetchOnWindowFocus: false,
	})

	const proposalsData = useMemo(() => {
		if (!infiniteData?.pages.length) return undefined
		const proposals = infiniteData.pages.flatMap((p) => p.data) as ProposalItem[]
		const total = infiniteData.pages[0]?.total ?? 0
		return { proposals, total }
	}, [infiniteData])

	const isBusy = isLoading || (isFetching && !isFetchingNextPage)

	useEffect(() => {
		if (isBusy) {
			if (!loadingStartedAt.current) {
				loadingStartedAt.current = Date.now()
			}
			setUiLoading(true)
			return
		}

		if (!loadingStartedAt.current) {
			setUiLoading(false)
			return
		}

		const elapsed = Date.now() - loadingStartedAt.current
		const remaining = Math.max(500 - elapsed, 0)
		const timeout = setTimeout(() => {
			setUiLoading(false)
			loadingStartedAt.current = null
		}, remaining)

		return () => clearTimeout(timeout)
	}, [isBusy])

	const filteredProposals = useMemo(() => {
		if (!proposalsData?.proposals) return []
		return proposalsData.proposals
	}, [proposalsData])

	const hasProposals = filteredProposals.length > 0
	const showSkeleton = uiLoading && !hasProposals

	const [nowMs, setNowMs] = useState(() => Date.now())

	useEffect(() => {
		const syncNow = () => setNowMs(Date.now())
		const onVisible = () => {
			if (document.visibilityState === "visible") syncNow()
		}
		document.addEventListener("visibilitychange", onVisible)
		const intervalId = window.setInterval(syncNow, 30_000)
		return () => {
			document.removeEventListener("visibilitychange", onVisible)
			window.clearInterval(intervalId)
		}
	}, [])

	const handleOpenDetail = (proposalId: number) => {
		navigate({
			to: "/catalog/proposals/$proposalId",
			params: { proposalId: proposalId.toString() },
		})
	}

	return (
		<div className="h-full bg-base-100">
			<header className="sticky top-0 z-50 w-full bg-base-100 border-b border-base-300">
				<div className="px-2 py-3">
					<h1 className="text-2xl font-bold">Мои предложения</h1>
				</div>
				<div className="px-2 pb-2">
					<div className="flex gap-3 items-center">
						<div className="flex-1 min-w-0">
							<ModalSelect
								id={statusFilterId}
								title="Фильтр по статусу"
								placeholder="Все предложения"
								options={STATUS_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
								value={statusFilter}
								onChange={(v) => setStatusFilter((v || "") as ProposalStatus | "")}
								size="sm"
							/>
						</div>

						<button
							type="button"
							onClick={() => refetch()}
							className={`group btn btn-outline btn-primary btn-xs gap-1.5 min-h-8 px-3 whitespace-nowrap transition-all duration-200 touch-manipulation active:scale-95 shrink-0 ${
								uiLoading ? "btn-disabled animate-pulse" : "hover:shadow-md hover:scale-105"
							}`}
							disabled={uiLoading}
							title="Обновить список предложений"
						>
							<FaSync
								className={`w-3.5 h-3.5 transition-transform duration-200 ${
									uiLoading ? "animate-spin" : "group-hover:rotate-180"
								}`}
							/>
							<span className="text-xs font-medium">
								{uiLoading ? "Обновление..." : "Обновить"}
							</span>
						</button>
					</div>
				</div>
			</header>

			<main className="w-full max-w-md mx-auto overflow-y-auto">
				<div className="p-2">
					<div className="space-y-4 sm:space-y-6">
						<div className="space-y-3">
							{showSkeleton ? (
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
												<div className="border-t border-base-300/50 pt-3">
													<div className="skeleton h-10 w-full rounded-field" />
												</div>
											</div>
										</div>
									))}
								</div>
							) : error ? (
								<div className="min-h-[50vh]">
									<ErrorPage error={error as Error} fullHeight={false} />
								</div>
							) : filteredProposals.length ? (
								filteredProposals.map((proposal) => {
									const statusCfg = getStatusConfig(proposal.status)
									const createdDate = new Date(proposal.created_at)
									const updatedDate = new Date(proposal.updated_at)
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
									const shortId = proposal.request_id
										? proposal.request_id.slice(0, 8)
										: proposal.id.toString()
									const title = proposal.pending_name?.trim() || "Предложение"
									const reviewNotes = proposal.review_notes?.trim()
									const displayDate = isUpdated ? proposal.updated_at : proposal.created_at

									return (
										<div key={proposal.id} className="card border border-base-300 bg-base-100">
											<div className="card-body gap-3 p-3 sm:p-4">
												<div className="flex items-start justify-between gap-3">
													<div className="flex min-w-0 items-center gap-3">
														<div
															className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${statusCfg.iconWrap}`}
														>
															{statusCfg.icon}
														</div>
														<div className="min-w-0">
															<h3 className="text-base font-semibold leading-tight truncate">
																{title}
															</h3>
															<p className="text-sm text-base-content/70">{timeAgo}</p>
														</div>
													</div>
													<span className={`shrink-0 ${statusCfg.badge}`}>{statusCfg.label}</span>
												</div>

												<div className="space-y-2">
													<div className="flex flex-wrap items-baseline gap-x-2 gap-y-1">
														<span className="text-sm text-base-content/70">Номер:</span>
														<span className="text-sm font-medium text-base-content">
															#{shortId}
														</span>
													</div>

													{reviewNotes && (
														<div className="rounded-lg border border-warning/30 bg-warning/5 px-3 py-2">
															<div className="flex items-start gap-2">
																<FaStickyNote className="w-4 h-4 text-warning shrink-0" />
																<div className="space-y-1">
																	<p className="text-sm font-semibold text-warning">
																		Комментарий модератора
																	</p>
																	<p className="text-sm text-base-content/80 whitespace-pre-line leading-snug">
																		{reviewNotes}
																	</p>
																</div>
															</div>
														</div>
													)}

													<div className="flex flex-col gap-2 border-t border-base-300/50 pt-3 sm:flex-row sm:flex-wrap sm:items-start sm:gap-x-4 sm:gap-y-2">
														<div className="flex min-w-0 items-center gap-1 text-xs text-base-content/70">
															<FaClock className="w-3 h-3 shrink-0" />
															<span className="min-w-0">
																{isUpdated ? "Обновлено" : "Создано"}: {formatDateTime(displayDate)}
															</span>
														</div>
													</div>
												</div>

												<div className="border-t border-base-300/50 pt-3">
													<button
														type="button"
														className="btn btn-primary w-full transition-all duration-200 touch-manipulation active:scale-[0.98]"
														onClick={() => handleOpenDetail(proposal.id)}
													>
														<span className="text-sm font-medium">Открыть карточку</span>
													</button>
												</div>
											</div>
										</div>
									)
								})
							) : (
								<div className="card bg-base-100 border border-base-300">
									<div className="card-body p-6 text-center">
										<FaInbox className="w-12 h-12 mx-auto mb-4 text-base-content/30" />
										<h4 className="text-lg font-semibold mb-2">Нет предложений</h4>
										<p className="text-sm text-base-content/70 leading-relaxed">
											У вас пока нет таких предложений
										</p>
									</div>
								</div>
							)}
							{hasNextPage && filteredProposals.length > 0 ? (
								<div className="pt-2 pb-4">
									<button
										type="button"
										className={`btn btn-outline btn-primary w-full ${isFetchingNextPage ? "loading" : ""}`}
										disabled={isFetchingNextPage}
										onClick={() => fetchNextPage()}
									>
										Загрузить ещё
									</button>
								</div>
							) : null}
						</div>
					</div>
				</div>
			</main>

			<footer className="fixed bottom-0 left-0 right-0 z-50 border-t bg-base-100/95 backdrop-blur-sm border-base-300">
				<div className="max-w-md mx-auto">
					<div className="p-2">
						<button type="button" className="w-full btn" onClick={handleBack}>
							Назад в каталог
						</button>
					</div>
				</div>
			</footer>
		</div>
	)
}

export default CatalogProposalsUser
