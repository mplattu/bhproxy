package feed

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (f *Feed) getFromBehold() error {
	resp, err := fetchFeedResponse(http.DefaultClient, f.ID)
	if err != nil {
		return fmt.Errorf("error fetching feed from Behold: %w", err)
	}
	err = parseFeedFromResponse(resp, f)
	if err != nil {
		return fmt.Errorf("error parsing feed response from Behold: %w", err)
	}
	return nil
}

const defaultApiBaseUrl = "https://feeds.behold.so/"

// ErrFeedNotExists indicates that feed doesn't exist even in Beholds database
var ErrFeedNotExists = errors.New("feed not found")

func fetchFeedResponse(client HTTPClient, id string) (feedResponse, error) {
	apiBaseUrl := os.Getenv("BHP_BASEURL")
	if apiBaseUrl == "" {
		apiBaseUrl = defaultApiBaseUrl
	}

	url := apiBaseUrl + id

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return feedResponse{}, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return feedResponse{}, fmt.Errorf("error fetching feed %s: %w", id, err)
	}

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return feedResponse{}, ErrFeedNotExists
	}
	if resp.StatusCode != http.StatusOK {
		return feedResponse{}, fmt.Errorf("error fetching feed %s: status code %d", id, resp.StatusCode)
	}
	var feed feedResponse
	feed.ID = id
	err = json.NewDecoder(resp.Body).Decode(&feed)
	if err != nil {
		return feedResponse{}, fmt.Errorf("error parsing feed %s: %w", id, err)
	}
	return feed, nil
}

func parseFeedFromResponse(feedResponse feedResponse, feed *Feed) error {
	feed.ID = feedResponse.ID
	feed.Username = feedResponse.Username
	feed.Biography = feedResponse.Biography
	feed.ProfilePictureUrl = feedResponse.ProfilePicURL
	feed.Website = feedResponse.Website
	feed.FollowersCount = feedResponse.FollowersCount
	feed.FollowsCount = feedResponse.FollowsCount
	feed.Posts = make([]Post, 0)
	for _, post := range feedResponse.Posts {
		layout := "2006-01-02T15:04:05Z0700" // correct time format for +0000 timezone

		parsedTime, err := time.Parse(layout, post.TimestampString)
		if err != nil {
			fmt.Println("Error parsing time:", err)
			return fmt.Errorf("error parsing time: %w", err)
		}

		fmt.Println("Parsed time:", parsedTime)
		feed.Posts = append(feed.Posts, Post{
			ID:                    post.ID,
			feedID:                feed.ID,
			Permalink:             post.Permalink,
			Timestamp:             parsedTime,
			MediaType:             post.MediaType,
			mediaSmallExternalURL: post.Sizes.Small.MediaURL,
			MediaSmallHeight:      post.Sizes.Small.Height,
			MediaSmallWidth:       post.Sizes.Small.Width,
			Caption:               post.Caption,
			PrunedCaption:         post.PrunedCaption,
		})
	}
	return nil
}

type feedResponse struct {
	ID             string
	Username       string         `json:"username"`
	Biography      string         `json:"biography"`
	ProfilePicURL  string         `json:"profilePictureUrl"`
	Website        string         `json:"website"`
	FollowersCount int            `json:"followersCount"`
	FollowsCount   int            `json:"followsCount"`
	Posts          []postResponse `json:"posts"`
}

type postResponse struct {
	ID              string        `json:"id"`
	TimestampString string        `json:"timestamp"`
	Permalink       string        `json:"permalink"`
	MediaType       string        `json:"mediaType"`
	Sizes           sizesResponse `json:"sizes"`
	Caption         string        `json:"caption"`
	PrunedCaption   string        `json:"prunedCaption"`
}

type sizesResponse struct {
	Small  smallResponse  `json:"small"`
	Medium mediumResponse `json:"medium"`
	Large  largeResponse  `json:"large"`
}

type smallResponse struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MediaURL string `json:"mediaUrl"`
}

type mediumResponse struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MediaURL string `json:"mediaUrl"`
}

type largeResponse struct {
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	MediaURL string `json:"mediaUrl"`
}
