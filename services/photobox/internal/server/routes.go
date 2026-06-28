package server

import (
	"context"
	"time"

	"csort.ru/coffeebot/internal/authz"
	"csort.ru/coffeebot/internal/middleware"
	"csort.ru/coffeebot/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

func setupRoutes(
	ctx context.Context,
	app *fiber.App,
	h *Handlers,
	identityClient authz.IdentityClient,
	log zerolog.Logger,
) {
	authLog := log.With().Str("component", "middleware.auth").Logger()
	jwt := middleware.JWTAuth(ctx, identityClient, authLog)
	adminMod := middleware.WithRoles(authLog, "admin", "moderator")

	app.Get("/health", healthCheckHandler)

	api := app.Group("/api/v1")
	api.Get("/classifications", h.ClassificationHandler.GetClassifications)

	weeds := api.Group("/weeds")
	weeds.Get("/", h.WeedHandler.ListWeeds)
	weeds.Get("/:id", h.WeedHandler.GetWeed)
	weeds.Get("/:id/details", h.WeedHandler.GetFullWeedDetails)
	weeds.Get("/:id/analysis-objects", h.WeedHandler.GetWeedAnalysisObjects)
	weeds.Get("/:id/images", h.WeedHandler.GetWeedImages)
	weeds.Get("/:id/analyses", h.WeedHandler.GetWeedAnalyses)
	weeds.Post("/:id/images", jwt, adminMod, h.WeedHandler.AddWeedImage)
	weeds.Delete("/:id/images/:imageId", jwt, adminMod, h.WeedHandler.DeleteWeedImage)
	weeds.Patch("/:id/images/:imageId/primary", jwt, adminMod, h.WeedHandler.SetPrimaryImage)
	weeds.Put("/:id", jwt, adminMod, h.WeedHandler.UpdateWeed)
	weeds.Delete("/:id", jwt, adminMod, h.WeedHandler.DeleteWeed)

	weedNotes := weeds.Group("/:id/notes", jwt)
	weedNotes.Get("/", h.NotesHandler.ListCoffeeNotes)
	weedNotes.Post("/", h.NotesHandler.CreateNote)

	notes := api.Group("/notes", jwt)
	notes.Put("/:noteId", h.NotesHandler.UpdateNote)
	notes.Delete("/:noteId", h.NotesHandler.DeleteNote)

	proposals := api.Group("/proposals", jwt)
	proposals.Post("/", h.ProposalHandler.CreateProposal)
	proposals.Get("/", h.ProposalHandler.ListProposals)
	proposals.Get("/:id", h.ProposalHandler.GetProposal)
	proposals.Patch("/:id", h.ProposalHandler.UpdateProposalDraft)
	proposals.Post("/:id/submit", h.ProposalHandler.SubmitProposal)
	proposals.Post("/:id/cancel", h.ProposalHandler.CancelProposal)
	proposals.Post("/:id/images", h.ProposalHandler.UploadProposalImage)
	proposals.Delete("/:id/images/:imageId", h.ProposalHandler.DeleteProposalImage)
	proposals.Post("/:id/request-changes", adminMod, h.ProposalHandler.RequestChanges)
	proposals.Post("/:id/reject", adminMod, h.ProposalHandler.RejectProposal)
	proposals.Post("/:id/apply", adminMod, h.ProposalHandler.ApplyProposal)
}

func healthCheckHandler(c fiber.Ctx) error {
	return response.OK(c, fiber.Map{
		"status":  "healthy",
		"service": "coffeebot",
		"time":    time.Now().Format(time.RFC3339),
	})
}
