import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Modal } from "@/components/Modal"
import { Toast } from "@/components/ui"
import { createBot } from "@/features/bots/api"
import type { BotPlatform } from "@/types/bot"

interface CreateBotModalProps {
	onClose: () => void
}

export function CreateBotModal({ onClose }: CreateBotModalProps) {
	const [botName, setBotName] = useState("")
	const [token, setToken] = useState("")
	const [platform, setPlatform] = useState<BotPlatform>("telegram")
	const [toastMessage, setToastMessage] = useState<{
		type: "success" | "error"
		message: string
	} | null>(null)
	const queryClient = useQueryClient()

	const handleSubmit = async (e: React.SubmitEvent<HTMLFormElement>) => {
		e.preventDefault()
		try {
			await createBot(botName, token, platform)
			queryClient.invalidateQueries({ queryKey: ["bots"] })
			onClose()
		} catch (error) {
			console.error("Failed to create bot:", error)
			setToastMessage({ type: "error", message: "Ошибка при создании бота." })
		}
	}

	return (
		<Modal title="Добавить нового бота" onClose={onClose}>
			{toastMessage && (
				<Toast
					type={toastMessage.type}
					message={toastMessage.message}
					onClose={() => setToastMessage(null)}
				/>
			)}
			<form onSubmit={handleSubmit}>
				<div className="form-control mb-4">
					<label className="label" htmlFor="platform">
						<span className="label-text">Платформа</span>
					</label>
					<select
						id="platform"
						className="select select-bordered w-full"
						value={platform}
						onChange={(e) => setPlatform(e.target.value as BotPlatform)}
					>
						<option value="telegram">Telegram</option>
						<option value="max">MAX</option>
					</select>
				</div>
				<div className="form-control mb-4">
					<label className="label" htmlFor="botName">
						<span className="label-text">Имя бота</span>
					</label>
					<input
						type="text"
						placeholder="Введите имя бота"
						className="input input-bordered w-full"
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
						placeholder="Введите токен"
						className="input input-bordered w-full"
						value={token}
						onChange={(e) => setToken(e.target.value)}
						id="token"
						required
					/>
				</div>
				<div className="flex justify-end gap-2">
					<button type="button" className="btn btn-ghost" onClick={onClose}>
						Отмена
					</button>
					<button
						type="submit"
						className="btn btn-primary"
						disabled={botName.length < 3 || token.length < 20}
					>
						Создать бота
					</button>
				</div>
			</form>
		</Modal>
	)
}
