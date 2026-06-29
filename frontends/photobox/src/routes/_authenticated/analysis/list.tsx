import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import type { Analysis } from "@/api/analysis/types"
import AnalysisSelectorDialog from "@/components/analysis/selectors/AnalysisSelector"

export const Route = createFileRoute("/_authenticated/analysis/list")({
	component: RouteComponent,
})

function RouteComponent() {
	const navigate = Route.useNavigate()
	const [selectedAnalysisIds, setSelectedAnalysisIds] = useState<string[]>([])

	const handleClose = () => {
		navigate({ to: "/menu" })
	}

	const handleAddAnalysis = (analysis: Analysis) => {
		setSelectedAnalysisIds((prev) => [...prev, analysis.id])
	}

	const handleRemoveAnalysis = (analysisId: string) => {
		setSelectedAnalysisIds((prev) => prev.filter((id) => id !== analysisId))
	}

	const handleRemoveAllAnalyses = () => {
		setSelectedAnalysisIds([])
	}

	return (
		<div className="flex min-h-0 flex-1 flex-col overflow-hidden bg-base-100">
			<AnalysisSelectorDialog
				isOpen
				usesGlobalMenuBackButton
				onClose={handleClose}
				selectedAnalysisIds={selectedAnalysisIds}
				onAddAnalysis={handleAddAnalysis}
				onRemoveAnalysis={handleRemoveAnalysis}
				onRemoveAllAnalyses={handleRemoveAllAnalyses}
				hasAddedAnalyses={selectedAnalysisIds.length > 0}
				selectionMode="multiple"
				mode="view-only"
				onOpenCreateDialog={() =>
					navigate({ to: "/analysis/create", search: { openRequest: undefined } })
				}
			/>
		</div>
	)
}
