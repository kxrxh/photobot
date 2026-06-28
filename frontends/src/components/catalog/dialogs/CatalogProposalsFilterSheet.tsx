import { useEffect, useId, useState } from "react"
import type { ProposalStatus } from "@/api/catalog/types"
import FilterBottomSheet from "@/components/common/dialogs/FilterBottomSheet"
import { ModalSelect } from "@/components/common/ui/ModalSelect"
import { STATUS_OPTIONS } from "../components/CatalogProposalsShared"

export type CatalogProposalsFilterSheetProps = {
	isOpen: boolean
	onClose: () => void
	status?: ProposalStatus | ""
	sortOrder: "asc" | "desc"
	authorId: string
	onApply: (filters: {
		status?: ProposalStatus | ""
		requester: string
		sort: "asc" | "desc"
	}) => void
	onClear: () => void
	/** When true, hide status filter — moderator sees only "on moderation" (submitted) proposals */
	moderatorMode?: boolean
}

const CatalogProposalsFilterSheet = ({
	isOpen,
	onClose,
	status = "",
	sortOrder,
	authorId,
	onApply,
	onClear,
	moderatorMode = false,
}: CatalogProposalsFilterSheetProps) => {
	const statusId = useId()
	const sortId = useId()
	const authorIdInput = useId()
	const [localStatus, setLocalStatus] = useState<ProposalStatus | "">(status)
	const [localSortOrder, setLocalSortOrder] = useState<"asc" | "desc">(sortOrder)
	const [localAuthorId, setLocalAuthorId] = useState(authorId)

	useEffect(() => {
		if (!isOpen) return
		setLocalStatus(status)
		setLocalSortOrder(sortOrder)
		setLocalAuthorId(authorId)
	}, [authorId, isOpen, sortOrder, status])

	const handleClear = () => {
		onClear()
		if (!moderatorMode) setLocalStatus(status)
		setLocalSortOrder("desc")
		setLocalAuthorId("")
	}

	const handleApply = () => {
		onApply({
			...(moderatorMode ? {} : { status: localStatus }),
			requester: localAuthorId.trim(),
			sort: localSortOrder,
		})
	}

	return (
		<FilterBottomSheet
			isOpen={isOpen}
			onClose={onClose}
			title="Фильтры предложений"
			onClear={handleClear}
			onApply={handleApply}
		>
			<div className="space-y-4">
				{!moderatorMode && (
					<div className="form-control">
						<label htmlFor={statusId} className="label">
							<span className="font-medium label-text text-base-content">Статус</span>
						</label>
						<ModalSelect
							id={statusId}
							title="Статус"
							placeholder="Все статусы"
							options={STATUS_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
							value={localStatus}
							onChange={(v) => setLocalStatus((v || "") as ProposalStatus | "")}
						/>
					</div>
				)}

				<div className="form-control">
					<label htmlFor={sortId} className="label">
						<span className="font-medium label-text text-base-content">Сортировка по дате</span>
					</label>
					<ModalSelect
						id={sortId}
						title="Сортировка по дате"
						placeholder="Сортировка по дате"
						options={[
							{ value: "desc", label: "Новые сверху" },
							{ value: "asc", label: "Старые сверху" },
						]}
						value={localSortOrder}
						onChange={(v) => setLocalSortOrder(v === "asc" ? "asc" : "desc")}
						clearable={false}
					/>
				</div>

				<div className="form-control">
					<label htmlFor={authorIdInput} className="label">
						<span className="font-medium label-text text-base-content">ID автора</span>
					</label>
					<input
						id={authorIdInput}
						type="number"
						min={0}
						value={localAuthorId}
						onChange={(e) => setLocalAuthorId(e.target.value)}
						className="input input-bordered w-full"
						placeholder="99999"
					/>
				</div>
			</div>
		</FilterBottomSheet>
	)
}

export default CatalogProposalsFilterSheet
