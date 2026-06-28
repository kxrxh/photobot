import type React from "react"
import { FaCheck } from "react-icons/fa"
import type { SimpleObject } from "@/api/analysis/types"

interface AnalysisConfirmationAlertProps {
	showConfirmModal: boolean
	setShowConfirmModal: (show: boolean) => void
	excludedObjects: string[]
	objects: SimpleObject[] | undefined
	onConfirmSave: () => void
	isSaving?: boolean
}

const AnalysisConfirmationAlert: React.FC<AnalysisConfirmationAlertProps> = ({
	showConfirmModal,
	setShowConfirmModal,
	excludedObjects,
	objects,
	onConfirmSave,
	isSaving = false,
}) => {
	return (
		<dialog className={`modal backdrop-blur-xs ${showConfirmModal ? "modal-open" : ""}`}>
			<div className="modal-box">
				<h3 className="text-lg font-bold">Подтверждение сохранения анализа</h3>
				<div className="py-4">
					<p className="mb-2">
						{excludedObjects.length > 0
							? `Вы исключили ${excludedObjects.length} из ${objects?.length || 0} объектов.`
							: `Все ${objects?.length || 0} объектов будут сохранены.`}
					</p>
					<p>Вы уверены, что хотите сохранить анализ?</p>
				</div>
				<div className="modal-action">
					<form method="dialog" className="flex flex-row gap-2 w-full">
						<button
							className="flex-1 btn"
							onClick={() => setShowConfirmModal(false)}
							type="button"
							disabled={isSaving}
						>
							Отмена
						</button>
						<button
							className="flex-1 btn btn-primary"
							onClick={onConfirmSave}
							type="button"
							disabled={isSaving}
						>
							{isSaving ? (
								<>
									<span className="loading loading-spinner loading-sm"></span>
									<span className="ml-2">Сохранение...</span>
								</>
							) : (
								<>
									<FaCheck className="w-4 h-4 mr-2" />
									Сохранить
								</>
							)}
						</button>
					</form>
				</div>
			</div>
		</dialog>
	)
}

export default AnalysisConfirmationAlert
