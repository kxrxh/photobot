package requests

import (
	"encoding/json"

	"csort.ru/analysis-service/internal/repository/requests"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/kxrxh/gopt"
)

func listRowToRequest(
	id, userID, platform, product string,
	status requests.RequestStatus,
	year *string,
	massLiter *float64,
	location *string,
	images []byte,
	mass1000, mass *float64,
	tempID, errorMessage *string,
	createdAt, updatedAt pgtype.Timestamptz,
) Request {
	return Request{
		ID:             id,
		UserID:         userID,
		Platform:       platform,
		Product:        product,
		Status:         RequestStatus(status),
		Year:           gopt.FromPtr(year).UnwrapOr(""),
		MassLiter:      massLiter,
		Location:       gopt.FromPtr(location).UnwrapOr(""),
		Images:         images,
		Classification: json.RawMessage("null"),
		Mass1000:       mass1000,
		Mass:           mass,
		TempID:         gopt.FromPtr(tempID).UnwrapOr(""),
		ErrorMessage:   gopt.FromPtr(errorMessage).UnwrapOr(""),
		CreatedAt:      createdAt.Time,
		UpdatedAt:      updatedAt.Time,
	}
}

func listRowsToRequests(rows []requests.ListRequestsByUserIDAndPlatformRow) []Request {
	out := make([]Request, len(rows))
	for i, r := range rows {
		out[i] = listRowToRequest(
			r.ID, r.UserID, r.Platform, r.Product, r.Status,
			r.Year, r.MassLiter, r.Location, r.Images,
			r.Mass1000, r.Mass, r.TempID, r.ErrorMessage,
			r.CreatedAt, r.UpdatedAt,
		)
	}
	return out
}

func listRowsToRequestsFromStatus(
	rows []requests.ListRequestsByUserIDAndPlatformAndStatusRow,
) []Request {
	out := make([]Request, len(rows))
	for i, r := range rows {
		out[i] = listRowToRequest(
			r.ID, r.UserID, r.Platform, r.Product, r.Status,
			r.Year, r.MassLiter, r.Location, r.Images,
			r.Mass1000, r.Mass, r.TempID, r.ErrorMessage,
			r.CreatedAt, r.UpdatedAt,
		)
	}
	return out
}

func multiPlatformListRowsToRequests(rows []requests.ListRequestsByUserPlatformPairsRow) []Request {
	out := make([]Request, len(rows))
	for i, r := range rows {
		out[i] = listRowToRequest(
			r.ID, r.UserID, r.Platform, r.Product, r.Status,
			r.Year, r.MassLiter, r.Location, r.Images,
			r.Mass1000, r.Mass, r.TempID, r.ErrorMessage,
			r.CreatedAt, r.UpdatedAt,
		)
	}
	return out
}

func multiPlatformListRowsToRequestsFromStatus(
	rows []requests.ListRequestsByUserPlatformPairsAndStatusRow,
) []Request {
	out := make([]Request, len(rows))
	for i, r := range rows {
		out[i] = listRowToRequest(
			r.ID, r.UserID, r.Platform, r.Product, r.Status,
			r.Year, r.MassLiter, r.Location, r.Images,
			r.Mass1000, r.Mass, r.TempID, r.ErrorMessage,
			r.CreatedAt, r.UpdatedAt,
		)
	}
	return out
}
