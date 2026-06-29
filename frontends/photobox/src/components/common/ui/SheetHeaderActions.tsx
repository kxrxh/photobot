import type { ButtonHTMLAttributes } from "react"
import { IoArrowBack, IoClose } from "react-icons/io5"

const backIconCircleClass = "btn btn-soft btn-circle btn-sm shrink-0"

const closeIconCircleClass = "btn btn-soft btn-circle btn-sm shrink-0 btn-neutral"

type HeaderIconButtonProps = Omit<ButtonHTMLAttributes<HTMLButtonElement>, "className" | "type"> & {
	className?: string
}

export function SheetHeaderBackButton({ className, children, ...props }: HeaderIconButtonProps) {
	return (
		<button
			type="button"
			className={[backIconCircleClass, className].filter(Boolean).join(" ")}
			{...props}
		>
			{children ?? <IoArrowBack size={18} aria-hidden />}
		</button>
	)
}

export function SheetHeaderCloseButton({
	className,
	"aria-label": ariaLabel = "Закрыть",
	title,
	...props
}: HeaderIconButtonProps & { title?: string }) {
	return (
		<button
			type="button"
			className={[closeIconCircleClass, className].filter(Boolean).join(" ")}
			{...props}
			aria-label={ariaLabel}
			title={title ?? ariaLabel}
		>
			<IoClose size={18} aria-hidden />
		</button>
	)
}
