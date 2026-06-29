import { createFileRoute, Link } from "@tanstack/react-router"
import { FaRobot, FaServer, FaTachometerAlt, FaUsers, FaUserTag } from "react-icons/fa"

export const Route = createFileRoute("/_protected/")({
	component: DashboardComponent,
})

function DashboardComponent() {
	return (
		<div className="p-4 md:p-6 lg:p-8">
			<div className="mb-8">
				<h1 className="text-2xl md:text-3xl font-bold mb-2 flex items-center gap-2">
					<FaTachometerAlt />
					Панель управления
				</h1>
				<p className="text-base-content/80">
					Добро пожаловать в панель администратора. Отсюда вы можете управлять пользователями,
					ролями, ботами и сервисами.
				</p>
			</div>

			<div className="grid grid-cols-1 md:grid-cols-2 gap-6">
				<div className="card bg-base-100 shadow-xl transition-all duration-300 hover:shadow-2xl hover:-translate-y-1">
					<div className="card-body flex flex-col">
						<div className="grow">
							<h2 className="card-title flex items-center gap-2">
								<FaUsers />
								Пользователи
							</h2>
							<p>Управление пользователями приложения.</p>
						</div>
						<div className="card-actions mt-4">
							<Link to="/users" className="btn btn-primary w-full sm:w-auto">
								<FaUsers />
								Перейти к пользователям
							</Link>
						</div>
					</div>
				</div>
				<div className="card bg-base-100 shadow-xl transition-all duration-300 hover:shadow-2xl hover:-translate-y-1">
					<div className="card-body flex flex-col">
						<div className="grow">
							<h2 className="card-title flex items-center gap-2">
								<FaUserTag />
								Роли
							</h2>
							<p>Управление ролями и правами.</p>
						</div>
						<div className="card-actions mt-4">
							<Link to="/roles" className="btn btn-primary w-full sm:w-auto">
								<FaUserTag />
								Перейти к ролям
							</Link>
						</div>
					</div>
				</div>
				<div className="card bg-base-100 shadow-xl transition-all duration-300 hover:shadow-2xl hover:-translate-y-1">
					<div className="card-body flex flex-col">
						<div className="grow">
							<h2 className="card-title flex items-center gap-2">
								<FaRobot />
								Боты
							</h2>
							<p>Управление ботами Telegram.</p>
						</div>
						<div className="card-actions mt-4">
							<Link to="/bots" className="btn btn-primary w-full sm:w-auto">
								<FaRobot />
								Перейти к ботам
							</Link>
						</div>
					</div>
				</div>
				<div className="card bg-base-100 shadow-xl transition-all duration-300 hover:shadow-2xl hover:-translate-y-1">
					<div className="card-body flex flex-col">
						<div className="grow">
							<h2 className="card-title flex items-center gap-2">
								<FaServer />
								Сервисы
							</h2>
							<p>Управление сервисами и ключами.</p>
						</div>
						<div className="card-actions mt-4">
							<Link to="/services" className="btn btn-primary w-full sm:w-auto">
								<FaServer />
								Перейти к сервисам
							</Link>
						</div>
					</div>
				</div>
			</div>
		</div>
	)
}
