import { useCallback, useMemo } from "react"
import { API_ENDPOINTS } from "@/api/config"
import { getStoredAccessToken } from "@/lib/auth/storage"
import { log } from "@/utils/log"
import type { WebSocketMessage } from "./useWebSocket"
import { useWebSocket } from "./useWebSocket"

export interface AnalysisWebSocketMessage extends WebSocketMessage {
	type: "analysis_update" | "request_update" | "objects_update" | "auth_ok" | "auth_error"
}

export interface RequestUpdateData {
	status: string
	temp_id?: string
	error_message?: string
	excluded_objects?: string[]
}

export interface RequestUpdatePayload {
	requestId: string
	data: RequestUpdateData
}

export interface UseAnalysisWebSocketOptions {
	userId: string
	requestId?: string
	enabled?: boolean
	onRequestUpdate?: (update: RequestUpdatePayload) => void
	onAnalysisUpdate?: (data: unknown) => void
	onObjectsUpdate?: (data: unknown) => void
}

function isRequestUpdateData(data: unknown): data is RequestUpdateData {
	return (
		typeof data === "object" &&
		data !== null &&
		typeof (data as Record<string, unknown>).status === "string"
	)
}

/**
 * Builds WebSocket URL for analysis service.
 * Auth is sent as first message after connect (first-message auth pattern).
 */
function buildWebSocketUrl(baseUrl: string): string {
	const wsBaseUrl = baseUrl.replace(/^http/, "ws")
	const cleanBaseUrl = wsBaseUrl.replace(/\/$/, "")
	const wsPath = cleanBaseUrl.endsWith("/api/v1")
		? `${cleanBaseUrl}/ws`
		: `${cleanBaseUrl}/api/v1/ws`

	return wsPath
}

/**
 * Specialized hook for analysis WebSocket connections
 */
export function useAnalysisWebSocket(options: UseAnalysisWebSocketOptions) {
	const {
		userId,
		requestId,
		enabled = true,
		onRequestUpdate,
		onAnalysisUpdate,
		onObjectsUpdate,
	} = options

	const wsUrl = useMemo(() => {
		if (!userId) return undefined

		const token = getStoredAccessToken()
		if (!token) {
			log.devWarn("[AnalysisWebSocket] No access token available, skipping WebSocket connection")
			return undefined
		}

		const baseUrl = API_ENDPOINTS.analysis
		if (!baseUrl) {
			log.devWarn("[AnalysisWebSocket] VITE_ANALYSIS_API_URL not configured")
			return undefined
		}

		try {
			return buildWebSocketUrl(baseUrl)
		} catch (error) {
			log.devError("[AnalysisWebSocket] Failed to build WebSocket URL:", error)
			return undefined
		}
	}, [userId])

	const handleMessage = useCallback(
		(message: WebSocketMessage) => {
			const analysisMessage = message as AnalysisWebSocketMessage

			switch (analysisMessage.type) {
				case "auth_ok":
					log.devWarn("[AnalysisWebSocket] Authenticated")
					return
				case "auth_error":
					log.devWarn("[AnalysisWebSocket] Auth failed:", analysisMessage.data)
					return
			}

			if (requestId && message.request_id && message.request_id !== requestId) {
				return
			}

			switch (analysisMessage.type) {
				case "request_update": {
					if (!isRequestUpdateData(analysisMessage.data)) {
						log.devWarn("[AnalysisWebSocket] Invalid request_update data:", analysisMessage.data)
						break
					}
					const rid = message.request_id
					if (!rid) {
						log.devWarn("[AnalysisWebSocket] request_update missing request_id")
						break
					}
					onRequestUpdate?.({ requestId: rid, data: analysisMessage.data })
					break
				}
				case "analysis_update":
					onAnalysisUpdate?.(analysisMessage.data)
					break
				case "objects_update":
					onObjectsUpdate?.(analysisMessage.data)
					break
				default:
					log.devWarn("[AnalysisWebSocket] Unknown message type:", analysisMessage.type)
			}
		},
		[requestId, onRequestUpdate, onAnalysisUpdate, onObjectsUpdate]
	)

	return useWebSocket({
		url: wsUrl,
		enabled: enabled && !!userId && !!wsUrl,
		onMessage: handleMessage,
		onOpen: (send) => {
			const token = getStoredAccessToken()
			if (token) {
				send({ type: "auth", token })
			}
		},
		onClose: () => log.devWarn("[AnalysisWebSocket] Disconnected"),
		onError: (error) => log.devError("[AnalysisWebSocket] Error:", error),
	})
}
