import { useQuery } from "@tanstack/react-query"
import type React from "react"
import { useEffect, useRef, useState } from "react"
import { createPortal } from "react-dom"
import { listMarkups } from "@/api/markup"
import type { Markup, SaveMarkup } from "@/api/markup/types"
import { queryKeys } from "@/api/queryKeys"
import LoadingSkeleton from "@/components/common/ui/LoadingSkeleton"
import SearchAndCreateBar from "@/components/common/ui/SearchAndCreateBar"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"
import type { FractionType } from "@/hooks/useFractions"
import type { SelectedAnalysis } from "@/routes/_authenticated/markup"
import CreateForm from "./CreateForm"
import MarkupCard from "./MarkupCard"
import useDialogState from "./useDialogState"

interface MarkupSaveSheetProps {
	isOpen: boolean
	onClose: () => void
	onSave: (markup: SaveMarkup, id?: string) => void
	currentMarkup?: Markup
	userRoles?: string[]
	fractions: FractionType[]
	analyses: SelectedAnalysis[]
}

const MarkupSaveSheet: React.FC<MarkupSaveSheetProps> = ({
	isOpen,
	onClose,
	onSave,
	fractions,
	analyses,
}) => {
	const dialogRef = useRef<HTMLDialogElement>(null)
	const markupsListRef = useRef<HTMLDivElement>(null)

	const {
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
	} = useDialogState(isOpen)

	const [debouncedSearchTerm, setDebouncedSearchTerm] = useState(searchTerm)
	useEffect(() => {
		const handler = setTimeout(() => {
			setDebouncedSearchTerm(searchTerm)
		}, 500)
		return () => clearTimeout(handler)
	}, [searchTerm])

	const {
		data: markups = [],
		isLoading: isMarkupsLoading,
		isFetching: isMarkupsFetching,
	} = useQuery({
		queryKey: queryKeys.markups({ name: debouncedSearchTerm }),
		queryFn: () => listMarkups({ name: debouncedSearchTerm }),
		staleTime: 5 * 60 * 1000,
		gcTime: 5 * 60 * 1000,
		enabled: isOpen && !showCreateForm,
	})
	const loading = isMarkupsLoading || isMarkupsFetching

	useEffect(() => {
		if (dialogRef.current) {
			if (isOpen) {
				dialogRef.current.showModal()
				if (markupsListRef.current) {
					markupsListRef.current.scrollTop = 0
				}
			} else {
				dialogRef.current.close()
			}
		}
	}, [isOpen])

	// Ensure state resets if dialog is closed by ESC or backdrop
	useEffect(() => {
		const dialog = dialogRef.current
		if (!dialog) return

		const handleDialogClose = () => {
			onClose()
		}
		dialog.addEventListener("close", handleDialogClose)

		return () => {
			dialog.removeEventListener("close", handleDialogClose)
		}
	}, [onClose])

	const toggleCardExpansion = (markupId: string) => {
		setExpandedCards((prev) => {
			const newSet = new Set<string>()
			if (!prev.has(markupId)) {
				newSet.add(markupId)
			}

			if (!newSet.has(markupId) && (editingMarkup as string | null) === markupId) {
				setEditingMarkup(null)
				setEditName("")
			}

			return newSet
		})
	}

	const handleStartEdit = (markup: Markup) => {
		setEditingMarkup(markup.id)
		setEditName(markup.name)
	}

	const handleCancelEdit = () => {
		setEditingMarkup(null)
		setEditName("")
	}

	const handleReplaceMarkup = (markup: Markup) => {
		const updatedMarkup: SaveMarkup = {
			name: editName.trim(),
			fractions: fractions.map((fraction) => ({
				object_ids: fraction.objects.map((o) => o.id),
				name: fraction.name,
			})),
			analyses_ids: analyses.map((a) => a.id),
		}
		onSave(updatedMarkup, markup.id.toString())
		onClose()
	}

	const handleCreateNew = () => {
		setShowCreateForm(true)
		setSelectedMarkup(null)
	}

	const handleCancelCreate = () => {
		setShowCreateForm(false)
		setNewMarkup({
			name: "",
		})
	}

	const handleSaveNew = () => {
		if (!newMarkup.name.trim()) {
			alert("Пожалуйста, введите название разметки")
			return
		}

		const newMarkupData: SaveMarkup = {
			name: newMarkup.name,
			fractions: fractions.map((fraction) => ({
				object_ids: fraction.objects.map((o) => o.id),
				name: fraction.name,
			})),
			analyses_ids: analyses.map((a) => a.id),
		}

		onSave(newMarkupData)
		onClose()
	}

	return createPortal(
		<dialog ref={dialogRef} className="modal">
			<div className="flex fixed inset-0 z-50 justify-center items-center">
				<div className="flex relative flex-col w-full h-full bg-base-100">
					<div className="flex sticky top-0 z-10 justify-between items-center p-2 border-b border-base-200 bg-base-100">
						<h2 className="text-xl font-bold">
							{showCreateForm ? "Новая разметка" : "Заменить разметку"}
						</h2>
						<div className="flex gap-4 items-center">
							<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
						</div>
					</div>

					{showCreateForm ? (
						<CreateForm
							form={newMarkup}
							onFormChange={setNewMarkup}
							onSave={handleSaveNew}
							onCancel={handleCancelCreate}
						/>
					) : (
						<>
							<SearchAndCreateBar
								searchTerm={searchTerm}
								onSearchChange={setSearchTerm}
								onCreateNew={handleCreateNew}
								searchPlaceholder="Поиск разметки..."
							/>

							<div className="overflow-y-auto flex-1 p-2 space-y-4" ref={markupsListRef}>
								{loading ? (
									<LoadingSkeleton itemCount={3} />
								) : (
									<>
										{markups
											.sort((a, b) => {
												if (a.id === selectedMarkup?.id) return -1
												if (b.id === selectedMarkup?.id) return 1
												return 0
											})
											.map((markup) => (
												<MarkupCard
													key={markup.id}
													markup={markup}
													isSelected={selectedMarkup?.id === markup.id}
													isExpanded={expandedCards.has(markup.id)}
													isEditing={editingMarkup === markup.id}
													editName={editName}
													onToggleExpansion={() => toggleCardExpansion(markup.id)}
													onStartEdit={() => handleStartEdit(markup)}
													onCancelEdit={handleCancelEdit}
													onNameChange={setEditName}
													onReplace={() => handleReplaceMarkup(markup)}
												/>
											))}

										{markups.length === 0 && (
											<div className="p-8 text-center text-base-content/70">
												Нет разметок, соответствующих поиску.
											</div>
										)}
									</>
								)}
							</div>
						</>
					)}
				</div>
			</div>

			<div className="modal-backdrop" />
		</dialog>,
		document.body
	)
}

export default MarkupSaveSheet
