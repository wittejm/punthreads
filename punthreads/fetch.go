package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func OnPage(url string, dataFile string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-agent", "threadgettbot0.01")
	res, err := client.Do(req)
	if err != nil {
		log.Fatal("err", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	check(err)

	f, err := os.Create(fmt.Sprintf("../data/%s", dataFile))
	check(err)

	n3, err := f.WriteString(string(body))
	check(err)
	fmt.Printf("wrote %d bytes\n", n3)

	return "done"

}

/*func main() {
	fmt.Println(OnPage("https://www.reddit.com/r/pics.json", "r.pics.0.json"))
}*/
