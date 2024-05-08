package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/wittejm/punthreads/chatgpt"
	"github.com/wittejm/punthreads/db"
	"github.com/wittejm/punthreads/scrape"
)

func getRatingFromResponse(response string) int {
	pattern := regexp.MustCompile(`Pun: ([0-9]+)/10`)
	result := pattern.FindStringSubmatch(response)
	if len(result) > 1 {
		convResult, err := strconv.Atoi(result[1])
		if err != nil {
			panic(err)
		}
		return convResult
	}
	return -1
}

func ConcurrentlyWalkPostsAndRate(subreddit string) {
	minScore := 10
	postData := scrape.GatherSavedPosts(subreddit)

	for _, post := range postData {
		if post.Post.Data.Children == nil {
			continue
		}
		time.Sleep(time.Millisecond * 200)
		go WalkPostAndRate(subreddit, post, minScore)
	}
}

func WalkPostAndRate(subreddit string, post scrape.PostAndCommentsContent, minScore int) {

	title := post.Post.Data.Children[0].Data.Title
	postId := post.Post.Data.Children[0].Data.Name[3:]
	fmt.Println(post.Post.Data.Children[0].Data.Title, post.Post.Data.Children[0].Data.Name)
	commentThreads := scrape.CommentsToBestCommentThreads(post.Comments, minScore)
	for _, commentThread := range commentThreads {
		scrape.PrintComments([]scrape.Comment{commentThread}, 1)
		if commentThread.ThreadLength() > 5 {

			threadText := commentThread.ThreadToString()

			response := chatgpt.GetGptResponse(threadText)
			rating := getRatingFromResponse(response)
			entry := db.Entry{
				Subreddit:  subreddit,
				Title:      title,
				PostId:     postId,
				ThreadText: threadText,
				Response:   response,
				Rating:     rating,
			}

			db.WriteThreadAndResult(entry)

			if rating >= 1 {
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
			}

		}
	}
}
