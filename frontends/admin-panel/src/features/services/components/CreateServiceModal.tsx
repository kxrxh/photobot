import { useMutation, useQueryClient } from "@tanstack/react-query"
import { useEffect, useState } from "react"
import { FaPlus } from "react-icons/fa"
import { Modal } from "@/components/Modal"
import { Toast } from "@/components/ui"
import { createService } from "@/features/services/api"

interface CreateServiceModalProps {
	onClose: () => void
}

export function CreateServiceModal({ onClose }: CreateServiceModalProps) {
	const [serviceId, setServiceId] = useState("")
	const [serviceSecret, setServiceSecret] = useState("")
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
		mutationFn: () => createService(serviceId, serviceSecret),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["services"] })
			onClose()
		},
		onError: (err: Error) => {
			setError(err.message)
		},
	})

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault()
		mutation.mutate()
	}

	return (
		<Modal title="Добавить новый сервис" onClose={onClose}>
			{error && (
				<div className="mb-4">
					<Toast type="error" message={error} onClose={() => setError(null)} />
				</div>
			)}

			<form onSubmit={handleSubmit} className="space-y-4">
				<div className="form-control">
					<label className="label" htmlFor="serviceId">
						<span className="label-text">ID сервиса</span>
					</label>
					<input
						type="text"
						id="serviceId"
						className="input input-bordered w-full"
						placeholder="Например: classification-service"
						value={serviceId}
						onChange={(e) => setServiceId(e.target.value)}
						required
					/>
				</div>
				<div className="form-control">
					<label className="label" htmlFor="serviceSecret">
						<span className="label-text">Секрет сервиса</span>
					</label>
					<input
						type="password"
						id="serviceSecret"
						className="input input-bordered w-full"
						placeholder="Введите секретный ключ"
						value={serviceSecret}
						onChange={(e) => setServiceSecret(e.target.value)}
						required
					/>
				</div>
				<div className="flex justify-end gap-2">
					<button type="button" className="btn" onClick={onClose}>
						Отмена
					</button>
					<button type="submit" className="btn btn-primary" disabled={mutation.isPending}>
						{mutation.isPending ? (
							<span className="loading loading-spinner" />
						) : (
							<>
								<FaPlus className="mr-2" />
								Создать
							</>
						)}
					</button>
				</div>
			</form>
		</Modal>
	)
}
