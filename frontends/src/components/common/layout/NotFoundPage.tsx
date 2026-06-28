import { Link } from "@tanstack/react-router"

const NotFoundPage = () => {
	return (
		<div className="flex items-center justify-center min-h-screen p-4">
			<div className="text-center">
				<div className="mb-8">
					<h1 className="font-bold text-9xl text-primary">404</h1>
					<div className="w-24 h-1 mx-auto my-4 bg-primary" />
				</div>
				<h2 className="mb-4 text-3xl font-semibold">Страница не найдена</h2>
				<p className="mb-8 text-base-content/70">
					Страница, которую вы ищете, не существует или была перемещена.
				</p>
				<Link to="/" className="btn btn-primary">
					Вернуться на главную
				</Link>
			</div>
		</div>
	)
}

export default NotFoundPage
