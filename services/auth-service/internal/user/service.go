package user

import (
	"context"
	"strings"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/role"
	"csort.ru/auth-service/internal/webauth"
	"csort.ru/auth-service/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type Service struct {
	db          *database.Queries
	roleService *role.Service
	logger      zerolog.Logger
}

// NewService returns a user service backed by db and roleService.
func NewService(db *database.Queries, roleService *role.Service) *Service {
	return &Service{
		db:          db,
		roleService: roleService,
		logger:      logger.GetLogger("user.service"),
	}
}

func (s *Service) CreateWeb(
	ctx context.Context,
	req *WebRegisterRequest,
	passwordHash string,
) (*User, error) {
	login := webauth.NormalizeLogin(req.Login)
	if !webauth.ValidateLoginFormat(login) {
		return nil, apperrors.New(
			fiber.StatusBadRequest,
			"login must be 3-32 characters: lowercase letters, digits, underscore, hyphen",
		)
	}

	if _, err := s.db.GetUserByLogin(ctx, login); err == nil {
		return nil, apperrors.New(fiber.StatusConflict, "login already taken")
	}

	dbUser, err := s.db.CreateWebUser(ctx, database.CreateWebUserParams{
		Login:            pgtype.Text{String: login, Valid: true},
		PasswordHash:     pgtype.Text{String: passwordHash, Valid: true},
		OrganizationName: toPgText(req.OrganizationName),
		Inn:              toPgText(req.INN),
		FullName:         toPgText(req.FullName),
		PhoneNumber:      toPgText(req.PhoneNumber),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create web user")
		if msg := utils.UniqueViolationMessage(err, map[string]string{
			"users_login_key":        "Логин уже занят",
			"users_phone_number_key": "Номер телефона уже зарегистрирован",
		}, "Данные уже используются. Проверьте введённые значения."); msg != "" {
			return nil, apperrors.New(fiber.StatusConflict, msg)
		}
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to create user")
	}

	s.logger.Info().Int32("user_id", dbUser.ID).Str("login", login).Msg("Web user created")
	return &User{User: dbUser, Roles: []string{}}, nil
}

func (s *Service) GetByLogin(ctx context.Context, login string) (*User, error) {
	login = webauth.NormalizeLogin(login)
	if login == "" {
		return nil, apperrors.New(fiber.StatusBadRequest, "login is required")
	}

	dbUser, err := s.db.GetUserByLogin(ctx, login)
	if err != nil {
		s.logger.Error().Err(err).Str("login", login).Msg("User not found by login")
		return nil, apperrors.New(fiber.StatusNotFound, "user not found")
	}

	userRoles, err := s.roleService.GetRolesForUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	u := &User{
		User:  dbUser,
		Roles: make([]string, 0, len(userRoles)),
	}
	for _, r := range userRoles {
		u.Roles = append(u.Roles, r.Name)
	}
	return u, nil
}

// Create registers a new user for the given Telegram ID.
func (s *Service) Create(
	ctx context.Context,
	telegramID int64,
	req *RegisterRequest,
) (*User, error) {
	existingUser, err := s.db.GetUserByTelegramId(ctx, pgtype.Int8{Int64: telegramID, Valid: true})
	if err == nil && existingUser.ID != 0 {
		s.logger.Warn().
			Int64("telegram_id", telegramID).
			Msg("User with this Telegram ID already exists")
		return nil, apperrors.Newf(
			fiber.StatusConflict,
			"user with Telegram ID %d already exists",
			telegramID,
		)
	}

	dbUser, err := s.db.CreateUser(ctx, database.CreateUserParams{
		TelegramID:       pgtype.Int8{Int64: telegramID, Valid: true},
		OrganizationName: toPgText(req.OrganizationName),
		Inn:              toPgText(req.INN),
		FullName:         toPgText(req.FullName),
		PhoneNumber:      toPgText(req.PhoneNumber),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user in database")
		if msg := utils.UniqueViolationMessage(err, map[string]string{
			"users_phone_number_key": "Номер телефона уже зарегистрирован",
			"users_telegram_id_key":  "Аккаунт с этим Telegram уже существует",
			"users_max_id_key":       "Аккаунт с этим MAX уже существует",
		}, "Данные уже используются. Проверьте введённые значения."); msg != "" {
			return nil, apperrors.New(fiber.StatusConflict, msg)
		}
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to create user")
	}

	s.logger.Info().Int32("user_id", dbUser.ID).Msg("User created successfully")

	return &User{User: dbUser, Roles: []string{}}, nil
}

// CreateWithMaxID creates a new user identified by MAX user ID.
func (s *Service) CreateWithMaxID(
	ctx context.Context,
	maxID int64,
	req *RegisterRequest,
) (*User, error) {
	existingUser, err := s.db.GetUserByMaxId(ctx, pgtype.Int8{Int64: maxID, Valid: true})
	if err == nil && existingUser.ID != 0 {
		s.logger.Warn().Int64("max_id", maxID).Msg("User with this MAX ID already exists")
		return nil, apperrors.Newf(
			fiber.StatusConflict,
			"user with MAX ID %d already exists",
			maxID,
		)
	}

	dbUser, err := s.db.CreateUserWithMaxId(ctx, database.CreateUserWithMaxIdParams{
		MaxID:            pgtype.Int8{Int64: maxID, Valid: true},
		OrganizationName: toPgText(req.OrganizationName),
		Inn:              toPgText(req.INN),
		FullName:         toPgText(req.FullName),
		PhoneNumber:      toPgText(req.PhoneNumber),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create MAX user in database")
		if msg := utils.UniqueViolationMessage(err, map[string]string{
			"users_phone_number_key": "Номер телефона уже зарегистрирован",
			"users_telegram_id_key":  "Аккаунт с этим Telegram уже существует",
			"users_max_id_key":       "Аккаунт с этим MAX уже существует",
		}, "Данные уже используются. Проверьте введённые значения."); msg != "" {
			return nil, apperrors.New(fiber.StatusConflict, msg)
		}
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to create MAX user",
		)
	}

	s.logger.Info().Int32("user_id", dbUser.ID).Msg("MAX user created successfully")

	return &User{User: dbUser, Roles: []string{}}, nil
}

// Get returns a user by internal ID, including roles.
func (s *Service) Get(ctx context.Context, id int32) (*User, error) {
	if id <= 0 {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid user ID: must be positive")
	}

	dbUser, err := s.db.GetUser(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", id).Msg("User not found")
		return nil, apperrors.Newf(fiber.StatusNotFound, "user with ID %d not found", id)
	}

	roles, err := s.roleService.GetRolesForUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	user := &User{
		User:  dbUser,
		Roles: make([]string, 0, len(roles)),
	}

	for _, role := range roles {
		user.Roles = append(user.Roles, role.Name)
	}

	return user, nil
}

// GetByTelegramId returns a user by Telegram ID, including roles.
func (s *Service) GetByTelegramId(ctx context.Context, telegramId int64) (*User, error) {
	if telegramId <= 0 {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid Telegram ID: must be positive")
	}

	dbUser, err := s.db.GetUserByTelegramId(ctx, pgtype.Int8{Int64: telegramId, Valid: true})
	if err != nil {
		s.logger.Error().
			Err(err).
			Int64("telegram_id", telegramId).
			Msg("User not found by Telegram ID")
		return nil, apperrors.Newf(
			fiber.StatusNotFound,
			"user with Telegram ID %d not found",
			telegramId,
		)
	}

	userRoles, err := s.roleService.GetRolesForUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	user := &User{
		User:  dbUser,
		Roles: make([]string, 0, len(userRoles)),
	}

	for _, role := range userRoles {
		user.Roles = append(user.Roles, role.Name)
	}
	return user, nil
}

// GetByMaxId returns a user by MAX messenger ID, including roles.
func (s *Service) GetByMaxId(ctx context.Context, maxId int64) (*User, error) {
	if maxId <= 0 {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid MAX ID: must be positive")
	}

	dbUser, err := s.db.GetUserByMaxId(ctx, pgtype.Int8{Int64: maxId, Valid: true})
	if err != nil {
		s.logger.Error().Err(err).Int64("max_id", maxId).Msg("User not found by MAX ID")
		return nil, apperrors.Newf(fiber.StatusNotFound, "user with MAX ID %d not found", maxId)
	}

	userRoles, err := s.roleService.GetRolesForUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	user := &User{
		User:  dbUser,
		Roles: make([]string, 0, len(userRoles)),
	}

	for _, role := range userRoles {
		user.Roles = append(user.Roles, role.Name)
	}
	return user, nil
}

// Update applies a partial profile update; empty request fields keep existing values.
func (s *Service) Update(ctx context.Context, id int32, req *UserRequest) (*User, error) {
	if id <= 0 {
		return nil, apperrors.New(fiber.StatusBadRequest, "invalid user ID: must be positive")
	}

	existing, err := s.db.GetUser(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", id).Msg("User not found for update")
		return nil, apperrors.Newf(fiber.StatusNotFound, "user with ID %d not found", id)
	}

	orgName := req.OrganizationName
	if strings.TrimSpace(orgName) == "" {
		orgName = fromPgText(existing.OrganizationName)
	}
	inn := req.INN
	if strings.TrimSpace(inn) == "" {
		inn = fromPgText(existing.Inn)
	}
	fullName := req.FullName
	if strings.TrimSpace(fullName) == "" {
		fullName = fromPgText(existing.FullName)
	}
	phoneNumber := req.PhoneNumber
	if strings.TrimSpace(phoneNumber) == "" {
		phoneNumber = fromPgText(existing.PhoneNumber)
	}

	dbUser, err := s.db.UpdateUser(ctx, database.UpdateUserParams{
		ID:               id,
		OrganizationName: toPgText(orgName),
		Inn:              toPgText(inn),
		FullName:         toPgText(fullName),
		PhoneNumber:      toPgText(phoneNumber),
	})
	if err != nil {
		s.logger.Error().Err(err).Int32("user_id", id).Msg("Failed to update user")
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to update user")
	}

	s.logger.Info().Int32("user_id", dbUser.ID).Msg("User updated successfully")

	roles, err := s.roleService.GetRolesForUser(ctx, dbUser.ID)
	if err != nil {
		return nil, err
	}

	user := &User{
		User:  dbUser,
		Roles: make([]string, 0, len(roles)),
	}

	for _, role := range roles {
		user.Roles = append(user.Roles, role.Name)
	}

	return user, nil
}

// List returns all users (without per-user roles).
func (s *Service) List(ctx context.Context) ([]User, error) {
	dbUsers, err := s.db.ListUsers(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list users")
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to list users")
	}

	users := make([]User, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		users = append(users, User{
			User:  dbUser,
			Roles: nil,
		})
	}

	return users, nil
}
