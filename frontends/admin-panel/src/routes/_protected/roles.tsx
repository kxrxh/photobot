import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { FaEdit, FaPlus, FaTrash } from "react-icons/fa"
import { FaShieldHalved } from "react-icons/fa6"
import { PageError, PageLoading, Toast } from "@/components/ui"
import { createRole, deleteRole, getRoles, updateRole } from "@/features/roles"
import { useConfirmDelete, useToast } from "@/hooks"
import type { Role } from "@/types/role"

export const Route = createFileRoute("/_protected/roles")({
	component: RolesComponent,
})

function RolesComponent() {
	const queryClient = useQueryClient()
	const { toast, showSuccess, showError, clear } = useToast()
	const [newRoleName, setNewRoleName] = useState("")
	const [editingRoleId, setEditingRoleId] = useState<number | null>(null)
	const [editingRoleName, setEditingRoleName] = useState("")

	const { secondsLeft, handleDelete, isPending } = useConfirmDelete<Role>({
		delaySeconds: 5,
		onConfirm: (roleId) => {
			deleteRoleMutation.mutate(roleId)
		},
	})

	const {
		data: roles = [],
		isLoading,
		error,
		refetch,
	} = useQuery<Role[]>({
		queryKey: ["roles"],
		queryFn: getRoles,
	})

	const createRoleMutation = useMutation({
		mutationFn: (role: { name: string }) => createRole(role),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["roles"] })
			setNewRoleName("")
			showSuccess("Роль успешно создана")
		},
		onError: (err: Error) => {
			showError(err.message)
		},
	})

	const updateRoleMutation = useMutation({
		mutationFn: ({ roleId, name }: { roleId: number; name: string }) =>
			updateRole(roleId, { name }),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["roles"] })
			setEditingRoleId(null)
			setEditingRoleName("")
		},
	})

	const deleteRoleMutation = useMutation({
		mutationFn: (roleId: number) => deleteRole(roleId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["roles"] })
		},
	})

	const handleCreateRole = () => {
		if (!newRoleName.trim()) return
		const existing = roles.find((r) => r.name.toLowerCase() === newRoleName.trim().toLowerCase())
		if (existing) {
			showError("Роль с таким названием уже существует")
			return
		}
		createRoleMutation.mutate({ name: newRoleName.trim() })
	}

	const handleEditRole = (role: Role) => {
		setEditingRoleId(role.id)
		setEditingRoleName(role.name)
	}

	const handleSaveRole = () => {
		if (editingRoleName.trim() && editingRoleId !== null) {
			updateRoleMutation.mutate({
				roleId: editingRoleId,
				name: editingRoleName.trim(),
			})
		}
	}

	const handleCancelEdit = () => {
		setEditingRoleId(null)
		setEditingRoleName("")
	}

	if (isLoading) {
		return <PageLoading message="Загрузка ролей…" />
	}

	if (error) {
		return (
			<PageError
				message={`Ошибка при загрузке ролей: ${error.message}`}
				onRetry={() => void refetch()}
			/>
		)
	}

	return (
		<div className="container mx-auto p-6 max-w-7xl">
			<div className="mb-8">
				<div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
					<div>
						<h1 className="text-4xl font-bold text-base-content mb-2">Роли</h1>
						<p className="text-base-content/70">Количество ролей: {roles.length}</p>
					</div>
					<div className="join w-full max-w-md max-sm:join-vertical">
						<input
							type="text"
							placeholder="Название новой роли"
							className="input input-bordered join-item flex-1 min-w-0"
							value={newRoleName}
							onChange={(e) => setNewRoleName(e.target.value)}
							onKeyDown={(e) => e.key === "Enter" && handleCreateRole()}
						/>
						<button
							type="button"
							className="btn btn-primary join-item gap-2"
							onClick={handleCreateRole}
							disabled={!newRoleName.trim() || createRoleMutation.isPending}
						>
							{createRoleMutation.isPending ? (
								<span className="loading loading-spinner loading-sm" />
							) : (
								<FaPlus className="h-4 w-4" />
							)}
							Добавить
						</button>
					</div>
				</div>
			</div>

			<div className="card bg-base-100 shadow-xl max-w-4xl mx-auto">
				<div className="card-body p-0">
					<div className="overflow-x-auto">
						<table className="table table-lg">
							<thead>
								<tr className="border-b border-base-300">
									<th className="bg-base-200/50 font-semibold text-base-content">ID</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Название</th>
									<th className="bg-base-200/50 font-semibold text-base-content text-right">
										Действия
									</th>
								</tr>
							</thead>
							<tbody>
								{roles.map((role) => (
									<tr
										key={role.id}
										className="hover:bg-base-200/30 transition-colors duration-200 border-b border-base-300/50"
									>
										<td className="font-mono text-sm">{role.id}</td>
										<td>
											{editingRoleId === role.id ? (
												<div className="flex items-center gap-2">
													<input
														type="text"
														className="input input-sm input-bordered flex-1 max-w-xs"
														value={editingRoleName}
														onChange={(e) => setEditingRoleName(e.target.value)}
														aria-label="Название роли"
													/>
													<button
														type="button"
														className="btn btn-sm btn-success"
														onClick={handleSaveRole}
														disabled={updateRoleMutation.isPending}
													>
														{updateRoleMutation.isPending ? (
															<span className="loading loading-spinner loading-xs" />
														) : (
															"Сохранить"
														)}
													</button>
													<button
														type="button"
														className="btn btn-sm btn-ghost"
														onClick={handleCancelEdit}
													>
														Отмена
													</button>
												</div>
											) : (
												<span className="font-medium">{role.name}</span>
											)}
										</td>
										<td>
											<div className="flex justify-end gap-2">
												{editingRoleId !== role.id && (
													<button
														type="button"
														className="btn btn-ghost btn-circle btn-sm text-info hover:text-info"
														onClick={() => handleEditRole(role)}
														title="Изменить"
													>
														<FaEdit className="h-4 w-4" />
													</button>
												)}
												{isPending(role) ? (
													<button
														type="button"
														className="btn btn-error btn-sm"
														onClick={() => handleDelete(role)}
														disabled={deleteRoleMutation.isPending}
													>
														{deleteRoleMutation.isPending ? (
															<span className="loading loading-spinner loading-xs" />
														) : (
															`Подтвердить? (${secondsLeft}с)`
														)}
													</button>
												) : (
													<button
														type="button"
														className="btn btn-ghost btn-circle btn-sm text-error hover:text-error"
														onClick={() => handleDelete(role)}
														title="Удалить"
													>
														<FaTrash className="h-4 w-4" />
													</button>
												)}
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{roles.length === 0 && (
						<div className="text-center py-12">
							<div className="flex flex-col items-center gap-4">
								<FaShieldHalved className="h-16 w-16 text-base-content/30" />
								<div>
									<h3 className="text-lg font-semibold text-base-content mb-2">Роли не найдены</h3>
									<p className="text-base-content/70 mb-4">Создайте первую роль выше</p>
								</div>
							</div>
						</div>
					)}
				</div>
			</div>

			{toast && (
				<div className="toast toast-end toast-bottom mt-4">
					<Toast type={toast.type} message={toast.message} onClose={clear} />
				</div>
			)}
		</div>
	)
}
