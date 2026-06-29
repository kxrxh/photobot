import { useQueryClient } from "@tanstack/react-query"
import type React from "react"
import { useState } from "react"
import { FaCheck, FaExchangeAlt, FaUndo } from "react-icons/fa"
import { confirmRequest } from "@/api/analysis"
import type { AnalysisRequest, SimpleObject } from "@/api/analysis/types"
import { queryKeys } from "@/api/queryKeys"
import { ObjectImage } from "@/components/markup/components/ObjectImage"
import { useAlert } from "@/hooks/useAlert"
import { useAnalysisWebSocket } from "@/hooks/websocket"
import { getUserFacingErrorMessage } from "@/utils/errors"
import { log } from "@/utils/log"
import AnalysisConfirmationAlert from "../components/modal/AnalysisConfirmationAlert"

interface ResultsTabProps {
	loadingAnalysis: boolean
	objects: SimpleObject[] | undefined
	selectedRequest: AnalysisRequest | null
	onBackToRequests: () => void
	userId: number | undefined
	excludedObjects: string[]
	onExcludeObject: (objectId: string) => void
	onIncludeObject: (objectId: string) => void
	onResetSelection: () => void
	onInvertSelection: () => void
}

const ResultsTab: React.FC<ResultsTabProps> = ({
	loadingAnalysis,
	objects,
	selectedRequest,
	onBackToRequests,
	userId,
	excludedObjects,
	onExcludeObject,
	onIncludeObject,
	onResetSelection,
	onInvertSelection,
}) => {
	const [showConfirmModal, setShowConfirmModal] = useState(false)
	const [isSaving, setIsSaving] = useState(false)
	const { showError } = useAlert()
	const queryClient = useQueryClient()

	const handleConfirmSave = async () => {
		if (!selectedRequest) return

		const requestId = selectedRequest.id
		setIsSaving(true)

		try {
			await confirmRequest({
				request_id: requestId,
				excluded_object_ids: excludedObjects || [],
			})

			await queryClient.invalidateQueries({
				queryKey: queryKeys.analyses.all,
			})
			await queryClient.invalidateQueries({
				queryKey: queryKeys.requests.all,
			})

			setShowConfirmModal(false)

			onBackToRequests()
		} catch (error) {
			log.error("Failed to confirm request:", error)
			showError(getUserFacingErrorMessage(error))
		} finally {
			setIsSaving(false)
		}
	}

	const requestId = selectedRequest?.id
	useAnalysisWebSocket({
		userId: userId?.toString() || "",
		requestId: requestId,
		enabled: !!userId && !!requestId,
		onRequestUpdate: ({ data }) => {
			if (data.status === "completed") {
				onBackToRequests()
			}
		},
	})

	return (
		<>
			<div className="space-y-2 animate-fadeIn">
				<div className="space-y-4">
					{loadingAnalysis ? (
						<div className="space-y-4">
							<div className="text-center">
								<div className="skeleton h-6 w-48 mx-auto mb-2"></div>
								<div className="skeleton h-4 w-32 mx-auto"></div>
							</div>
							<div className="grid grid-cols-4 gap-2 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10">
								{Array.from({ length: 20 }, (_, i) => `skeleton-${i}`).map((key) => (
									<div key={key} className="aspect-square rounded-md border border-base-300 p-1">
										<div className="skeleton w-full h-full rounded"></div>
									</div>
								))}
							</div>
						</div>
					) : objects ? (
						<div className="space-y-4">
							{objects.length > 0 ? (
								<div className="space-y-4">
									<div className="grid grid-cols-4 gap-2 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-10 place-items-center">
										{objects.map((obj) => {
											const objectId = obj.id
											const isExcluded = excludedObjects.includes(objectId)

											return (
												<ObjectImage
													key={obj.id}
													object={obj}
													isSelected={false}
													isExcluded={isExcluded}
													isControlModeActive={true}
													mode="exclude"
													onClick={
														isExcluded
															? () => onIncludeObject(objectId)
															: () => onExcludeObject(objectId)
													}
													size="lg"
												/>
											)
										})}
									</div>

									<div className="pb-12">
										{excludedObjects.length === objects.length ? (
											<div className="rounded-lg border border-warning/25 bg-warning/5 px-3 py-2.5 text-center">
												<p className="text-sm text-warning">
													Нельзя сохранить анализ, исключив все объекты
												</p>
											</div>
										) : (
											<button
												type="button"
												className="btn btn-primary w-full"
												onClick={() => setShowConfirmModal(true)}
											>
												<FaCheck className="mr-2 h-4 w-4" aria-hidden />
												Сохранить анализ
											</button>
										)}
									</div>
								</div>
							) : (
								<div className="py-8 text-center">
									<div className="card bg-base-100 border border-base-300">
										<div className="card-body p-6 text-center">
											<h4 className="text-lg font-semibold mb-2">Объекты не найдены</h4>
											<p className="text-sm opacity-70">
												В этом анализе не было обнаружено объектов
											</p>
										</div>
									</div>
								</div>
							)}
						</div>
					) : (
						<div className="py-8 text-center">
							<div className="card bg-base-100 border border-base-300">
								<div className="card-body p-6 text-center">
									<h4 className="text-lg font-semibold mb-2">Результаты не найдены</h4>
									<p className="text-sm opacity-70 mb-4">Не удалось загрузить результаты анализа</p>
								</div>
							</div>
						</div>
					)}
				</div>

				{objects && objects.length > 0 && (
					<div className="fixed right-0 bottom-0 left-0 z-50 border-t border-base-200 bg-base-100">
						<div className="mx-auto flex max-w-2xl gap-2 px-2 py-2 pb-[max(0.5rem,env(safe-area-inset-bottom))]">
							<button
								type="button"
								onClick={onInvertSelection}
								className="btn btn-ghost btn-sm min-h-9 flex-1 gap-2 font-normal text-base-content"
								title="Инвертировать выбор"
							>
								<FaExchangeAlt className="h-4 w-4 shrink-0 opacity-70" aria-hidden />
								Инвертировать
							</button>
							{excludedObjects.length > 0 ? (
								<button
									type="button"
									onClick={onResetSelection}
									className="btn btn-ghost btn-sm min-h-9 flex-1 gap-2 font-normal text-base-content"
									title="Сбросить выбор"
								>
									<FaUndo className="h-4 w-4 shrink-0 opacity-70" aria-hidden />
									Сбросить
								</button>
							) : null}
						</div>
					</div>
				)}
			</div>

			<AnalysisConfirmationAlert
				showConfirmModal={showConfirmModal}
				setShowConfirmModal={setShowConfirmModal}
				excludedObjects={excludedObjects}
				objects={objects}
				onConfirmSave={handleConfirmSave}
				isSaving={isSaving}
			/>
		</>
	)
}

export default ResultsTab
