package main_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"

	"github.com/lattots/bhproxy/pkg/feed"
)

const testBhproxyUrl = "http://localhost:8080/cgi-bin/bhproxy?id=JYK0bzSTZPConDbzq1GL"
const testFilesPath = "../test-data/%d"
const destinationPath = "../webroot/docs"

var originalFeedFilepath = "../webroot/docs/JYK0bzSTZPConDbzq1GL"

var testFilesToRemove = []string{
	originalFeedFilepath,
	"../webroot/docs/LwSatAQuhZR5y9aDE3dIDATQLKH2",
}

type BhFeed struct {
	Username          string   `json:"username"`
	Biography         string   `json:"biography"`
	ProfilePictureURL string   `json:"profilePictureUrl"`
	Website           string   `json:"website"`
	FollowersCount    int      `json:"followersCount"`
	FollowsCount      int      `json:"followsCount"`
	Posts             []BhPost `json:"posts"`
}

type BhPost struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Permalink string `json:"permalink"`
	MediaType string `json:"mediaType"`
	MediaURL  string `json:"mediaUrl"`
	Sizes     struct {
		Small struct {
			MediaURL string `json:"mediaUrl"`
			Height   int    `json:"height"`
			Width    int    `json:"width"`
		} `json:"small"`
		Medium struct {
			MediaURL string `json:"mediaUrl"`
			Height   int    `json:"height"`
			Width    int    `json:"width"`
		} `json:"medium"`
		Large struct {
			MediaURL string `json:"mediaUrl"`
			Height   int    `json:"height"`
			Width    int    `json:"width"`
		} `json:"large"`
		Full struct {
			MediaURL string `json:"mediaUrl"`
			Height   int    `json:"height"`
			Width    int    `json:"width"`
		} `json:"full"`
	} `json:"sizes"`
	Caption       string `json:"caption"`
	PrunedCaption string `json:"prunedCaption"`
	ColorPalette  struct {
		Dominant     string `json:"dominant"`
		Muted        string `json:"muted"`
		MutedLight   string `json:"mutedLight"`
		MutedDark    string `json:"mutedDark"`
		Vibrant      string `json:"vibrant"`
		VibrantLight string `json:"vibrantLight"`
		VibrantDark  string `json:"vibrantDark"`
	} `json:"colorPalette"`
}

func prepareTestFiles(t *testing.T, scene int) {
	for _, thisFileToRemove := range testFilesToRemove {
		os.RemoveAll(thisFileToRemove)
	}

	sourcePath := fmt.Sprintf(testFilesPath, scene)
	err := copy.Copy(sourcePath, destinationPath)
	if err != nil {
		t.Errorf("Copying test files recursively returns error: %v", err)
	}
}

func executeRequest(t *testing.T, url string) []byte {
	if url == "" {
		var emptyBody []byte
		t.Errorf("executeRequest called with empty url")
		return emptyBody
	}

	res, err := http.Get(url)
	if err != nil {
		t.Errorf("HTTP GET returns error: %v", err)
	}

	assert.Equal(t, 200, res.StatusCode, "HTTP GET response code is not 200, but instead %d", res.StatusCode)

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("HTTP GET got error reading response body: %v", err)
	}

	return body
}

func testFeedHeaders(t *testing.T, feed feed.Feed) {
	assert.Equal(t, "JYK0bzSTZPConDbzq1GL", feed.ID)
	assert.Equal(t, "foobar", feed.Username)
	assert.Equal(t, "This is foobar biography.", feed.Biography)
	assert.Equal(t, "https://foobar.com/", feed.Website)
	assert.Equal(t, 111, feed.FollowersCount)
	assert.Equal(t, 222, feed.FollowsCount)
}

func getExpectedBhFeed(expectedFeed *BhFeed) error {
	fileContentStr, err := os.ReadFile(originalFeedFilepath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile("(\".+)\\+\\d+\"")
	fileContentStrWithFixedDatetime := re.ReplaceAllString(string(fileContentStr), "${1}+00:00\"")
	err = json.Unmarshal([]byte(fileContentStrWithFixedDatetime), expectedFeed)

	if err != nil {
		return err
	}

	return nil
}

func getUrlResponseBodyHash(t *testing.T, url string) string {
	body := executeRequest(t, url)

	hash := sha256.New()
	hash.Write(body)
	hashBytes := hash.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)

	return hashString
}

func TestRequest(t *testing.T) {
	for numberOfPosts := 1; numberOfPosts <= 8; numberOfPosts++ {
		fmt.Printf("Testing: %d\n", numberOfPosts)
		prepareTestFiles(t, numberOfPosts)

		time.Sleep(2 * time.Second)

		responseBody := executeRequest(t, testBhproxyUrl)

		var retrievedFeed feed.Feed
		assert.Nil(t, json.Unmarshal(responseBody, &retrievedFeed))

		testFeedHeaders(t, retrievedFeed)

		assert.Equal(t, numberOfPosts, len(retrievedFeed.Posts))

		var expectedBhFeed BhFeed
		err := getExpectedBhFeed(&expectedBhFeed)
		assert.Nil(t, err)

		assert.Equal(t, len(expectedBhFeed.Posts), len(retrievedFeed.Posts))

		for postIndex := 0; postIndex < len(expectedBhFeed.Posts); postIndex++ {
			roundString := fmt.Sprintf("numberOfPosts: %d, postIndex: %d", numberOfPosts, postIndex)
			expectedPost := expectedBhFeed.Posts[postIndex]
			observedPost := retrievedFeed.Posts[postIndex]

			assert.Equal(t, expectedPost.ID, observedPost.ID, roundString)
			assert.Equal(t, expectedPost.Permalink, observedPost.Permalink, roundString)
			assert.Equal(t, expectedPost.Sizes.Small.Width, observedPost.MediaSmallWidth, roundString)
			assert.Equal(t, expectedPost.Sizes.Small.Height, observedPost.MediaSmallHeight, roundString)
			assert.Equal(t, expectedPost.Caption, observedPost.Caption, roundString)
			assert.Equal(t, expectedPost.PrunedCaption, observedPost.PrunedCaption, roundString)

			assert.NotEqual(t, "", expectedPost.Sizes.Small.MediaURL, roundString)
			assert.NotEqual(t, "", observedPost.MediaSmallUrl, roundString)

			expectedHash := getUrlResponseBodyHash(t, expectedPost.Sizes.Small.MediaURL)
			observedHash := getUrlResponseBodyHash(t, observedPost.MediaSmallUrl)

			assert.Equal(t, expectedHash, observedHash, roundString)
		}
	}
}
