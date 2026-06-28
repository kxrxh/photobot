package utils

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func PgTimestampToTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}
