import { useEffect, useRef, useState } from "react"
import type { Classification } from "@/api/classification/types"
import type { NewClassificationForm } from "./CreateForm"

export const useDialogState = (isOpen: boolean, classifications: Classification[]) => {
	const [searchTerm, setSearchTerm] = useState("")
	const [selectedClassification, setSelectedClassification] = useState<Classification | null>(null)
	const [showCreateForm, setShowCreateForm] = useState(false)
	const [newClassification, setNewClassification] = useState<NewClassificationForm>({ name: "" })
	const [expandedCards, setExpandedCards] = useState<Set<string>>(new Set())
	const [selectedParams, setSelectedParams] = useState<Record<string, Set<string>>>({})

	const prevIsOpen = useRef(isOpen)

	useEffect(() => {
		if (isOpen && !prevIsOpen.current) {
			setSearchTerm("")
			setSelectedClassification(null)
			setShowCreateForm(false)
			setNewClassification({ name: "" })
			setExpandedCards(new Set())
			// Initialize selected params for each classification
			const initialParams: Record<string, Set<string>> = {}
			for (const classification of classifications) {
				initialParams[classification.id] = new Set(["all"])
			}
			setSelectedParams(initialParams)
		}
		prevIsOpen.current = isOpen
	}, [isOpen, classifications])

	return {
		searchTerm,
		setSearchTerm,
		selectedClassification,
		setSelectedClassification,
		showCreateForm,
		setShowCreateForm,
		newClassification,
		setNewClassification,
		expandedCards,
		setExpandedCards,
		selectedParams,
		setSelectedParams,
	}
}

export const useParameterToggle = (
	setSelectedParams: React.Dispatch<React.SetStateAction<Record<string, Set<string>>>>
) => {
	const toggleParameter = (classificationId: string, paramId: string) => {
		setSelectedParams((prev) => {
			const newParams = { ...prev }
			const currentSet = new Set(newParams[classificationId] || new Set())

			if (paramId === "all") {
				if (currentSet.has("all")) {
					newParams[classificationId] = new Set()
				} else {
					newParams[classificationId] = new Set(["all"])
				}
			} else {
				if (currentSet.has("all")) {
					currentSet.delete("all")
				}

				// Toggle the selected parameter
				if (currentSet.has(paramId)) {
					currentSet.delete(paramId)
				} else {
					currentSet.add(paramId)
				}

				newParams[classificationId] = currentSet
			}

			return newParams
		})
	}

	return { toggleParameter }
}
