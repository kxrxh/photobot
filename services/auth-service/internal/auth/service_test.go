package auth

import (
	"context"
	"net/url"
	"testing"
	"time"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/user"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubServicesValidator struct {
	validateFn func(ctx context.Context, serviceID, serviceSecret string) error
}

func (s stubServicesValidator) ValidateCredentials(
	ctx context.Context,
	serviceID, serviceSecret string,
) error {
	return s.validateFn(ctx, serviceID, serviceSecret)
}

type stubTokenStore struct {
	setFn       func(ctx context.Context, key, value string, ttl time.Duration) error
	getFn       func(ctx context.Context, key string) (string, error)
	delFn       func(ctx context.Context, key string) error
	getAndDelFn func(ctx context.Context, key string) (string, error)
}

func (s stubTokenStore) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if s.setFn == nil {
		return nil
	}
	return s.setFn(ctx, key, value, ttl)
}

func (s stubTokenStore) Get(ctx context.Context, key string) (string, error) {
	if s.getFn == nil {
		return "", nil
	}
	return s.getFn(ctx, key)
}

func (s stubTokenStore) Del(ctx context.Context, key string) error {
	if s.delFn == nil {
		return nil
	}
	return s.delFn(ctx, key)
}

func (s stubTokenStore) GetAndDel(ctx context.Context, key string) (string, error) {
	if s.getAndDelFn == nil {
		return "", nil
	}
	return s.getAndDelFn(ctx, key)
}

type stubBotProvider struct {
	getTokenFn func(ctx context.Context, name, platform string) (string, error)
}

func (s stubBotProvider) GetTokenByNameAndPlatform(
	ctx context.Context,
	name, platform string,
) (string, error) {
	return s.getTokenFn(ctx, name, platform)
}

func (s stubBotProvider) ListTokensByPlatform(
	ctx context.Context,
	platform string,
) ([]string, error) {
	return nil, nil
}

type stubUserProvider struct {
	getFn          func(context.Context, int32) (*user.User, error)
	getByTgFn      func(context.Context, int64) (*user.User, error)
	getByMaxIDFunc func(context.Context, int64) (*user.User, error)
	getByLoginFn   func(context.Context, string) (*user.User, error)
}

func (s stubUserProvider) Get(ctx context.Context, id int32) (*user.User, error) {
	if s.getFn == nil {
		return nil, nil
	}
	return s.getFn(ctx, id)
}

func (s stubUserProvider) GetByTelegramId(ctx context.Context, id int64) (*user.User, error) {
	if s.getByTgFn == nil {
		return nil, nil
	}
	return s.getByTgFn(ctx, id)
}

func (s stubUserProvider) GetByMaxId(ctx context.Context, id int64) (*user.User, error) {
	if s.getByMaxIDFunc == nil {
		return nil, nil
	}
	return s.getByMaxIDFunc(ctx, id)
}

func (s stubUserProvider) GetByLogin(ctx context.Context, login string) (*user.User, error) {
	if s.getByLoginFn == nil {
		return nil, nil
	}
	return s.getByLoginFn(ctx, login)
}

type stubRoleProvider struct {
	getRolesFn func(ctx context.Context, userID int32) ([]database.Role, error)
}

func (s stubRoleProvider) GetRolesForUser(
	ctx context.Context,
	userID int32,
) ([]database.Role, error) {
	if s.getRolesFn == nil {
		return nil, nil
	}
	return s.getRolesFn(ctx, userID)
}

type stubOwnershipClient struct{}

func (stubOwnershipClient) TransferOwnership(context.Context, int32, int32) error {
	return nil
}

func TestNewService(t *testing.T) {
	t.Run("creates service with config", func(t *testing.T) {
		params := &Params{
			Config: &Config{
				AccessExpiryMinutes:  15,
				RefreshExpiryMinutes: 10080,
				AdminLogin:           "admin",
				AdminPassword:        "secret",
				Debug:                true,
			},
			OwnershipClient: stubOwnershipClient{},
		}
		svc := NewService(params)
		require.NotNil(t, svc)
		assert.NotNil(t, svc.config)
		assert.Equal(t, 15, svc.config.AccessExpiryMinutes)
		assert.Equal(t, "admin", svc.config.AdminLogin)
		assert.True(t, svc.config.Debug)
	})

	t.Run("creates service with nil dependencies", func(t *testing.T) {
		params := &Params{Config: &Config{}}
		svc := NewService(params)
		require.NotNil(t, svc)
		assert.Nil(t, svc.db)
		assert.Nil(t, svc.tokenStore)
		assert.Nil(t, svc.botService)
		assert.Nil(t, svc.ownershipClient)
	})
}

func TestService_Login_NilParams(t *testing.T) {
	svc := NewService(&Params{
		Config:          &Config{},
		OwnershipClient: stubOwnershipClient{},
	})
	ctx := context.Background()

	_, _, err := svc.Login(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing login params")
}

func TestService_Login_UnsupportedGrantType(t *testing.T) {
	svc := NewService(&Params{
		Config:          &Config{},
		OwnershipClient: stubOwnershipClient{},
	})
	ctx := context.Background()

	params := &LoginParams{GTY: GrantType("unsupported")}
	_, _, err := svc.Login(ctx, params)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported grant type")
}

func TestService_Login_ServiceGrantType_Success(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	serviceID := "svc1"
	serviceSecret := "secret"

	svc := NewService(&Params{
		Config: &Config{AccessExpiryMinutes: 15, RefreshExpiryMinutes: 60, Debug: true},
		ServicesService: stubServicesValidator{
			validateFn: func(_ context.Context, gotID, gotSecret string) error {
				assert.Equal(t, serviceID, gotID)
				assert.Equal(t, serviceSecret, gotSecret)
				return nil
			},
		},
		TokenStore: stubTokenStore{
			setFn: func(_ context.Context, _ string, value string, _ time.Duration) error {
				assert.Equal(t, "1", value)
				return nil
			},
		},
		OwnershipClient: stubOwnershipClient{},
	})

	params := &LoginParams{
		GTY:           GrantTypeService,
		ServiceID:     &serviceID,
		ServiceSecret: &serviceSecret,
	}
	pair, roles, err := svc.Login(ctx, params)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Equal(t, []string{ServiceRole}, roles)
}

func TestService_Login_InitDataGrantType_TelegramPlatform_Success(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	botToken := "test-bot-token"
	botName := "mybot"
	userJSON := `{"id":12345,"first_name":"John","last_name":"Doe","username":"johndoe"}`
	initData := "user=" + url.QueryEscape(userJSON)

	userID := int32(10)

	svc := NewService(&Params{
		Config: &Config{
			AccessExpiryMinutes:   15,
			RefreshExpiryMinutes:  60,
			Debug:                 true,
			DebugBypassSignatures: true,
		},
		BotService: stubBotProvider{
			getTokenFn: func(_ context.Context, name, platform string) (string, error) {
				assert.Equal(t, botName, name)
				assert.Equal(t, "telegram", platform)
				return botToken, nil
			},
		},
		UserService: stubUserProvider{
			getByTgFn: func(_ context.Context, id int64) (*user.User, error) {
				assert.Equal(t, int64(12345), id)
				return &user.User{
					User:  database.User{ID: userID, CreatedAt: time.Now()},
					Roles: []string{"user"},
				}, nil
			},
		},
		TokenStore: stubTokenStore{
			setFn: func(_ context.Context, _ string, value string, _ time.Duration) error {
				assert.Equal(t, "1", value)
				return nil
			},
		},
		OwnershipClient: stubOwnershipClient{},
	})

	platform := "telegram"
	params := &LoginParams{
		GTY:               GrantTypeInitData,
		InitData:          &initData,
		BotName:           &botName,
		MessengerPlatform: &platform,
	}
	pair, roles, err := svc.Login(ctx, params)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.Contains(t, roles, "user")
}

func TestService_RequestLinkCode_Success(t *testing.T) {
	ctx := context.Background()
	userID := int32(42)

	svc := NewService(&Params{
		Config: &Config{},
		TokenStore: stubTokenStore{
			setFn: func(_ context.Context, _ string, value string, _ time.Duration) error {
				assert.Equal(t, "42", value)
				return nil
			},
		},
		OwnershipClient: stubOwnershipClient{},
	})

	result, err := svc.RequestLinkCode(ctx, userID)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Code)
	assert.Len(t, result.Code, 6)
	assert.Equal(t, 300, result.ExpiresInSeconds)
}

func TestService_Refresh_Success(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	userID := int32(100)
	roles := []string{"user", "editor"}

	genParams := &GenerationParams{
		UserID: &userID,
		Roles:  roles,
		GTY:    GrantTypeInitData,
		JTI:    "test-jti-123",
	}
	refreshToken, err := GenerateJWT(genParams, RefreshToken, 60*time.Minute)
	require.NoError(t, err)

	svc := NewService(&Params{
		Config: &Config{AccessExpiryMinutes: 15, RefreshExpiryMinutes: 60},
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				assert.Equal(t, "refresh_token:test-jti-123", key)
				return "1", nil
			},
			delFn: func(_ context.Context, key string) error {
				assert.Equal(t, "refresh_token:test-jti-123", key)
				return nil
			},
			setFn: func(_ context.Context, _ string, value string, _ time.Duration) error {
				assert.Equal(t, "1", value)
				return nil
			},
		},
		RoleService: stubRoleProvider{
			getRolesFn: func(_ context.Context, gotID int32) ([]database.Role, error) {
				assert.Equal(t, userID, gotID)
				return []database.Role{{ID: 1, Name: "user"}, {ID: 2, Name: "editor"}}, nil
			},
		},
		OwnershipClient: stubOwnershipClient{},
	})

	pair, err := svc.Refresh(ctx, refreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, pair.AccessToken)
	assert.NotEmpty(t, pair.RefreshToken)
	assert.NotEqual(t, refreshToken, pair.RefreshToken)
}

func TestService_Refresh_SecondUseRejected(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	userID := int32(100)
	genParams := &GenerationParams{
		UserID: &userID,
		Roles:  []string{"user"},
		GTY:    GrantTypeInitData,
		JTI:    "reuse-jti",
	}
	refreshToken, err := GenerateJWT(genParams, RefreshToken, 60*time.Minute)
	require.NoError(t, err)

	var getCalls int
	svc := NewService(&Params{
		Config: &Config{AccessExpiryMinutes: 15, RefreshExpiryMinutes: 60},
		TokenStore: stubTokenStore{
			getFn: func(_ context.Context, key string) (string, error) {
				getCalls++
				assert.Equal(t, "refresh_token:reuse-jti", key)
				if getCalls == 1 {
					return "1", nil
				}
				return "", nil
			},
			delFn: func(_ context.Context, key string) error {
				assert.Equal(t, "refresh_token:reuse-jti", key)
				return nil
			},
			setFn: func(_ context.Context, _ string, value string, _ time.Duration) error {
				return nil
			},
		},
		RoleService: stubRoleProvider{
			getRolesFn: func(_ context.Context, gotID int32) ([]database.Role, error) {
				assert.Equal(t, userID, gotID)
				return []database.Role{{ID: 1, Name: "user"}}, nil
			},
		},
		OwnershipClient: stubOwnershipClient{},
	})

	pair, err := svc.Refresh(ctx, refreshToken)
	require.NoError(t, err)
	require.NotNil(t, pair)
	assert.NotEmpty(t, pair.AccessToken)

	_, err = svc.Refresh(ctx, refreshToken)
	require.Error(t, err)
	ae, ok := apperrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, fiber.StatusUnauthorized, ae.Code)
}

func TestService_Refresh_WrongTokenType(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	userID := int32(50)
	accessToken, err := GenerateJWT(&GenerationParams{
		UserID: &userID,
		Roles:  []string{"user"},
		GTY:    GrantTypeInitData,
	}, AccessToken, 15*time.Minute)
	require.NoError(t, err)

	svc := NewService(&Params{
		Config:          &Config{},
		OwnershipClient: stubOwnershipClient{},
	})

	_, err = svc.Refresh(ctx, accessToken)
	require.Error(t, err)
	ae, ok := apperrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, fiber.StatusUnauthorized, ae.Code)
}

func TestService_Refresh_MissingJTI(t *testing.T) {
	InitTestKeys(t)

	ctx := context.Background()
	userID := int32(77)
	refreshToken, err := GenerateJWT(&GenerationParams{
		UserID: &userID,
		Roles:  []string{"user"},
		GTY:    GrantTypeInitData,
		JTI:    "",
	}, RefreshToken, 60*time.Minute)
	require.NoError(t, err)

	svc := NewService(&Params{
		Config:          &Config{},
		OwnershipClient: stubOwnershipClient{},
	})

	_, err = svc.Refresh(ctx, refreshToken)
	require.Error(t, err)
	ae, ok := apperrors.FromError(err)
	require.True(t, ok)
	assert.Equal(t, fiber.StatusUnauthorized, ae.Code)
}
