import { toast } from "sonner"

const DEFAULT_DURATION = 3000

function toSonnerPosition(position?: "top" | "bottom"): "top-center" | "bottom-center" | undefined {
	if (!position) return undefined
	return position === "top" ? "top-center" : "bottom-center"
}

function safeDuration(duration?: number): number {
	if (typeof duration === "number" && duration > 0 && duration !== Number.POSITIVE_INFINITY) {
		return duration
	}
	return DEFAULT_DURATION
}

export function useAlert() {
	const showSuccess = (message: string, duration?: number, position?: "top" | "bottom") => {
		toast.success(message, {
			duration: safeDuration(duration),
			position: toSonnerPosition(position),
		})
	}

	const showError = (message: string, duration?: number, position?: "top" | "bottom") => {
		toast.error(message, {
			duration: safeDuration(duration),
			position: toSonnerPosition(position),
		})
	}

	const showWarning = (message: string, duration?: number, position?: "top" | "bottom") => {
		toast.warning(message, {
			duration: safeDuration(duration),
			position: toSonnerPosition(position),
		})
	}

	const showInfo = (message: string, duration?: number, position?: "top" | "bottom") => {
		toast.info(message, {
			duration: safeDuration(duration),
			position: toSonnerPosition(position),
		})
	}

	const showAlert = (alert: {
		type: "success" | "error" | "warning" | "info"
		message: string
		duration?: number
		position?: "top" | "bottom"
	}) => {
		switch (alert.type) {
			case "success":
				showSuccess(alert.message, alert.duration, alert.position)
				break
			case "error":
				showError(alert.message, alert.duration, alert.position)
				break
			case "warning":
				showWarning(alert.message, alert.duration, alert.position)
				break
			case "info":
				showInfo(alert.message, alert.duration, alert.position)
				break
		}
	}

	const removeAlert = (id: string) => {
		toast.dismiss(id)
	}

	const clearAlerts = () => {
		toast.dismiss()
	}

	return {
		alerts: [] as Array<{ id: string; type: string; message: string }>,
		showSuccess,
		showError,
		showWarning,
		showInfo,
		showAlert,
		removeAlert,
		clearAlerts,
	}
}
