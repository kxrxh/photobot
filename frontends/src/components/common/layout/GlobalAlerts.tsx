import { useEffect } from "react"
import { Toaster, toast } from "sonner"
import { useMessenger } from "@/hooks/useMessenger"

export default function GlobalAlerts() {
	const { isDark } = useMessenger()

	useEffect(() => {
		const handleClick = (event: MouseEvent) => {
			const target = event.target as HTMLElement
			const toastEl = target.closest("[data-sonner-toast]")
			if (!toastEl) return
			if (target.closest("[data-close-button]") || target.closest("button")) return

			const toastList = toastEl.parentElement?.querySelectorAll("[data-sonner-toast]")
			if (!toastList?.length) return
			const index = Array.from(toastList).indexOf(toastEl as Element)
			const toasts = toast.getToasts()
			const visible = toasts.filter((t) => "id" in t && !("dismiss" in t && t.dismiss))
			const targetToast = visible[index]
			if (targetToast && "id" in targetToast) {
				toast.dismiss(targetToast.id)
			}
		}

		document.addEventListener("click", handleClick, true)
		return () => document.removeEventListener("click", handleClick, true)
	}, [])

	return (
		<Toaster
			position="bottom-center"
			visibleToasts={5}
			expand
			richColors
			theme={isDark ? "dark" : "light"}
			mobileOffset={{ bottom: 16 }}
			duration={3000}
			toastOptions={{
				duration: 3000,
				closeButton: true,
			}}
		/>
	)
}
