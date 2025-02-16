package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/lattots/bhproxy/pkg/db"
	"github.com/lattots/bhproxy/pkg/feed"
)

type Handler interface {
	HandleGetFeed(http.ResponseWriter, *http.Request)
}

type sqliteHandler struct {
	db *sql.DB
}

func (h *sqliteHandler) HandleGetFeed(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Printf("HandleGetFeed for %s", id)

	f, err := feed.GetFeedWithID(h.db, id)
	if errors.Is(err, feed.ErrFeedNotExists) {
		w.WriteHeader(http.StatusNotFound)
		log.Println("feed doesn't exist")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error getting feed:", err)
		return
	}
	if f == nil {
		w.WriteHeader(http.StatusNotFound)
		log.Println("feed not found with id:", id)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(f); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println("error encoding feed to response:", err)
		return
	}
}

func NewSqliteHandler(filename string) (Handler, error) {
	database, err := db.OpenSqliteDB(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}
	err = db.InitSqliteDB(database)
	if err != nil {
		return nil, fmt.Errorf("error initializing sqlite db: %w", err)
	}
	return &sqliteHandler{db: database}, nil
}
