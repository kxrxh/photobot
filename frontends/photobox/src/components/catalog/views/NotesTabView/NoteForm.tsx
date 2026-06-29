import { useEffect, useState } from "react"
import type { Note } from "@/api/notes/types"

export default function NoteForm({
	note,
	onSave,
	onCancel,
}: {
	note?: Note | null
	onSave: (noteData: { text: string }) => void
	onCancel: () => void
}) {
	const [text, setText] = useState(note?.note || "")

	useEffect(() => {
		setText(note?.note || "")
	}, [note])

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault()
		if (text.trim()) {
			onSave({ text })
		}
	}

	return (
		<form onSubmit={handleSubmit} className="space-y-4">
			<textarea
				className="w-full textarea textarea-bordered"
				rows={5}
				value={text}
				onChange={(e) => setText(e.target.value)}
				placeholder="Введите текст заметки..."
				required
			/>
			<div className="flex w-full gap-2">
				<button type="button" className="flex-1 btn" onClick={onCancel}>
					Отмена
				</button>
				<button type="submit" className="flex-1 btn btn-primary">
					Сохранить
				</button>
			</div>
		</form>
	)
}
