import { useEffect, useRef } from "react"

import type { AnalysisRequest } from "@/api/analysis/types"

type OpenRequestDeepLinkOptions = {
	isOpen: boolean
	/** `search.openRequest` from `/analysis/create` */
	openRequestId: string | undefined
	userId: number | null
	/** First `useQuery` result for the list; stay `undefined` until loaded */
	requestsList: { requests: AnalysisRequest[]; total: number } | undefined
	scrollToTop: () => void
	onOpenConfirmation: (request: AnalysisRequest) => void
}

/**
 * If the URL names a request (`openRequest=`), open the confirmation/results tab when that
 * row is `waiting_for_confirmation`.
 *
 * Reads only from the existing requests list query — never calls `getRequests` again — so
 * opening the page triggers a single list request. Refetch updates do not keep forcing navigation:
 * we remember the last handled `openRequest` until the param changes or the sheet closes.
 */
export function useOpenRequestDeepLink({
	isOpen,
	openRequestId,
	userId,
	requestsList,
	scrollToTop,
	onOpenConfirmation,
}: OpenRequestDeepLinkOptions): void {
	const lastHandledOpenRequestIdRef = useRef<string | null>(null)
	const previousOpenRequestIdRef = useRef<string | undefined>(undefined)

	useEffect(() => {
		if (!isOpen) {
			lastHandledOpenRequestIdRef.current = null
			previousOpenRequestIdRef.current = undefined
			return
		}

		if (previousOpenRequestIdRef.current !== openRequestId) {
			lastHandledOpenRequestIdRef.current = null
			previousOpenRequestIdRef.current = openRequestId
		}

		if (!openRequestId || !userId || requestsList === undefined) return
		if (lastHandledOpenRequestIdRef.current === openRequestId) return

		lastHandledOpenRequestIdRef.current = openRequestId

		const request = requestsList.requests.find((r) => r.id === openRequestId)
		if (request?.status !== "waiting_for_confirmation") return

		onOpenConfirmation(request)
		const scrollTid = window.setTimeout(() => scrollToTop(), 0)
		return () => clearTimeout(scrollTid)
	}, [isOpen, openRequestId, userId, requestsList, scrollToTop, onOpenConfirmation])
}
