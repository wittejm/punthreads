package scrape

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/mitchellh/mapstructure"
)

// Top level of the JSON content from the subreddit page

type SubredditContent struct {
	Kind string
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
				Name  string `json:"name"`
				Score int    `json:"score"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type PostAndCommentsContent struct {
	Post struct {
		Kind string `json:"kind"`
		Data struct {
			Modhash  string `json:"modhash"`
			Children []struct {
				Kind string `json:"kind"`
				Data struct {
					Title string `json:"title"`
					Name  string `json:"name"`
				} `json:"data"`
			} `json:"children"`
		} `json:"data"`
	}
	Comments CommentsContent
}

type CommentsContent struct {
	Kind string `json:"kind"`
	Data struct {
		Children []struct {
			Kind string `json:"kind"`
			Data struct {
				Body    string          `json:"body"`
				Score   int             `json:"score"`
				Replies CommentsContent `json:"replies"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func (r *PostAndCommentsContent) UnmarshalJSON(p []byte) error {
	var tmp []interface{}
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}
	postContent, ok := tmp[0].(map[string]interface{})
	if !ok {
		return errors.New("failed to cast postContent raw")
	}
	commentsContent, ok := tmp[1].(map[string]interface{})
	if !ok {
		return errors.New("failed to cast commentsContent raw")
	}

	if err := mapstructure.Decode(postContent, &r.Post); err != nil {
		return err
	}
	if err := mapstructure.Decode(commentsContent, &r.Comments); err != nil {
		return err
	}
	return nil
}

func getBestShapedThread(comment Comment, minScore int) *Comment {

	if len(comment.Text) > 200 || strings.ContainsAny(comment.Text, "\n\r") || comment.Score < minScore || strings.Contains(comment.Text, "[removed]") || strings.Contains(comment.Text, "[deleted]") {
		return nil
	}

	lenCmp := func(a, b Comment) int {
		return cmp.Compare(b.Score, a.Score)
	}
	slices.SortFunc(comment.Replies, lenCmp)

	var replies []Comment
	if len(comment.Replies) != 0 {
		repliesResult := getBestShapedThread(comment.Replies[0], minScore)
		if repliesResult != nil {
			replies = []Comment{*repliesResult}
		}
	}

	return &Comment{

		Score:   comment.Score,
		Text:    comment.Text,
		Replies: replies,
	}
}

/*
Check if local file exists, and if so load and return the string content
otherwise do the request and save the resulting string into the local file, then return the string.
currently our parser loads the json content into a map data structure.
*/

func fileExists(name string) (bool, error) {
	_, err := os.Stat(name)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return false, err
	}
	return err == nil, nil
}

func loadOrFetchSubreddit(subreddit string, order string, pageNum int, after string) (*SubredditContent, error) {

	v := url.Values{}
	v.Add("limit", "20")
	if order != "" {
		v.Add("t", order)
	}
	if after != "" {
		v.Add("after", after)
	}

	var extensionString string
	if order == "" {
		extensionString = ".json"
	} else {
		extensionString = "top/.json"
	}

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s", subreddit, extensionString)
	fmt.Println(url)
	filename := fmt.Sprintf("./data/r.%s.%s.%d.json", subreddit, order, pageNum)

	exists, err := fileExists(filename)
	if err != nil {
		return nil, err
	}
	if !exists {
		fmt.Println("fetching from reddit")

		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.URL.RawQuery = v.Encode()
		req.Header.Set("User-agent", "threadgettbot0.0.0")
		res, err := client.Do(req)

		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		f, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		defer res.Body.Close()
		val, err := io.Copy(f, res.Body)
		if err != nil {
			return nil, err
		}

		fmt.Printf("wrote %d bytes\n", val)
	}
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data SubredditContent
	json.Unmarshal(body, &data) // eat the error here. This currently generates an error on incoming data because "replies" will be an empty string instead of a comment tree if there are no replies. The unmarshaller then just leaves the field as a zero value (no children). This runs as expected.

	return &data, nil
}

func LoadOrFetchPost(subreddit string, postId string, order string, offset int, waitGroup *sync.WaitGroup) (*PostAndCommentsContent, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/comments/%s.json", subreddit, postId)

	filename := fmt.Sprintf("./data/%s.%s.%s.%d.json", subreddit, postId, order, offset)
	var body []byte
	defer waitGroup.Done()
	_, err := os.ReadFile(filename)
	if err == nil {
		fmt.Println("read the file")
	} else {
		fmt.Println("fetching from reddit")

		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-agent", "threadgettbot0.0.2")
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusOK {
			err = fmt.Errorf("bad status: %s", res.Status)
			return nil, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		f, err := os.Create(filename)
		if err != nil {
			return nil, err
		}
		n3, err := f.Write(body)
		if err != nil {
			return nil, err
		}
		fmt.Printf("wrote %d bytes\n", n3)
	}

	body, err = os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var data PostAndCommentsContent

	json.Unmarshal(body, &data) // Eat the error, as above in the load Subreddit code.
	return &data, nil
}

func pageGenerator(subreddit string, order string) func() (*SubredditContent, error) {
	lastPostId := ""
	counter := 1
	return func() (*SubredditContent, error) {
		nextPageContent, err := loadOrFetchSubreddit(subreddit, order, counter, lastPostId)
		if err != nil {
			return nil, err
		}
		if len(nextPageContent.Data.Children) == 0 {
			lastPostId = ""
		} else {
			lastPostId = nextPageContent.Data.Children[len(nextPageContent.Data.Children)-1].Data.Name
		}
		counter = counter + 1
		return nextPageContent, err
	}
}
