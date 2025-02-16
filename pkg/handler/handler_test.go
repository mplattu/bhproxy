package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/lattots/bhproxy/pkg/db"
	"github.com/lattots/bhproxy/pkg/feed"
)

func TestNewSqliteHandler(t *testing.T) {
	tempDB := "test.db"
	defer os.Remove(tempDB)
	handler, err := NewSqliteHandler(tempDB)
	if err != nil {
		t.Errorf("NewSqliteHandler returned an error: %s", err)
	}
	if handler == nil {
		t.Errorf("NewSqliteHandler returned nil handler")
	}
}

func TestSqliteHandler(t *testing.T) {
	testDBfilepath := "data/db.sqlite"

	handler, err := NewSqliteHandler(testDBfilepath)
	if err != nil {
		t.Errorf("NewSqliteHandler returned an error: %s", err)
	}

	err = populateTestDB(testDBfilepath)
	if err != nil {
		t.Fatalf("populateTestDB returned an error: %s", err)
	}

	http.HandleFunc("GET /", handler.HandleGetFeed)

	go func() {
		fmt.Println("Test server listening on port 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			t.Errorf("NewSqliteHandler returned an error: %s", err)
		}
	}()

	resp, err := http.Get("http://localhost:8080?id=1234")
	if err != nil {
		t.Errorf("error making http request to test server: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	f := &feed.Feed{}
	err = json.NewDecoder(resp.Body).Decode(f)
	if err != nil {
		t.Errorf("error decoding json: %s", err)
	}

	if f == nil {
		t.Errorf("server returned nil feed")
	}
	if f.ID != "1234" {
		t.Errorf("expected feed with id 1234, got %s", f.ID)
	}
	if len(f.Posts) != 1 {
		t.Errorf("expected 1 post, got %d", len(f.Posts))
	}
	if f.Posts[0].ID != "abcd" {
		t.Errorf("expected post with id abcd, got %s", f.Posts[0].ID)
	}
}

func populateTestDB(filepath string) error {
	testDB, err := db.OpenSqliteDB(filepath)
	if err != nil {
		return err
	}

	upsertFeedsQuery := `
	INSERT OR REPLACE INTO feeds (
		feed_id,
		username,
		biography,
		profile_picture_url,
		website,
		followers_count,
		follows_count,
		last_fetched
	) VALUES (
		'1234',
		'johndoe',
		'Software Engineer',
		'https://example.com/johndoe.jpg',
		'https://johndoe.com',
		1500,
		500,
		CURRENT_TIMESTAMP
	);
	`

	_, err = testDB.Exec(upsertFeedsQuery)
	if err != nil {
		return err
	}

	upsertPostsQuery := `
	INSERT OR REPLACE INTO posts (
		post_id,
		feed_id,
		permalink,
		timestamp,
		media_type,
		media_small_url,
		media_small_height,
		media_small_width,
		caption,
		pruned_caption
	) VALUES (
		'abcd',
		'1234',
		'https://example.com/posts/abcd',
		'2024-07-26 10:00:00',
		'image',
		'https://www.gstatic.com/webp/gallery/1.webp',
		300,
		300,
		'Beautiful sunset',
		'Beautiful sunset...'
	);
	`

	_, err = testDB.Exec(upsertPostsQuery)
	return err
}
