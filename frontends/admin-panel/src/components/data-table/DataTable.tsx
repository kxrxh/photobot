import type { ReactNode } from "react"

interface DataTableProps {
	children: ReactNode
	className?: string
}

export function DataTable({ children, className }: DataTableProps) {
	return (
		<div className={`card bg-base-100 shadow-xl ${className ?? ""}`}>
			<div className="card-body p-0">
				<div className="overflow-x-auto">
					<table className="table w-full">{children}</table>
				</div>
			</div>
		</div>
	)
}
