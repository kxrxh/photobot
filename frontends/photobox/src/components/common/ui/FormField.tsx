import type { ReactNode } from "react"

export interface FormFieldProps {
	id: string
	label?: ReactNode
	hint?: ReactNode
	error?: ReactNode
	required?: boolean
	className?: string
	children: ReactNode
}

export function FormField({
	id,
	label,
	hint,
	error,
	required,
	className,
	children,
}: FormFieldProps) {
	const describedById = hint || error ? `${id}__help` : undefined

	return (
		<div className={["w-full form-control", className ?? ""].filter(Boolean).join(" ")}>
			{label ? (
				<label htmlFor={id} className="label py-1">
					<span className="font-medium label-text">
						{label}
						{required ? <span className="text-error"> *</span> : null}
					</span>
				</label>
			) : null}

			<div aria-describedby={describedById}>{children}</div>

			{hint && !error ? (
				<div id={describedById} className="mt-1 text-xs text-base-content/70">
					{hint}
				</div>
			) : null}
			{error ? (
				<div id={describedById} className="mt-1 text-xs text-error">
					{error}
				</div>
			) : null}
		</div>
	)
}
