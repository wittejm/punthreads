package scrape

import (
	"fmt"
	"os"
	"strings"
)

func ClearBadFiles(subreddit string) {
	filenames := getPostFilenames()

	for _, filename := range filenames {
		if strings.Contains(filename, fmt.Sprintf("r.%s", subreddit)) || !strings.Contains(filename, subreddit) {
			continue
		}
		body, err := os.ReadFile(fmt.Sprintf("./data/%s", filename))
		if err != nil {
			panic(err)
		}

		bodyStr := string(body)
		if strings.Contains(bodyStr, "<title>Too Many Requests</title>") {
			os.Remove(fmt.Sprintf("./data/%s", filename))
		}

	}
}
