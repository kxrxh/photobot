import { FaCalendar, FaEdit, FaRegTrashAlt, FaUser } from "react-icons/fa"
import type { Note } from "@/api/notes/types"

export default function NoteCard({
	note,
	onEdit,
	onDelete,
	onView,
}: {
	note: Note
	onEdit: (note: Note) => void
	onDelete: (note: Note) => void
	onView: (note: Note) => void
}) {
	return (
		<div className="transition-all duration-300 border shadow-md card bg-base-100 hover:shadow-lg border-base-content/10">
			<div className="card-body p-4">
				<p className="text-base-content/90 wrap-break-word truncate">{note.note}</p>
				<button
					type="button"
					className="mt-2 text-sm link link-primary self-start"
					onClick={() => onView(note)}
				>
					Читать полностью
				</button>
				<div className="my-2 divider" />
				<div className="flex items-center justify-between text-xs text-base-content/70">
					<div className="grid gap-2">
						<div className="flex items-center gap-2">
							<FaUser className="w-4 h-4" />
							<span className="font-medium">
								{note.created_by != null ? `User ${note.created_by}` : "—"}
							</span>
						</div>
						<div className="flex items-center gap-2">
							<FaCalendar className="w-4 h-4" />
							<span>
								{(() => {
									if (!note.created_at) return "—"
									const d = new Date(note.created_at)
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
						</div>
					</div>
					<div className="self-end card-actions">
						<button
							type="button"
							aria-label="Редактировать заметку"
							className="btn btn-ghost btn-sm btn-circle"
							onClick={() => onEdit(note)}
						>
							<FaEdit className="w-4 h-4" />
						</button>
						<button
							type="button"
							aria-label="Удалить заметку"
							className="btn btn-ghost btn-sm btn-circle hover:text-error"
							onClick={() => onDelete(note)}
						>
							<FaRegTrashAlt className="w-4 h-4" />
						</button>
					</div>
				</div>
			</div>
		</div>
	)
}
