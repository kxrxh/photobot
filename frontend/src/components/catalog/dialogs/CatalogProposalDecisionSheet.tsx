import type React from "react"
import { useId } from "react"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"

type Props = {
	isOpen: boolean
	reviewNote: string
	onReviewNoteChange: (value: string) => void
	onClose: () => void
	onApprove: () => void
	onRequestChanges: () => void
	onReject: () => void
	onCancel: () => void
	disableApprove: boolean
	disableNoteActions: boolean
	isProcessing: boolean
}

const CatalogProposalDecisionSheet: React.FC<Props> = ({
	isOpen,
	reviewNote,
	onReviewNoteChange,
	onClose,
	onApprove,
	onRequestChanges,
	onReject,
	onCancel,
	disableApprove,
	disableNoteActions,
	isProcessing,
}) => {
	const reviewNoteId = useId()

	if (!isOpen) return null

	return (
		<div className="fixed inset-0 z-50 flex items-end justify-center bg-base-300/60 backdrop-blur-sm">
			<div className="w-full max-w-md rounded-t-2xl bg-base-100 shadow-xl border border-base-300">
				<div className="p-4 space-y-3">
					<div className="flex items-center justify-between">
						<h3 className="text-lg font-semibold">Решение по заявке</h3>
						<SheetHeaderCloseButton onClick={onClose} disabled={isProcessing} />
					</div>

					<p className="text-sm text-base-content/70">
						Добавьте комментарий к решению: он нужен для одобрения, запроса правок и отклонения.
					</p>

					<div className="space-y-2">
						<label className="label" htmlFor={reviewNoteId}>
							<span className="label-text">Комментарий</span>
						</label>
						<textarea
							id={reviewNoteId}
							className="w-full h-24 textarea textarea-bordered"
							placeholder="Кратко опишите решение"
							value={reviewNote}
							onChange={(e) => onReviewNoteChange(e.target.value)}
							disabled={isProcessing}
						/>
					</div>

					<div className="grid grid-cols-2 gap-2">
						<button type="button" className="btn" onClick={onCancel} disabled={isProcessing}>
							Отменить
						</button>
						<button
							type="button"
							className="btn btn-success"
							onClick={onApprove}
							disabled={disableApprove}
						>
							Одобрить
						</button>
						<button
							type="button"
							className="btn btn-warning"
							onClick={onRequestChanges}
							disabled={disableNoteActions}
						>
							Запросить правки
						</button>
						<button
							type="button"
							className="btn btn-error"
							onClick={onReject}
							disabled={disableNoteActions}
						>
							Отклонить
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}

export default CatalogProposalDecisionSheet
