import { handleApiResponse } from "@/api/helpers"
import type { Note } from "@/api/notes/types"
import { API_ENDPOINTS } from "../config"
import { createAuthenticatedClient } from "../createAuthenticatedClient"
import type { ApiResponse } from "../types"

const client = createAuthenticatedClient({
	prefixUrl: API_ENDPOINTS.catalog,
	timeout: 10000,
})

export const fetchNotes = async (catalogItemId: string): Promise<Note[]> => {
	const response = await client.get(`weeds/${catalogItemId}/notes`).json<ApiResponse<Note[]>>()

	return handleApiResponse(response)
}

export const createNote = async (catalogItemId: string, text: string): Promise<Note> => {
	const response = await client
		.post(`weeds/${catalogItemId}/notes`, {
			json: { note: text },
		})
		.json<ApiResponse<Note>>()

	return handleApiResponse(response)
}

export const updateNote = async (note: Note): Promise<Note> => {
	const response = await client
		.put(`notes/${note.id}`, {
			json: { note: note.note },
		})
		.json<ApiResponse<Note>>()

	return handleApiResponse(response)
}

export const deleteNote = async (noteId: number): Promise<void> => {
	await client.delete(`notes/${noteId}`)
}
