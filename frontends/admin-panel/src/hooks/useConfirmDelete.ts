import { useCallback, useEffect, useState } from "react"

const DEFAULT_DELAY_SECONDS = 5

interface UseConfirmDeleteOptions {
	delaySeconds?: number
	onConfirm: (id: number) => void
	onCancel?: () => void
}

/**
 * Hook for delete-with-confirmation flow (e.g. "click again within 5s to confirm").
 * Returns state and handlers for the two-step delete pattern.
 */
export function useConfirmDelete<T extends { id: number }>({
	delaySeconds = DEFAULT_DELAY_SECONDS,
	onConfirm,
	onCancel,
}: UseConfirmDeleteOptions) {
	const [pendingId, setPendingId] = useState<number | null>(null)
	const [secondsLeft, setSecondsLeft] = useState(0)
	const [timerId, setTimerId] = useState<ReturnType<typeof setTimeout> | null>(null)

	useEffect(() => {
		if (secondsLeft <= 0) return
		const interval = setInterval(() => {
			setSecondsLeft((prev) => {
				if (prev <= 1) {
					clearInterval(interval)
					return 0
				}
				return prev - 1
			})
		}, 1000)
		return () => clearInterval(interval)
	}, [secondsLeft])

	const handleDelete = useCallback(
		(item: T) => {
			if (pendingId === item.id) {
				if (timerId) {
					clearTimeout(timerId)
					setTimerId(null)
				}
				onConfirm(item.id)
				setPendingId(null)
				setSecondsLeft(0)
			} else {
				setPendingId(item.id)
				setSecondsLeft(delaySeconds)
				const timer = setTimeout(() => {
					setPendingId(null)
					setTimerId(null)
					setSecondsLeft(0)
					onCancel?.()
				}, delaySeconds * 1000)
				setTimerId(timer)
			}
		},
		[delaySeconds, onCancel, onConfirm, pendingId, timerId]
	)

	const cancel = useCallback(() => {
		if (timerId) {
			clearTimeout(timerId)
			setTimerId(null)
		}
		setPendingId(null)
		setSecondsLeft(0)
		onCancel?.()
	}, [onCancel, timerId])

	return {
		pendingId,
		secondsLeft,
		handleDelete,
		cancel,
		isPending: (item: T) => pendingId === item.id,
	}
}
