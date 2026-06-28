package dto

import (
	"time"

	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/utils"
)

type SaveNoteRequest struct {
	Note string `json:"note" validate:"required,min=1,max=1000"`
}

type NotesListResponse struct {
	Data []NoteListItem `json:"data"`
}

type NoteListItem struct {
	ID        int32     `json:"id"`
	WeedID    int32     `json:"weed_id"`
	Note      string    `json:"note"`
	CreatedBy int32     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NoteListItemFromDB(n database.WeedNote) NoteListItem {
	return NoteListItem{
		ID:        n.ID,
		WeedID:    n.WeedID,
		Note:      n.Note,
		CreatedBy: n.CreatedBy,
		CreatedAt: utils.PgTimestampToTime(n.CreatedAt),
		UpdatedAt: utils.PgTimestampToTime(n.UpdatedAt),
	}
}

type CreateNoteResponse struct {
	ID        int32     `json:"id"`
	WeedID    int32     `json:"weed_id"`
	Note      string    `json:"note"`
	CreatedBy int32     `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateNoteResponseFromDB(n database.WeedNote) CreateNoteResponse {
	return CreateNoteResponse{
		ID:        n.ID,
		WeedID:    n.WeedID,
		Note:      n.Note,
		CreatedBy: n.CreatedBy,
		CreatedAt: utils.PgTimestampToTime(n.CreatedAt),
		UpdatedAt: utils.PgTimestampToTime(n.UpdatedAt),
	}
}
