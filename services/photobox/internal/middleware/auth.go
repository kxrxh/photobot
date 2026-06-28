package middleware

import (
	"context"
	"crypto/sha256"
	"slices"
	"strings"
	"sync"
	"time"

	"csort.ru/coffeebot/internal/authz"
	"csort.ru/coffeebot/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

const (
	AuthorizationHeader = "Authorization"
	UserDataKey         = "user_data"
	UserRolesKey        = "user_roles"
)

type cachedTokenPayload struct {
	Identity *authz.Identity
	Roles    []string
}

type cachedValidation struct {
	payload   cachedTokenPayload
	expiresAt time.Time
}

var (
	tokenCache           = make(map[[sha256.Size]byte]cachedValidation)
	cacheRWMutex         sync.RWMutex
	cacheExpiration      = 1 * time.Minute
	maxTokenCacheEntries = 8192
)

func init() {
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			evictExpiredTokenCache()
		}
	}()
}

func evictExpiredTokenCache() {
	now := time.Now()
	cacheRWMutex.Lock()
	defer cacheRWMutex.Unlock()
	for k, v := range tokenCache {
		if now.After(v.expiresAt) {
			delete(tokenCache, k)
		}
	}
}

func trimTokenCacheIfNeeded() {
	for len(tokenCache) > maxTokenCacheEntries {
		removed := false
		for k := range tokenCache {
			delete(tokenCache, k)
			removed = true
			if len(tokenCache) <= maxTokenCacheEntries {
				break
			}
		}
		if !removed {
			break
		}
	}
}

func storeTokenValidation(cacheKey [sha256.Size]byte, payload cachedTokenPayload) {
	cacheRWMutex.Lock()
	defer cacheRWMutex.Unlock()
	tokenCache[cacheKey] = cachedValidation{
		payload:   payload,
		expiresAt: time.Now().Add(cacheExpiration),
	}
	trimTokenCacheIfNeeded()
}

func loadCachedTokenValidation(
	cacheKey [sha256.Size]byte,
	now time.Time,
) (cachedTokenPayload, bool) {
	cacheRWMutex.RLock()
	cached, found := tokenCache[cacheKey]
	cacheRWMutex.RUnlock()
	if !found {
		return cachedTokenPayload{}, false
	}
	if now.Before(cached.expiresAt) {
		return cached.payload, true
	}
	cacheRWMutex.Lock()
	delete(tokenCache, cacheKey)
	cacheRWMutex.Unlock()
	return cachedTokenPayload{}, false
}

// JWTAuth validates Bearer JWTs via the identity client.
func JWTAuth(
	parentCtx context.Context,
	identityClient authz.IdentityClient,
	log zerolog.Logger,
) fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() == fiber.MethodOptions {
			log.Debug().
				Str("method", c.Method()).
				Str("path", c.Path()).
				Msg("OPTIONS request, skipping auth")
			return c.Next()
		}

		authHeader := c.Get(AuthorizationHeader)
		if authHeader == "" {
			log.Warn().Str("path", c.Path()).Msg("Missing Authorization header")
			return response.Fail(c, fiber.StatusUnauthorized,
				"Authentication required: Missing Authorization header", nil)
		}

		scheme, token, ok := strings.Cut(authHeader, " ")
		if !ok || scheme != "Bearer" || token == "" {
			log.Warn().
				Str("authHeader", authHeader).
				Msg("Invalid Authorization header format")
			return response.Fail(c, fiber.StatusUnauthorized,
				"Invalid Authorization header format. Expected: Bearer <token>", nil)
		}

		cacheKey := sha256.Sum256([]byte(token))
		if payload, ok := loadCachedTokenValidation(cacheKey, time.Now()); ok {
			log.Debug().Msg("Token found in cache, using cached validation")
			c.Locals(UserDataKey, payload.Identity)
			c.Locals(UserRolesKey, payload.Roles)
			return c.Next()
		}

		callCtx, cancel := context.WithTimeout(
			ContextWithRequestSpan(parentCtx, c.Context()),
			10*time.Second,
		)
		defer cancel()

		validationResp, err := identityClient.ValidateToken(callCtx, token)
		if err != nil {
			log.Warn().
				Err(err).
				Str("path", c.Path()).
				Str("error", err.Error()).
				Msg("Token validation failed")
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid or expired token", nil)
		}

		if !validationResp.Valid {
			log.Warn().Str("path", c.Path()).Msg("Token validation returned invalid")
			return response.Fail(c, fiber.StatusUnauthorized, "Invalid token", nil)
		}

		storeTokenValidation(cacheKey, cachedTokenPayload{
			Identity: validationResp.Identity,
			Roles:    validationResp.Roles,
		})

		c.Locals(UserDataKey, validationResp.Identity)
		c.Locals(UserRolesKey, validationResp.Roles)

		return c.Next()
	}
}

func WithRoles(log zerolog.Logger, requiredRoles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRoles, ok := c.Locals(UserRolesKey).([]string)
		if !ok {
			log.Error().Msg("Could not determine user roles")
			return response.Fail(c, fiber.StatusForbidden, "Could not determine user roles", nil)
		}

		if len(requiredRoles) == 0 {
			return c.Next()
		}

		for _, role := range requiredRoles {
			if slices.Contains(userRoles, role) {
				return c.Next()
			}
		}

		return response.Fail(c, fiber.StatusForbidden, "Insufficient permissions", nil)
	}
}
