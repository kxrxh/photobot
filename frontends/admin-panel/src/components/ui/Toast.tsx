import { useEffect } from "react"
import { BiSolidCheckCircle, BiSolidError, BiX } from "react-icons/bi"

interface ToastProps {
	type: "success" | "error"
	message: string
	onClose: () => void
}

export function Toast({ type, message, onClose }: ToastProps) {
	useEffect(() => {
		const timer = setTimeout(() => {
			onClose()
		}, 3000)
		return () => clearTimeout(timer)
	}, [onClose])

	const baseClasses =
		"flex items-center gap-3 p-4 rounded-xl shadow-lg border backdrop-blur-sm transition-all duration-300 ease-in-out transform"

	const typeClasses =
		type === "success"
			? "bg-success/10 border-success/20 text-success"
			: "bg-error/10 border-error/20 text-error"

	const iconClasses = type === "success" ? "text-success" : "text-error"

	const icon =
		type === "success" ? (
			<BiSolidCheckCircle className={`h-5 w-5 ${iconClasses}`} />
		) : (
			<BiSolidError className={`h-5 w-5 ${iconClasses}`} />
		)

	return (
		<div
			className={`${baseClasses} ${typeClasses} animate-in slide-in-from-top-2 fade-in duration-300`}
		>
			{icon}
			<span className="flex-1 text-sm font-medium">{message}</span>
			<button
				type="button"
				className={`p-1 rounded-lg transition-colors duration-200 ${type === "success" ? "text-success" : "text-error"}`}
				onClick={onClose}
				aria-label="Закрыть"
			>
				<BiX className="h-4 w-4" />
			</button>
		</div>
	)
}
