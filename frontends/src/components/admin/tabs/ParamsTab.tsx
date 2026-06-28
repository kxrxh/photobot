import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { FiPlus, FiTrash2 } from "react-icons/fi"
import { createParam, deleteParam, getParams } from "@/api/params"
import type { ClassificationParam } from "@/api/params/types"
import { queryKeys } from "@/api/queryKeys"
import AddParameterAlert from "@/components/admin/dialogs/AddParameterAlert"

function ParamsTab() {
	const queryClient = useQueryClient()
	const { data: params, isLoading } = useQuery({
		queryKey: queryKeys.params,
		queryFn: getParams,
	})

	const [newParamName, setNewParamName] = useState("")
	const [isAddModalOpen, setIsAddModalOpen] = useState(false)

	const { mutate: createParamMutation, isPending: isCreating } = useMutation({
		mutationFn: createParam,
		onSuccess: () => {
			setNewParamName("")
			queryClient.invalidateQueries({ queryKey: queryKeys.params })
		},
	})

	const { mutate: deleteParamMutation } = useMutation({
		mutationFn: deleteParam,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: queryKeys.params })
		},
	})

	const handleOpenAdd = () => setIsAddModalOpen(true)
	const handleCloseAdd = () => {
		setIsAddModalOpen(false)
		setNewParamName("")
	}
	const handleConfirmAdd = () => {
		const value = newParamName.trim()
		if (!value) return
		createParamMutation(value, { onSuccess: handleCloseAdd })
	}

	const handleDelete = (id: string) => {
		deleteParamMutation(id)
	}

	return (
		<div className="p-2 space-y-4 h-full">
			<button
				type="button"
				className="btn btn-primary btn-block gap-2 p-2"
				onClick={handleOpenAdd}
				disabled={isCreating}
			>
				<FiPlus className="w-4 h-4" /> Добавить параметр
			</button>

			{!isLoading ? (
				<div className="space-y-2 overflow-y-auto p-2">
					{params?.length ? (
						params.map((p: ClassificationParam) => (
							<div key={p.id} className="card bg-base-100 shadow-sm">
								<div className="card-body p-4">
									<div className="flex justify-between items-start">
										<div className="flex-1">
											<h3 className="font-medium text-base-content">{p.name}</h3>
											<p className="text-sm text-base-content/60 mt-1">
												{new Date(p.created_at).toLocaleTimeString("ru-RU", {
													day: "2-digit",
													month: "2-digit",
													year: "numeric",
													hour: "2-digit",
													minute: "2-digit",
													second: "2-digit",
												})}
											</p>
										</div>
										<div className="flex gap-2 ml-4">
											<button
												type="button"
												className="btn btn-ghost btn-sm btn-square text-error"
												onClick={() => handleDelete(p.id)}
											>
												<FiTrash2 className="w-4 h-4" />
											</button>
										</div>
									</div>
								</div>
							</div>
						))
					) : (
						<div className="text-center py-8 text-base-content/60">Здесь пока пусто</div>
					)}
				</div>
			) : (
				<div className="flex justify-center py-8">
					<span className="loading loading-spinner loading-md"></span>
				</div>
			)}

			<AddParameterAlert
				isOpen={isAddModalOpen}
				onClose={handleCloseAdd}
				onConfirm={handleConfirmAdd}
				newParamName={newParamName}
				setNewParamName={setNewParamName}
			/>
		</div>
	)
}

export default ParamsTab
