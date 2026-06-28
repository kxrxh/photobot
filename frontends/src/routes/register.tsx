import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { loginWithPassword, registerWeb } from "@/api/auth"
import { AuthPageLayout } from "@/components/auth/AuthPageLayout"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { useAuth } from "@/contexts/AuthContext"
import { getAuthErrorMessage } from "@/utils/errors"

export const Route = createFileRoute("/register")({
	component: RouteComponent,
})

function RouteComponent() {
	const navigate = useNavigate({ from: "/register" })
	const { applyTokenResponse } = useAuth()
	const [login, setLogin] = useState("")
	const [password, setPassword] = useState("")
	const [confirmPassword, setConfirmPassword] = useState("")
	const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null)
	const [error, setError] = useState<string | null>(null)
	const [loading, setLoading] = useState(false)

	const handleRegister = async (e: React.FormEvent) => {
		e.preventDefault()
		setError(null)
		if (password !== confirmPassword) {
			setError("Пароли не совпадают")
			return
		}
		setLoading(true)
		try {
			const result = await registerWeb({ login: login.trim(), password })
			setRecoveryCodes(result.recovery_codes)
		} catch (err) {
			setError(await getAuthErrorMessage(err))
		} finally {
			setLoading(false)
		}
	}

	const handleContinue = async () => {
		if (!recoveryCodes) return
		setLoading(true)
		try {
			const tokens = await loginWithPassword(login.trim(), password)
			await applyTokenResponse(tokens)
			navigate({ to: "/menu", replace: true })
		} catch (err) {
			setError(await getAuthErrorMessage(err))
		} finally {
			setLoading(false)
		}
	}

	if (recoveryCodes) {
		return (
			<AuthPageLayout
				title="Сохраните коды восстановления"
				subtitle="Они понадобятся для сброса пароля, если нет привязанного мессенджера"
			>
				<div className="rounded-box bg-base-200 p-4 font-mono text-sm space-y-1">
					{recoveryCodes.map((code) => (
						<div key={code}>{code}</div>
					))}
				</div>
				<p className="text-sm text-warning">
					Сохраните коды в надёжном месте. Повторно они не будут показаны.
				</p>
				<Button
					type="button"
					variant="primary"
					fullWidth
					loading={loading}
					onClick={handleContinue}
				>
					Я сохранил коды — продолжить
				</Button>
			</AuthPageLayout>
		)
	}

	return (
		<AuthPageLayout
			title="Регистрация"
			subtitle="Новый аккаунт только для браузера. Уже пользуетесь ботом — создайте логин в Профиле мини-приложения."
		>
			<form onSubmit={handleRegister} className="flex flex-col gap-3">
				<Input
					type="text"
					autoComplete="username"
					placeholder="Логин (a-z, 0-9, _, -)"
					value={login}
					onChange={(e) => setLogin(e.target.value)}
					required
					minLength={3}
					maxLength={32}
				/>
				<Input
					type="password"
					autoComplete="new-password"
					placeholder="Пароль"
					value={password}
					onChange={(e) => setPassword(e.target.value)}
					required
					minLength={6}
				/>
				<Input
					type="password"
					autoComplete="new-password"
					placeholder="Повторите пароль"
					value={confirmPassword}
					onChange={(e) => setConfirmPassword(e.target.value)}
					required
					minLength={6}
				/>
				{error && <p className="text-sm text-error">{error}</p>}
				<Button type="submit" variant="primary" fullWidth loading={loading}>
					Зарегистрироваться
				</Button>
			</form>
		</AuthPageLayout>
	)
}
