package main

import (
	"fmt"
	"os"

	"github.com/wittejm/punthreads/db"
	"github.com/wittejm/punthreads/scrape"
)

func launch() {

	if len(os.Args) == 1 {
		fmt.Println("commands: all, gather, rate, clear, review")
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
		postIds, err := scrape.GatherPostIds(subreddit, "all")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s", err)
			return
		}
		fmt.Println("len(postIds):", len(postIds))
		scrape.ConcurrentlyFetchPosts("all", postIds)

	} else if command == "rate" {
		ConcurrentlyWalkPostsAndRate(subreddit)
	} else if command == "clear" {
		scrape.ClearBadFiles("all")
	} else if command == "review" {
		db.Review()
	} else {
		fmt.Println("Unrecognized command")
	}
}
