import { createFileRoute, useNavigate } from "@tanstack/react-router"
import { useState } from "react"
import { resetPassword } from "@/api/auth"
import { AuthPageLayout } from "@/components/auth/AuthPageLayout"
import { Button } from "@/components/common/ui/Button"
import { Input } from "@/components/common/ui/Input"
import { getAuthErrorMessage } from "@/utils/errors"

export const Route = createFileRoute("/reset-password")({
	component: RouteComponent,
})

function RouteComponent() {
	const navigate = useNavigate({ from: "/reset-password" })
	const [login, setLogin] = useState("")
	const [otp, setOtp] = useState("")
	const [newPassword, setNewPassword] = useState("")
	const [error, setError] = useState<string | null>(null)
	const [loading, setLoading] = useState(false)

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		setError(null)
		setLoading(true)
		try {
			await resetPassword(login.trim(), otp.trim(), newPassword)
			navigate({ to: "/login" })
		} catch (err) {
			setError(await getAuthErrorMessage(err))
		} finally {
			setLoading(false)
		}
	}

	return (
		<AuthPageLayout title="Новый пароль" subtitle="Введите код из мессенджера">
			<form onSubmit={handleSubmit} className="flex flex-col gap-3">
				<Input
					type="text"
					placeholder="Логин"
					value={login}
					onChange={(e) => setLogin(e.target.value)}
					required
				/>
				<Input
					type="text"
					inputMode="numeric"
					placeholder="Код из Telegram/MAX"
					value={otp}
					onChange={(e) => setOtp(e.target.value.replace(/\D/g, "").slice(0, 6))}
					required
					maxLength={6}
				/>
				<Input
					type="password"
					placeholder="Новый пароль"
					value={newPassword}
					onChange={(e) => setNewPassword(e.target.value)}
					required
					minLength={6}
				/>
				{error && <p className="text-sm text-error">{error}</p>}
				<Button type="submit" variant="primary" fullWidth loading={loading}>
					Сохранить пароль
				</Button>
			</form>
		</AuthPageLayout>
	)
}
