package weed

import "csort.ru/coffeebot/internal/database"

//go:generate go run github.com/jmattheis/goverter/cmd/goverter@v1.9.4 gen -build-tags "" .

// goverter:converter
// goverter:output:file ./query_converter_gen.go
// goverter:output:package csort.ru/coffeebot/internal/weed
// goverter:extend csort.ru/coffeebot/internal/pgxconv:NonEmptyStringPtr
type queryConverter interface {
	// goverter:map PaginatedRequest.Limit Limit
	// goverter:map PaginatedRequest.Offset Offset
	// goverter:map LWMin LwMin
	// goverter:map LWMax LwMax
	ListWeedsParamsToListDB(params ListWeedsParams) database.ListWeedsParams
	ListWeedsParamsToCountDB(source database.ListWeedsParams) database.CountWeedsParams
}
