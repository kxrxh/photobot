package auth

const (
	AdminRole   = "admin"
	ServiceRole = "service"
)

const (
	LocalsUserID     = "user_id"
	LocalsTelegramID = "telegram_id"
	LocalsMaxID      = "max_id"
	LocalsUserRoles  = "user_roles"
	LocalsGrantType  = "grant_type"
)

type GrantType string

const (
	GrantTypeInitData     GrantType = "initdata"
	GrantTypeService      GrantType = "client_credentials"
	GrantTypePassword     GrantType = "password"
	GrantTypeUserPassword GrantType = "user_password"
)

func (gt GrantType) String() string {
	return string(gt)
}

func IsValidGrantType(s string) bool {
	switch GrantType(s) {
	case GrantTypeInitData, GrantTypeService, GrantTypePassword, GrantTypeUserPassword:
		return true
	}
	return false
}

type LoginParams struct {
	InitData          *string   `json:"initData"`
	BotName           *string   `json:"botName"`
	MessengerPlatform *string   `json:"messenger_platform"`
	ServiceID         *string   `json:"service_id"`
	ServiceSecret     *string   `json:"service_secret"`
	Audience          *string   `json:"audience"`
	Login             *string   `json:"login"`
	Password          *string   `json:"password"`
	GTY               GrantType `json:"gty"`
}

type KeyPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AdminLoginRequest represents the request body for admin login
type AdminLoginRequest struct {
	Login    string `json:"login"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

// ServiceLoginRequest represents the request body for service login
type ServiceLoginRequest struct {
	ServiceID     string `json:"service_id"     validate:"required"`
	ServiceSecret string `json:"service_secret" validate:"required"`
	Audience      string `json:"audience"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginResponse represents the response body for user login
type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Roles        []string `json:"roles"`
}

// AdminLoginResponse represents the response body for admin login
type AdminLoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Roles        []string `json:"roles"`
}
