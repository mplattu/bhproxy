package utility

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "bhproxy-test")
	defer os.Remove(f.Name())
	if err != nil {
		t.Errorf("Could not create temporary file: %s", err)
	}

	if !FileExists(f.Name()) {
		t.Errorf("FileExists returns true although file should exist")
	}

	os.Remove(f.Name())

	if FileExists(f.Name()) {
		t.Errorf("FileExists returns true although file does not exist")
	}
}

func TestFileIsWriteable(t *testing.T) {
	f, err := os.CreateTemp("", "bhproxy-test")
	defer os.Remove(f.Name())
	if err != nil {
		t.Errorf("Could not create temporary file: %s", err)
	}

	err = os.Chmod(f.Name(), 0400)
	if err != nil {
		t.Errorf("Could not change file permissions: %s", err)
	}

	if FileIsWriteable(f.Name()) {
		t.Errorf("FileIsWriteable returns true although file should not be writeable")
	}
}
