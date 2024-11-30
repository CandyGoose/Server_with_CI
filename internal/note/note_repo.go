package note

import (
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrNoteNotFound = errors.New("note not found")
)

type MemoryRepo struct {
	sync.RWMutex
	notes  map[uint64]*Note
	lastID uint64
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		notes: make(map[uint64]*Note),
	}
}

func (m *MemoryRepo) GetAll() ([]*Note, error) {
	m.RLock()
	defer m.RUnlock()

	allNotes := make([]*Note, 0, len(m.notes))
	for _, note := range m.notes {
		allNotes = append(allNotes, note)
	}
	return allNotes, nil
}

func (m *MemoryRepo) GetByID(id uint64) (*Note, error) {
	m.RLock()
	defer m.RUnlock()

	note, exists := m.notes[id]
	if !exists {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

func (m *MemoryRepo) Add(note *Note) (uint64, error) {
	m.Lock()
	defer m.Unlock()

	m.lastID++
	note.ID = m.lastID
	note.CreatedAt = time.Now()
	note.UpdatedAt = time.Now()
	m.notes[note.ID] = note
	return note.ID, nil
}

func (m *MemoryRepo) Update(note *Note) (*Note, error) {
	m.Lock()
	defer m.Unlock()

	existingNote, exists := m.notes[note.ID]
	if !exists {
		return nil, ErrNoteNotFound
	}

	existingNote.Text = note.Text
	existingNote.UpdatedAt = time.Now()
	return existingNote, nil
}

func (m *MemoryRepo) Delete(id uint64) error {
	m.Lock()
	defer m.Unlock()

	if _, exists := m.notes[id]; !exists {
		return ErrNoteNotFound
	}

	delete(m.notes, id)
	return nil
}
