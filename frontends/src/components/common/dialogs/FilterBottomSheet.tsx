import { AnimatePresence, motion } from "framer-motion"
import type { ReactNode } from "react"
import { useEffect } from "react"
import { createPortal } from "react-dom"
import { FaFilter } from "react-icons/fa"
import { IoClose } from "react-icons/io5"
import { SheetHeaderCloseButton } from "@/components/common/ui/SheetHeaderActions"

export interface FilterBottomSheetProps {
	isOpen: boolean
	onClose: () => void
	title: string
	icon?: ReactNode
	children: ReactNode
	onClear: () => void
	onApply: () => void
}

const defaultIcon = <FaFilter className="mr-3 text-primary" aria-hidden="true" />

const FilterBottomSheet = ({
	isOpen,
	onClose,
	title,
	icon = defaultIcon,
	children,
	onClear,
	onApply,
}: FilterBottomSheetProps) => {
	useEffect(() => {
		if (!isOpen) return
		const handleKeyDown = (event: KeyboardEvent) => {
			if (event.key === "Escape") {
				onClose()
			}
		}
		window.addEventListener("keydown", handleKeyDown)
		return () => window.removeEventListener("keydown", handleKeyDown)
	}, [isOpen, onClose])

	return createPortal(
		<AnimatePresence>
			{isOpen && (
				<>
					<motion.div
						initial={{ opacity: 0 }}
						animate={{ opacity: 1 }}
						exit={{ opacity: 0 }}
						transition={{ duration: 0.2 }}
						className="fixed inset-0 z-50 bg-black/50 backdrop-blur-md"
						onClick={onClose}
						aria-label="Закрыть фильтры"
					/>

					<motion.div
						initial={{ y: "100%" }}
						animate={{ y: 0 }}
						exit={{ y: "100%" }}
						transition={{
							type: "tween",
							damping: 25,
							stiffness: 300,
							duration: 0.3,
						}}
						className="fixed bottom-0 left-0 right-0 z-50 w-full mx-auto bg-base-100 rounded-t-3xl shadow-2xl border-t border-base-200"
					>
						<div className="flex justify-center pt-4 pb-3">
							<div className="w-12 h-1.5 bg-base-300 rounded-full" />
						</div>

						<header className="flex items-center justify-between px-6 py-4 border-b border-base-200">
							<h3 className="flex items-center text-xl font-semibold text-base-content">
								{icon}
								{title}
							</h3>
							<SheetHeaderCloseButton onClick={onClose} />
						</header>

						<div className="flex-1 overflow-y-auto max-h-[65vh]">
							<div className="px-6 py-4 space-y-6">{children}</div>
						</div>

						<footer className="px-4 py-3 border-t border-base-200 bg-base-100/95 backdrop-blur-sm">
							<div className="flex flex-col gap-3">
								<button
									type="button"
									onClick={onClear}
									className="w-full btn btn-soft btn-ghost gap-2 rounded-2xl py-3"
								>
									<IoClose size={16} />
									<span className="font-medium">Сбросить фильтры</span>
								</button>
								<button
									type="button"
									onClick={onApply}
									className="w-full btn btn-primary gap-2 rounded-2xl py-3"
								>
									<span className="font-medium">Применить</span>
								</button>
							</div>
						</footer>
					</motion.div>
				</>
			)}
		</AnimatePresence>,
		document.body
	)
}

export default FilterBottomSheet
