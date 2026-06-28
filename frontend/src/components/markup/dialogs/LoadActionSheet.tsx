import { useId } from "react"
import { FaBook, FaDatabase, FaImage } from "react-icons/fa"

interface LoadActionSheetProps {
	isOpen: boolean
	onClose: () => void
	openAnalysisModal: () => void
	openCatalogSelectorModal: () => void
	openPhotoModal: () => void
}

const LoadActionSheet: React.FC<LoadActionSheetProps> = ({
	isOpen,
	onClose,
	openAnalysisModal,
	openCatalogSelectorModal,
	openPhotoModal,
}) => {
	const modalId = useId()

	return (
		<dialog id={modalId} className={`modal backdrop-blur-xs ${isOpen ? "modal-open" : ""}`}>
			<div className="max-w-md modal-box">
				<h3 className="mb-6 text-xl font-semibold">Загрузка данных</h3>
				<div className="space-y-4">
					<button
						className="flex gap-4 items-center p-4 w-full text-left rounded-xl border transition-all duration-200 border-base-300 active:scale-95 active:bg-primary/20 transform"
						onClick={() => {
							onClose()
							openAnalysisModal()
						}}
						type="button"
					>
						<div className="p-2 rounded-lg bg-primary/10 transition-transform duration-200">
							<FaDatabase className="w-6 h-6 text-primary" />
						</div>
						<div className="flex flex-col items-start">
							<span className="text-sm font-medium">Анализ</span>
							<span className="text-xs text-base-content/70">Загрузка объектов из анализа</span>
						</div>
					</button>
					<button
						className="flex gap-4 items-center p-4 w-full text-left rounded-xl border transition-all duration-200 border-base-300 active:scale-95 active:bg-primary/20 transform"
						onClick={() => {
							onClose()
							openCatalogSelectorModal()
						}}
						type="button"
					>
						<div className="p-2 rounded-lg bg-primary/10 transition-transform duration-200">
							<FaBook className="w-6 h-6 text-primary" />
						</div>
						<div className="flex flex-col items-start">
							<span className="text-sm font-medium">Запись из каталога</span>
							<span className="text-xs text-base-content/70">Загрузка фракции целиком</span>
						</div>
					</button>
					<button
						className="flex gap-4 items-center p-4 w-full text-left rounded-xl border transition-all duration-200 border-base-300 active:scale-95 active:bg-primary/20 transform"
						onClick={() => {
							onClose()
							openPhotoModal()
						}}
						type="button"
					>
						<div className="p-2 rounded-lg bg-primary/10 transition-transform duration-200">
							<FaImage className="w-6 h-6 text-primary" />
						</div>
						<div className="flex flex-col items-start">
							<span className="text-sm font-medium">Из фото</span>
							<span className="text-xs text-base-content/70">Загрузка объектов из фотографии</span>
						</div>
					</button>
				</div>
				<div className="mt-6 modal-action">
					<button
						className="w-full btn active:scale-95 transition-transform duration-200"
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

export default LoadActionSheet
