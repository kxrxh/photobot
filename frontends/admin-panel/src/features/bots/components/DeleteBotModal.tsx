import { useQueryClient } from "@tanstack/react-query"
import { useState } from "react"
import { Modal } from "@/components/Modal"
import { Toast } from "@/components/ui"
import { deleteBot } from "@/features/bots/api"
import type { Bot } from "@/types/bot"

interface DeleteBotModalProps {
	bot: Bot
	onClose: () => void
}

export function DeleteBotModal({ bot, onClose }: DeleteBotModalProps) {
	const [toastMessage, setToastMessage] = useState<{
		type: "success" | "error"
		message: string
	} | null>(null)
	const queryClient = useQueryClient()

	const handleDelete = async () => {
		try {
			await deleteBot(bot.id)
			queryClient.invalidateQueries({ queryKey: ["bots"] })
			onClose()
		} catch (error) {
			console.error("Failed to delete bot:", error)
			setToastMessage({ type: "error", message: "Ошибка при удалении бота." })
		}
	}

	return (
		<Modal title="Удалить бота" onClose={onClose}>
			{toastMessage && (
				<Toast
					type={toastMessage.type}
					message={toastMessage.message}
					onClose={() => setToastMessage(null)}
				/>
			)}
			<p className="mb-6">
				Вы уверены, что хотите удалить бота "{bot.name || "Не указано"}"? Это действие необратимо.
			</p>
			<div className="flex justify-end gap-2">
				<button type="button" className="btn btn-ghost" onClick={onClose}>
					Отмена
				</button>
				<button type="button" className="btn btn-error" onClick={handleDelete}>
					Удалить
				</button>
			</div>
		</Modal>
	)
}
