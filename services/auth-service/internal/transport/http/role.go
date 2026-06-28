package http

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"csort.ru/auth-service/internal/auth"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/role"
	"csort.ru/auth-service/internal/transport/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

const httpRoleComponent = "transport.http.role"

type RoleHandler struct {
	roleService *role.Service
	validator   *validator.Validate
}

func NewRoleHandler(roleService *role.Service, v *validator.Validate) *RoleHandler {
	return &RoleHandler{
		roleService: roleService,
		validator:   v,
	}
}

// CreateRole creates a new role.
func (h *RoleHandler) CreateRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	var req dto.RoleRequest

	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("parse request body failed")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed")
	}
	domainReq := RoleRequestToDomain(req)
	roleResp, err := h.roleService.Create(c.Context(), domainReq.Name)
	if err != nil {
		log.Error().Err(err).Msg("create role failed")
		if err.Error() == fmt.Sprintf("role with name '%s' already exists", req.Name) {
			return fiber.NewError(fiber.StatusConflict, "Role with this name already exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create role")
	}

	return response.Created(c, roleResp)
}

// GetRole retrieves a role by ID.
func (h *RoleHandler) GetRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	id64, err := strconv.ParseInt(c.Req().Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid role ID")
	}
	id := int32(id64)

	roleResp, err := h.roleService.Get(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Int32("role_id", id).Msg("get role failed")
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Role not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get role")
	}

	return response.OK(c, roleResp)
}

// GetRoleByName retrieves a role by name.
func (h *RoleHandler) GetRoleByName(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	name := c.Req().Params("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Role name cannot be empty")
	}

	roleResp, err := h.roleService.GetByName(c.Context(), name)
	if err != nil {
		log.Error().Err(err).Str("role_name", name).Msg("get role by name failed")
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Role not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get role by name")
	}

	return response.OK(c, roleResp)
}

// ListRoles retrieves all roles.
func (h *RoleHandler) ListRoles(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	roleList, err := h.roleService.List(c.Context())
	if err != nil {
		log.Error().Err(err).Msg("list roles failed")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to list roles")
	}

	return response.OK(c, roleList)
}

// UpdateRole updates an existing role.
func (h *RoleHandler) UpdateRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	id64, err := strconv.ParseInt(c.Req().Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid role ID")
	}
	id := int32(id64)

	var req dto.RoleRequest

	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("parse request body failed")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed")
	}

	domainReq := RoleRequestToDomain(req)
	roleResp, err := h.roleService.Update(c.Context(), id, domainReq.Name)
	if err != nil {
		log.Error().Err(err).Int32("role_id", id).Msg("update role failed")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update role")
	}

	return response.OK(c, roleResp)
}

// DeleteRole deletes a role by ID.
func (h *RoleHandler) DeleteRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	id64, err := strconv.ParseInt(c.Req().Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid role ID")
	}
	id := int32(id64)

	err = h.roleService.Delete(c.Context(), id)
	if err != nil {
		log.Error().Err(err).Int32("role_id", id).Msg("delete role failed")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete role")
	}

	return response.NoContent(c)
}

// GetUserRoles retrieves all roles for a specific user.
func (h *RoleHandler) GetUserRoles(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	userId64, err := strconv.ParseInt(c.Req().Params("user_id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid user ID")
	}
	targetUserID := int32(userId64)

	roleList, err := h.roleService.GetRolesForUser(c.Context(), targetUserID)
	if err != nil {
		log.Error().Err(err).Int32("user_id", targetUserID).Msg("get user roles failed")
		if err.Error() == fmt.Sprintf("user with ID %d not found", targetUserID) {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get user roles")
	}

	return response.OK(c, roleList)
}

// GetMyRoles retrieves all roles for the current authenticated user.
func (h *RoleHandler) GetMyRoles(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	userID, ok := c.Locals(auth.LocalsUserID).(*int32)
	if !ok || userID == nil {
		return fiber.NewError(fiber.StatusUnauthorized, "User not authenticated")
	}

	roleList, err := h.roleService.GetRolesForUser(c.Context(), *userID)
	if err != nil {
		log.Error().Err(err).Int32("user_id", *userID).Msg("get user roles failed")
		if err.Error() == fmt.Sprintf("user with ID %d not found", *userID) {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to get user roles")
	}

	return response.OK(c, roleList)
}

// AssignRole assigns a role to a user.
func (h *RoleHandler) AssignRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	var req dto.AssignRevokeRoleRequest

	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("parse request body failed")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed")
	}

	domainReq := AssignRevokeRoleRequestToDomain(req)
	err := h.roleService.AssignRole(c.Context(), domainReq.UserID, domainReq.RoleID)
	if err != nil {
		log.Error().
			Err(err).
			Int32("user_id", domainReq.UserID).
			Int32("role_id", domainReq.RoleID).
			Msg("assign role failed")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to add user role")
	}

	return response.NoContent(c)
}

// RevokeRole removes a role from a user.
func (h *RoleHandler) RevokeRole(c fiber.Ctx) error {
	log := logger.Component(c, httpRoleComponent)
	var req dto.AssignRevokeRoleRequest

	if err := c.Bind().Body(&req); err != nil {
		log.Warn().Err(err).Msg("parse request body failed")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.validator.Struct(req); err != nil {
		log.Warn().Err(err).Msg("validation rejected")
		return fiber.NewError(fiber.StatusBadRequest, "Validation failed")
	}

	domainReq := AssignRevokeRoleRequestToDomain(req)
	err := h.roleService.RevokeRole(c.Context(), domainReq.UserID, domainReq.RoleID)
	if err != nil {
		log.Error().
			Err(err).
			Int32("user_id", domainReq.UserID).
			Int32("role_id", domainReq.RoleID).
			Msg("revoke role failed")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to remove user role")
	}

	return response.NoContent(c)
}
