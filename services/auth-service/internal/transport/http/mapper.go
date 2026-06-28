package http

import (
	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/role"
	"csort.ru/auth-service/internal/service"
	"csort.ru/auth-service/internal/user"
)

func AdminLoginResponseFromDomain(pair *auth.KeyPair, roles []string) dto.AdminLoginResponse {
	return dto.AdminLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		Roles:        roles,
	}
}

func LoginResponseFromDomain(pair *auth.KeyPair, roles []string) dto.LoginResponse {
	return dto.LoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		Roles:        roles,
	}
}

func RegisterRequestToDomain(req dto.RegisterRequest) user.RegisterRequest {
	return user.RegisterRequest{
		OrganizationName: req.OrganizationName,
		INN:              req.INN,
		FullName:         req.FullName,
		PhoneNumber:      req.PhoneNumber,
	}
}

func UserRequestToDomain(req dto.UserRequest) user.UserRequest {
	return user.UserRequest{
		OrganizationName: req.OrganizationName,
		INN:              req.INN,
		FullName:         req.FullName,
		PhoneNumber:      req.PhoneNumber,
	}
}

func BotCreateRequestToDomain(req dto.CreateBotRequest) bot.CreateRequest {
	return bot.CreateRequest{
		Name:     req.Name,
		Token:    req.Token,
		Platform: req.Platform,
	}
}

func BotUpdateRequestToDomain(req dto.UpdateBotRequest) bot.UpdateRequest {
	return bot.UpdateRequest{
		Name:     req.Name,
		Token:    req.Token,
		Platform: req.Platform,
	}
}

func BotResponseFromDomain(item *bot.Response) dto.BotResponse {
	return dto.BotResponse{
		ID:        item.ID,
		Name:      item.Name,
		Platform:  item.Platform,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func BotResponsesFromDomain(items []bot.Response) []dto.BotResponse {
	out := make([]dto.BotResponse, len(items))
	for i := range items {
		out[i] = BotResponseFromDomain(&items[i])
	}
	return out
}

func RoleRequestToDomain(req dto.RoleRequest) role.RoleRequest {
	return role.RoleRequest{Name: req.Name}
}

func AssignRevokeRoleRequestToDomain(req dto.AssignRevokeRoleRequest) role.AssignRevokeRoleRequest {
	return role.AssignRevokeRoleRequest{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}
}

func CreateServiceRequestToDomain(req dto.CreateServiceRequest) service.CreateRequest {
	return service.CreateRequest{
		ServiceID:     req.ServiceID,
		ServiceSecret: req.ServiceSecret,
	}
}
