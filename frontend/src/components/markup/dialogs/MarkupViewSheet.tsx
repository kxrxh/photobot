import type React from "react"
import { useEffect, useRef } from "react"
import { createPortal } from "react-dom"
import type { Markup } from "@/api/markup/types"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"

interface MarkupViewSheetProps {
	isOpen: boolean
	onClose: () => void
	markup: Markup | null
}

const MarkupViewSheet: React.FC<MarkupViewSheetProps> = ({ isOpen, onClose, markup }) => {
	const dialogRef = useRef<HTMLDialogElement>(null)

	useEffect(() => {
		const dialog = dialogRef.current
		if (dialog) {
			if (isOpen) {
				dialog.showModal()
			} else {
				dialog.close()
			}
		}
	}, [isOpen])

	useEffect(() => {
		const dialog = dialogRef.current
		if (!dialog) return

		const handleClose = () => {
			if (isOpen) {
				onClose()
			}
		}

		dialog.addEventListener("close", handleClose)
		return () => {
			dialog.removeEventListener("close", handleClose)
		}
	}, [isOpen, onClose])

	if (!markup) {
		return null
	}

	return createPortal(
		<dialog ref={dialogRef} className="modal backdrop-blur-xs">
			<div className="max-w-lg modal-box">
				<div className="flex justify-between items-center gap-3 mb-6">
					<div className="min-w-0">
						<h3 className="text-xl font-semibold text-base-content">Список фракций</h3>
						<p className="mt-1 text-sm text-base-content/70">{markup.fractions.length} фр.</p>
					</div>
					<SheetHeaderCloseButton onClick={onClose} aria-label="Закрыть" />
				</div>

				<div className="overflow-y-auto space-y-3 max-h-96">
					{markup.fractions.map((fraction, index) => (
						<div
							key={fraction.id}
							className="p-4 bg-linear-to-r rounded-xl border transition-all duration-200 group from-base-100 to-base-50 hover:from-primary/5 hover:to-primary/10 border-base-200 hover:border-primary/20 hover:shadow-sm"
						>
							<div className="flex justify-between items-center">
								<div className="flex gap-3 items-center">
									<div className="flex justify-center items-center w-8 h-8 rounded-lg transition-colors duration-200 bg-primary/10 group-hover:bg-primary/20">
										<span className="text-sm font-medium text-primary">{index + 1}</span>
									</div>
									<div>
										<h4 className="font-medium transition-colors duration-200 text-base-content group-hover:text-primary/90">
											{fraction.name}
										</h4>
									</div>
								</div>
								<div className="flex gap-2 items-center">
									<div className="font-medium badge badge-primary badge-outline">
										{fraction.object_ids.length}
									</div>
								</div>
							</div>
						</div>
					))}
				</div>

				<div className="flex justify-center pt-4 mt-6 w-full border-t border-base-200">
					<button type="button" className="w-full btn btn-primary" onClick={onClose}>
						Закрыть
					</button>
				</div>
			</div>
			<button type="button" className="modal-backdrop" onClick={onClose} aria-label="Закрыть" />
		</dialog>,
		document.body
	)
}

export default MarkupViewSheet
