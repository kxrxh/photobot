import type { InputHTMLAttributes, ReactNode } from "react"

type InputSize = "sm" | "md"

export interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "size"> {
	size?: InputSize
	error?: boolean
	startAdornment?: ReactNode
	endAdornment?: ReactNode
}

function getInputSizeClass(size: InputSize) {
	switch (size) {
		case "sm":
			return "input-sm"
		case "md":
			return ""
		default: {
			const _exhaustive: never = size
			return _exhaustive
		}
	}
}

export function Input({
	size = "md",
	error = false,
	startAdornment,
	endAdornment,
	className,
	...rest
}: InputProps) {
	const sizeClass = getInputSizeClass(size)

	// DaisyUI input groups are flexible; keep layout compact and predictable.
	if (startAdornment || endAdornment) {
		return (
			<label
				className={[
					"input input-bordered pbx-field",
					error ? "input-error" : "",
					sizeClass,
					"flex items-center gap-2",
					className ?? "",
				]
					.filter(Boolean)
					.join(" ")}
			>
				{startAdornment ? <span aria-hidden="true">{startAdornment}</span> : null}
				<input className="grow" {...rest} />
				{endAdornment ? <span aria-hidden="true">{endAdornment}</span> : null}
			</label>
		)
	}

	return (
		<input
			className={[
				"w-full input input-bordered pbx-field",
				error ? "input-error" : "",
				sizeClass,
				className ?? "",
			]
				.filter(Boolean)
				.join(" ")}
			{...rest}
		/>
	)
}
