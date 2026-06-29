import { useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { IoClose } from "react-icons/io5"
import { Modal } from "@/components/Modal"
import { Toast } from "@/components/ui"
import { updateBot } from "@/features/bots/api"
import type { Bot } from "@/types/bot"

interface EditBotModalProps {
	bot: Bot
	onClose: () => void
}

export function EditBotModal({ bot, onClose }: EditBotModalProps) {
	const [botName, setBotName] = useState(bot.name || "")
	const [token, setToken] = useState(bot.token || "")
	const [toastMessage, setToastMessage] = useState<{
		type: "success" | "error"
		message: string
	} | null>(null)
	const queryClient = useQueryClient()

	const isNameChanged = botName !== (bot.name || "")
	const isTokenChanged = token !== (bot.token || "")

	const isFormValid =
		((isNameChanged && botName.length >= 3) ||
			(isTokenChanged && (token.length === 0 || token.length >= 20))) &&
		(isNameChanged || isTokenChanged)

	useEffect(() => {
		setBotName(bot.name || "")
		setToken(bot.token || "")
	}, [bot])

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault()
		try {
			const updatedName = isNameChanged ? botName : undefined
			const updatedToken = isTokenChanged ? token : undefined
			await updateBot(bot.id, updatedName, updatedToken)
			queryClient.invalidateQueries({ queryKey: ["bots"] })
			onClose()
		} catch (error) {
			console.error("Failed to update bot:", error)
			setToastMessage({
				type: "error",
				message: "Ошибка при обновлении бота.",
			})
		}
	}

	return (
		<Modal onClose={onClose} className="max-w-md">
			{toastMessage && (
				<Toast
					type={toastMessage.type}
					message={toastMessage.message}
					onClose={() => setToastMessage(null)}
				/>
			)}
			<div className="flex justify-between items-center mb-4">
				<div>
					<h2 className="text-2xl font-bold">Редактировать бота</h2>
					<span
						className={`badge badge-sm mt-1 ${
							(bot.platform ?? "telegram") === "max" ? "badge-primary" : "badge-ghost"
						}`}
					>
						{(bot.platform ?? "telegram") === "max" ? "MAX" : "Telegram"}
					</span>
				</div>
				<button
					type="button"
					className="btn btn-ghost btn-sm"
					onClick={onClose}
					aria-label="Закрыть"
				>
					<IoClose className="size-5" />
				</button>
			</div>
			<form onSubmit={handleSubmit}>
				<div className="form-control mb-4">
					<label className="label" htmlFor="botName">
						<span className="label-text">Имя бота</span>
					</label>
					<input
						type="text"
						placeholder="Введите имя бота"
						className="input input-bordered w-full focus:outline-none focus:ring-0 focus:border-primary"
						value={botName}
						onChange={(e) => setBotName(e.target.value)}
						id="botName"
						required
					/>
				</div>
				<div className="form-control mb-6">
					<label className="label" htmlFor="token">
						<span className="label-text">Токен</span>
					</label>
					<input
						type="text"
						placeholder="Если поле пустое, токен не будет изменен"
						className="input input-bordered w-full focus:outline-none focus:ring-0 focus:border-primary"
						value={token}
						onChange={(e) => setToken(e.target.value)}
						id="token"
					/>
				</div>
				<div className="flex gap-2">
					<button type="button" className="btn flex-1" onClick={onClose}>
						Отмена
					</button>
					<button type="submit" className="btn btn-primary flex-1" disabled={!isFormValid}>
						Сохранить изменения
					</button>
				</div>
			</form>
		</Modal>
	)
}
