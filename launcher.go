package main

import (
	"fmt"
	"os"

	"github.com/wittejm/punthreads/scrape"
)

func main() {

	if len(os.Args) == 1 {
		fmt.Println("commands: gather, rate")
		return
	}

	command := os.Args[1]
	var subreddit string
	if len(os.Args) > 2 {
		subreddit = os.Args[2]
	} else {
		subreddit = "all"
	}

	if command == "gather" {
		scrape.GatherPosts(subreddit, "all")
	} else if command == "rate" {
		WalkPostsAndRate(subreddit)
	}
}
