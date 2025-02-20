package main

import (
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/lattots/bhproxy/pkg/handler"
	"github.com/lattots/bhproxy/pkg/utility"
)

func main() {
	databaseFilename := os.Getenv("BHP_DB_FILENAME")
	if databaseFilename == "" {
		log.Fatal("required environment variable db_filename is not set or is empty")
	}

	if !utility.FileExists(databaseFilename) {
		log.Fatalf("database file %s does not exist", databaseFilename)
	}
	if !utility.FileIsWriteable(databaseFilename) {
		log.Fatalf("database file %s is not writeable", databaseFilename)
	}

	h, err := handler.NewSqliteHandler(databaseFilename)
	if err != nil {
		log.Fatalf("failed to create sqlite handler: %s", err)
	}
	http.HandleFunc("GET /", h.HandleGetFeed)
	if err := cgi.Serve(nil); err != nil {
		log.Fatalf("failed to serve cgi request: %s", err)
	}
}
