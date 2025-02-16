package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenSqliteDB(filename string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filename)
	if err != nil {
		return nil, fmt.Errorf("error opening sqlite database: %w", err)
	}
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("error pinging sqlite database: %w", err)
	}
	return db, nil
}

func InitSqliteDB(db *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS feeds 
		(feed_id TEXT PRIMARY KEY, 
		username TEXT,
		biography TEXT,
		profile_picture_url TEXT,
		website TEXT,
		followers_count INT,
		follows_count INT,
		last_fetched TIMESTAMP)`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating feeds table: %w", err)
	}

	query = `CREATE TABLE IF NOT EXISTS posts
		(post_id TEXT PRIMARY KEY,
		feed_id TEXT,
		permalink TEXT,
		timestamp TIMESTAMP,
		media_type TEXT,
		media_small_url TEXT,
		media_small_height INT,
		media_small_width INT,
		caption TEXT,
		pruned_caption TEXT)`

	_, err = db.Exec(query)
	if err != nil {
		return fmt.Errorf("error creating posts table: %w", err)
	}
	return nil
}
