import CatalogProposalsModerator from "@/components/catalog/components/CatalogProposalsModerator"
import { isModeratorRole } from "@/components/catalog/components/CatalogProposalsShared"
import CatalogProposalsUser from "@/components/catalog/components/CatalogProposalsUser"
import { useAuth } from "@/contexts/AuthContext"

export default function CatalogProposals() {
	const { roles, userId } = useAuth()
	const isModerator = isModeratorRole(roles)
	const effectiveUserId = userId != null ? Number(userId) : undefined

	return (
		<div className="flex flex-col gap-2">
			{isModerator ? (
				<CatalogProposalsModerator userId={effectiveUserId} />
			) : (
				<CatalogProposalsUser userId={effectiveUserId} />
			)}
		</div>
	)
}
