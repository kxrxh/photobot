import { Fragment, useId, useRef } from "react"
import { FaChevronDown } from "react-icons/fa"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"

export type ModalSelectOption = { value: string; label: string }

export type ModalSelectGroup = { label: string; options: ModalSelectOption[] }

export type ModalSelectProps = {
	id: string
	title: string
	placeholder: string
	value: string
	onChange: (value: string) => void
	disabled?: boolean
	/** Default true: first row sets `value` to `""`. */
	clearable?: boolean
	size?: "md" | "sm"
} & (
	| { options: ModalSelectOption[]; groupedOptions?: never }
	| { groupedOptions: ModalSelectGroup[]; options?: never }
)

const TRIGGER_MD =
	"flex h-12 min-h-12 w-full items-center justify-between gap-2 rounded-xl border border-base-300 bg-base-100 px-3 text-left text-sm shadow-sm transition-[border-color,box-shadow] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/25 disabled:cursor-not-allowed disabled:opacity-50"

const TRIGGER_SM =
	"flex h-10 min-h-10 w-full items-center justify-between gap-2 rounded-lg border border-base-300 bg-base-100 px-2.5 text-left text-sm shadow-sm transition-[border-color,box-shadow] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/25 disabled:cursor-not-allowed disabled:opacity-50"

const ROW_MD =
	"flex min-h-12 w-full items-center rounded-xl px-3 text-left text-sm transition-colors hover:bg-base-200/90 active:bg-base-300/60"

const ROW_SM =
	"flex min-h-10 w-full items-center rounded-lg px-2.5 text-left text-sm transition-colors hover:bg-base-200/90 active:bg-base-300/60"

/**
 * Mobile-friendly select: opens a scrollable modal instead of a native &lt;select&gt;
 * (native pickers often clip or scroll the whole page in WebViews / Telegram).
 */
export function ModalSelect({
	id,
	title,
	placeholder,
	value,
	onChange,
	disabled = false,
	clearable = true,
	size = "md",
	...source
}: ModalSelectProps) {
	const dialogRef = useRef<HTMLDialogElement>(null)
	const titleId = useId()
	const groupedOptions = "groupedOptions" in source ? source.groupedOptions : undefined
	const flatOptions: ModalSelectOption[] =
		"groupedOptions" in source ? [] : (("options" in source ? source.options : undefined) ?? [])
	const isGrouped = groupedOptions != null && groupedOptions.length > 0

	const triggerClass = size === "sm" ? TRIGGER_SM : TRIGGER_MD
	const rowClass = size === "sm" ? ROW_SM : ROW_MD

	const selectedLabel = (() => {
		if (value === "") return undefined
		if (isGrouped) {
			for (const g of groupedOptions) {
				const hit = g.options.find((o) => o.value === value)
				if (hit) return hit.label
			}
			return undefined
		}
		return flatOptions.find((o) => o.value === value)?.label
	})()

	const openModal = () => {
		if (disabled) return
		dialogRef.current?.showModal()
	}

	const closeModal = () => {
		dialogRef.current?.close()
	}

	const pick = (next: string) => {
		onChange(next)
		closeModal()
	}

	const optionCount = isGrouped
		? groupedOptions.reduce((n, g) => n + g.options.length, 0)
		: flatOptions.length

	return (
		<>
			<button
				id={id}
				type="button"
				className={triggerClass}
				disabled={disabled}
				aria-haspopup="dialog"
				title={title}
				onClick={openModal}
			>
				<span
					className={`min-w-0 flex-1 truncate ${selectedLabel ? "text-base-content" : "text-base-content/50"}`}
				>
					{selectedLabel ?? placeholder}
				</span>
				<FaChevronDown className="h-3.5 w-3.5 shrink-0 text-base-content/45" aria-hidden />
			</button>

			<dialog ref={dialogRef} className="modal" aria-labelledby={titleId}>
				<div className="modal-box pbx-dialog flex h-auto max-h-[min(75dvh,36rem)] w-[calc(100vw-1.25rem)] max-w-lg flex-col gap-0 overflow-hidden p-0 sm:w-full">
					<div className="flex shrink-0 items-center justify-between gap-2 border-b border-base-300 px-3 py-2.5">
						<h2 id={titleId} className="min-w-0 flex-1 text-base font-semibold leading-snug">
							{title}
						</h2>
						<SheetHeaderCloseButton
							aria-label="Закрыть"
							title="Закрыть"
							onClick={closeModal}
							className="shrink-0"
						/>
					</div>

					<div className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-2">
						<div className="flex flex-col gap-2">
							<div className="flex flex-col gap-1" role="listbox" aria-label={title}>
								{clearable ? (
									<button
										type="button"
										role="option"
										aria-selected={value === ""}
										className={`${rowClass} ${value === "" ? "bg-primary/10 font-medium text-primary" : "text-base-content/70"}`}
										onClick={() => pick("")}
									>
										{placeholder}
									</button>
								) : null}

								{isGrouped
									? groupedOptions.map((group) => (
											<Fragment key={group.label}>
												<div
													className="bg-base-100/95 px-2 py-1.5 text-xs font-semibold uppercase tracking-wide text-base-content/45"
													role="presentation"
												>
													{group.label}
												</div>
												{group.options.map((opt) => {
													const selected = value === opt.value
													return (
														<button
															key={opt.value}
															type="button"
															role="option"
															aria-selected={selected}
															className={`${rowClass} ${selected ? "border border-primary/25 bg-primary/10 font-medium text-primary" : "text-base-content"}`}
															onClick={() => pick(opt.value)}
														>
															<span className="wrap-break-word">{opt.label}</span>
														</button>
													)
												})}
											</Fragment>
										))
									: flatOptions.map((opt) => {
											const selected = value === opt.value
											return (
												<button
													key={opt.value}
													type="button"
													role="option"
													aria-selected={selected}
													className={`${rowClass} ${selected ? "border border-primary/25 bg-primary/10 font-medium text-primary" : "text-base-content"}`}
													onClick={() => pick(opt.value)}
												>
													<span className="wrap-break-word">{opt.label}</span>
												</button>
											)
										})}
							</div>

							{optionCount === 0 ? (
								<p className="px-2 py-1 text-center text-sm text-base-content/60">
									Нет вариантов для выбора
								</p>
							) : null}
						</div>
					</div>
				</div>

				<form method="dialog" className="modal-backdrop pbx-dialog-backdrop">
					<button type="button" onClick={closeModal}>
						close
					</button>
				</form>
			</dialog>
		</>
	)
}
