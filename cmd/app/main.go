package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"main/internal/handlers"
	"main/internal/note"
)

type App struct {
	Router   *mux.Router
	NoteRepo note.NoteRepo
	Port     int
	Server   *http.Server
}

func NewApp(port int) *App {
	noteRepo := note.NewMemoryRepo()

	noteHandler := &handlers.NoteHandler{
		NoteRepo: noteRepo,
	}

	router := mux.NewRouter()
	router.HandleFunc("/note", noteHandler.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/note", noteHandler.Add).Methods(http.MethodPost)
	router.HandleFunc("/note/{id:[0-9]+}", noteHandler.GetByID).Methods(http.MethodGet)
	router.HandleFunc("/note/{id:[0-9]+}", noteHandler.Update).Methods(http.MethodPut)
	router.HandleFunc("/note/{id:[0-9]+}", noteHandler.Delete).Methods(http.MethodDelete)

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: router,
	}

	return &App{
		Router:   router,
		NoteRepo: noteRepo,
		Port:     port,
		Server:   server,
	}
}

func (app *App) Run() {
	go func() {
		log.Printf("Сервер запущен на порту %d\n", app.Port)
		if err := app.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Ошибка ListenAndServe: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Получен сигнал завершения работы. Завершаем сервер...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при завершении работы сервера: %v\n", err)
	}

	log.Println("Сервер успешно завершил работу.")
}

func main() {
	port := flag.Int("port", 80, "установить порт для прослушивания")
	flag.Parse()

	app := NewApp(*port)

	app.Run()
}
