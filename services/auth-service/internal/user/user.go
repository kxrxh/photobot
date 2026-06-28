package user

import (
	"strings"

	"csort.ru/auth-service/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	database.User
	Roles []string `json:"roles,omitempty"`
}

type UserList []User

type WebRegisterRequest struct {
	Login            string `json:"login"             validate:"required,min=3,max=32"`
	Password         string `json:"password"          validate:"required,min=6"`
	OrganizationName string `json:"organization_name" validate:"omitempty,max=128"`
	INN              string `json:"inn"               validate:"omitempty,min=10,max=12,numeric"`
	FullName         string `json:"full_name"         validate:"omitempty,max=128"`
	PhoneNumber      string `json:"phone_number"      validate:"omitempty,min=10,max=32"`
}

type RegisterRequest struct {
	OrganizationName string `json:"organization_name" validate:"omitempty,max=128"`
	INN              string `json:"inn"               validate:"omitempty,min=10,max=12,numeric"`
	FullName         string `json:"full_name"         validate:"omitempty,max=128"`
	PhoneNumber      string `json:"phone_number"      validate:"omitempty,min=10,max=32"`
}

type UserRequest struct {
	OrganizationName string `json:"organization_name" validate:"omitempty,max=128"`
	INN              string `json:"inn"               validate:"omitempty,min=10,max=12,numeric"`
	FullName         string `json:"full_name"         validate:"omitempty,max=128"`
	PhoneNumber      string `json:"phone_number"      validate:"omitempty,min=10,max=32"`
}

func toPgText(s string) pgtype.Text {
	s = strings.TrimSpace(s)
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

func fromPgText(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}
