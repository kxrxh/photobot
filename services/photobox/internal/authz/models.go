package authz

import "time"

type User struct {
	ID               int32     `json:"id"`
	TelegramID       int64     `json:"telegram_id"`
	OrganizationName string    `json:"organization_name"`
	INN              string    `json:"inn"`
	FullName         string    `json:"full_name"`
	PhoneNumber      string    `json:"phone_number"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type Role struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

// TokenValidationResponse represents the response from token validation
type TokenValidationResponse struct {
	Valid    bool      `json:"valid"`
	User     *User     `json:"user,omitempty"`
	Identity *Identity `json:"identity,omitempty"`
	Roles    []string  `json:"roles,omitempty"`
	Error    *string   `json:"error,omitempty"`
}

type Identity struct {
	UserID     int32    `json:"user_id"`
	TelegramID *int64   `json:"telegram_id,omitempty"`
	MaxID      *int64   `json:"max_id,omitempty"`
	Roles      []string `json:"roles"`
}

type CustomClaims struct {
	UserID     int32    `json:"user_id"`
	TelegramID *int64   `json:"telegram_id,omitempty"`
	MaxID      *int64   `json:"max_id,omitempty"`
	Roles      []string `json:"roles"`
	Type       string   `json:"type"`
}

// ExternalUserID returns the messenger-specific user ID (Telegram or MAX), if available.
func (i *Identity) ExternalUserID() (int64, bool) {
	if i == nil {
		return 0, false
	}
	if i.TelegramID != nil {
		return *i.TelegramID, true
	}
	if i.MaxID != nil {
		return *i.MaxID, true
	}
	return 0, false
}
