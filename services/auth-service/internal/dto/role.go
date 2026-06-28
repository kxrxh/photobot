package dto

type RoleRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

type AssignRevokeRoleRequest struct {
	UserID int32 `json:"user_id" validate:"required,gt=0"`
	RoleID int32 `json:"role_id" validate:"required,gt=0"`
}
