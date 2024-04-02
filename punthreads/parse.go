package main

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"

	"github.com/mitchellh/mapstructure"
)

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

func check(e error) {
	if e != nil {
		panic(e)
	}
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

func LoadOrFetchSubreddit(subreddit string, order string, pageNum int, after string) SubredditContent {

	// Check if local file exists, and if so load and return the string content
	// otherwise do the request and save the resulting string into the local file, then return the string.
	// currently our parser loads the json content into a map data structure.
	// this approach passes
	var orderAndExtensionString string
	if order == "" {
		orderAndExtensionString = ".json?"
	} else {
		orderAndExtensionString = fmt.Sprintf("top/.json?t=%s", order)
	}

	var paginationString string
	if after == "" {
		paginationString = "&limit=100"
	} else {
		paginationString = fmt.Sprintf("&after=%s&limit=100", after)
	}

	url := fmt.Sprintf("https://www.reddit.com/r/%s/%s%s", subreddit, orderAndExtensionString, paginationString)
	fmt.Println(url)
	filename := fmt.Sprintf("../data/r.%s.%s.%d.json", subreddit, order, pageNum)

	var body []byte
	_, err := os.ReadFile(filename)
	if err == nil {
		fmt.Printf("read the file: %s\n", filename)
	} else {
		fmt.Println("fetching from reddit")

		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-agent", "threadgettbot0.0.0")
		res, err := client.Do(req)

		check(err)
		defer res.Body.Close() // Close the response body when we're done with it

		body, err := io.ReadAll(res.Body)
		check(err)

		f, err := os.Create(filename)
		check(err)

		n3, err := f.Write(body)
		check(err)
		fmt.Printf("wrote %d bytes\n", n3)
	}
	body, err = os.ReadFile(filename)
	check(err)
	var data SubredditContent
	json.Unmarshal(body, &data)

	fmt.Println("pretty print:")
	var printabledata map[string]interface{}
	jsonData, _ := json.Marshal(data)
	json.Unmarshal(jsonData, &printabledata)
	// PrintJSON(printabledata, 0)
	return data
}
func LoadOrFetchPost(subreddit string, postId string, order string, offset int) PostAndCommentsContent {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/comments/%s.json", subreddit, postId)

	filename := fmt.Sprintf("../data/%s.%s.%s.%d.json", subreddit, postId, order, offset)
	var body []byte
	_, err := os.ReadFile(filename)
	if err == nil {
		fmt.Println("read the file")
	} else {
		fmt.Println("fetching from reddit")

		client := &http.Client{}
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("User-agent", "threadgettbot0.0.0")
		res, err := client.Do(req)

		check(err)
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		check(err)

		f, err := os.Create(filename)
		check(err)

		n3, err := f.Write(body)
		check(err)
		fmt.Printf("wrote %d bytes\n", n3)
	}

	body, err = os.ReadFile(filename)
	check(err)

	var data PostAndCommentsContent
	json.Unmarshal(body, &data)

	return data
}

func PrintCommentsTree(comments CommentsData, indent int) {

}

func PrintJSON(data map[string]interface{}, indents int) {
	for k, v := range data {
		fmt.Println(strings.Repeat("  ", indents), "k:", k)

		if vdata, ok := v.([]interface{}); ok {
			for _, item := range vdata {
				if vitem, ok := item.(map[string]interface{}); ok {
					PrintJSON(vitem, indents+1)
				}
			}
		} else if vdata, ok := v.(map[string]interface{}); ok {
			PrintJSON(vdata, indents+1)
		} else {
			fmt.Printf("%s T of v: %T v: %s v: %d\n", strings.Repeat("  ", indents), v, v, v)
		}
	}
}

/*
	func main2() {
		/*
			pageOneContent := LoadOrFetchSubreddit("funny", "", 0, "")
			var titles []string
			for i, p := range pageOneContent.Data.Children {
				fmt.Println(i, p.Title)
				titles = append(titles, p.Title)
			}
			lastPostId := pageOneContent.Data.Children[len(pageOneContent.Data.Children)-1].Name
			pageTwoContent := LoadOrFetchSubreddit("funny", "", 1, lastPostId)
			for i, p := range pageTwoContent.Data.Children {
				fmt.Println(i, p.Title)
				titles = append(titles, p.Title)
			}
		//
		generator := PageGenerator("funny", "day")
		//var titles []string
		for i := 0; i < 5; i++ {
			nextPage := generator()

			for j, p := range nextPage.Data.Children {
				fmt.Println(j, p.Title)
			}
		}

}
*/
func PageGenerator(subreddit string, order string) func() SubredditContent {
	lastPostId := ""
	var nextPageContent SubredditContent
	counter := 1
	return func() SubredditContent {
		//for { this doesn't need to loop.
		nextPageContent = LoadOrFetchSubreddit(subreddit, order, counter, lastPostId)

		if len(nextPageContent.Data.Children) == 0 {
			lastPostId = ""
		} else {
			lastPostId = nextPageContent.Data.Children[len(nextPageContent.Data.Children)-1].Name
		}
		counter = counter + 1
		return nextPageContent
		//}
	}
}
