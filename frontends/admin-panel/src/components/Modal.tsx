import { type ReactNode, useEffect } from "react"
import { FaTimes } from "react-icons/fa"

interface ModalProps {
	title?: string
	onClose: () => void
	children: ReactNode
	className?: string
}

export function Modal({ title, onClose, children, className }: ModalProps) {
	useEffect(() => {
		const handler = (e: KeyboardEvent) => {
			if (e.key === "Escape") onClose()
		}
		window.addEventListener("keydown", handler)
		return () => window.removeEventListener("keydown", handler)
	}, [onClose])

	return (
		<div
			className="fixed inset-0 bg-black/50 flex items-center justify-center z-50"
			onClick={onClose}
			onKeyDown={(e) => {
				if (e.key === "Enter" || e.key === " ") {
					e.preventDefault()
					onClose()
				}
			}}
			role="dialog"
			aria-modal="true"
		>
			<div
				role="document"
				className={`bg-base-100 rounded-lg shadow-xl p-6 w-full max-w-md ${className ?? ""}`}
				onClick={(e) => e.stopPropagation()}
				onKeyDown={(e) => e.stopPropagation()}
			>
				{title && (
					<div className="flex justify-between items-center mb-4">
						<h2 className="text-2xl font-bold">{title}</h2>
						<button
							type="button"
							className="btn btn-ghost btn-sm"
							onClick={onClose}
							aria-label="Закрыть"
						>
							<FaTimes className="h-5 w-5" />
						</button>
					</div>
				)}
				{children}
			</div>
		</div>
	)
}
