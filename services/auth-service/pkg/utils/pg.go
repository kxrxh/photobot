package utils

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// UniqueViolationMessage maps Postgres unique violations (SQLSTATE 23505) to a message; returns "" if err is not 23505.
func UniqueViolationMessage(err error, messages map[string]string, defaultMsg string) string {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return ""
	}
	if msg, ok := messages[pgErr.ConstraintName]; ok {
		return msg
	}
	return defaultMsg
}
