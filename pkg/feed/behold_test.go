package feed

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type MockHTTPClient struct {
	Response *http.Response
	Error    error
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.Response, m.Error
}

func TestGetFeed(t *testing.T) {
	sampleTime := "2025-01-29T18:34:09+0000"
	sampleResponse := feedResponse{
		ID:       "123",
		Username: "test account name",
		Posts: []postResponse{
			{ID: "post1", TimestampString: sampleTime, Permalink: "link", MediaType: "photo"},
		},
	}

	responseBody, _ := json.Marshal(sampleResponse)

	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(responseBody)),
		},
		Error: nil,
	}

	feedResponse, err := fetchFeedResponse(mockClient, "123")
	if err != nil {
		t.Errorf("fetchFeedResponse returned an error: %v", err)
	}

	if feedResponse.ID != "123" {
		t.Errorf("Expected feed ID to be 123, got %s", feedResponse.ID)
	}
	if len(feedResponse.Posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(feedResponse.Posts))
	}

	feed := &Feed{}

	err = parseFeedFromResponse(feedResponse, feed)
	if err != nil {
		t.Errorf("parseFeedFromResponse returned an error: %v", err)
	}
	if feed.ID != "123" {
		t.Errorf("Expected feed ID to be 123, got %s", feed.ID)
	}
	if feed.Username != "test account name" {
		t.Errorf("Expected feed username to be %s, got %s", "test account name", feed.Username)
	}
	if len(feed.Posts) != 1 {
		t.Errorf("Expected 1 post, got %d", len(feed.Posts))
	}
	if feed.Posts[0].Timestamp.String() == "" {
		t.Errorf("Expected timestamp to be set, got %s", feed.Posts[0].Timestamp)
	}
}
