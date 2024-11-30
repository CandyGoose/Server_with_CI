package handlers

import (
	"encoding/json"
	"io"
	"log"
	"main/internal/note"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	TestErrReadResponseBody   = "failed to read response body"
	TestErrUnmarshalResponse  = "failed to unmarshal response body"
	TestErrStatusCodeMismatch = "status code mismatch"
	TestErrResultMismatch     = "result mismatch"
)

func setupHandler(initialData []*note.Note) *NoteHandler {
	repo := initNoteRepo(initialData)
	return &NoteHandler{
		NoteRepo: repo,
	}
}

func initNoteRepo(notes []*note.Note) *note.MemoryRepo {
	noteRepo := note.NewMemoryRepo()

	for _, n := range notes {
		noteCopy := &note.Note{}
		*noteCopy = *n
		_, err := noteRepo.Add(noteCopy)
		if err != nil {
			log.Printf("Ошибка при добавлении заметки: %v", err)
		}
	}

	return noteRepo
}

func TestGetAll(t *testing.T) {
	initialData := []*note.Note{
		{
			Text: "send a message",
		},
		{
			Text: "buy food",
		},
	}
	handler := setupHandler(initialData)

	expectedData := []*note.Note{
		{
			ID:   1,
			Text: "send a message",
		},
		{
			ID:   2,
			Text: "buy food",
		},
	}

	testCases := []struct {
		Name           string
		URL            string
		ExpectedStatus int
		ExpectedBody   JSONResponse
	}{
		{
			Name:           "Success",
			URL:            "/note",
			ExpectedStatus: http.StatusOK,
			ExpectedBody: JSONResponse{
				Data: expectedData,
			},
		},
		{
			Name:           "Success_OrderByText",
			URL:            "/note?order_field=text",
			ExpectedStatus: http.StatusOK,
			ExpectedBody: JSONResponse{
				Data: []*note.Note{
					expectedData[1],
					expectedData[0],
				},
			},
		},
		{
			Name:           "InvalidOrderField",
			URL:            "/note?order_field=unknown",
			ExpectedStatus: http.StatusBadRequest,
			ExpectedBody: JSONResponse{
				Error: "OrderField: unknown does not validate as in(id|text|created_at|updated_at)",
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.URL, nil)
			w := httptest.NewRecorder()

			handler.GetAll(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err, TestErrReadResponseBody)

			var jsonResp JSONResponse
			err = json.Unmarshal(body, &jsonResp)
			assert.NoError(t, err, TestErrUnmarshalResponse)

			assert.Equal(t, tc.ExpectedStatus, resp.StatusCode, TestErrStatusCodeMismatch)

			if tc.ExpectedStatus == http.StatusOK {
				notesInterface, ok := jsonResp.Data.([]interface{})
				assert.True(t, ok, "expected Data to be a slice")

				var receivedNotes []*note.Note
				for _, n := range notesInterface {
					noteMap, ok := n.(map[string]interface{})
					assert.True(t, ok, "expected each note to be a map")

					noteBytes, err := json.Marshal(noteMap)
					assert.NoError(t, err, TestErrUnmarshalResponse)

					var note note.Note
					err = json.Unmarshal(noteBytes, &note)
					assert.NoError(t, err, TestErrUnmarshalResponse)

					receivedNotes = append(receivedNotes, &note)
				}

				assert.Len(t, receivedNotes, len(tc.ExpectedBody.Data.([]*note.Note)), "unexpected number of notes")

				for i, expectedNote := range tc.ExpectedBody.Data.([]*note.Note) {
					assert.Equal(t, expectedNote.ID, receivedNotes[i].ID, "note ID mismatch")
					assert.Equal(t, expectedNote.Text, receivedNotes[i].Text, "note Text mismatch")
				}
			} else {
				assert.Equal(t, tc.ExpectedBody.Error, jsonResp.Error, TestErrResultMismatch)
			}
		})
	}
}
