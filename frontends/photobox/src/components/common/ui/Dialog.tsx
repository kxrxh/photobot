import type { ReactNode } from "react"
import { useId, useMemo } from "react"

type DialogSize = "sm" | "md" | "lg"

export interface DialogProps {
	open: boolean
	onClose: () => void
	title?: ReactNode
	children: ReactNode
	footer?: ReactNode
	size?: DialogSize
	closeOnBackdrop?: boolean
	className?: string
}

function getDialogSizeClass(size: DialogSize) {
	switch (size) {
		case "sm":
			return "max-w-sm"
		case "md":
			return "max-w-md"
		case "lg":
			return "max-w-2xl"
		default: {
			const _exhaustive: never = size
			return _exhaustive
		}
	}
}

export function Dialog({
	open,
	onClose,
	title,
	children,
	footer,
	size = "md",
	closeOnBackdrop = true,
	className,
}: DialogProps) {
	const reactId = useId()
	const titleId = useMemo(() => `dialog_${reactId}__title`, [reactId])
	const sizeClass = getDialogSizeClass(size)

	return (
		<dialog
			aria-modal="true"
			aria-labelledby={title ? titleId : undefined}
			className={["modal backdrop-blur-xs", open ? "modal-open" : "", className ?? ""]
				.filter(Boolean)
				.join(" ")}
			onKeyDown={(e) => {
				if (e.key === "Escape") onClose()
			}}
		>
			<div
				className={["modal-box", "pbx-dialog", sizeClass, "p-4", "gap-3"].filter(Boolean).join(" ")}
			>
				{title ? (
					<h3 id={titleId} className="text-base font-semibold">
						{title}
					</h3>
				) : null}

				<div className="pt-1">{children}</div>

				{footer ? <div className="modal-action mt-2">{footer}</div> : null}
			</div>

			{closeOnBackdrop ? (
				<form method="dialog" className="modal-backdrop pbx-dialog-backdrop">
					<button type="button" onClick={onClose}>
						close
					</button>
				</form>
			) : null}
		</dialog>
	)
}
