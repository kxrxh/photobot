package pgxconv

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// TimestampToPtr converts a nullable Postgres timestamp to *time.Time.
func TimestampToPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// TimestamptzToPtr converts a nullable Postgres timestamptz to *time.Time.
func TimestamptzToPtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

// TimestampToTime converts a Postgres timestamp to time.Time (matches sqlc row .Time usage).
func TimestampToTime(t pgtype.Timestamp) time.Time {
	return t.Time
}
