package utility

import "os"

func FileExists(filepath string) bool {
	_, err := os.Stat(filepath)

	return err == nil
}

func FileIsWriteable(filepath string) bool {
	_, err := os.OpenFile(filepath, os.O_RDWR, 0666)

	return err == nil
}
