package http

import (
	"context"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/sharelink"
	"github.com/gofiber/fiber/v3"
)

type ReportPackAuthorizer struct {
	secret          []byte
	skew            time.Duration
	authClient      *auth.Client
	analysisService *analysis.Service
}

func NewReportPackAuthorizer(
	cfg config.ShareLinkConfig,
	authClient *auth.Client,
	analysisService *analysis.Service,
) *ReportPackAuthorizer {
	skew := time.Duration(cfg.MaxClockSkewSec) * time.Second
	if skew < 0 {
		skew = 0
	}
	return &ReportPackAuthorizer{
		secret:          []byte(cfg.HMACSecret),
		skew:            skew,
		authClient:      authClient,
		analysisService: analysisService,
	}
}

func (a *ReportPackAuthorizer) EnsureReportPackAccess(c fiber.Ctx, analysisID string) error {
	expQ := c.Query("exp")
	sigQ := c.Query("sig")
	if sharelink.HasShareQuery(expQ, sigQ) {
		if err := sharelink.Verify(
			a.secret,
			analysisID,
			expQ,
			sigQ,
			time.Now(),
			a.skew,
		); err != nil {
			return apierrors.Unauthorized("invalid or expired share link")
		}
		return nil
	}

	const bearerPrefix = "Bearer "
	hdr := c.Get("Authorization")
	if !strings.HasPrefix(hdr, bearerPrefix) || len(hdr) <= len(bearerPrefix) {
		return apierrors.Unauthorized("authentication required")
	}
	tok := strings.TrimSpace(hdr[len(bearerPrefix):])

	ctx, cancel := context.WithTimeout(c.Context(), 10*time.Second)
	defer cancel()

	resp, err := a.authClient.ValidateToken(ctx, tok)
	if err != nil || !resp.Valid || resp.Identity == nil {
		return apierrors.Unauthorized("invalid or expired token")
	}

	id := resp.Identity
	if sharelink.IdentityHasServiceRole(id) {
		return nil
	}

	domain, err := a.analysisService.GetByID(c.Context(), analysisID)
	if err != nil {
		return err
	}
	if !sharelink.IdentityOwnsAnalysis(id, domain.UserID) {
		return apierrors.Forbidden("you do not have access to this analysis")
	}
	return nil
}
