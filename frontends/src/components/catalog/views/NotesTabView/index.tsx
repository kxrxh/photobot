import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { FaPlus } from "react-icons/fa"
import { MdOutlineNotes } from "react-icons/md"
import { createNote, deleteNote, fetchNotes, updateNote } from "@/api/notes"
import type { Note } from "@/api/notes/types"
import { queryKeys } from "@/api/queryKeys"
import { getUserFacingErrorMessage } from "@/utils/errors"
import ConfirmDeleteAlert from "./ConfirmDeleteAlert"
import NoteCard from "./NoteCard"
import NoteEditorSheet from "./NoteEditorSheet"
import ViewNoteSheet from "./ViewNoteSheet"

type NotesTabViewProps = {
	catalogItemId: string
}

export default function NotesTabView({ catalogItemId }: NotesTabViewProps) {
	const queryClient = useQueryClient()

	const {
		data: notes,
		isPending,
		error,
	} = useQuery({
		queryKey: queryKeys.notes(catalogItemId),
		queryFn: () => fetchNotes(catalogItemId),
	})

	const createNoteMutation = useMutation({
		mutationFn: (newNoteData: { catalogItemId: string; text: string }) =>
			createNote(newNoteData.catalogItemId, newNoteData.text),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.notes(catalogItemId) })
		},
	})

	const updateNoteMutation = useMutation({
		mutationFn: (updatedNote: Note) => updateNote(updatedNote),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.notes(catalogItemId) })
		},
	})

	const deleteNoteMutation = useMutation({
		mutationFn: (noteId: number) => deleteNote(noteId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.notes(catalogItemId) })
		},
	})

	const [editingNote, setEditingNote] = useState<Note | null | undefined>(undefined)
	const [isEditorOpen, setIsEditorOpen] = useState<boolean>(false)
	const [viewingNote, setViewingNote] = useState<Note | null>(null)
	const [isRemoveModalOpen, setIsRemoveModalOpen] = useState<boolean>(false)
	const [noteToRemove, setNoteToRemove] = useState<Note | null>(null)

	const handleAddNew = () => {
		setEditingNote(null)
		setIsEditorOpen(true)
	}

	const handleEdit = (note: Note) => {
		setEditingNote(note)
		setIsEditorOpen(true)
	}

	const handleView = (note: Note) => {
		setViewingNote(note)
	}

	const handleAskDelete = (note: Note) => {
		setNoteToRemove(note)
		setIsRemoveModalOpen(true)
	}

	const handleConfirmDelete = () => {
		if (noteToRemove) {
			deleteNoteMutation.mutate(noteToRemove.id)
		}
		setIsRemoveModalOpen(false)
		setNoteToRemove(null)
	}

	const handleCloseRemoveModal = () => {
		setIsRemoveModalOpen(false)
		setNoteToRemove(null)
	}

	const handleSave = (noteData: { text: string }) => {
		if (editingNote) {
			updateNoteMutation.mutate({
				...editingNote,
				note: noteData.text,
			})
		} else {
			createNoteMutation.mutate({
				catalogItemId,
				...noteData,
			})
		}
		setIsEditorOpen(false)
	}

	const handleCloseEditor = () => {
		setIsEditorOpen(false)
	}

	// Clear editing state shortly after modal is closed to avoid header text flicker during close animation
	useEffect(() => {
		if (!isEditorOpen) {
			const timeoutId = setTimeout(() => {
				setEditingNote(undefined)
			}, 250)
			return () => clearTimeout(timeoutId)
		}
		return undefined
	}, [isEditorOpen])

	const handleCloseViewModal = () => {
		setViewingNote(null)
	}

	if (isPending) return <div>Загрузка заметок...</div>
	if (error) return <div>Ошибка загрузки заметок: {getUserFacingErrorMessage(error)}</div>

	return (
		<div className="relative p-4 space-y-4">
			{notes && notes.length > 0 ? (
				<div className="space-y-4">
					{notes.map((note: Note) => (
						<NoteCard
							key={note.id}
							note={note}
							onEdit={handleEdit}
							onDelete={handleAskDelete}
							onView={handleView}
						/>
					))}
				</div>
			) : (
				<div className="flex flex-col items-center justify-center py-16 text-center text-base-content/60">
					<MdOutlineNotes size="4em" className="mb-4" />
					<h3 className="text-lg font-semibold">Заметок пока нет</h3>
					<p>Нажмите «+», чтобы добавить первую.</p>
				</div>
			)}

			<div className="fixed bottom-28 right-4">
				<button
					type="button"
					aria-label="Добавить заметку"
					className="btn btn-primary btn-lg btn-circle"
					onClick={handleAddNew}
				>
					<FaPlus size="1em" />
				</button>
			</div>

			<NoteEditorSheet
				isOpen={isEditorOpen}
				editingNote={editingNote}
				onSave={handleSave}
				onCancel={handleCloseEditor}
			/>

			<ViewNoteSheet viewingNote={viewingNote} onClose={handleCloseViewModal} />

			<ConfirmDeleteAlert
				isOpen={isRemoveModalOpen}
				onCancel={handleCloseRemoveModal}
				onConfirm={handleConfirmDelete}
			/>
		</div>
	)
}
