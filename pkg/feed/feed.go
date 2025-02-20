package feed

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Feed struct {
	ID                string `json:"id"`
	Username          string `json:"username"`
	Biography         string `json:"biography"`
	ProfilePictureUrl string `json:"profilePictureUrl"`
	Website           string `json:"website"`
	FollowersCount    int    `json:"followersCount"`
	FollowsCount      int    `json:"followsCount"`
	Posts             []Post `json:"posts"`

	lastFetched time.Time
}

type Post struct {
	ID               string `json:"id"`
	feedID           string
	Permalink        string    `json:"permalink"`
	Timestamp        time.Time `json:"timestamp"`
	MediaType        string    `json:"mediaType"`
	MediaSmallUrl    string    `json:"mediaSmallUrl"`
	MediaSmallHeight int       `json:"mediaSmallHeight"`
	MediaSmallWidth  int       `json:"mediaSmallWidth"`
	Caption          string    `json:"caption"`
	PrunedCaption    string    `json:"prunedCaption"`

	mediaSmallExternalURL string
}

func GetFeedWithID(db *sql.DB, id string) (*Feed, error) {
	if !isAllowedFeedId(id) {
		return nil, fmt.Errorf("given feed id %s is not in the whitelist", id)
	}

	feed := &Feed{ID: id}

	err := feed.fetchOrCreateFeed(db)
	if err != nil {
		return nil, fmt.Errorf("error fetching feed: %w", err)
	}

	err = feed.populatePostImages(db)
	if err != nil {
		return nil, fmt.Errorf("error populating post images: %w", err)
	}

	// server is left cleaning up deprecated posts on its own
	go feed.removeDeprecatedPosts(db)

	return feed, nil
}

func isAllowedFeedId(feedID string) bool {
	allowedFeedIdsStr := os.Getenv("BHP_ALLOWED_FEED_IDS")
	allowedFeedIdsStrWithoutSpaces := strings.ReplaceAll(allowedFeedIdsStr, " ", "")

	if allowedFeedIdsStrWithoutSpaces == "" {
		return true
	}

	allowedFeedIds := strings.Split(allowedFeedIdsStrWithoutSpaces, ",")

	return slices.Contains(allowedFeedIds, feedID)
}

func (f *Feed) fetchOrCreateFeed(db *sql.DB) error {
	queryRes, err := queryFeed(db, f.ID)
	if err != nil {
		return fmt.Errorf("failed to fetch feed from db: %w", err)
	}

	err = parseFeedRows(queryRes, f)
	if err == nil {
		log.Println("found feed from local database")
	} else if errors.Is(err, ErrFeedNotFound) {
		log.Println("feed not found from local database")
		err = f.getFromBehold()
		if err != nil {
			return fmt.Errorf("failed to get feed from Behold: %w", err)
		}

		err = f.insertToDB(db)
		if err != nil {
			return fmt.Errorf("failed to insert feed in database: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to parse feed from rows: %w", err)
	}

	return nil
}

func (f *Feed) populatePostImages(db *sql.DB) error {
	relevantPostIDs, err := f.getRelevantPosts(db)
	if err != nil {
		return fmt.Errorf("failed to get relevant post ids: %w", err)
	}

	imageURLs, err := ensurePostImagesExist(db, relevantPostIDs)
	if err != nil {
		return fmt.Errorf("failed to ensure post images exist: %w", err)
	}

	for i := range imageURLs {
		f.Posts[i].MediaSmallUrl = imageURLs[i]
	}

	return nil
}

func (f *Feed) insertToDB(db *sql.DB) error {
	f.lastFetched = time.Now()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(
		`INSERT INTO feeds 
		(feed_id, username, biography, profile_picture_url, website, followers_count, follows_count, last_fetched) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(feed_id) DO UPDATE SET
		username = excluded.username,
		biography = excluded.biography,
		profile_picture_url = excluded.profile_picture_url,
		website = excluded.website,
		followers_count = excluded.followers_count,
		follows_count = excluded.follows_count,
		last_fetched = excluded.last_fetched;`,
		f.ID, f.Username, f.Biography, f.ProfilePictureUrl, f.Website,
		f.FollowersCount, f.FollowsCount, f.lastFetched,
	)
	if err != nil {
		return fmt.Errorf("failed to insert feed: %w", err)
	}

	for _, post := range f.Posts {
		_, err = tx.Exec(
			`INSERT INTO posts 
			(post_id, feed_id, permalink, timestamp, media_type, media_small_url, media_small_height, media_small_width, caption, pruned_caption) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(post_id) DO UPDATE SET
			feed_id = excluded.feed_id,
			permalink = excluded.permalink,
			timestamp = excluded.timestamp,
			media_type = excluded.media_type,
			media_small_url = excluded.media_small_url,
			media_small_height = excluded.media_small_height,
			media_small_width = excluded.media_small_width,
			caption = excluded.caption,
			pruned_caption = excluded.pruned_caption;`,
			post.ID, post.feedID, post.Permalink, post.Timestamp, post.MediaType,
			post.mediaSmallExternalURL, post.MediaSmallHeight, post.MediaSmallWidth,
			post.Caption, post.PrunedCaption,
		)
		if err != nil {
			return fmt.Errorf("failed to insert post: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func getImageDirectory() (string, error) {
	imageDirectory := os.Getenv("BHP_IMAGE_DIRECTORY")
	if imageDirectory == "" {
		return "", errors.New("required environment variable image_directory is not set or is empty")
	}

	return imageDirectory, nil
}

func getImageURL() string {
	return os.Getenv("BHP_IMAGE_URL")
}

func ensurePostImagesExist(db *sql.DB, postIDs []string) ([]string, error) {
	internalURLs := make([]string, len(postIDs))
	imageDirectory, err := getImageDirectory()
	if err != nil {
		return nil, fmt.Errorf("failed to check image file: %w", err)
	}

	for i, postID := range postIDs {
		fileNameMediaSmall := postID + ".webp"
		filePathMediaSmall := filepath.Join(imageDirectory, fileNameMediaSmall)
		internalURLs[i] = fmt.Sprintf("%s/%s", getImageURL(), fileNameMediaSmall)

		// If the image exists, skip download
		if _, err := os.Stat(filePathMediaSmall); err == nil {
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to check image file: %w", err)
		}
		// if image is not found, it is downloaded from external source
		dlURL, err := getImageExternalURL(db, postID)
		if errors.Is(err, ErrPostNotFound) {
			return nil, fmt.Errorf("post %s not found: %w", postID, err)
		}
		if err != nil {
			return nil, fmt.Errorf("error getting images external url: %w", err)
		}
		err = downloadImage(filePathMediaSmall, dlURL)
		if err != nil {
			return nil, fmt.Errorf("error downloading image: %w", err)
		}
	}
	return internalURLs, nil
}

// ErrPostNotFound means that post with given ID can't be found in the database
var ErrPostNotFound = errors.New("post not found")

// getImageExternalURL returns the external download URL of image
func getImageExternalURL(db *sql.DB, postID string) (string, error) {
	query := `SELECT media_small_url FROM posts WHERE post_id = ?`
	row := db.QueryRow(query, postID)
	var url string
	err := row.Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrPostNotFound
	} else if err != nil {
		return "", fmt.Errorf("failed to query post image url: %w", err)
	}
	return url, nil
}

// downloadImage downloads the image from external source and saves it to image directory with filename:
// image-directory/IMAGE_ID.webp
func downloadImage(filePath, url string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download image, status: %d", resp.StatusCode)
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write image to file: %w", err)
	}

	return nil
}

// ErrFeedNotFound means that feed with given ID can't be found in the database
var ErrFeedNotFound = errors.New("feed not found")

// queryFeed tries to fetch feed and all of its posts from database
func queryFeed(db *sql.DB, id string) (*sql.Rows, error) {
	rows, err := db.Query(
		`SELECT
    	feeds.feed_id,
        username,
        biography,
        profile_picture_url,
        website,
        followers_count,
        follows_count,
        post_id,
        posts.feed_id,
        permalink,
        timestamp,
        media_type,
        media_small_url,
        media_small_height,
        media_small_width,
        caption,
        pruned_caption
    FROM feeds
    INNER JOIN posts ON feeds.feed_id = posts.feed_id
    WHERE feeds.feed_id = ? AND last_fetched >= DATETIME('now', '-1 day');`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("error querying feeds: %w", err)
	}
	return rows, nil
}

// parseFeedRows tries to parse database query result to the receiver pointer "feed"
func parseFeedRows(rows *sql.Rows, feed *Feed) error {
	posts := make([]Post, 0)
	resultCount := 0
	for rows.Next() {
		resultCount++
		post := Post{}
		err := rows.Scan(
			&feed.ID,
			&feed.Username,
			&feed.Biography,
			&feed.ProfilePictureUrl,
			&feed.Website,
			&feed.FollowersCount,
			&feed.FollowsCount,
			&post.ID,
			&post.feedID,
			&post.Permalink,
			&post.Timestamp,
			&post.MediaType,
			&post.mediaSmallExternalURL,
			&post.MediaSmallHeight,
			&post.MediaSmallWidth,
			&post.Caption,
			&post.PrunedCaption,
		)
		if err != nil {
			return fmt.Errorf("error scanning row: %w", err)
		}
		posts = append(posts, post)
	}
	if resultCount == 0 {
		return ErrFeedNotFound
	}
	feed.Posts = posts
	return nil
}

func (f *Feed) removeDeprecatedPosts(db *sql.DB) {
	imageDirectory, err := getImageDirectory()
	if err != nil {
		log.Printf("could not remove deprecated posts: %w", err)
		return
	}

	ids, err := f.getIrrelevantPosts(db)
	if err != nil {
		log.Printf("error getting post ids for feed %s: %s", f.ID, err)
		return
	}
	// if feed has no irrelevant posts, function returns
	if len(ids) == 0 {
		return
	}

	query := `DELETE FROM posts WHERE post_id IN (?)`
	_, err = db.Exec(query, ids)
	if err != nil {
		log.Printf("error deleting posts for feed %s: %s", f.ID, err)
	}

	for _, postID := range ids {
		filePath := filepath.Join(imageDirectory, postID+".webp")
		// check if file already doesn't exist
		if _, err := os.Stat(filePath); err != nil {
			continue
		}
		// if file exists, it is removed
		err = os.Remove(filePath)
		if err != nil {
			log.Printf("error removing file %s: %s", filePath, err)
		}
	}
}

// getRelevantPosts returns the IDs of all relevant posts that belong to the Feed
func (f *Feed) getRelevantPosts(db *sql.DB) ([]string, error) {
	// get 6 most recent post IDs from feed
	query := `SELECT post_id FROM posts WHERE feed_id = ? ORDER BY timestamp DESC LIMIT 6;`
	rows, err := db.Query(query, f.ID)
	if err != nil {
		return nil, fmt.Errorf("error fetching last 6 posts from feed: %w", err)
	}
	defer rows.Close()
	postIDs := make([]string, 0)
	for rows.Next() {
		var postID string
		err = rows.Scan(&postID)
		if err != nil {
			return nil, fmt.Errorf("error scanning post ID from row: %w", err)
		}
		postIDs = append(postIDs, postID)
	}
	return postIDs, nil
}

// getIrrelevantPosts returns the IDs of all irrelevant (very old) posts that belong to the Feed
func (f *Feed) getIrrelevantPosts(db *sql.DB) ([]string, error) {
	// get all post IDs from the feed
	query := `SELECT post_id FROM posts WHERE feed_id = ? ORDER BY timestamp;`
	rows, err := db.Query(query, f.ID)
	if err != nil {
		return nil, fmt.Errorf("error fetching last 6 posts from feed: %w", err)
	}
	defer rows.Close()
	postIDs := make([]string, 0)
	for rows.Next() {
		var postID string
		err = rows.Scan(&postID)
		if err != nil {
			return nil, fmt.Errorf("error scanning post ID from row: %w", err)
		}
		postIDs = append(postIDs, postID)
	}
	if len(postIDs) > 6 {
		// 6 most recent posts are still relevant so skip them
		return postIDs[6:], nil
	}
	return []string{}, nil
}
