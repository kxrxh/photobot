import { createFileRoute } from "@tanstack/react-router"
import CatalogProposals from "@/components/catalog/components/CatalogProposals"

export const Route = createFileRoute("/_authenticated/catalog/proposals/")({
	component: () => <CatalogProposals />,
})
