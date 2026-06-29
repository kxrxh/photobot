import type { ReactNode } from "react"

interface EmptyStateProps {
	icon: ReactNode
	title: string
	description?: string
	action?: ReactNode
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
	return (
		<div className="text-center py-12">
			<div className="flex flex-col items-center gap-4">
				{icon}
				<div>
					<h3 className="text-lg font-semibold text-base-content mb-2">{title}</h3>
					{description && <p className="text-sm text-base-content/70 mb-4">{description}</p>}
					{action}
				</div>
			</div>
		</div>
	)
}
