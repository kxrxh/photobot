import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { FaPlus, FaServer, FaTrash } from "react-icons/fa"
import { PageError, PageLoading } from "@/components/ui"
import { CreateServiceModal, DeleteServiceModal, getServices } from "@/features/services"
import { useEntityQuery } from "@/hooks"
import type { Service } from "@/types/service"

export const Route = createFileRoute("/_protected/services")({
	component: ServicesComponent,
})

function ServicesComponent() {
	const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
	const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false)
	const [selectedService, setSelectedService] = useState<Service | null>(null)

	const {
		data: services,
		isLoading,
		error,
		refetch,
	} = useEntityQuery({
		queryKey: ["services"],
		queryFn: getServices,
	})

	if (isLoading) {
		return <PageLoading message="Загрузка сервисов…" />
	}

	if (error) {
		return (
			<PageError
				message={`Ошибка при загрузке сервисов: ${error.message}`}
				onRetry={() => void refetch()}
			/>
		)
	}

	return (
		<div className="container mx-auto p-6 max-w-7xl">
			<div className="mb-8">
				<div className="flex flex-col sm:flex-row sm:justify-between sm:items-center gap-4">
					<div>
						<h1 className="text-4xl font-bold text-base-content mb-2">Сервисы</h1>
						<p className="text-base-content/70">Количество сервисов: {services?.length || 0}</p>
					</div>
					<button
						type="button"
						className="btn btn-primary btn-md gap-2"
						onClick={() => setIsCreateModalOpen(true)}
					>
						<FaPlus className="h-4 w-4" />
						Добавить новый сервис
					</button>
				</div>
			</div>

			<div className="card bg-base-100 shadow-xl max-w-4xl mx-auto">
				<div className="card-body p-0">
					<div className="overflow-x-auto">
						<table className="table table-lg">
							<thead>
								<tr className="border-b border-base-300">
									<th className="bg-base-200/50 font-semibold text-base-content">ID Сервиса</th>
									<th className="bg-base-200/50 font-semibold text-base-content">Дата создания</th>
									<th className="bg-base-200/50 font-semibold text-base-content">
										Последнее обновление
									</th>
									<th className="bg-base-200/50 font-semibold text-base-content text-right"></th>
								</tr>
							</thead>
							<tbody>
								{services?.map((service) => (
									<tr
										key={service.service_id}
										className="hover:bg-base-200/30 transition-colors duration-200 border-b border-base-300/50"
									>
										<td className="font-mono text-sm">
											<div className="flex items-center gap-3">{service.service_id}</div>
										</td>
										<td className="text-sm">
											{new Date(service.created_at).toLocaleString("ru-RU")}
										</td>
										<td className="text-sm">
											{new Date(service.updated_at).toLocaleString("ru-RU")}
										</td>
										<td>
											<div className="flex justify-end gap-2">
												<button
													type="button"
													className="btn btn-ghost btn-circle btn-sm text-base-content/70 hover:text-error transition-all duration-200"
													onClick={() => {
														setSelectedService(service)
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

					{services?.length === 0 && (
						<div className="text-center py-12">
							<div className="flex flex-col items-center gap-4">
								<FaServer className="h-16 w-16 text-base-content/30" />
								<div>
									<h3 className="text-lg font-semibold text-base-content mb-2">
										Сервисы не найдены
									</h3>
									<p className="text-base-content/70 mb-4">Начните добавлять сервисы в систему</p>
									<button
										type="button"
										className="btn btn-primary"
										onClick={() => setIsCreateModalOpen(true)}
									>
										Добавить первый сервис
									</button>
								</div>
							</div>
						</div>
					)}
				</div>
			</div>

			{isCreateModalOpen && <CreateServiceModal onClose={() => setIsCreateModalOpen(false)} />}

			{isDeleteModalOpen && selectedService && (
				<DeleteServiceModal
					service={selectedService}
					onClose={() => {
						setIsDeleteModalOpen(false)
						setSelectedService(null)
					}}
				/>
			)}
		</div>
	)
}
