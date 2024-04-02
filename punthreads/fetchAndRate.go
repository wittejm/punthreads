package main

import (
	"fmt"
	"regexp"
	"strconv"
)

func (c Comment) threadLength() int {
	if c.Replies == nil || len(c.Replies) == 0 {
		return 1
	} else {
		return 1 + c.Replies[0].threadLength()
	}
}

func (c Comment) threadToString() string {
	var repliesString string
	if c.Replies == nil || len(c.Replies) == 0 {
		repliesString = ""
	} else {
		repliesString = c.Replies[0].threadToString()
	}
	result := fmt.Sprintf("%d %s \n%s", c.Score, c.Text, repliesString)
	return result
}

func getRatingFromResponse(response string) int {
	pattern := regexp.MustCompile(`Pun: ([0-9]+)/10`)
	//fmt.Println(pattern.FindStringSubmatch("the result is\nPun: 2/10\nOther: 4/10")[1])
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

func WalkPostsAndRate(subreddit string) {
	minScore := 10
	postData := GatherSavedPosts(subreddit)

	for _, post := range postData {
		if post.Post.Data.Children == nil {
			continue
		}
		title := post.Post.Data.Children[0].Data.Title
		postId := post.Post.Data.Children[0].Data.Name[3:]
		fmt.Println(post.Post.Data.Children[0].Data.Title, post.Post.Data.Children[0].Data.Name)
		commentThreads := CommentsToBestCommentThreads(post.Comments, minScore)
		for _, commentThread := range commentThreads {
			PrintComments([]Comment{commentThread}, 1)
			if commentThread.threadLength() > 5 {

				threadText := commentThread.threadToString()

				response := getGptResponse(threadText)
				rating := getRatingFromResponse(response)
				entry := Entry{
					Subreddit:  subreddit,
					Title:      title,
					PostId:     postId,
					ThreadText: threadText,
					Response:   response,
					Rating:     rating,
				}

				fmt.Println(entry)
				writeThreadAndResult(entry)
				//if rating >= 7 {
				//scanner := bufio.NewScanner(os.Stdin)
				//scanner.Scan()
				//}
			}
		}
	}
}
