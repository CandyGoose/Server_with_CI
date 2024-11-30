package note

import "time"

type Note struct {
	ID        uint64
	Text      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type NoteRepo interface {
	GetAll() ([]*Note, error)
	GetByID(id uint64) (*Note, error)
	Add(note *Note) (uint64, error)
	Update(newNote *Note) (*Note, error)
	Delete(id uint64) error
}
