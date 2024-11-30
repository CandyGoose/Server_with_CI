package note

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRepo_GetByID(t *testing.T) {
	repo := NewMemoryRepo()

	note := &Note{Text: "Unique note"}
	id, err := repo.Add(note)
	assert.NoError(t, err)

	retrievedNote, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, id, retrievedNote.ID)
	assert.Equal(t, note.Text, retrievedNote.Text)

	_, err = repo.GetByID(999)
	assert.Equal(t, ErrNoteNotFound, err)
}

func TestMemoryRepo_Add(t *testing.T) {
	repo := NewMemoryRepo()

	note := &Note{Text: "New note"}
	id, err := repo.Add(note)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), id)

	retrievedNote, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, id, retrievedNote.ID)
	assert.Equal(t, note.Text, retrievedNote.Text)
	assert.WithinDuration(t, time.Now(), retrievedNote.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), retrievedNote.UpdatedAt, time.Second)
}

func TestMemoryRepo_Update_NonExistent(t *testing.T) {
	repo := NewMemoryRepo()

	update := &Note{
		ID:   999,
		Text: "Non-existent note",
	}
	_, err := repo.Update(update)
	assert.Equal(t, ErrNoteNotFound, err)
}

func TestMemoryRepo_Delete(t *testing.T) {
	repo := NewMemoryRepo()

	note := &Note{Text: "Note to delete"}
	id, err := repo.Add(note)
	assert.NoError(t, err)

	err = repo.Delete(id)
	assert.NoError(t, err)

	_, err = repo.GetByID(id)
	assert.Equal(t, ErrNoteNotFound, err)

	err = repo.Delete(id)
	assert.Equal(t, ErrNoteNotFound, err)
}

func TestMemoryRepo_Delete_NonExistent(t *testing.T) {
	repo := NewMemoryRepo()

	err := repo.Delete(999)
	assert.Equal(t, ErrNoteNotFound, err)
}
