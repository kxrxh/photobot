package note

import (
	"context"
	"slices"

	"csort.ru/coffeebot/internal/database"
)

type NotesService struct {
	db database.Querier
}

func NewNotesService(db database.Querier) *NotesService {
	return &NotesService{db: db}
}

func (s *NotesService) CreateCoffeeNote(
	ctx context.Context,
	weedID int32,
	note string,
	createdBy int32,
) (database.WeedNote, error) {
	return s.db.CreateWeedNote(ctx, database.CreateWeedNoteParams{
		WeedID:    weedID,
		Note:      note,
		CreatedBy: createdBy,
	})
}

func (s *NotesService) GetCoffeeNotes(
	ctx context.Context,
	weedID int32,
) ([]database.WeedNote, error) {
	return s.db.GetWeedNotes(ctx, weedID)
}

func (s *NotesService) UpdateCoffeeNote(
	ctx context.Context,
	id int32,
	note string,
) (database.WeedNote, error) {
	return s.db.EditWeedNote(ctx, database.EditWeedNoteParams{
		ID:   id,
		Note: note,
	})
}

func (s *NotesService) DeleteCoffeeNote(ctx context.Context, id int32) error {
	return s.db.DeleteWeedNote(ctx, id)
}

// CheckIsAuthorized is true if userID owns the note or roles (from JWT) include admin.
func (s *NotesService) CheckIsAuthorized(
	ctx context.Context,
	noteID int32,
	userID int32,
	rolesFromJWT []string,
) (bool, error) {
	isAuthor, err := s.db.CheckIsAuthorized(ctx, database.CheckIsAuthorizedParams{
		ID:        noteID,
		CreatedBy: userID,
	})
	if err != nil {
		return false, err
	}
	if isAuthor {
		return true, nil
	}

	return slices.Contains(rolesFromJWT, "admin"), nil
}
