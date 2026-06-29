import { useCallback, useState } from "react"

export type ToastType = "success" | "error"

export interface ToastState {
	type: ToastType
	message: string
}

export function useToast() {
	const [toast, setToast] = useState<ToastState | null>(null)

	const showSuccess = useCallback((message: string) => {
		setToast({ type: "success", message })
	}, [])

	const showError = useCallback((message: string) => {
		setToast({ type: "error", message })
	}, [])

	const clear = useCallback(() => {
		setToast(null)
	}, [])

	return { toast, showSuccess, showError, clear }
}
