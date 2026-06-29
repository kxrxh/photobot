import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { FaEdit, FaPlus, FaTrash, FaUserPlus } from "react-icons/fa"
import { PageError, PageLoading } from "@/components/ui"
import { CreateBotModal, DeleteBotModal, EditBotModal, getBots } from "@/features/bots"
import { useEntityQuery } from "@/hooks"
import type { Bot } from "@/types/bot"

export const Route = createFileRoute("/_protected/bots")({
	component: BotsComponent,
})

function BotsComponent() {
	const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
	const [isEditModalOpen, setIsEditModalOpen] = useState(false)
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
	const [selectedBot, setSelectedBot] = useState<Bot | null>(null)

	const {
		data: bots,
		isLoading,
		error,
		refetch,
	} = useEntityQuery({
		queryKey: ["bots"],
		queryFn: getBots,
	})

	if (isLoading) {
		return <PageLoading message="Загрузка ботов…" />
	}

	if (error) {
		return (
			<PageError
				message={`Ошибка при загрузке ботов: ${error.message}`}
				onRetry={() => void refetch()}
			/>
		)
	}

	return (
		<div className="container mx-auto p-6 max-w-7xl">
			<div className="mb-8">
				<div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
					<div>
						<h1 className="text-4xl font-bold text-base-content mb-2">Боты</h1>
						<p className="text-base-content/70">Количество ботов: {bots?.length || 0}</p>
					</div>
					<button
						type="button"
						className="btn btn-primary btn-md gap-2"
						onClick={() => setIsCreateModalOpen(true)}
					>
						<FaPlus className="h-4 w-4" />
						Добавить нового бота
					</button>
				</div>
			</div>

			<div className="card bg-base-100 shadow-xl max-w-4xl mx-auto">
				<div className="card-body p-0">
					<div className="overflow-x-auto">
						<table className="table table-lg">
							<thead>
								<tr className="border-b border-base-300">
									<th className="bg-base-200/50 font-semibold text-base-content">ID</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Платформа</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Имя</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Дата создания</th>
									<th className="bg-base-200/50 font-semibold text-base-content text-right"></th>
								</tr>
							</thead>
							<tbody>
								{bots?.map((bot) => (
									<tr
										key={bot.id}
										className="hover:bg-base-200/30 transition-colors duration-200 border-b border-base-300/50"
									>
										<td className="font-mono text-sm">
											<div className="flex items-center gap-3">{bot.id}</div>
										</td>
										<td>
											<span
												className={`badge badge-sm ${
													(bot.platform ?? "telegram") === "max" ? "badge-primary" : "badge-ghost"
												}`}
											>
												{(bot.platform ?? "telegram") === "max" ? "MAX" : "Telegram"}
											</span>
										</td>
										<td>
											<div className="flex items-center gap-3">
												<div className="font-semibold text-base-content text-sm">
													{bot.name || "Не указано"}
												</div>
											</div>
										</td>
										<td className="text-sm">
											{new Date(bot.created_at).toLocaleString("ru-RU", {
												year: "numeric",
												month: "2-digit",
												day: "2-digit",
												hour: "2-digit",
												minute: "2-digit",
												second: "2-digit",
												hour12: false,
											})}
										</td>
										<td>
											<div className="flex justify-end gap-2">
												<button
													type="button"
													className="btn btn-ghost btn-circle btn-sm text-base-content/70 hover:text-primary transition-all duration-200"
													onClick={() => {
														setSelectedBot(bot)
														setIsEditModalOpen(true)
													}}
												>
													<FaEdit className="h-4 w-4" />
												</button>
												<button
													type="button"
													className="btn btn-ghost btn-circle btn-sm text-base-content/70 hover:text-primary transition-all duration-200"
													onClick={() => {
														setSelectedBot(bot)
														setIsDeleteModalOpen(true)
													}}
												>
													<FaTrash className="h-4 w-4" />
												</button>
											</div>
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>

					{bots?.length === 0 && (
						<div className="text-center py-12">
							<div className="flex flex-col items-center gap-4">
								<FaUserPlus className="h-16 w-16 text-base-content/30" />
								<div>
									<h3 className="text-lg font-semibold text-base-content mb-2">Боты не найдены</h3>
									<p className="text-base-content/70 mb-4">Начните добавлять ботов в систему</p>
									<button
										type="button"
										className="btn btn-primary"
										onClick={() => setIsCreateModalOpen(true)}
									>
										Добавить первого бота
									</button>
								</div>
							</div>
						</div>
					)}
				</div>
			</div>

			{isCreateModalOpen && <CreateBotModal onClose={() => setIsCreateModalOpen(false)} />}

			{isEditModalOpen && selectedBot && (
				<EditBotModal
					bot={selectedBot}
					onClose={() => {
						setIsEditModalOpen(false)
						setSelectedBot(null)
					}}
				/>
			)}

			{isDeleteModalOpen && selectedBot && (
				<DeleteBotModal
					bot={selectedBot}
					onClose={() => {
						setIsDeleteModalOpen(false)
						setSelectedBot(null)
					}}
				/>
			)}
		</div>
	)
}
