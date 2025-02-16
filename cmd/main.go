package main

import (
	"log"
	"net/http"
	"net/http/cgi"

	"github.com/lattots/bhproxy/pkg/handler"
)

const databaseFilename = "data/db.sqlite"

func main() {
	h, err := handler.NewSqliteHandler(databaseFilename)
	if err != nil {
		log.Fatalf("failed to create sqlite handler: %s", err)
	}
	http.HandleFunc("GET /", h.HandleGetFeed)
	if err := cgi.Serve(nil); err != nil {
		log.Fatalf("failed to serve cgi request: %s", err)
	}
}
