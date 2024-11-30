package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"

	"main/internal/note"
)

const (
	ErrDecodeQueryParams = "decode query params failed"
	ErrUnmarshalPayload  = "unmarshal payload failed"
	ErrReadPayload       = "failed to read payload"
	ErrGetNotes          = "failed to get notes"
	ErrAddNewNote        = "failed to add new note"
	ErrUpdateNote        = "failed to update note"
	ErrDeleteNote        = "failed to delete note"
	ErrNoteDoesNotExist  = "note does not exist"
	ErrBadNoteID         = "bad note id"
	ErrEncodeResponse    = "failed to encode response"
	ErrCloseBody         = "failed to close request body"
)

type NoteHandler struct {
	NoteRepo note.NoteRepo
}

type NoteReq struct {
	Text string `json:"text" valid:"required"`
}

type Options struct {
	OrderField string `schema:"order_field" valid:"in(id|text|created_at|updated_at), optional"`
}

type JSONResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
}

func parseAndValidateQueryParams(r *http.Request) (*Options, *JSONResponse) {
	decoder := schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)

	opt := &Options{}
	if err := decoder.Decode(opt, r.URL.Query()); err != nil {
		log.Printf("Error: %s", ErrDecodeQueryParams)
		return nil, &JSONResponse{Error: ErrDecodeQueryParams}
	}

	if _, err := govalidator.ValidateStruct(opt); err != nil {
		log.Printf("Error: %s", err.Error())
		return nil, &JSONResponse{Error: err.Error()}
	}

	return opt, nil
}

func parseAndValidateBody(r *http.Request, dst interface{}) *JSONResponse {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error: %s", ErrReadPayload)
		return &JSONResponse{Error: ErrReadPayload}
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("Error: %s", ErrCloseBody)
		}
	}()

	if err := json.Unmarshal(body, dst); err != nil {
		log.Printf("Error: %s", ErrUnmarshalPayload)
		return &JSONResponse{Error: ErrUnmarshalPayload}
	}

	if _, err := govalidator.ValidateStruct(dst); err != nil {
		log.Printf("Error: %s", err.Error())
		return &JSONResponse{Error: err.Error()}
	}

	return nil
}

func respond(w http.ResponseWriter, status int, resp *JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error: %s", ErrEncodeResponse)
	}
}

func (h *NoteHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	opt, errResp := parseAndValidateQueryParams(r)
	if errResp != nil {
		respond(w, http.StatusBadRequest, errResp)
		return
	}

	notes, err := h.NoteRepo.GetAll()
	if err != nil {
		log.Printf("Error: %s", ErrGetNotes)
		respond(w, http.StatusInternalServerError, &JSONResponse{Error: ErrGetNotes})
		return
	}

	if opt.OrderField != "" {
		sortNotes(notes, opt.OrderField)
	}

	respond(w, http.StatusOK, &JSONResponse{Data: notes})
}

func (h *NoteHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	noteID, err := extractNoteID(r)
	if err != nil {
		log.Printf("Error: %s", ErrBadNoteID)
		respond(w, http.StatusBadRequest, &JSONResponse{Error: ErrBadNoteID})
		return
	}

	n, err := h.NoteRepo.GetByID(noteID)
	if err != nil {
		if errors.Is(err, note.ErrNoteNotFound) {
			respond(w, http.StatusUnprocessableEntity, &JSONResponse{Error: ErrNoteDoesNotExist})
		} else {
			log.Printf("Error: %s", ErrGetNotes)
			respond(w, http.StatusInternalServerError, &JSONResponse{Error: ErrGetNotes})
		}
		return
	}

	respond(w, http.StatusOK, &JSONResponse{Data: n})
}

func (h *NoteHandler) Add(w http.ResponseWriter, r *http.Request) {
	req := &NoteReq{}
	if errResp := parseAndValidateBody(r, req); errResp != nil {
		respond(w, http.StatusBadRequest, errResp)
		return
	}

	n := &note.Note{
		Text:      req.Text,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	createdNote, err := h.NoteRepo.Add(n)
	if err != nil {
		log.Printf("Error: %s", ErrAddNewNote)
		respond(w, http.StatusInternalServerError, &JSONResponse{Error: ErrAddNewNote})
		return
	}

	respond(w, http.StatusCreated, &JSONResponse{Data: createdNote})
}

func (h *NoteHandler) Update(w http.ResponseWriter, r *http.Request) {
	noteID, err := extractNoteID(r)
	if err != nil {
		log.Printf("Error: %s", ErrBadNoteID)
		respond(w, http.StatusBadRequest, &JSONResponse{Error: ErrBadNoteID})
		return
	}

	req := &NoteReq{}
	if errResp := parseAndValidateBody(r, req); errResp != nil {
		respond(w, http.StatusBadRequest, errResp)
		return
	}

	n := &note.Note{
		ID:        noteID,
		Text:      req.Text,
		UpdatedAt: time.Now(),
	}

	updatedNote, err := h.NoteRepo.Update(n)
	if err != nil {
		if errors.Is(err, note.ErrNoteNotFound) {
			respond(w, http.StatusUnprocessableEntity, &JSONResponse{Error: ErrNoteDoesNotExist})
		} else {
			log.Printf("Error: %s", ErrUpdateNote)
			respond(w, http.StatusInternalServerError, &JSONResponse{Error: ErrUpdateNote})
		}
		return
	}

	respond(w, http.StatusOK, &JSONResponse{Data: updatedNote})
}

func (h *NoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	noteID, err := extractNoteID(r)
	if err != nil {
		log.Printf("Error: %s", ErrBadNoteID)
		respond(w, http.StatusBadRequest, &JSONResponse{Error: ErrBadNoteID})
		return
	}

	if err := h.NoteRepo.Delete(noteID); err != nil {
		if errors.Is(err, note.ErrNoteNotFound) {
			respond(w, http.StatusUnprocessableEntity, &JSONResponse{Error: ErrNoteDoesNotExist})
		} else {
			log.Printf("Error: %s", ErrDeleteNote)
			respond(w, http.StatusInternalServerError, &JSONResponse{Error: ErrDeleteNote})
		}
		return
	}

	respond(w, http.StatusOK, &JSONResponse{Data: map[string]string{"message": "success"}})
}

func extractNoteID(r *http.Request) (uint64, error) {
	vars := mux.Vars(r)
	return strconv.ParseUint(vars["id"], 10, 64)
}

func sortNotes(notes []*note.Note, orderField string) {
	sort.SliceStable(notes, func(i, j int) bool {
		switch orderField {
		case "id":
			return notes[i].ID < notes[j].ID
		case "text":
			return strings.ToLower(notes[i].Text) < strings.ToLower(notes[j].Text)
		case "created_at":
			return notes[i].CreatedAt.Before(notes[j].CreatedAt)
		case "updated_at":
			return notes[i].UpdatedAt.Before(notes[j].UpdatedAt)
		default:
			return notes[i].ID < notes[j].ID
		}
	})
}
