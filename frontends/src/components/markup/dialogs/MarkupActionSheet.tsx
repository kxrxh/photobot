import { useId } from "react"
import { FaSave, FaUpload } from "react-icons/fa"

interface MarkupActionSheetProps {
	isOpen: boolean
	onClose: () => void
	onOpenMarkupSelector: () => void
	onOpenSaveDialog: () => void
}

const MarkupActionSheet: React.FC<MarkupActionSheetProps> = ({
	isOpen,
	onClose,
	onOpenMarkupSelector,
	onOpenSaveDialog,
}) => {
	const handleSaveClick = () => {
		onClose()
		onOpenSaveDialog()
	}

	return (
		<dialog
			id={`markup_modal_${useId()}`}
			className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}
		>
			<div className="max-w-md modal-box">
				<h3 className="mb-6 text-xl font-semibold">Разметка</h3>
				<div className="space-y-4">
					<button
						className="flex gap-4 items-center p-4 w-full text-left rounded-xl border transition-all duration-200 transform border-base-300 active:scale-95 active:bg-primary/20"
						type="button"
						onClick={handleSaveClick}
					>
						<div className="p-2 rounded-lg transition-transform duration-200 bg-primary/10">
							<FaSave className="w-6 h-6 text-primary" />
						</div>
						<div className="flex flex-col items-start">
							<span className="text-sm font-medium">Сохранить разметку</span>
							<span className="text-xs text-base-content/70">Сохранить текущую разметку</span>
						</div>
					</button>
					<button
						className="flex gap-4 items-center p-4 w-full text-left rounded-xl border transition-all duration-200 transform border-base-300 active:scale-95 active:bg-primary/20"
						type="button"
						onClick={() => {
							onClose()
							onOpenMarkupSelector()
						}}
					>
						<div className="p-2 rounded-lg transition-transform duration-200 bg-primary/10">
							<FaUpload className="w-6 h-6 text-primary" />
						</div>
						<div className="flex flex-col items-start">
							<span className="text-sm font-medium">Загрузить разметку</span>
							<span className="text-xs text-left text-base-content/70">
								Загрузить разметку с фракциями и объектами
							</span>
						</div>
					</button>
				</div>
				<div className="mt-6 modal-action">
					<button
						className="w-full transition-transform duration-200 btn active:scale-95"
						onClick={onClose}
						type="button"
					>
						Закрыть
					</button>
				</div>
			</div>
		</dialog>
	)
}

export default MarkupActionSheet
