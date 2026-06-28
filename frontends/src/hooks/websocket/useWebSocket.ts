import { useCallback, useEffect, useRef, useState } from "react"
import { log } from "@/utils/log"

export interface WebSocketMessage {
	type: string
	request_id?: string
	data?: unknown
}

export interface UseWebSocketOptions {
	url?: string
	enabled?: boolean
	/** Base delay (ms) before first reconnect attempt; grows exponentially with a cap. */
	reconnectInterval?: number
	/** Max delay (ms) between reconnect attempts. */
	maxReconnectDelayMs?: number
	/** Max reconnect attempts; use `Infinity` for no cap (default). */
	reconnectAttempts?: number
	onMessage?: (message: WebSocketMessage) => void
	onOpen?: (send: (message: Record<string, unknown>) => void) => void
	onClose?: () => void
	onError?: (error: Event) => void
}

export interface UseWebSocketReturn {
	isConnected: boolean
	connect: () => void
	disconnect: () => void
	send: (message: Record<string, unknown>) => void
}

function nextReconnectDelayMs(attemptIndex: number, baseMs: number, maxMs: number): number {
	const exp = Math.min(attemptIndex, 20)
	const backoff = Math.min(baseMs * 2 ** exp, maxMs)
	const jitter = Math.floor(Math.random() * 1000)
	return backoff + jitter
}

/**
 * Custom hook for managing WebSocket connections with automatic reconnection
 */
export function useWebSocket(options: UseWebSocketOptions): UseWebSocketReturn {
	const {
		url,
		enabled = true,
		reconnectInterval = 5000,
		maxReconnectDelayMs = 60_000,
		reconnectAttempts = Number.POSITIVE_INFINITY,
		onMessage,
		onOpen,
		onClose,
		onError,
	} = options

	const wsRef = useRef<WebSocket | null>(null)
	const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
	const connectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
	const reconnectCountRef = useRef(0)
	const shouldReconnectRef = useRef(true)
	const isConnectingRef = useRef(false)
	const connectFnRef = useRef<(() => void) | null>(null)
	const [isConnected, setIsConnected] = useState(false)

	const callbacksRef = useRef({ onMessage, onOpen, onClose, onError })
	useEffect(() => {
		callbacksRef.current = { onMessage, onOpen, onClose, onError }
	}, [onMessage, onOpen, onClose, onError])

	const clearReconnectTimeout = useCallback(() => {
		if (reconnectTimeoutRef.current) {
			clearTimeout(reconnectTimeoutRef.current)
			reconnectTimeoutRef.current = null
		}
		if (connectTimeoutRef.current) {
			clearTimeout(connectTimeoutRef.current)
			connectTimeoutRef.current = null
		}
	}, [])

	const send = useCallback((message: Record<string, unknown>) => {
		if (wsRef.current?.readyState === WebSocket.OPEN) {
			try {
				wsRef.current.send(JSON.stringify(message))
			} catch (error) {
				log.devError("[WebSocket] Failed to send message:", error)
			}
		} else {
			log.devWarn("[WebSocket] Cannot send message: not connected")
		}
	}, [])

	const connect = useCallback(() => {
		if (!url) {
			log.devWarn("[WebSocket] URL not provided")
			return
		}

		if (isConnectingRef.current) {
			return
		}

		if (wsRef.current?.readyState === WebSocket.OPEN) {
			return
		}

		if (wsRef.current?.readyState === WebSocket.CONNECTING) {
			return
		}

		if (wsRef.current) {
			try {
				wsRef.current.close()
			} catch {}
			wsRef.current = null
		}

		isConnectingRef.current = true

		try {
			const ws = new WebSocket(url)
			wsRef.current = ws

			ws.onopen = () => {
				isConnectingRef.current = false
				setIsConnected(true)
				reconnectCountRef.current = 0
				callbacksRef.current.onOpen?.(send)
			}

			ws.onmessage = (event) => {
				try {
					const message: WebSocketMessage = JSON.parse(event.data)
					callbacksRef.current.onMessage?.(message)
				} catch (error) {
					log.devError("[WebSocket] Failed to parse message:", error)
				}
			}

			ws.onclose = () => {
				isConnectingRef.current = false
				setIsConnected(false)
				callbacksRef.current.onClose?.()

				if (shouldReconnectRef.current && reconnectCountRef.current < reconnectAttempts) {
					const attemptIndex = reconnectCountRef.current
					reconnectCountRef.current++

					const delay = nextReconnectDelayMs(attemptIndex, reconnectInterval, maxReconnectDelayMs)

					reconnectTimeoutRef.current = setTimeout(() => {
						connectFnRef.current?.()
					}, delay)
				} else if (
					Number.isFinite(reconnectAttempts) &&
					reconnectCountRef.current >= reconnectAttempts
				) {
					log.devWarn("[WebSocket] Max reconnection attempts reached")
				}
			}

			ws.onerror = (error) => {
				isConnectingRef.current = false
				log.devError("[WebSocket] Connection error:", error)
				callbacksRef.current.onError?.(error)
			}
		} catch (error) {
			isConnectingRef.current = false
			log.devError("[WebSocket] Failed to create connection:", error)
			setIsConnected(false)
		}
	}, [url, reconnectInterval, maxReconnectDelayMs, reconnectAttempts, send])

	useEffect(() => {
		connectFnRef.current = connect
	}, [connect])

	const disconnect = useCallback(() => {
		shouldReconnectRef.current = false
		clearReconnectTimeout()
		reconnectCountRef.current = 0

		if (wsRef.current) {
			try {
				wsRef.current.close()
			} catch {}
			wsRef.current = null
		}

		setIsConnected(false)
	}, [clearReconnectTimeout])

	/** Close any existing socket and connect again (e.g. after tab resume). */
	const forceReconnect = useCallback(() => {
		if (!url || !enabled) return
		clearReconnectTimeout()
		reconnectCountRef.current = 0
		shouldReconnectRef.current = false
		if (wsRef.current) {
			try {
				wsRef.current.close()
			} catch {}
			wsRef.current = null
		}
		isConnectingRef.current = false
		setIsConnected(false)
		shouldReconnectRef.current = true
		connectFnRef.current?.()
	}, [url, enabled, clearReconnectTimeout])

	useEffect(() => {
		if (!enabled || !url) return

		const onResume = () => {
			if (document.visibilityState !== "visible") return
			reconnectCountRef.current = 0
			forceReconnect()
		}

		const onOnline = () => {
			reconnectCountRef.current = 0
			forceReconnect()
		}

		document.addEventListener("visibilitychange", onResume)
		window.addEventListener("online", onOnline)

		return () => {
			document.removeEventListener("visibilitychange", onResume)
			window.removeEventListener("online", onOnline)
		}
	}, [enabled, url, forceReconnect])

	useEffect(() => {
		if (reconnectTimeoutRef.current) {
			clearTimeout(reconnectTimeoutRef.current)
			reconnectTimeoutRef.current = null
		}
		if (connectTimeoutRef.current) {
			clearTimeout(connectTimeoutRef.current)
			connectTimeoutRef.current = null
		}

		if (!enabled || !url) {
			shouldReconnectRef.current = false
			if (wsRef.current) {
				try {
					wsRef.current.close()
				} catch {}
				wsRef.current = null
			}
			setIsConnected(false)
			return
		}

		shouldReconnectRef.current = true
		reconnectCountRef.current = 0

		connectTimeoutRef.current = setTimeout(() => {
			connectFnRef.current?.()
		}, 100)

		return () => {
			if (reconnectTimeoutRef.current) {
				clearTimeout(reconnectTimeoutRef.current)
				reconnectTimeoutRef.current = null
			}
			if (connectTimeoutRef.current) {
				clearTimeout(connectTimeoutRef.current)
				connectTimeoutRef.current = null
			}
			shouldReconnectRef.current = false
			if (wsRef.current) {
				try {
					wsRef.current.close()
				} catch {}
				wsRef.current = null
			}
			setIsConnected(false)
		}
	}, [url, enabled])

	return {
		isConnected,
		connect,
		disconnect,
		send,
	}
}
