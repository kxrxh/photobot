import { useId } from "react"
import type { Note } from "@/api/notes/types"
import NoteForm from "./NoteForm"

export default function NoteEditorSheet({
	isOpen,
	editingNote,
	onSave,
	onCancel,
}: {
	isOpen: boolean
	editingNote: Note | null | undefined
	onSave: (noteData: { text: string }) => void
	onCancel: () => void
}) {
	const dialogId = useId()
	return (
		<dialog
			id={dialogId}
			className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}
			onClose={onCancel}
		>
			<div className="modal-box">
				<h3 className="text-lg font-bold">
					{editingNote ? "Редактировать заметку" : "Добавить заметку"}
				</h3>
				<div>
					<NoteForm
						key={editingNote ? editingNote.id : "new"}
						note={editingNote}
						onSave={onSave}
						onCancel={onCancel}
					/>
				</div>
			</div>
			<form method="dialog" className="modal-backdrop" />
		</dialog>
	)
}
