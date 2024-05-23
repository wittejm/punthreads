/*
Package scrape has functions to get the subreddit page, then get all the posts in the result.

Do I really need to lead package comments with "Package scrape"? VSCode tells me to and it seems unnecessary.
*/
package scrape

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"sync"
	"time"
)

type Comment struct {
	Score   int
	Text    string
	Replies []Comment // I wanted to replace this with an array of pointers following Chris's feedback, but my call to slices.SortFunc(comment.Replies) in parse.go seems to need structs, not pointers. I don't know how to reconcile those two things.
}

func (c Comment) ThreadLength() int {
	if c.Replies == nil || len(c.Replies) == 0 {
		return 1
	} else {
		return 1 + c.Replies[0].ThreadLength()
	}
}

func (c Comment) ThreadToString() string {
	var repliesString string
	if len(c.Replies) != 0 {
		repliesString = c.Replies[0].ThreadToString()
	}
	result := fmt.Sprintf("%d %s \n%s", c.Score, c.Text, repliesString)
	return result
}

func commentsContentToBareComments(commentContent CommentsContent) []Comment {
	var comments []Comment
	for _, c := range commentContent.Data.Children {
		comment := c.Data.Body
		score := c.Data.Score
		replies := commentsContentToBareComments(c.Data.Replies)
		comments = append(comments,
			Comment{
				Score:   score,
				Text:    comment,
				Replies: replies,
			})
	}
	return comments
}

func PrintComments(comments []Comment, indent int) {
	for _, c := range comments {

		fmt.Printf("%s %d %5d %s\n", strings.Repeat("  ", indent), c.ThreadLength(), c.Score, c.Text)
		PrintComments(c.Replies, indent+1)
	}
}

func GatherPostIds(subreddit string, period string) ([]string, error) {
	var allPostIds []string
	generator := pageGenerator(subreddit, period)

	for i := 0; i < 2; i++ {
		subredditContent, err := generator()
		if err != nil {
			return nil, err
		}
		if len(subredditContent.Data.Children) < 2 {
			break
		}
		for _, c := range subredditContent.Data.Children[2:] {
			postId := c.Data.Name[3:]
			allPostIds = append(allPostIds, postId)
		}
	}
	return allPostIds, nil
}

func ConcurrentlyFetchPosts(subreddit string, postIds []string) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(postIds))
	var tokens = make(chan struct{}, 20)

	for i, postId := range postIds {
		fmt.Println("Fetching post:", i, postId)
		time.Sleep(time.Millisecond * 100)
		tokens <- struct{}{}
		go func() {
			_, err := LoadOrFetchPost(subreddit, postId, "", 0, &waitGroup)
			if err != nil {
				panic(err) // in a multithreaded environment, let's failfast and kill everything while we are debugging errors.
			}
			<-tokens
		}()

		// TODO: This code runs, but I don't think the concurrency is currect. It doesn't seem to run any faster, or wait for all threads to finish.

	}
	waitGroup.Wait()
}

func getPostFilenames() []string {
	dataContents, err := os.ReadDir("./data")
	if err != nil {
		panic(err)
	}
	var filenames []string

	for ind, item := range dataContents {
		fmt.Printf("%d %T %s\n", ind, item, item)
		var val = item.Name()
		filenames = append(filenames, val)
	}
	return filenames
}
func GatherSavedPosts(subreddit string) ([]PostAndCommentsContent, error) {
	var allPostData []PostAndCommentsContent

	filenames := getPostFilenames()
	var err error
	for _, filename := range filenames {
		// Skip subreddit pages
		if strings.Contains(filename, fmt.Sprintf("r.%s", subreddit)) || !strings.Contains(filename, subreddit) {
			continue
		}
		body, err := os.ReadFile(fmt.Sprintf("./data/%s", filename))
		if err != nil {
			return nil, err
		}
		var data PostAndCommentsContent
		err = json.Unmarshal(body, &data)
		// if err != nil {
		// 	return nil, err
		// }
		allPostData = append(allPostData, data)
	}

	return allPostData, err
}

func CommentsToBestCommentThreads(comments CommentsContent, minScore int) []Comment {

	var resultCommentThreads []Comment
	cleanedComments := commentsContentToBareComments(comments)

	lenCmp := func(a, b Comment) int {
		return cmp.Compare(b.Score, a.Score)
	}

	slices.SortFunc(cleanedComments, lenCmp)
	if len(cleanedComments) == 0 {
		return nil
	}
	for topIndex := 0; topIndex < min(len(cleanedComments), 5); topIndex++ {
		resultThread := getBestShapedThread(cleanedComments[topIndex], minScore)
		if resultThread == nil || (*resultThread).Score < minScore {
			continue
		}

		resultCommentThreads = append(resultCommentThreads, *resultThread)
	}

	return resultCommentThreads
}
