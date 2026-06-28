import type { ReactNode } from "react"
import { FaCheck, FaClock, FaEdit, FaEye, FaTimes } from "react-icons/fa"
import type { ProposalListItem, ProposalStatus } from "@/api/catalog/types"

/** Row shape for proposal list endpoints (moderation + user lists). */
export type ProposalItem = ProposalListItem & {
	request_id?: string
}

export const STATUS_OPTIONS: { value: ProposalStatus; label: string }[] = [
	{ value: "submitted", label: "На рассмотрении" },
	{ value: "changes_requested", label: "Требуются изменения" },
	{ value: "applied", label: "Применено" },
	{ value: "rejected", label: "Отклонено" },
	{ value: "cancelled", label: "Отменено" },
]

export const isModeratorRole = (roles: Set<string>) => roles.has("admin") || roles.has("moderator")

/** Can open catalog item edit page (admin/moderator apply directly, catalog_editor creates proposal). */
export const canEditCatalogItem = (roles: Set<string>) =>
	roles.has("admin") || roles.has("moderator") || roles.has("catalog_editor")

export type ProposalStatusListConfig = {
	icon: ReactNode
	label: string
	iconWrap: string
	badge: string
}

export const getStatusConfig = (status: ProposalStatus): ProposalStatusListConfig => {
	switch (status) {
		case "submitted":
			return {
				icon: <FaEye className="h-4 w-4" />,
				label: "На рассмотрении",
				iconWrap: "bg-info/10 text-info",
				badge: "badge badge-sm badge-info badge-outline",
			}
		case "changes_requested":
			return {
				icon: <FaEdit className="h-4 w-4" />,
				label: "Требуются изменения",
				iconWrap: "bg-warning/10 text-warning",
				badge: "badge badge-sm badge-warning badge-outline",
			}
		case "applied":
			return {
				icon: <FaCheck className="h-4 w-4" />,
				label: "Применено",
				iconWrap: "bg-success/10 text-success",
				badge: "badge badge-sm badge-success badge-outline",
			}
		case "rejected":
			return {
				icon: <FaTimes className="h-4 w-4" />,
				label: "Отклонено",
				iconWrap: "bg-error/10 text-error",
				badge: "badge badge-sm badge-error badge-outline",
			}
		case "cancelled":
			return {
				icon: <FaTimes className="h-4 w-4" />,
				label: "Отменено",
				iconWrap: "bg-base-200 text-base-content/70",
				badge: "badge badge-sm badge-ghost",
			}
		default:
			return {
				icon: <FaClock className="h-4 w-4" />,
				label: "Неизвестно",
				iconWrap: "bg-base-200 text-base-content/70",
				badge: "badge badge-sm badge-ghost",
			}
	}
}

export const formatTimeAgo = (dateString: string) => {
	const createdDate = new Date(dateString)
	const diffMs = Date.now() - createdDate.getTime()
	const diffMins = Math.floor(diffMs / (1000 * 60))
	const diffHours = Math.floor(diffMins / 60)
	const diffDays = Math.floor(diffHours / 24)

	if (diffMins < 1) return "только что"
	if (diffMins < 60) return `${diffMins} мин назад`
	if (diffHours < 24) return `${diffHours} ч назад`
	return `${diffDays} д назад`
}

export const formatDateTime = (dateString: string) => new Date(dateString).toLocaleString("ru-RU")
