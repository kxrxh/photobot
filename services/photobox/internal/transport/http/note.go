package http

import (
	"strconv"

	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/middleware"
	"csort.ru/coffeebot/internal/note"
	"csort.ru/coffeebot/internal/transport/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type NotesHandler struct {
	notesService *note.NotesService
	validator    *validator.Validate
	log          zerolog.Logger
}

func NewNotesHandler(notesService *note.NotesService, zlog zerolog.Logger) *NotesHandler {
	return &NotesHandler{
		notesService: notesService,
		validator:    validator.New(),
		log:          zlog,
	}
}

func (h *NotesHandler) ListCoffeeNotes(c fiber.Ctx) error {
	coffeeIdStr := c.Params("id")
	if coffeeIdStr == "" {
		h.log.Error().Msg("Missing id parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid coffee ID", nil)
	}

	coffeeId, err := strconv.ParseInt(coffeeIdStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse coffee ID")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid coffee ID", nil)
	}

	notes, err := h.notesService.GetCoffeeNotes(c.Context(), int32(coffeeId))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("coffeeID", int32(coffeeId)).
			Msg("Failed to get notes")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to get notes", nil)
	}

	data := make([]dto.NoteListItem, len(notes))
	for i, n := range notes {
		data[i] = dto.NoteListItemFromDB(n)
	}
	return response.OK(c, data)
}

func (h *NotesHandler) CreateNote(c fiber.Ctx) error {
	coffeeIdStr := c.Params("id")
	if coffeeIdStr == "" {
		h.log.Error().Msg("Missing id parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid coffee ID", nil)
	}

	coffeeId, err := strconv.ParseInt(coffeeIdStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse coffee ID")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid coffee ID", nil)
	}
	var request dto.SaveNoteRequest
	if err := c.Bind().Body(&request); err != nil {
		h.log.Error().Err(err).Msg("Invalid request body")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(request); err != nil {
		h.log.Error().Err(err).Msg("Note validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	created, err := h.notesService.CreateCoffeeNote(
		c.Context(),
		int32(coffeeId),
		request.Note,
		currentIdentity.UserID,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("coffeeID", int32(coffeeId)).
			Msg("Failed to create note")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to create note", nil)
	}
	return response.Created(c, dto.CreateNoteResponseFromDB(created))
}

func (h *NotesHandler) UpdateNote(c fiber.Ctx) error {
	noteIdStr := c.Params("noteId")
	if noteIdStr == "" {
		h.log.Error().Msg("Missing noteId parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid note ID", nil)
	}

	noteId, err := strconv.ParseInt(noteIdStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse note ID")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid note ID", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userRoles, err := middleware.GetUserRoles(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	isAuthorized, err := h.notesService.CheckIsAuthorized(
		c.Context(),
		int32(noteId),
		currentIdentity.UserID,
		userRoles,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("noteID", int32(noteId)).
			Msg("Failed to check authorization")
		return response.Fail(
			c,
			fiber.StatusInternalServerError,
			"Failed to check authorization",
			nil,
		)
	}
	if !isAuthorized {
		return response.Fail(c, fiber.StatusForbidden, "You can only edit your own notes", nil)
	}

	var request dto.SaveNoteRequest
	if err := c.Bind().Body(&request); err != nil {
		h.log.Error().Err(err).Msg("Invalid request body")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}

	if err := h.validator.Struct(request); err != nil {
		h.log.Error().Err(err).Msg("Note validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}

	updatedNote, err := h.notesService.UpdateCoffeeNote(c.Context(), int32(noteId), request.Note)
	if err != nil {
		h.log.Error().Err(err).Int32("noteID", int32(noteId)).Msg("Failed to update note")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to update note", nil)
	}

	return response.OK(c, updatedNote)
}

func (h *NotesHandler) DeleteNote(c fiber.Ctx) error {
	noteIdStr := c.Params("noteId")
	if noteIdStr == "" {
		h.log.Error().Msg("Missing noteId parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid note ID", nil)
	}

	noteId, err := strconv.ParseInt(noteIdStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse note ID")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid note ID", nil)
	}

	currentIdentity, err := middleware.GetCurrentIdentity(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	userRoles, err := middleware.GetUserRoles(c)
	if err != nil {
		return response.Fail(c, fiber.StatusUnauthorized, "Unauthorized", nil)
	}

	isAuthorized, err := h.notesService.CheckIsAuthorized(
		c.Context(),
		int32(noteId),
		currentIdentity.UserID,
		userRoles,
	)
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("noteID", int32(noteId)).
			Msg("Failed to check authorization")
		return response.Fail(
			c,
			fiber.StatusInternalServerError,
			"Failed to check authorization",
			nil,
		)
	}
	if !isAuthorized {
		return response.Fail(c, fiber.StatusForbidden, "You can only delete your own notes", nil)
	}

	err = h.notesService.DeleteCoffeeNote(c.Context(), int32(noteId))
	if err != nil {
		h.log.Error().Err(err).Int32("noteID", int32(noteId)).Msg("Failed to delete note")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to delete note", nil)
	}

	return response.NoContent(c)
}
