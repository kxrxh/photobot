import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { FaEdit, FaPhone, FaUserPlus, FaUsers } from "react-icons/fa"
import { PageError, PageLoading, Toast } from "@/components/ui"
import { getUsers, ManageRolesModal } from "@/features/users"
import { useEntityQuery } from "@/hooks"
import type { User } from "@/types/user"

export const Route = createFileRoute("/_protected/users")({
	component: UsersComponent,
})

function UsersComponent() {
	const [selectedUser, setSelectedUser] = useState<User | null>(null)
	const [showToast, setShowToast] = useState(false)
	const [toastMessage, setToastMessage] = useState("")
	const [toastType, setToastType] = useState<"success" | "error">("success")

	const {
		data: users,
		isLoading,
		error,
		refetch,
	} = useEntityQuery({
		queryKey: ["users"],
		queryFn: getUsers,
	})

	if (isLoading) {
		return <PageLoading message="Загрузка пользователей…" />
	}

	if (error) {
		return (
			<PageError
				message={`Ошибка при загрузке пользователей: ${error.message}`}
				onRetry={() => void refetch()}
			/>
		)
	}

	return (
		<div className="container mx-auto p-6 max-w-7xl">
			<div className="mb-8">
				<div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
					<div>
						<h1 className="text-4xl font-bold text-base-content mb-2">Пользователи</h1>
						<p className="text-sm text-base-content/70">
							Количество пользователей: {users?.length || 0}
						</p>
					</div>
				</div>
			</div>

			<div className="card bg-base-100 shadow-xl">
				<div className="card-body p-0">
					<div className="overflow-x-auto">
						<table className="table w-full">
							<thead>
								<tr className="border-b border-base-300">
									<th className="bg-base-200/50 font-semibold text-base-content">Полное имя</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Организация</th>
									<th className="bg-base-200/50 font-semibold text-base-content">ИНН</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Телефон</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Telegram ID</th>
									<th className="bg-base-200/50 font-semibold text-base-content">MAX ID</th>
									<th className="bg-base-200/50 font-semibold text-base-content text-center min-w-[140px]">
										Действия
									</th>
								</tr>
							</thead>
							<tbody>
								{users?.map((user) => (
									<tr
										key={user.id}
										className="hover:bg-base-200/30 transition-colors duration-200 border-b border-base-300/50"
									>
										<td>
											<div className="flex items-center gap-3">
												<div className="avatar avatar-placeholder">
													<div className="bg-primary/10 text-primary w-10 rounded-full">
														<span className="text-xs font-bold">
															{user.full_name
																?.split(" ")
																.slice(0, 2)
																.map((n) => n[0]?.toUpperCase())
																.join("") || "UN"}
														</span>
													</div>
												</div>
												<div>
													<div className="font-semibold text-sm text-base-content">
														{user.full_name || "Не указано"}
													</div>
													<div className="text-xs text-base-content/70">ID: #{user.id}</div>
												</div>
											</div>
										</td>
										<td>
											<div className="max-w-xs">
												<div
													className="font-medium text-xs text-base-content truncate"
													title={user.organization_name}
												>
													{user.organization_name || "Не указано"}
												</div>
											</div>
										</td>
										<td>
											<button
												type="button"
												className="font-mono text-xs bg-base-200 px-2 py-1 rounded inline-block cursor-pointer"
												onClick={() => {
													if (user.inn) {
														navigator.clipboard.writeText(user.inn)
														setToastMessage("ИНН скопирован в буфер обмена")
														setToastType("success")
														setShowToast(true)
													}
												}}
												title={user.inn ? "Нажмите, чтобы скопировать" : undefined}
											>
												{user.inn || "Не указан"}
											</button>
										</td>
										<td>
											<div className="flex items-center gap-2">
												<FaPhone className="h-4 w-4 text-base-content/50" />
												<span className="font-mono text-xs">
													{user.phone_number || "Не указан"}
												</span>
											</div>
										</td>
										<td>
											{user.telegram_id ? (
												<div className="flex items-center gap-2">
													<a
														href={`tg://user?id=${user.telegram_id}`}
														target="_blank"
														rel="noopener noreferrer"
														className="font-mono text-xs text-blue-500 hover:underline"
													>
														@{user.telegram_id}
													</a>
												</div>
											) : (
												<div className="flex items-center gap-2">
													<div className="w-2 h-2 bg-base-300 rounded-full"></div>
													<span className="text-base-content/50 text-xs">Не подключен</span>
												</div>
											)}
										</td>
										<td>
											{user.max_id != null && user.max_id !== 0 ? (
												<span className="font-mono text-xs">{user.max_id}</span>
											) : (
												<div className="flex items-center gap-2">
													<div className="w-2 h-2 bg-base-300 rounded-full"></div>
													<span className="text-base-content/50 text-xs">Не подключен</span>
												</div>
											)}
										</td>
										<td className="whitespace-nowrap align-top">
											<div className="flex flex-col gap-2">
												<button
													type="button"
													className="btn btn-xs btn-outline btn-primary transition-all duration-200 justify-start"
													disabled
													title="Скоро"
												>
													<FaEdit className="h-3 w-3 shrink-0" />
													Изменить
												</button>
												<button
													type="button"
													className="btn btn-xs btn-outline btn-secondary transition-all duration-200 justify-start"
													onClick={() => setSelectedUser(user)}
												>
													<FaUsers className="h-3 w-3 shrink-0" />
													Изменить роли
												</button>
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{users?.length === 0 && (
						<div className="text-center py-12">
							<div className="flex flex-col items-center gap-4">
								<FaUserPlus className="h-16 w-16 text-base-content/30" />
								<div>
									<h3 className="text-lg font-semibold text-base-content mb-2">
										Пользователи не найдены
									</h3>
									<p className="text-sm text-base-content/70 mb-4">
										Пользователи регистрируются через приложение.
									</p>
								</div>
							</div>
						</div>
					)}
				</div>
			</div>

			{selectedUser && (
				<ManageRolesModal user={selectedUser} onClose={() => setSelectedUser(null)} />
			)}

			{showToast && (
				<div className="toast toast-end toast-bottom">
					<Toast type={toastType} message={toastMessage} onClose={() => setShowToast(false)} />
				</div>
			)}
		</div>
	)
}
