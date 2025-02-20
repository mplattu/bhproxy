package utility

import (
	"log"
	"os"
	"path/filepath"
)

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	return err == nil
}

func FileIsWriteable(filepath string) bool {
	_, err := os.OpenFile(filepath, os.O_RDWR, 0666)

	return err == nil
}

func GetDotEnvPath() string {
	executable, err := os.Executable()
	if err != nil {
		log.Printf("could not get path of the executable script: %s", err)
		return ".env"
	}

	return filepath.Join(filepath.Dir(executable), ".env")
}
