import { BiSolidError } from "react-icons/bi"

interface PageErrorProps {
	title?: string
	message: string
	onRetry?: () => void
}

export function PageError({ title = "Ошибка загрузки", message, onRetry }: PageErrorProps) {
	return (
		<div className="p-4">
			<div role="alert" className="alert alert-error max-w-2xl mx-auto">
				<BiSolidError className="h-6 w-6 shrink-0 stroke-current" />
				<div>
					<h3 className="font-bold">{title}</h3>
					<div className="text-xs">{message}</div>
					{onRetry && (
						<button type="button" className="btn btn-sm btn-ghost mt-2" onClick={onRetry}>
							Повторить
						</button>
					)}
				</div>
			</div>
		</div>
	)
}
