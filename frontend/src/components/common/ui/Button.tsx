import type { ButtonHTMLAttributes, ReactNode } from "react"

type ButtonVariant = "primary" | "neutral" | "ghost" | "outline" | "outlinePrimary"
type ButtonSize = "sm" | "md"

export interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
	variant?: ButtonVariant
	size?: ButtonSize
	loading?: boolean
	startIcon?: ReactNode
	endIcon?: ReactNode
	fullWidth?: boolean
}

function getButtonVariantClass(variant: ButtonVariant) {
	switch (variant) {
		case "primary":
			return "btn-primary"
		case "neutral":
			return "btn-neutral"
		case "ghost":
			return "btn-ghost"
		case "outline":
			return "btn-outline"
		case "outlinePrimary":
			return "btn-outline btn-primary"
		default: {
			const _exhaustive: never = variant
			return _exhaustive
		}
	}
}

function getButtonSizeClass(size: ButtonSize) {
	switch (size) {
		case "sm":
			return "btn-sm"
		case "md":
			return ""
		default: {
			const _exhaustive: never = size
			return _exhaustive
		}
	}
}

export function Button({
	variant = "primary",
	size = "md",
	loading = false,
	startIcon,
	endIcon,
	fullWidth = false,
	className,
	disabled,
	children,
	type = "button",
	...rest
}: ButtonProps) {
	const isDisabled = disabled || loading
	const variantClass = getButtonVariantClass(variant)
	const sizeClass = getButtonSizeClass(size)

	return (
		<button
			type={type}
			className={[
				"btn",
				variantClass,
				sizeClass,
				fullWidth ? "w-full" : "",
				"pbx-btn",
				"inline-flex items-center justify-center",
				"gap-2",
				loading ? "cursor-wait" : "",
				className ?? "",
			]
				.filter(Boolean)
				.join(" ")}
			disabled={isDisabled}
			{...rest}
		>
			{loading ? <span className="loading loading-spinner loading-sm" aria-hidden="true" /> : null}
			{startIcon ? (
				<span aria-hidden="true" className="shrink-0">
					{startIcon}
				</span>
			) : null}
			{children}
			{endIcon ? (
				<span aria-hidden="true" className="shrink-0">
					{endIcon}
				</span>
			) : null}
		</button>
	)
}
