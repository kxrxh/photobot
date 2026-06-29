import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { setupWebAccess } from "@/api/auth"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"

export const Route = createFileRoute("/_authenticated/settings/web-access")({
	component: WebAccessPage,
})

function WebAccessPage() {
	const navigate = useNavigate()
	const [login, setLogin] = useState("")
	const [password, setPassword] = useState("")
	const [confirmPassword, setConfirmPassword] = useState("")
	const [recoveryCodes, setRecoveryCodes] = useState<string[] | null>(null)
	const [error, setError] = useState<string | null>(null)
	const [loading, setLoading] = useState(false)

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		setError(null)
		if (password !== confirmPassword) {
			setError("Пароли не совпадают")
			return
		}
		setLoading(true)
		try {
			const result = await setupWebAccess(login.trim(), password)
			setRecoveryCodes(result.recovery_codes)
		} catch (err) {
			setError(err instanceof Error ? err.message : "Не удалось создать веб-доступ")
		} finally {
			setLoading(false)
		}
	}

	if (recoveryCodes) {
		return (
			<div className="flex flex-col w-full max-w-md gap-4 p-4 mx-auto">
				<h2 className="text-xl font-semibold text-center">Веб-доступ создан</h2>
				<p className="text-sm text-base-content/70 text-center">
					Теперь можно входить в браузере с логином <strong>{login}</strong>
				</p>
				<div className="rounded-box bg-base-200 p-4 font-mono text-sm space-y-1">
					{recoveryCodes.map((code) => (
						<div key={code}>{code}</div>
					))}
				</div>
				<p className="text-sm text-warning text-center">
					Сохраните коды восстановления — они показываются один раз.
				</p>
				<Button
					type="button"
					variant="primary"
					fullWidth
					onClick={() => navigate({ to: "/profile" })}
				>
					Вернуться в профиль
				</Button>
			</div>
		)
	}

	return (
		<div className="flex flex-col w-full max-w-md gap-4 p-4 mx-auto">
			<h2 className="text-xl font-semibold text-center">Вход через браузер</h2>
			<p className="text-sm text-base-content/70 text-center">
				Придумайте логин и пароль для этого аккаунта. Telegram/MAX останутся привязаны.
			</p>
			<form onSubmit={handleSubmit} className="flex flex-col gap-3">
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
					Создать веб-доступ
				</Button>
				<Button
					type="button"
					variant="ghost"
					fullWidth
					onClick={() => navigate({ to: "/profile" })}
				>
					Отмена
				</Button>
			</form>
		</div>
	)
}
