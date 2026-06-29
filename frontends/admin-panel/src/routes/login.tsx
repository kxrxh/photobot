import { useQueryClient } from "@tanstack/react-query"
import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { HTTPError } from "ky"
import { useCallback, useState } from "react"
import { FaEye, FaEyeSlash, FaLock, FaUser } from "react-icons/fa"
import { Toast } from "@/components/ui"
import { login, useAuth } from "@/features/auth"

export const Route = createFileRoute("/login")({
	component: LoginComponent,
})

function LoginComponent() {
	const { login: loginContext } = useAuth()
	const [loginField, setLogin] = useState("")
	const [password, setPassword] = useState("")
	const [errorMessage, setErrorMessage] = useState<string | null>(null)
	const [showPassword, setShowPassword] = useState(false)
	const [isLoading, setIsLoading] = useState(false)
	const navigate = useNavigate()
	const queryClient = useQueryClient()

	const onLogin = useCallback(async () => {
		setIsLoading(true)
		try {
			const response = await login(loginField, password)
			setErrorMessage(null)
			loginContext(response.result.access_token, response.result.refresh_token)
			await queryClient.invalidateQueries()

			setTimeout(() => {
				navigate({ to: "/" })
			}, 0)
		} catch (error) {
			if (error instanceof HTTPError && error.response.status === 401) {
				setErrorMessage("Неверный логин или пароль")
			} else {
				setErrorMessage("Произошла неизвестная ошибка")
			}
		} finally {
			setIsLoading(false)
		}
	}, [loginField, password, loginContext, navigate, queryClient])

	return (
		<div className="min-h-screen bg-gradient-to-br from-primary/10 via-secondary/5 to-accent/10 flex items-center justify-center p-4">
			<div className="card bg-base-100/80 backdrop-blur-md w-full max-w-md shadow-2xl border border-base-300/50">
				<div className="card-body p-8">
					{/* Header */}
					<div className="text-center mb-8">
						<div className="w-16 h-16 bg-gradient-to-r from-primary to-secondary rounded-full flex items-center justify-center mx-auto mb-4 shadow-lg">
							<FaLock className="text-2xl text-white" />
						</div>
						<h1 className="text-3xl font-bold bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
							Добро пожаловать
						</h1>
					</div>

					{/* Error Message */}
					{errorMessage && (
						<div className="mb-6">
							<Toast type="error" message={errorMessage} onClose={() => setErrorMessage(null)} />
						</div>
					)}

					<form
						onSubmit={(e) => {
							e.preventDefault()
							onLogin()
						}}
						className="space-y-6"
					>
						{/* Login Field */}
						<div className="form-control">
							<label className="label" htmlFor="login">
								<span className="label-text font-medium flex items-center gap-2">Логин</span>
							</label>
							<div className="relative">
								<input
									type="text"
									id="login"
									placeholder="Введите логин администратора"
									className="input input-bordered w-full pl-12 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200"
									required
									value={loginField}
									onChange={(e) => {
										setLogin(e.target.value)
										setErrorMessage(null)
									}}
								/>
								<FaUser className="absolute left-4 top-1/2 transform -translate-y-1/2 text-base-content/40 pointer-events-none z-10" />
							</div>
						</div>

						{/* Password Field */}
						<div className="form-control">
							<label className="label" htmlFor="password">
								<span className="label-text font-medium flex items-center gap-2">Пароль</span>
							</label>
							<div className="relative">
								<input
									type={showPassword ? "text" : "password"}
									id="password"
									placeholder="Введите пароль администратора"
									className="input input-bordered w-full pl-12 pr-12 focus:outline-none focus:ring-2 focus:ring-primary/20 focus:border-primary transition-all duration-200"
									required
									value={password}
									onChange={(e) => {
										setPassword(e.target.value)
										setErrorMessage(null)
									}}
								/>
								<FaLock className="absolute left-4 top-1/2 transform -translate-y-1/2 text-base-content/40 pointer-events-none z-10" />
								<button
									type="button"
									onClick={() => setShowPassword(!showPassword)}
									className="absolute right-4 top-1/2 transform -translate-y-1/2 text-base-content/40 hover:text-base-content transition-colors duration-200 z-10"
								>
									{showPassword ? <FaEyeSlash /> : <FaEye />}
								</button>
							</div>
						</div>

						{/* Submit Button */}
						<div className="form-control mt-8">
							<button
								type="submit"
								className={`btn btn-primary w-full h-12 text-lg font-medium transition-all duration-200 ${
									isLoading ? "loading" : "hover:scale-105 hover:shadow-lg"
								}`}
								disabled={isLoading}
							>
								{isLoading ? (
									<>
										<span className="loading loading-spinner loading-sm"></span>
										Вход...
									</>
								) : (
									"Войти"
								)}
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	)
}
