package testing

type IdentityTokenProfile struct {
	UserID   int32
	Roles    []string
	FullName string
}

//nolint:gosec // G101
const (
	IntegrationTokenAdmin     = "integration-test-token"
	IntegrationTokenModerator = "moderator-integration-token"
	IntegrationTokenCatalogA  = "catalog-user-a-token"
	IntegrationTokenCatalogB  = "catalog-user-b-token"
)

func DefaultIntegrationIdentityTokens() map[string]IdentityTokenProfile {
	return map[string]IdentityTokenProfile{
		IntegrationTokenAdmin: {
			UserID:   1,
			Roles:    []string{"admin", "moderator"},
			FullName: "Admin User",
		},
		IntegrationTokenModerator: {
			UserID:   2,
			Roles:    []string{"moderator"},
			FullName: "Moderator User",
		},
		IntegrationTokenCatalogA: {
			UserID:   10,
			Roles:    []string{"user"},
			FullName: "Catalog User A",
		},
		IntegrationTokenCatalogB: {
			UserID:   11,
			Roles:    []string{"user"},
			FullName: "Catalog User B",
		},
	}
}
