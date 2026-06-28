import type React from "react"
import { useId } from "react"
import { ModalSelect } from "@/components/common/ui/ModalSelect"

/** Matches AnalysisCreationSheet `UploadTab` section titles */
const SECTION_TITLE_CLASS =
	"mb-1.5 block text-xs font-semibold uppercase tracking-wide text-base-content/45"

interface Classification {
	id: string
	name: string
}

interface GeneralInfoTabProps {
	productName: string
	onNameChange: (name: string) => void
	primaryClassifications: Classification[]
	secondaryClassifications: Classification[]
	tertiaryClassifications?: Classification[]
	selectedPrimaryClassification: string | null
	onPrimaryClassificationChange: (id: string | null) => void
	selectedSecondaryClassification: string | null
	onSecondaryClassificationChange: (id: string | null) => void
	selectedTertiaryClassification?: string | null
	onTertiaryClassificationChange?: (id: string | null) => void
	description: string
	onDescriptionChange: (description: string) => void
	harmfulness: string
	onHarmfulnessChange: (harmfulness: string) => void
	onDelete?: () => void
	/** Показать подпись «Редактирование» над заголовком (страница редактирования) */
	editModeLabel?: string
	/** Только просмотр (например, заявка на рассмотрении) */
	readOnly?: boolean
}

const GeneralInfoTab: React.FC<GeneralInfoTabProps> = ({
	productName,
	onNameChange,
	primaryClassifications,
	secondaryClassifications,
	tertiaryClassifications,
	selectedPrimaryClassification,
	onPrimaryClassificationChange,
	selectedSecondaryClassification,
	onSecondaryClassificationChange,
	selectedTertiaryClassification,
	onTertiaryClassificationChange,
	description,
	onDescriptionChange,
	harmfulness,
	onHarmfulnessChange,
	onDelete,
	editModeLabel,
	readOnly = false,
}) => {
	const productNameId = useId()
	const primaryClassificationId = useId()
	const secondaryClassificationId = useId()
	const tertiaryClassificationId = useId()
	const productDescriptionId = useId()
	const productHarmfulnessId = useId()
	const sectionTitleId = useId()

	return (
		<div className="animate-fadeIn space-y-6 px-4 pb-8">
			<section aria-labelledby={sectionTitleId} className="flex flex-col gap-3">
				<div className="pt-1">
					{editModeLabel ? (
						<p className="mb-1.5 text-xs font-medium text-base-content/55">{editModeLabel}</p>
					) : null}
					<h2 id={sectionTitleId} className={SECTION_TITLE_CLASS}>
						Основная информация
					</h2>
				</div>

				<div className="space-y-4 rounded-2xl border border-base-200 bg-base-200/25 p-4">
					<div className="space-y-2">
						<label className="label" htmlFor={productNameId}>
							Название
						</label>
						<input
							id={productNameId}
							type="text"
							placeholder="Введите название"
							className="w-full input input-bordered focus:input-primary"
							value={productName}
							readOnly={readOnly}
							disabled={readOnly}
							onChange={(e) => onNameChange(e.target.value)}
						/>
					</div>

					<div className="space-y-2">
						<label className="label" htmlFor={primaryClassificationId}>
							Основная классификация
						</label>
						<ModalSelect
							id={primaryClassificationId}
							title="Основная классификация"
							placeholder="Выберите основную классификацию"
							options={primaryClassifications.map((c) => ({ value: c.id, label: c.name }))}
							value={selectedPrimaryClassification ?? ""}
							onChange={(v) => onPrimaryClassificationChange(v === "" ? null : v)}
							disabled={readOnly}
						/>
					</div>

					<div className="space-y-2">
						<label className="label" htmlFor={secondaryClassificationId}>
							Дополнительная классификация
						</label>
						<ModalSelect
							id={secondaryClassificationId}
							title="Дополнительная классификация"
							placeholder="Выберите дополнительную классификацию"
							options={secondaryClassifications.map((c) => ({ value: c.id, label: c.name }))}
							value={selectedSecondaryClassification ?? ""}
							onChange={(v) => onSecondaryClassificationChange(v === "" ? null : v)}
							disabled={readOnly || !selectedPrimaryClassification}
						/>
					</div>

					{typeof onTertiaryClassificationChange === "function" && (
						<div className="space-y-2">
							<label className="label" htmlFor={tertiaryClassificationId}>
								Подклассификация
							</label>
							<ModalSelect
								id={tertiaryClassificationId}
								title="Подклассификация"
								placeholder="Выберите подклассификацию"
								options={(tertiaryClassifications ?? []).map((c) => ({
									value: c.id,
									label: c.name,
								}))}
								value={selectedTertiaryClassification ?? ""}
								onChange={(v) => onTertiaryClassificationChange(v === "" ? null : v)}
								disabled={
									readOnly || !selectedPrimaryClassification || !selectedSecondaryClassification
								}
							/>
						</div>
					)}

					<div className="space-y-2">
						<label className="label" htmlFor={productDescriptionId}>
							Описание
						</label>
						<textarea
							id={productDescriptionId}
							className="w-full h-24 resize-none textarea textarea-bordered focus:textarea-primary"
							placeholder="Описание..."
							value={description}
							readOnly={readOnly}
							disabled={readOnly}
							onChange={(e) => onDescriptionChange(e.target.value)}
						/>
					</div>

					<div className="space-y-2">
						<label className="label" htmlFor={productHarmfulnessId}>
							Вредоносность
						</label>
						<textarea
							id={productHarmfulnessId}
							className="w-full h-24 resize-none textarea textarea-bordered focus:textarea-primary"
							placeholder="Вредоносность..."
							value={harmfulness}
							readOnly={readOnly}
							disabled={readOnly}
							onChange={(e) => onHarmfulnessChange(e.target.value)}
						/>
					</div>
				</div>
			</section>

			{onDelete && !readOnly && (
				<div className="pt-4 space-y-2">
					<p className="text-sm text-center">Внимание! Удаление записи необратимо.</p>
					<button type="button" className="w-full btn btn-primary btn-sm" onClick={onDelete}>
						Удалить запись
					</button>
				</div>
			)}
		</div>
	)
}

export default GeneralInfoTab
