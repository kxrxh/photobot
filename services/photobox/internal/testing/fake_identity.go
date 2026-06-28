package testing

import (
	"context"
	"errors"
	"fmt"

	"csort.ru/coffeebot/internal/authz"
)

type FakeIdentityClient struct {
	byToken  map[string]IdentityTokenProfile
	byUserID map[int32]IdentityTokenProfile
}

var _ authz.IdentityClient = (*FakeIdentityClient)(nil)

func NewFakeIdentityClient(byToken map[string]IdentityTokenProfile) *FakeIdentityClient {
	byUserID := make(map[int32]IdentityTokenProfile, len(byToken))
	for _, p := range byToken {
		byUserID[p.UserID] = p
	}
	return &FakeIdentityClient{byToken: byToken, byUserID: byUserID}
}

func (f *FakeIdentityClient) ValidateToken(
	_ context.Context,
	token string,
) (*authz.TokenValidationResponse, error) {
	p, ok := f.byToken[token]
	if !ok {
		return &authz.TokenValidationResponse{Valid: false}, nil
	}
	roles := append([]string(nil), p.Roles...)
	id := &authz.Identity{UserID: p.UserID, Roles: roles}
	return &authz.TokenValidationResponse{
		Valid:    true,
		Identity: id,
		Roles:    roles,
	}, nil
}

func (f *FakeIdentityClient) GetUser(_ context.Context, userID int32) (*authz.User, error) {
	p, ok := f.byUserID[userID]
	name := p.FullName
	if !ok || name == "" {
		name = fmt.Sprintf("User %d", userID)
	}
	return &authz.User{ID: userID, FullName: name}, nil
}

func (f *FakeIdentityClient) GetUserRoles(_ context.Context, userID int32) ([]authz.Role, error) {
	p, ok := f.byUserID[userID]
	if !ok {
		return nil, nil
	}
	out := make([]authz.Role, len(p.Roles))
	for i, n := range p.Roles {
		out[i] = authz.Role{ID: int32(i + 1), Name: n}
	}
	return out, nil
}

func (f *FakeIdentityClient) GetUserByTelegramID(context.Context, int64) (*authz.User, error) {
	return nil, errors.New("GetUserByTelegramID: not configured in test fake")
}
