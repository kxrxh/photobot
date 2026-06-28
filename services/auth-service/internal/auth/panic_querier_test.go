package auth

import (
	"context"

	"csort.ru/auth-service/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

type panicQuerier struct{}

var _ database.Querier = panicQuerier{}

func (panicQuerier) AddUserRole(context.Context, database.AddUserRoleParams) error {
	panic("unexpected AddUserRole")
}

func (panicQuerier) CountUsersByRole(context.Context, int32) (int64, error) {
	panic("unexpected CountUsersByRole")
}

func (panicQuerier) CreateBot(context.Context, database.CreateBotParams) (database.Bot, error) {
	panic("unexpected CreateBot")
}

func (panicQuerier) CreateRole(context.Context, string) (database.Role, error) {
	panic("unexpected CreateRole")
}

func (panicQuerier) CreateService(
	context.Context,
	database.CreateServiceParams,
) (database.Service, error) {
	panic("unexpected CreateService")
}

func (panicQuerier) CreateUser(context.Context, database.CreateUserParams) (database.User, error) {
	panic("unexpected CreateUser")
}

func (panicQuerier) CreateUserWithMaxId(
	context.Context,
	database.CreateUserWithMaxIdParams,
) (database.User, error) {
	panic("unexpected CreateUserWithMaxId")
}

func (panicQuerier) DeleteBot(context.Context, int32) error {
	panic("unexpected DeleteBot")
}

func (panicQuerier) DeleteRole(context.Context, int32) error {
	panic("unexpected DeleteRole")
}

func (panicQuerier) DeleteService(context.Context, string) error {
	panic("unexpected DeleteService")
}

func (panicQuerier) DeleteUser(context.Context, int32) error {
	panic("unexpected DeleteUser")
}

func (panicQuerier) GetBot(context.Context, int32) (database.GetBotRow, error) {
	panic("unexpected GetBot")
}

func (panicQuerier) GetBotByName(context.Context, string) (database.Bot, error) {
	panic("unexpected GetBotByName")
}

func (panicQuerier) GetBotByNameAndPlatform(
	context.Context,
	database.GetBotByNameAndPlatformParams,
) (database.Bot, error) {
	panic("unexpected GetBotByNameAndPlatform")
}

func (panicQuerier) GetRole(context.Context, int32) (database.Role, error) {
	panic("unexpected GetRole")
}

func (panicQuerier) GetRoleByName(context.Context, string) (database.Role, error) {
	panic("unexpected GetRoleByName")
}

func (panicQuerier) GetServiceByServiceID(context.Context, string) (database.Service, error) {
	panic("unexpected GetServiceByServiceID")
}

func (panicQuerier) GetUser(context.Context, int32) (database.User, error) {
	panic("unexpected GetUser")
}

func (panicQuerier) GetUserByMaxId(context.Context, pgtype.Int8) (database.User, error) {
	panic("unexpected GetUserByMaxId")
}

func (panicQuerier) GetUserByTelegramId(context.Context, pgtype.Int8) (database.User, error) {
	panic("unexpected GetUserByTelegramId")
}

func (panicQuerier) GetUserRoles(context.Context, int32) ([]database.Role, error) {
	panic("unexpected GetUserRoles")
}

func (panicQuerier) IsServiceExists(context.Context, string) (bool, error) {
	panic("unexpected IsServiceExists")
}

func (panicQuerier) LinkUserMaxId(context.Context, database.LinkUserMaxIdParams) error {
	panic("unexpected LinkUserMaxId")
}

func (panicQuerier) LinkUserTelegramId(context.Context, database.LinkUserTelegramIdParams) error {
	panic("unexpected LinkUserTelegramId")
}

func (panicQuerier) ListBots(context.Context) ([]database.ListBotsRow, error) {
	panic("unexpected ListBots")
}

func (panicQuerier) ListBotsByPlatform(context.Context, string) ([]database.Bot, error) {
	panic("unexpected ListBotsByPlatform")
}

func (panicQuerier) ListRoles(context.Context) ([]database.Role, error) {
	panic("unexpected ListRoles")
}

func (panicQuerier) ListServices(context.Context) ([]database.ListServicesRow, error) {
	panic("unexpected ListServices")
}

func (panicQuerier) ListUsers(context.Context) ([]database.User, error) {
	panic("unexpected ListUsers")
}

func (panicQuerier) RemoveUserRole(context.Context, database.RemoveUserRoleParams) error {
	panic("unexpected RemoveUserRole")
}

func (panicQuerier) UpdateBotName(
	context.Context,
	database.UpdateBotNameParams,
) (database.Bot, error) {
	panic("unexpected UpdateBotName")
}

func (panicQuerier) UpdateBotToken(
	context.Context,
	database.UpdateBotTokenParams,
) (database.Bot, error) {
	panic("unexpected UpdateBotToken")
}

func (panicQuerier) UpdateRole(context.Context, database.UpdateRoleParams) (database.Role, error) {
	panic("unexpected UpdateRole")
}

func (panicQuerier) UpdateService(
	context.Context,
	database.UpdateServiceParams,
) (database.Service, error) {
	panic("unexpected UpdateService")
}

func (panicQuerier) UpdateUser(context.Context, database.UpdateUserParams) (database.User, error) {
	panic("unexpected UpdateUser")
}

func (panicQuerier) CopyWebCredentials(context.Context, database.CopyWebCredentialsParams) error {
	panic("unexpected CopyWebCredentials")
}

func (panicQuerier) TransferWebCredentials(
	context.Context,
	database.TransferWebCredentialsParams,
) error {
	panic("unexpected TransferWebCredentials")
}

func (panicQuerier) ClearWebCredentials(context.Context, int32) error {
	panic("unexpected ClearWebCredentials")
}

func (panicQuerier) CreateRecoveryCode(
	context.Context,
	database.CreateRecoveryCodeParams,
) (database.UserRecoveryCode, error) {
	panic("unexpected CreateRecoveryCode")
}

func (panicQuerier) CreateWebUser(
	context.Context,
	database.CreateWebUserParams,
) (database.User, error) {
	panic("unexpected CreateWebUser")
}

func (panicQuerier) GetUserByLogin(context.Context, string) (database.User, error) {
	panic("unexpected GetUserByLogin")
}

func (panicQuerier) ListUnusedRecoveryCodesByUser(
	context.Context,
	int32,
) ([]database.UserRecoveryCode, error) {
	panic("unexpected ListUnusedRecoveryCodesByUser")
}

func (panicQuerier) MarkRecoveryCodeUsed(context.Context, int32) error {
	panic("unexpected MarkRecoveryCodeUsed")
}

func (panicQuerier) ReassignRecoveryCodes(
	context.Context,
	database.ReassignRecoveryCodesParams,
) error {
	panic("unexpected ReassignRecoveryCodes")
}

func (panicQuerier) SetPasswordHash(context.Context, database.SetPasswordHashParams) error {
	panic("unexpected SetPasswordHash")
}

func (panicQuerier) SetWebCredentials(
	context.Context,
	database.SetWebCredentialsParams,
) (database.User, error) {
	panic("unexpected SetWebCredentials")
}

func (panicQuerier) UpsertDevUser(
	context.Context,
	database.UpsertDevUserParams,
) (database.User, error) {
	panic("unexpected UpsertDevUser")
}
