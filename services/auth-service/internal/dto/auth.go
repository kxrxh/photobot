package dto

type AdminLoginRequest struct {
	Login    string `json:"login"    validate:"required"`
	Password string `json:"password" validate:"required"`
}

type ServiceLoginRequest struct {
	ServiceID     string `json:"service_id"     validate:"required"`
	ServiceSecret string `json:"service_secret" validate:"required"`
	Audience      string `json:"audience"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Roles        []string `json:"roles"`
}

type AdminLoginResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	Roles        []string `json:"roles"`
}

type LinkWithCodeRequest struct {
	Code string `json:"code" validate:"required,len=6"`
}

type WebRegisterRequest struct {
	Login            string `json:"login"             validate:"required,min=3,max=32"`
	Password         string `json:"password"          validate:"required,min=6"`
	OrganizationName string `json:"organization_name" validate:"omitempty,max=128"`
	INN              string `json:"inn"               validate:"omitempty,min=10,max=12,numeric"`
	FullName         string `json:"full_name"         validate:"omitempty,max=128"`
	PhoneNumber      string `json:"phone_number"      validate:"omitempty,min=10,max=32"`
}

type ForgotPasswordRequest struct {
	Login string `json:"login" validate:"required,min=3,max=32"`
}

type ResetPasswordRequest struct {
	Login       string `json:"login"        validate:"required,min=3,max=32"`
	Otp         string `json:"otp"          validate:"required,len=6"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type ResetPasswordRecoveryRequest struct {
	Login        string `json:"login"         validate:"required,min=3,max=32"`
	RecoveryCode string `json:"recovery_code" validate:"required,min=8,max=12"`
	NewPassword  string `json:"new_password"  validate:"required,min=6"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password"     validate:"required,min=6"`
}

type SetupWebAccessRequest struct {
	Login    string `json:"login"    validate:"required,min=3,max=32"`
	Password string `json:"password" validate:"required,min=6"`
}

type JWKSResponse struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	Kid string   `json:"kid"`
	Kty string   `json:"kty"`
	Use string   `json:"use,omitempty"`
	Alg string   `json:"alg,omitempty"`
	N   string   `json:"n,omitempty"`
	E   string   `json:"e,omitempty"`
	Crv string   `json:"crv,omitempty"`
	X   string   `json:"x,omitempty"`
	Y   string   `json:"y,omitempty"`
	X5c []string `json:"x5c,omitempty"`
}
