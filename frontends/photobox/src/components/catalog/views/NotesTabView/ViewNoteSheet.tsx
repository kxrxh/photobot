import { useId } from "react"
import { FaCalendar, FaUser } from "react-icons/fa"
import type { Note } from "@/api/notes/types"

export default function ViewNoteSheet({
	viewingNote,
	onClose,
}: {
	viewingNote: Note | null
	onClose: () => void
}) {
	const dialogId = useId()
	return (
		<dialog
			id={dialogId}
			className={`modal backdrop-blur-xs ${viewingNote ? "modal-open" : ""}`}
			onClose={onClose}
		>
			<div className="modal-box max-w-3xl max-h-[85vh] overflow-hidden">
				<h3 className="text-lg font-bold">Полный текст заметки</h3>
				{viewingNote && (
					<div className="mt-3 space-y-4">
						<div className="text-sm text-base-content/70 flex flex-wrap gap-4">
							<span className="inline-flex items-center gap-2">
								<FaUser className="w-4 h-4" />
								<span className="font-medium">
									{viewingNote.created_by != null ? `User ${viewingNote.created_by}` : "—"}
								</span>
							</span>
							<span className="inline-flex items-center gap-2">
								<FaCalendar className="w-4 h-4" />
								<span>
									{(() => {
										if (!viewingNote.created_at) return "—"
										const d = new Date(viewingNote.created_at)
										if (Number.isNaN(d.getTime())) return "—"
										return d.toLocaleDateString("ru-RU", {
											day: "2-digit",
											month: "2-digit",
											year: "numeric",
											hour: "2-digit",
											minute: "2-digit",
										})
									})()}
								</span>
							</span>
						</div>
						<div className="p-4 rounded bg-base-200 whitespace-pre-wrap wrap-break-word max-h-[60vh] overflow-auto">
							{viewingNote.note}
						</div>
						<div className="modal-action mt-2">
							<button type="button" className="btn btn-block" onClick={onClose}>
								Закрыть
							</button>
						</div>
					</div>
				)}
			</div>
			<form method="dialog" className="modal-backdrop" />
		</dialog>
	)
}
