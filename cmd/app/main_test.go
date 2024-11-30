package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"syscall"
	"testing"
	"time"

	"main/internal/handlers"
	"main/internal/note"

	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	port := 80
	app := NewApp(port)

	assert.NotNil(t, app)
	assert.Equal(t, port, app.Port)
	assert.NotNil(t, app.Router)
	assert.NotNil(t, app.Server)
}

func TestRunServer(t *testing.T) {
	port := 80
	app := NewApp(port)

	go app.Run()

	req := httptest.NewRequest(http.MethodGet, "/note", nil)
	w := httptest.NewRecorder()

	handler := handlers.NoteHandler{
		NoteRepo: note.NewMemoryRepo(),
	}
	handler.GetAll(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestShutdownOnSignal(t *testing.T) {
	port := 80
	app := NewApp(port)

	quit := make(chan os.Signal, 1)
	go func() {
		quit <- syscall.SIGINT
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := app.Server.Shutdown(ctx)
	assert.NoError(t, err, "Error during server shutdown")
}
