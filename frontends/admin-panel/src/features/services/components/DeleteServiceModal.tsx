import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { FaTrash } from "react-icons/fa"
import { Modal } from "@/components/Modal"
import { Toast } from "@/components/ui"
import { deleteService } from "@/features/services/api"
import type { Service } from "@/types/service"

interface DeleteServiceModalProps {
	service: Service
	onClose: () => void
}

export function DeleteServiceModal({ service, onClose }: DeleteServiceModalProps) {
	const [error, setError] = useState<string | null>(null)
	const queryClient = useQueryClient()

	useEffect(() => {
		const handler = (e: KeyboardEvent) => {
			if (e.key === "Escape") onClose()
		}
		window.addEventListener("keydown", handler)
		return () => window.removeEventListener("keydown", handler)
	}, [onClose])

	const mutation = useMutation({
		mutationFn: () => deleteService(service.service_id),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["services"] })
			onClose()
		},
		onError: (err: Error) => {
			setError(err.message)
		},
	})

	return (
		<Modal title="Удалить сервис" onClose={onClose}>
			{error && (
				<div className="mb-4">
					<Toast type="error" message={error} onClose={() => setError(null)} />
				</div>
			)}

			<p className="mb-6">
				Вы уверены, что хотите удалить сервис{" "}
				<span className="font-bold">{service.service_id}</span>? Это действие необратимо.
			</p>

			<div className="flex justify-end gap-2">
				<button type="button" className="btn" onClick={onClose}>
					Отмена
				</button>
				<button
					type="button"
					className="btn btn-error"
					onClick={() => mutation.mutate()}
					disabled={mutation.isPending}
				>
					{mutation.isPending ? (
						<span className="loading loading-spinner" />
					) : (
						<>
							<FaTrash className="mr-2" />
							Удалить
						</>
					)}
				</button>
			</div>
		</Modal>
	)
}
