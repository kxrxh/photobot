package dto

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
