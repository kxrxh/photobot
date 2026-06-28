package auth

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/user"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyOwnership struct {
	transfers [][2]int32
}

func (s *spyOwnership) TransferOwnership(_ context.Context, from, to int32) error {
	s.transfers = append(s.transfers, [2]int32{from, to})
	return nil
}

type mergeLinkQuerier struct {
	panicQuerier
	tEarly, tLate     time.Time
	spy               *spyOwnership
	deletedUserID     *int32
	lastTelegramLink  *database.LinkUserTelegramIdParams
	addRoleAssertions []database.AddUserRoleParams
}

func (m *mergeLinkQuerier) GetUser(_ context.Context, id int32) (database.User, error) {
	switch id {
	case 10:
		return database.User{
			ID:         10,
			TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
			CreatedAt:  m.tEarly,
		}, nil
	case 20:
		return database.User{
			ID:         20,
			TelegramID: pgtype.Int8{Int64: 222222, Valid: true},
			CreatedAt:  m.tLate,
		}, nil
	default:
		return database.User{}, errors.New("not found")
	}
}

func (m *mergeLinkQuerier) CopyWebCredentials(
	context.Context,
	database.CopyWebCredentialsParams,
) error {
	return nil
}

func (m *mergeLinkQuerier) TransferWebCredentials(
	context.Context,
	database.TransferWebCredentialsParams,
) error {
	return nil
}

func (m *mergeLinkQuerier) ClearWebCredentials(context.Context, int32) error {
	return nil
}

func (m *mergeLinkQuerier) ReassignRecoveryCodes(
	context.Context,
	database.ReassignRecoveryCodesParams,
) error {
	return nil
}

func (m *mergeLinkQuerier) GetUserRoles(_ context.Context, userID int32) ([]database.Role, error) {
	if userID != 20 {
		panic("GetUserRoles: expected user 20 (user being removed)")
	}
	return []database.Role{{ID: 7, Name: "extra"}}, nil
}

func (m *mergeLinkQuerier) AddUserRole(_ context.Context, arg database.AddUserRoleParams) error {
	m.addRoleAssertions = append(m.addRoleAssertions, arg)
	return nil
}

func (m *mergeLinkQuerier) DeleteUser(_ context.Context, id int32) error {
	if id != 20 {
		panic("DeleteUser: expected id 20")
	}
	i := id
	m.deletedUserID = &i
	return nil
}

func (m *mergeLinkQuerier) LinkUserTelegramId(
	_ context.Context,
	arg database.LinkUserTelegramIdParams,
) error {
	cp := arg
	m.lastTelegramLink = &cp
	return nil
}

var _ database.Querier = (*mergeLinkQuerier)(nil)

func TestService_LinkWithCode_MergeTelegram_KeepsOlderAccount(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	tEarly := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tLate := time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC)
	spy := &spyOwnership{}
	db := &mergeLinkQuerier{
		panicQuerier: panicQuerier{},
		tEarly:       tEarly,
		tLate:        tLate,
		spy:          spy,
	}
	const code = "654321"
	userJSON := `{"id":111111,"first_name":"Test"}`
	initData := "user=" + url.QueryEscape(userJSON)

	codeConsumed := false
	svc := NewService(&Params{
		DB: db,
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				require.Equal(t, linkCodeKey(code), key)
				return "20", nil
			},
			delFn: func(_ context.Context, key string) error {
				require.Equal(t, linkCodeKey(code), key)
				codeConsumed = true
				return nil
			},
			setFn: func(context.Context, string, string, time.Duration) error { return nil },
		},
		BotService: stubBotProvider{
			getTokenFn: func(_ context.Context, name, platform string) (string, error) {
				assert.Equal(t, "mybot", name)
				assert.Equal(t, "telegram", platform)
				return "not-used-in-debug", nil
			},
		},
		UserService: stubUserProvider{
			getFn: func(_ context.Context, id int32) (*user.User, error) {
				require.Equal(t, int32(10), id)
				return &user.User{
					User: database.User{
						ID:         10,
						TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
					},
					Roles: []string{"user"},
				}, nil
			},
		},
		RoleService: stubRoleProvider{
			getRolesFn: func(_ context.Context, userID int32) ([]database.Role, error) {
				require.Equal(t, int32(10), userID)
				return []database.Role{{ID: 1, Name: "user"}}, nil
			},
		},
		OwnershipClient: spy,
		Config: &Config{
			Debug:                 true,
			DebugBypassSignatures: true,
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
		},
	})

	res, err := svc.LinkWithCode(ctx, 10, code, initData, "mybot", "telegram")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.Message, "linked")
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)

	require.Len(t, spy.transfers, 1)
	assert.Equal(t, int32(20), spy.transfers[0][0])
	assert.Equal(t, int32(10), spy.transfers[0][1])

	require.NotNil(t, db.deletedUserID)
	assert.Equal(t, int32(20), *db.deletedUserID)

	require.NotNil(t, db.lastTelegramLink)
	assert.Equal(t, int32(10), db.lastTelegramLink.ID)
	require.True(t, db.lastTelegramLink.TelegramID.Valid)
	assert.Equal(t, int64(222222), db.lastTelegramLink.TelegramID.Int64)

	require.Len(t, db.addRoleAssertions, 1)
	assert.Equal(t, int32(10), db.addRoleAssertions[0].UserID)
	assert.Equal(t, int32(7), db.addRoleAssertions[0].RoleID)
	assert.True(t, codeConsumed)
}

type webMergeLinkQuerier struct {
	panicQuerier
	tEarly, tLate      time.Time
	spy                *spyOwnership
	keepHasLogin       bool
	deletedUserID      *int32
	copyWebCredentials *database.CopyWebCredentialsParams
}

func (m *webMergeLinkQuerier) GetUser(_ context.Context, id int32) (database.User, error) {
	switch id {
	case 10:
		u := database.User{
			ID:         10,
			TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
			CreatedAt:  m.tEarly,
		}
		if m.keepHasLogin {
			u.Login = pgtype.Text{String: "existing", Valid: true}
			u.PasswordHash = pgtype.Text{String: "hash", Valid: true}
		}
		return u, nil
	case 30:
		return database.User{
			ID:        30,
			Login:     pgtype.Text{String: "webuser", Valid: true},
			CreatedAt: m.tLate,
		}, nil
	default:
		return database.User{}, errors.New("not found")
	}
}

func (m *webMergeLinkQuerier) CopyWebCredentials(
	_ context.Context,
	arg database.CopyWebCredentialsParams,
) error {
	cp := arg
	m.copyWebCredentials = &cp
	return nil
}

func (m *webMergeLinkQuerier) TransferWebCredentials(
	context.Context,
	database.TransferWebCredentialsParams,
) error {
	return nil
}

func (m *webMergeLinkQuerier) ClearWebCredentials(context.Context, int32) error {
	return nil
}

func (m *webMergeLinkQuerier) ReassignRecoveryCodes(
	context.Context,
	database.ReassignRecoveryCodesParams,
) error {
	return nil
}

func (m *webMergeLinkQuerier) GetUserRoles(context.Context, int32) ([]database.Role, error) {
	return nil, nil
}

func (m *webMergeLinkQuerier) DeleteUser(_ context.Context, id int32) error {
	m.deletedUserID = &id
	return nil
}

var _ database.Querier = (*webMergeLinkQuerier)(nil)

func TestService_LinkWithCode_MergeWeb_KeepsOlderTelegramAccount(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	tEarly := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tLate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	spy := &spyOwnership{}
	db := &webMergeLinkQuerier{
		panicQuerier: panicQuerier{},
		tEarly:       tEarly,
		tLate:        tLate,
		spy:          spy,
	}
	const code = "654321"
	userJSON := `{"id":111111,"first_name":"Test"}`
	initData := "user=" + url.QueryEscape(userJSON)

	svc := NewService(&Params{
		DB: db,
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				require.Equal(t, linkCodeKey(code), key)
				return "30", nil
			},
			delFn: func(context.Context, string) error { return nil },
		},
		BotService: stubBotProvider{
			getTokenFn: func(context.Context, string, string) (string, error) {
				return "not-used-in-debug", nil
			},
		},
		UserService: stubUserProvider{
			getFn: func(_ context.Context, id int32) (*user.User, error) {
				require.Equal(t, int32(10), id)
				return &user.User{
					User: database.User{
						ID:         10,
						TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
						Login:      pgtype.Text{String: "webuser", Valid: true},
					},
					Roles: []string{"user"},
				}, nil
			},
		},
		RoleService:     stubRoleProvider{},
		OwnershipClient: spy,
		Config: &Config{
			Debug:                 true,
			DebugBypassSignatures: true,
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
		},
	})

	res, err := svc.LinkWithCode(ctx, 10, code, initData, "mybot", "telegram")
	require.NoError(t, err)
	require.NotNil(t, res)

	require.NotNil(t, db.deletedUserID)
	assert.Equal(t, int32(30), *db.deletedUserID)
	require.NotNil(t, db.copyWebCredentials)
	assert.Equal(t, int32(10), db.copyWebCredentials.ID)
	require.True(t, db.copyWebCredentials.Login.Valid)
	assert.Equal(t, "webuser", db.copyWebCredentials.Login.String)
	require.Len(t, spy.transfers, 1)
	assert.Equal(t, int32(30), spy.transfers[0][0])
	assert.Equal(t, int32(10), spy.transfers[0][1])
}

func TestService_LinkWithCode_MergeWeb_SkipsTransferWhenKeepAlreadyHasLogin(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	tEarly := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tLate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	spy := &spyOwnership{}
	db := &webMergeLinkQuerier{
		panicQuerier: panicQuerier{},
		tEarly:       tEarly,
		tLate:        tLate,
		spy:          spy,
		keepHasLogin: true,
	}
	const code = "654321"
	userJSON := `{"id":111111,"first_name":"Test"}`
	initData := "user=" + url.QueryEscape(userJSON)

	svc := NewService(&Params{
		DB: db,
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				require.Equal(t, linkCodeKey(code), key)
				return "30", nil
			},
			delFn: func(context.Context, string) error { return nil },
		},
		BotService: stubBotProvider{
			getTokenFn: func(context.Context, string, string) (string, error) {
				return "not-used-in-debug", nil
			},
		},
		UserService: stubUserProvider{
			getFn: func(_ context.Context, id int32) (*user.User, error) {
				require.Equal(t, int32(10), id)
				return &user.User{
					User: database.User{
						ID:           10,
						TelegramID:   pgtype.Int8{Int64: 111111, Valid: true},
						Login:        pgtype.Text{String: "existing", Valid: true},
						PasswordHash: pgtype.Text{String: "hash", Valid: true},
					},
					Roles: []string{"user"},
				}, nil
			},
		},
		RoleService:     stubRoleProvider{},
		OwnershipClient: spy,
		Config: &Config{
			Debug:                 true,
			DebugBypassSignatures: true,
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
		},
	})

	res, err := svc.LinkWithCode(ctx, 10, code, initData, "mybot", "telegram")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Nil(t, db.copyWebCredentials)
	require.NotNil(t, db.deletedUserID)
	assert.Equal(t, int32(30), *db.deletedUserID)
}

func TestService_LinkWithCode_RecoversWhenCodeUserAlreadyMerged(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	tEarly := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	spy := &spyOwnership{}
	db := &deletedCodeUserQuerier{tEarly: tEarly}
	const code = "654321"
	userJSON := `{"id":111111,"first_name":"Test"}`
	initData := "user=" + url.QueryEscape(userJSON)
	codeConsumed := false

	svc := NewService(&Params{
		DB: db,
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				require.Equal(t, linkCodeKey(code), key)
				return "30", nil
			},
			delFn: func(_ context.Context, key string) error {
				require.Equal(t, linkCodeKey(code), key)
				codeConsumed = true
				return nil
			},
		},
		BotService: stubBotProvider{
			getTokenFn: func(context.Context, string, string) (string, error) {
				return "not-used-in-debug", nil
			},
		},
		UserService: stubUserProvider{
			getFn: func(_ context.Context, id int32) (*user.User, error) {
				require.Equal(t, int32(10), id)
				return &user.User{
					User: database.User{
						ID:         10,
						TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
						Login:      pgtype.Text{String: "webuser", Valid: true},
					},
					Roles: []string{"user"},
				}, nil
			},
		},
		RoleService:     stubRoleProvider{},
		OwnershipClient: spy,
		Config: &Config{
			Debug:                 true,
			DebugBypassSignatures: true,
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
		},
	})

	res, err := svc.LinkWithCode(ctx, 10, code, initData, "mybot", "telegram")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, res.Message, "linked")
	assert.NotEmpty(t, res.AccessToken)
	assert.NotEmpty(t, res.RefreshToken)
	require.Len(t, spy.transfers, 1)
	assert.Equal(t, int32(30), spy.transfers[0][0])
	assert.Equal(t, int32(10), spy.transfers[0][1])
	assert.True(t, codeConsumed)
}

type deletedCodeUserQuerier struct {
	panicQuerier
	tEarly time.Time
}

func (m *deletedCodeUserQuerier) GetUser(_ context.Context, id int32) (database.User, error) {
	if id == 10 {
		return database.User{
			ID:         10,
			TelegramID: pgtype.Int8{Int64: 111111, Valid: true},
			CreatedAt:  m.tEarly,
		}, nil
	}
	return database.User{}, errors.New("not found")
}

var _ database.Querier = (*deletedCodeUserQuerier)(nil)

func TestService_LinkWithCode_ForbiddenWhenInitDataDoesNotMatchJWTUser(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	tEarly := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	tLate := time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC)
	db := &mergeLinkQuerier{
		panicQuerier: panicQuerier{},
		tEarly:       tEarly,
		tLate:        tLate,
		spy:          &spyOwnership{},
	}
	initData := "user=" + url.QueryEscape(`{"id":999999,"first_name":"X"}`)

	svc := NewService(&Params{
		DB:         db,
		TokenStore: stubTokenStore{},
		BotService: stubBotProvider{
			getTokenFn: func(context.Context, string, string) (string, error) { return "x", nil },
		},
		Config: &Config{
			Debug:                 true,
			DebugBypassSignatures: true,
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
		},
		OwnershipClient: stubOwnershipClient{},
	})

	_, err := svc.LinkWithCode(ctx, 10, "654321", initData, "mybot", "telegram")
	require.Error(t, err)
	ae, ok := apperrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, 403, ae.Code)
}
