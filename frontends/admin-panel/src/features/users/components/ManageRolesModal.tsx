import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { FaArrowLeft, FaArrowRight } from "react-icons/fa"
import { FaPeopleGroup } from "react-icons/fa6"
import { IoClose } from "react-icons/io5"
import { addUserRole, getRoles, getUserRoles, removeUserRole } from "@/features/roles"
import type { Role } from "@/types/role"
import type { User } from "@/types/user"

interface ManageRolesModalProps {
	user: User
	onClose: () => void
}

export function ManageRolesModal({ user, onClose }: ManageRolesModalProps) {
	const queryClient = useQueryClient()

	const { data: allRoles = [] } = useQuery<Role[]>({
		queryKey: ["roles"],
		queryFn: getRoles,
	})

	const { data: userRoles = [] } = useQuery<Role[]>({
		queryKey: ["userRoles", user.id],
		queryFn: () => getUserRoles(user.id),
	})

	const [assignedRoles, setAssignedRoles] = useState<Role[]>([])
	const [availableRoles, setAvailableRoles] = useState<Role[]>([])

	useEffect(() => {
		setAssignedRoles(userRoles)
	}, [userRoles])

	useEffect(() => {
		const existingRoleIds = new Set(allRoles.map((r) => r.id))
		setAssignedRoles((prev) => prev.filter((role) => existingRoleIds.has(role.id)))
	}, [allRoles])

	useEffect(() => {
		if (allRoles.length > 0) {
			const assignedRoleIds = new Set(assignedRoles.map((r) => r.id))
			setAvailableRoles(allRoles.filter((r) => !assignedRoleIds.has(r.id)))
		} else {
			setAvailableRoles([])
		}
	}, [allRoles, assignedRoles])

	const addUserRoleMutation = useMutation({
		mutationFn: ({ userId, roleId }: { userId: number; roleId: number }) =>
			addUserRole(userId, roleId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["userRoles", user.id] })
		},
	})

	const removeUserRoleMutation = useMutation({
		mutationFn: ({ userId, roleId }: { userId: number; roleId: number }) =>
			removeUserRole(userId, roleId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["userRoles", user.id] })
		},
	})

	const handleAddRole = (role: Role) => {
		addUserRoleMutation.mutate({ userId: user.id, roleId: role.id })
	}

	const handleRemoveRole = (role: Role) => {
		removeUserRoleMutation.mutate({ userId: user.id, roleId: role.id })
	}

	useEffect(() => {
		const handler = (e: KeyboardEvent) => {
			if (e.key === "Escape") onClose()
		}
		window.addEventListener("keydown", handler)
		return () => window.removeEventListener("keydown", handler)
	}, [onClose])

	return (
		<div
			className="modal modal-open"
			onClick={onClose}
			onKeyDown={(e) => {
				if (e.key === "Enter" || e.key === " ") {
					e.preventDefault()
					onClose()
				}
			}}
			role="dialog"
			aria-modal="true"
		>
			<div className="modal-backdrop backdrop-blur-sm bg-black/20" aria-hidden />
			<div
				role="document"
				className="modal-box w-11/12 max-w-4xl relative backdrop-blur-md bg-base-100/95"
				onClick={(e) => e.stopPropagation()}
				onKeyDown={(e) => e.stopPropagation()}
			>
				<div className="flex items-center justify-between mb-6">
					<div>
						<h3 className="text-xl font-bold">Управление ролями пользователя #{user.id}</h3>
						<p className="text-sm text-base-content/70 mt-1">
							Telegram ID:{" "}
							<span className="font-mono">{user.telegram_id ? String(user.telegram_id) : "—"}</span>
							{" · "}
							MAX ID:{" "}
							<span className="font-mono">
								{user.max_id != null && user.max_id !== 0 ? String(user.max_id) : "—"}
							</span>
						</p>
					</div>
					<button
						type="button"
						className="btn btn-sm btn-circle btn-ghost"
						onClick={onClose}
						aria-label="Закрыть"
					>
						<IoClose className="size-6" />
					</button>
				</div>

				<div className="flex items-center gap-2 mb-4">
					<FaPeopleGroup className="size-5 text-base-content/70" />
					<span className="font-medium">Назначение ролей</span>
				</div>

				<div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
					<div className="card bg-base-100 shadow-lg">
						<div className="card-body p-3">
							<h4 className="card-title text-sm mb-2">Доступные роли</h4>
							<div className="space-y-2 max-h-48 overflow-y-auto">
								{availableRoles.length > 0 ? (
									availableRoles.map((role) => (
										<div
											key={role.id}
											className="flex items-center justify-between p-3 bg-base-200/50 rounded-xl hover:bg-base-300/70 transition-all duration-200 backdrop-blur-sm"
										>
											<span className="text-sm font-medium">{role.name}</span>
											<button
												type="button"
												className="btn btn-sm btn-primary btn-circle"
												onClick={() => handleAddRole(role)}
												disabled={addUserRoleMutation.isPending}
											>
												{addUserRoleMutation.isPending ? (
													<span className="loading loading-spinner loading-xs" />
												) : (
													<FaArrowRight className="text-xs" />
												)}
											</button>
										</div>
									))
								) : (
									<div className="flex items-center justify-center h-16 text-center">
										<div className="text-base-content/50">
											<p className="text-sm">Нет доступных ролей</p>
										</div>
									</div>
								)}
							</div>
						</div>
					</div>

					<div className="card bg-base-100 shadow-lg">
						<div className="card-body p-3">
							<h4 className="card-title text-sm mb-2">Назначенные роли</h4>
							<div className="space-y-2 max-h-48 overflow-y-auto">
								{assignedRoles.length > 0 ? (
									assignedRoles.map((role) => (
										<div
											key={role.id}
											className="flex items-center justify-between p-3 bg-secondary/10 rounded-xl hover:bg-secondary/20 transition-all duration-200 backdrop-blur-sm"
										>
											<button
												type="button"
												className="btn btn-sm btn-secondary btn-circle"
												onClick={() => handleRemoveRole(role)}
												disabled={removeUserRoleMutation.isPending}
											>
												{removeUserRoleMutation.isPending ? (
													<span className="loading loading-spinner loading-xs" />
												) : (
													<FaArrowLeft className="text-xs" />
												)}
											</button>
											<span className="text-sm font-medium text-secondary">{role.name}</span>
										</div>
									))
								) : (
									<div className="flex items-center justify-center h-16 text-center">
										<div className="text-base-content/50">
											<p className="text-sm">Нет назначенных ролей</p>
										</div>
									</div>
								)}
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}
