package proposal

import "csort.ru/coffeebot/internal/database"

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen -build-tags "" .

// goverter:converter
// goverter:output:file ./row_converter_gen.go
// goverter:output:package csort.ru/coffeebot/internal/proposal
// goverter:extend csort.ru/coffeebot/internal/pgxconv:TimestampToPtr csort.ru/coffeebot/internal/pgxconv:TimestamptzToPtr csort.ru/coffeebot/internal/pgxconv:TimestampToTime csort.ru/coffeebot/internal/pgxconv:NonEmptyStringPtr
type rowConverter interface {
	ListRowToItem(source database.ListCatalogProposalsWithPendingWeedRow) ProposalListItem
	// goverter:map PaginatedRequest.Limit Limit
	// goverter:map PaginatedRequest.Offset Offset
	ListProposalsParamsToListDB(
		params ListProposalsParams,
	) database.ListCatalogProposalsWithPendingWeedParams
	ListProposalsParamsToCountDB(params ListProposalsParams) database.CountCatalogProposalsParams
}
