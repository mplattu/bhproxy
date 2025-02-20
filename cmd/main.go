package main

import (
	"log"
	"net/http"
	"net/http/cgi"
	"os"

	"github.com/joho/godotenv"

	"github.com/lattots/bhproxy/pkg/handler"
	"github.com/lattots/bhproxy/pkg/utility"
)

func routeLogMessages(logFileName string) *os.File {
	if logFileName != "" {
		logFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("could not open log file %s for append: %s", logFileName, err)
		}

		return logFile
	}

	return nil
}

func getDatabaseFilename() string {
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

	return databaseFilename
}

func main() {
	dotEnvPath := utility.GetDotEnvPath()
	if utility.FileExists(dotEnvPath) {
		err := godotenv.Load(utility.GetDotEnvPath())
		if err != nil {
			log.Fatalf("could not read .env file at %s: %s", dotEnvPath, err)
		}
	}

	logFile := routeLogMessages(os.Getenv("BHP_LOGFILE"))
	if logFile != nil {
		defer logFile.Close()
		log.SetOutput(logFile)
	}

	databaseFilename := getDatabaseFilename()

	h, err := handler.NewSqliteHandler(databaseFilename)
	if err != nil {
		log.Fatalf("failed to create sqlite handler: %s", err)
	}
	http.HandleFunc("GET /", h.HandleGetFeed)
	if err := cgi.Serve(nil); err != nil {
		log.Fatalf("failed to serve cgi request: %s", err)
	}
}
