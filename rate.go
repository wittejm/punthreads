package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/wittejm/punthreads/chatgpt"
	"github.com/wittejm/punthreads/db"
	"github.com/wittejm/punthreads/scrape"
)

func getRatingFromResponse(response string) (int, error) {
	pattern := regexp.MustCompile(`Pun: ([0-9]+)/10`)
	result := pattern.FindStringSubmatch(response)
	if len(result) > 1 {
		convResult, err := strconv.Atoi(result[1])
		if err != nil {
			return -1, err
		}
		return convResult, nil
	}
	return -1, errors.New("Expected text pattern not in GPT response")
}

func ConcurrentlyWalkPostsAndRate(subreddit string) error {
	fmt.Println("ConcurrentlyWalkPostsAndRate")
	minScore := 10
	postData, err := scrape.GatherSavedPosts(subreddit)
	if err != nil {
		return err
	}
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(postData))
	var tokens = make(chan struct{}, 20)
	for _, post := range postData {
		fmt.Println("postData")
		if post.Post.Data.Children == nil {
			fmt.Println("post.Post.Data.Children is nil")
			continue
		}
		time.Sleep(time.Millisecond * 200)
		tokens <- struct{}{}
		go func() {
			err := walkPostAndRate(subreddit, post, minScore, &waitGroup)
			if err != nil {
				panic(err) // in a multithreaded environment, let's failfast and kill everything while we are debugging errors.
			}
		}()
		<-tokens
	}
	fmt.Println("waiting")
	waitGroup.Wait()
	return nil
}

func walkPostAndRate(subreddit string, post scrape.PostAndCommentsContent, minScore int, waitGroup *sync.WaitGroup) error {
	fmt.Println("in walkPostAndRate")
	defer waitGroup.Done()
	title := post.Post.Data.Children[0].Data.Title
	postId := post.Post.Data.Children[0].Data.Name[3:]
	fmt.Println(post.Post.Data.Children[0].Data.Title, post.Post.Data.Children[0].Data.Name)
	commentThreads := scrape.CommentsToBestCommentThreads(post.Comments, minScore)
	for _, commentThread := range commentThreads {
		scrape.PrintComments([]scrape.Comment{commentThread}, 1)
		if commentThread.ThreadLength() > 5 {

			threadText := commentThread.ThreadToString()

			response, err := chatgpt.GetGPTResponse(threadText)
			if err != nil {
				return err
			}
			rating, err := getRatingFromResponse(response)
			if err != nil {
				return err
			}
			entry := db.Entry{
				Subreddit:  subreddit,
				Title:      title,
				PostId:     postId,
				ThreadText: threadText,
				Response:   response,
				Rating:     rating,
			}

			err = db.WriteThreadAndResult(entry)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
