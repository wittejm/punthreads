/*
Get the subreddit page, then get all the posts in the result.
*/
package main

import (
	//"bufio"
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Comment struct {
	Score   int
	Text    string
	Replies []Comment
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

func count(comments []Comment) int {
	i := 0
	for _, c := range comments {
		i += 1 + count(c.Replies)
	}
	return i
}

func PrintComments(comments []Comment, indent int) {
	for _, c := range comments {

		fmt.Printf("%s %d %5d %s\n", strings.Repeat("  ", indent), c.threadLength(), c.Score, c.Text)
		PrintComments(c.Replies, indent+1)
	}
}

func LocatePunThread(comments []Comment) []string {
	var result []string

	return result
}

func GatherPosts(subreddit string, period string) []PostAndCommentsContent {
	var allPostData []PostAndCommentsContent
	generator := PageGenerator(subreddit, period)

	for i := 0; i < 100; i++ {
		subredditContent := generator()
		for _, c := range subredditContent.Data.Children[2:] {
			postId := c.Name[3:]
			postData := LoadOrFetchPost(subreddit, postId, "", 0)
			allPostData = append(allPostData, postData)

		}
	}
	return allPostData
}

func getPostFilenames() []string {
	dataContents, err := os.ReadDir("../data")
	if err != nil {
		panic(err)
	}
	var filenames []string

	fmt.Println(dataContents)
	for ind, item := range dataContents {
		fmt.Printf("%d %T %s\n", ind, item, item)
		var val = item.Name()
		filenames = append(filenames, val)
	}
	return filenames
}
func GatherSavedPosts(subreddit string) []PostAndCommentsContent {
	var allPostData []PostAndCommentsContent

	filenames := getPostFilenames()
	for _, filename := range filenames {
		if strings.Contains(filename, fmt.Sprintf("r.%s", subreddit)) || !strings.Contains(filename, subreddit) {
			continue
		}
		body, err := os.ReadFile(fmt.Sprintf("../data/%s", filename))
		if err != nil {
			panic(err)
		}
		//scanner := bufio.NewScanner(os.Stdin)
		//scanner.Scan()
		var data PostAndCommentsContent
		json.Unmarshal(body, &data)

		allPostData = append(allPostData, data)
	}

	return allPostData
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

		//scanner := bufio.NewScanner(os.Stdin)
		//scanner.Scan()
		// PrintComments(cleanedComments, 1)
		// PrintComments(cleanedComments, 1)
		resultCommentThreads = append(resultCommentThreads, *resultThread)
	}

	// PrintComments(resultCommentThreads, 1)

	return resultCommentThreads
}

/*
func main_gather() {
	GatherPosts("holdmybeer", "all")
	//getGptRating()

}
*/
