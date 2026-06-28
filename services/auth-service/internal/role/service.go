package role

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"github.com/rs/zerolog"
)

type Service struct {
	db     *database.Queries
	logger zerolog.Logger
}

// NewService returns a role service backed by db.
func NewService(db *database.Queries) *Service {
	return &Service{
		db:     db,
		logger: logger.GetLogger("role.service"),
	}
}

// Create adds a role after validating the name and checking for duplicates.
func (s *Service) Create(ctx context.Context, name string) (*database.Role, error) {
	roleName := strings.TrimSpace(name)
	if err := ValidateRoleName(roleName); err != nil {
		return nil, err
	}

	existingRole, err := s.db.GetRoleByName(ctx, roleName)
	if err == nil && existingRole.ID != 0 {
		s.logger.Warn().Str("role_name", roleName).Msg("Role with this name already exists")
		return nil, fmt.Errorf("role with name '%s' already exists", roleName)
	}

	role, err := s.db.CreateRole(ctx, roleName)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("role_name", roleName).
			Msg("Failed to create role in database")
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	s.logger.Info().
		Int32("role_id", role.ID).
		Str("role_name", roleName).
		Msg("Role created successfully")

	return &role, nil
}

// Get returns a role by ID.
func (s *Service) Get(ctx context.Context, id int32) (*database.Role, error) {
	if id <= 0 {
		return nil, errors.New("invalid role ID: must be positive")
	}

	role, err := s.db.GetRole(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Role not found")
		return nil, err
	}

	return &role, nil
}

// GetByName returns a role by name.
func (s *Service) GetByName(ctx context.Context, name string) (*database.Role, error) {
	if name == "" {
		return nil, errors.New("role name cannot be empty")
	}

	role, err := s.db.GetRoleByName(ctx, name)
	if err != nil {
		s.logger.Error().Err(err).Str("role_name", name).Msg("Role not found by name")
		return nil, err
	}

	return &role, nil
}

// List returns all roles.
func (s *Service) List(ctx context.Context) ([]database.Role, error) {
	roles, err := s.db.ListRoles(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list roles")
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}

	return roles, nil
}

// Update renames a role after validation and conflict checks.
func (s *Service) Update(ctx context.Context, id int32, name string) (*database.Role, error) {
	if id <= 0 {
		return nil, errors.New("invalid role ID: must be positive")
	}

	name = strings.TrimSpace(name)
	if err := ValidateRoleName(name); err != nil {
		return nil, err
	}

	_, err := s.db.GetRole(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Role not found for update")
		return nil, fmt.Errorf("role with ID %d not found", id)
	}

	existingRole, err := s.db.GetRoleByName(ctx, name)
	if err == nil && existingRole.ID != id {
		s.logger.Warn().
			Str("role_name", name).
			Int32("existing_role_id", existingRole.ID).
			Msg("Role with this name already exists")
		return nil, fmt.Errorf("role with name '%s' already exists", name)
	}

	role, err := s.db.UpdateRole(ctx, database.UpdateRoleParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Failed to update role")
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	return &role, nil
}

// Delete removes a role if it exists and no users are assigned.
func (s *Service) Delete(ctx context.Context, id int32) error {
	if id <= 0 {
		return errors.New("invalid role ID: must be positive")
	}

	role, err := s.db.GetRole(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Role not found for deletion")
		return fmt.Errorf("role with ID %d not found", id)
	}

	userCount, err := s.db.CountUsersByRole(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Failed to count users for role")
		return fmt.Errorf("failed to check role assignments: %w", err)
	}

	if userCount > 0 {
		s.logger.Warn().
			Int32("role_id", id).
			Int64("user_count", userCount).
			Msg("Cannot delete role: users are still assigned")
		return fmt.Errorf("cannot delete role: %d users are still assigned to it", userCount)
	}

	err = s.db.DeleteRole(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", id).Msg("Failed to delete role")
		return fmt.Errorf("failed to delete role: %w", err)
	}

	s.logger.Info().
		Int32("role_id", id).
		Str("role_name", role.Name).
		Msg("Role deleted successfully")
	return nil
}

// GetRolesForUser returns all roles assigned to the user.
func (s *Service) GetRolesForUser(ctx context.Context, userID int32) ([]database.Role, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user ID: must be positive")
	}

	_, err := s.db.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", userID).Msg("User not found when getting roles")
		return nil, fmt.Errorf("user with ID %d not found", userID)
	}

	roles, err := s.db.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", userID).Msg("Failed to get user roles")
		return nil, fmt.Errorf("failed to get roles for user %d: %w", userID, err)
	}

	return roles, nil
}

// AssignRole assigns a role to a user if both exist and the assignment is new.
func (s *Service) AssignRole(ctx context.Context, userID int32, roleID int32) error {
	_, err := s.db.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", userID).Msg("User not found for role assignment")
		return fmt.Errorf("user with ID %d not found", userID)
	}

	_, err = s.db.GetRole(ctx, roleID)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", roleID).Msg("Role not found for assignment")
		return fmt.Errorf("role with ID %d not found", roleID)
	}

	userRoles, err := s.db.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("user_id", userID).
			Msg("Failed to retrieve user roles for duplicate check")
		return fmt.Errorf("failed to verify existing roles: %w", err)
	}

	for _, role := range userRoles {
		if role.ID == roleID {
			s.logger.Warn().
				Int32("user_id", userID).
				Int32("role_id", roleID).
				Msg("User already has this role")
			return fmt.Errorf("user already has role with ID %d", roleID)
		}
	}

	err = s.db.AddUserRole(ctx, database.AddUserRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("user_id", userID).
			Int32("role_id", roleID).
			Msg("Failed to add user role")
		return fmt.Errorf("failed to add role to user: %w", err)
	}

	s.logger.Info().
		Int32("user_id", userID).
		Int32("role_id", roleID).
		Msg("User role added successfully")
	return nil
}

// RevokeRole removes a role from a user if the assignment exists.
func (s *Service) RevokeRole(ctx context.Context, userID int32, roleID int32) error {
	if userID <= 0 {
		return errors.New("invalid user ID: must be positive")
	}

	if roleID <= 0 {
		return errors.New("invalid role ID: must be positive")
	}

	_, err := s.db.GetUser(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", userID).Msg("User not found for role removal")
		return fmt.Errorf("user with ID %d not found", userID)
	}

	_, err = s.db.GetRole(ctx, roleID)
	if err != nil {
		s.logger.Error().Err(err).Int32("role_id", roleID).Msg("Role not found for removal")
		return fmt.Errorf("role with ID %d not found", roleID)
	}

	userRoles, err := s.db.GetUserRoles(ctx, userID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("user_id", userID).
			Msg("Failed to get user roles for removal check")
		return fmt.Errorf("failed to verify user roles: %w", err)
	}

	hasRole := false
	for _, role := range userRoles {
		if role.ID == roleID {
			hasRole = true
			break
		}
	}

	if !hasRole {
		s.logger.Warn().
			Int32("user_id", userID).
			Int32("role_id", roleID).
			Msg("User doesn't have this role")
		return fmt.Errorf("user doesn't have role with ID %d", roleID)
	}

	err = s.db.RemoveUserRole(ctx, database.RemoveUserRoleParams{
		UserID: userID,
		RoleID: roleID,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("user_id", userID).
			Int32("role_id", roleID).
			Msg("Failed to remove user role")
		return fmt.Errorf("failed to remove role from user: %w", err)
	}

	s.logger.Info().
		Int32("user_id", userID).
		Int32("role_id", roleID).
		Msg("User role removed successfully")
	return nil
}
