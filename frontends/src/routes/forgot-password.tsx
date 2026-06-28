import { createFileRoute } from "@tanstack/react-router"
import { useState } from "react"
import { forgotPassword } from "@/api/auth"
import { AuthPageLayout } from "@/components/auth/AuthPageLayout"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { getAuthErrorMessage } from "@/utils/errors"

export const Route = createFileRoute("/forgot-password")({
	component: RouteComponent,
})

function RouteComponent() {
	const [login, setLogin] = useState("")
	const [sent, setSent] = useState(false)
	const [error, setError] = useState<string | null>(null)
	const [loading, setLoading] = useState(false)

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		setError(null)
		setLoading(true)
		try {
			await forgotPassword(login.trim())
			setSent(true)
		} catch (err) {
			setError(await getAuthErrorMessage(err))
		} finally {
			setLoading(false)
		}
	}

	return (
		<AuthPageLayout
			title="Сброс пароля"
			subtitle={
				sent
					? "Если аккаунт существует, код отправлен в привязанный мессенджер"
					: "Введите логин. Код придёт в Telegram/MAX, если аккаунт привязан"
			}
		>
			{sent ? (
				<p className="text-sm text-base-content/80 text-center">
					Нет привязанного мессенджера?
					<br />
					Используйте{" "}
					<a href="/reset-password-recovery" className="link link-primary">
						код восстановления
					</a>
					.
				</p>
			) : (
				<form onSubmit={handleSubmit} className="flex flex-col gap-3">
					<Input
						type="text"
						placeholder="Логин"
						value={login}
						onChange={(e) => setLogin(e.target.value)}
						required
					/>
					{error && <p className="text-sm text-error">{error}</p>}
					<Button type="submit" variant="primary" fullWidth loading={loading}>
						Отправить код
					</Button>
				</form>
			)}
		</AuthPageLayout>
	)
}
