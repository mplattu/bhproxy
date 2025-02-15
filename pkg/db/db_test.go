package db

import (
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestOpenSqliteDB(t *testing.T) {
	tempDB := "test.db"
	defer os.Remove(tempDB)

	db, err := OpenSqliteDB(tempDB)
	if err != nil {
		t.Errorf("OpenSqliteDB returned an error: %s", err)
	}
	defer db.Close()

	_, err = OpenSqliteDB("./")
	if err == nil {
		t.Errorf("OpenSqliteDB should have returned an error for invalid filename")
	}

	noPermDB := "no_perms.db"
	f, err := os.Create(noPermDB)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	err = os.Chmod(noPermDB, 0000)
	if err != nil {
		t.Logf("Warning: Could not change file permissions for test (may require root): %v", err)
	}

	defer os.Remove(noPermDB)

	_, err = OpenSqliteDB(noPermDB)
	if err == nil {
		t.Errorf("OpenSqliteDB should have returned an error for file with no permissions")
	}
}

func TestInitSqliteDB(t *testing.T) {
	tempDB := "test.db"
	defer os.Remove(tempDB)

	db, err := OpenSqliteDB(tempDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	err = InitSqliteDB(db)
	if err != nil {
		t.Errorf("InitSqliteDB returned an error: %v", err)
	}

	_, err = db.Exec("INSERT INTO feeds (feed_id) VALUES (?)", "testfeed")
	if err != nil {
		t.Errorf("Could not insert into feeds table, table may not have been created: %v", err)
	}

	_, err = db.Exec("INSERT INTO posts (post_id, feed_id) VALUES (?, ?)", "testpost", "testfeed")
	if err != nil {
		t.Errorf("Could not insert into posts table, table may not have been created: %v", err)
	}

	err = InitSqliteDB(db)
	if err != nil {
		t.Errorf("InitSqliteDB returned an error when tables already exist: %v", err)
	}
}
