import { createFileRoute, Link, useNavigate } from "@tanstack/react-router"
import { useEffect, useState } from "react"
import { AuthPageLayout } from "@/components/auth/AuthPageLayout"
import { MessengerReloadPrompt } from "@/components/auth/MessengerReloadPrompt"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { useAuth } from "@/contexts/AuthContext"
import { isWebAuthMode } from "@/lib/auth/mode"
import { getAuthErrorMessage } from "@/utils/errors"

export const Route = createFileRoute("/login")({
	component: RouteComponent,
})

function RouteComponent() {
	const { state, loginWithPassword, error: authError } = useAuth()
	const navigate = useNavigate({ from: "/login" })
	const [loginName, setLoginName] = useState("")
	const [password, setPassword] = useState("")
	const [error, setError] = useState<string | null>(null)
	const [loading, setLoading] = useState(false)

	useEffect(() => {
		if (state === "authenticated") {
			navigate({ to: "/menu", replace: true })
		}
	}, [state, navigate])

	if (state === "loading") {
		return (
			<div className="flex items-center justify-center h-screen">
				<span className="loading loading-spinner loading-lg" />
			</div>
		)
	}

	if (state === "authenticated") {
		return null
	}

	if (isWebAuthMode()) {
		const handleSubmit = async (e: React.FormEvent) => {
			e.preventDefault()
			setError(null)
			setLoading(true)
			try {
				await loginWithPassword(loginName.trim(), password)
			} catch (err) {
				setError(await getAuthErrorMessage(err))
			} finally {
				setLoading(false)
			}
		}

		return (
			<AuthPageLayout
				title="Вход"
				subtitle="Введите логин и пароль"
				footer={
					<div className="flex flex-col gap-2 text-center text-sm">
						<Link to="/register" className="link link-primary">
							Создать аккаунт
						</Link>
						<Link to="/forgot-password" className="link link-neutral">
							Забыли пароль?
						</Link>
					</div>
				}
			>
				<form onSubmit={handleSubmit} className="flex flex-col gap-3">
					<Input
						type="text"
						autoComplete="username"
						placeholder="Логин"
						value={loginName}
						onChange={(e) => setLoginName(e.target.value)}
						required
					/>
					<Input
						type="password"
						autoComplete="current-password"
						placeholder="Пароль"
						value={password}
						onChange={(e) => setPassword(e.target.value)}
						required
					/>
					{(error || authError) && <p className="text-sm text-error">{error ?? authError}</p>}
					<Button type="submit" variant="primary" fullWidth loading={loading}>
						Войти
					</Button>
				</form>
			</AuthPageLayout>
		)
	}

	if (state === "invalid_data") {
		throw new Error(authError || "Ошибка валидации данных Telegram")
	}

	return <MessengerReloadPrompt />
}
