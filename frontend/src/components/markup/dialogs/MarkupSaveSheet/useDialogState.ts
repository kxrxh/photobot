import { useEffect, useState } from "react"
import type { Markup } from "@/api/markup/types"
import type { NewMarkupForm } from "./CreateForm"

const useDialogState = (isOpen: boolean) => {
	const [searchTerm, setSearchTerm] = useState("")
	const [selectedMarkup, setSelectedMarkup] = useState<Markup | null>(null)
	const [showCreateForm, setShowCreateForm] = useState(false)
	const [newMarkup, setNewMarkup] = useState<NewMarkupForm>({
		name: "",
	})
	const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set())
	const [editingMarkup, setEditingMarkup] = useState<string | null>(null)
	const [editName, setEditName] = useState("")

	useEffect(() => {
		if (isOpen) {
			setSearchTerm("")
			setSelectedMarkup(null)
			setShowCreateForm(false)
			setNewMarkup({
				name: "",
			})
			setExpandedCards(new Set())
			setEditingMarkup(null)
			setEditName("")
		}
	}, [isOpen])

	return {
		searchTerm,
		setSearchTerm,
		selectedMarkup,
		setSelectedMarkup,
		showCreateForm,
		setShowCreateForm,
		newMarkup,
		setNewMarkup,
		expandedCards,
		setExpandedCards,
		editingMarkup,
		setEditingMarkup,
		editName,
		setEditName,
	}
}

export default useDialogState
