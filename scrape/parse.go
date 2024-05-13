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
	Data SubredditContentData `json:"data"`
}

type SubredditContentData struct {
	Children []SubredditPostChild `json:"children"`
}

type SubredditPostChild struct {
	SubredditPostChildData `json:"data"`
}

type SubredditPostChildData struct {
	Title string `json:"title"`
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type PostAndCommentsContent struct {
	Post     PostContent
	Comments CommentsContent
}

type PostContent struct {
	Kind string   `json:"kind"`
	Data PostData `json:"data"`
}

type PostData struct {
	Modhash  string      `json:"modhash"`
	Children []PostChild `json:"children"`
}

type PostChild struct {
	Kind string        `json:"kind"`
	Data PostChildData `json:"data"`
}

type PostChildData struct {
	Title string `json:"title"`
	Name  string `json:"name"`
}

type CommentsContent struct {
	Kind string       `json:"kind"`
	Data CommentsData `json:"data"`
}

type CommentsData struct {
	Children []CommentChild `json:"children"`
}
type CommentChild struct {
	Kind string           `json:"kind"`
	Data CommentChildData `json:"data"`
}

type CommentChildData struct {
	Body    string          `json:"body"`
	Score   int             `json:"score"`
	Replies CommentsContent `json:"replies"`
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
	if comment.Replies == nil {
		replies = []Comment{}
	} else {
		repliesResult := getBestShapedThread(comment.Replies[0], minScore)
		if repliesResult == nil {
			replies = []Comment{}

		} else {
			replies = []Comment{*repliesResult}
		}
	}

	result := Comment{

		Score:   comment.Score,
		Text:    comment.Text,
		Replies: replies,
	}

	return &result
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
	v.Add("limit", "5")
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
	body, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data SubredditContent
	err = json.Unmarshal(body, &data)

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
		req.Header.Set("User-agent", "threadgettbot0.0.1")
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		fmt.Println(res.Status)
		if res.StatusCode != 200 {
			err = fmt.Errorf("bad status: %s", res.Status)
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

	err = json.Unmarshal(body, &data)
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
			lastPostId = nextPageContent.Data.Children[len(nextPageContent.Data.Children)-1].Name
		}
		counter = counter + 1
		return nextPageContent, err
	}
}
