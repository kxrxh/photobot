package analysis

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func idFilterParams(filter string) (idExact pgtype.UUID, idPrefix string) {
	if filter == "" {
		return pgtype.UUID{}, ""
	}
	if id, err := uuid.Parse(filter); err == nil {
		return pgtype.UUID{Bytes: id, Valid: true}, ""
	}
	return pgtype.UUID{}, filter
}
