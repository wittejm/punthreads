package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}
type ResponseBody struct {
	Choices []Choice `json:"choices"`
}

func getGptResponse(threadText string) string {

	localFetchResult, err := fetchThreadByText(threadText)
	if err == nil {
		fmt.Println("found existing response")
		return localFetchResult.Response
	}

	fmt.Println("fetching from gpt")

	// look for the existing mongo entry with this text.
	// if it exists, return what's already there.

	OPENAI_API_KEY := os.Getenv("OPENAI_API_KEY")

	body := RequestBody{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You rate comment threads on how much they contain puns, other humorous content, or something else, each on a scale of 0 to 10, and you only respond in the format:\n`Pun: 2/10\nOther humor: 7/10\nOther: 1/10`\n This example thread contains puns: `2242 You can tell this guy does push UPS\n  1069 Definitely an Alpha Mail\n   229 What can Brown do for you today?\n   117 One of the worst slogans ever. Upvoted!\n   125 Wonder if he's got a nice package.` It would get a response of `Pun: 10/10\nOther humor: 0/10\nOther: 0/10`. This next thread contains one pun but mostly other humor. `21543 You could see her thought process after she knocked down her sister. She just thought “fuck it” and decided to take it all.\n 13067 The bloodlust had taken her\n  3448 BLOOD FOR THE BLOOD GOD!\n  1702 SKULLS FOR THE SKULL THRONE!\n   1949 MILK FOR THE KHORNE FLAKES!` It would get a response of: `Pun: 3/10\nOther humor: 8/10\nOther: 0/10` This thread is mostly puns: `7529 That bin definitely flipped its lid\n 1472 Almost lost its top\n 894 r/trashy \n   518 *In Mother Russia, can kicks you*\n    402 In mother Russia, trash takes you out\n 67 *In Mother Russia you don't have pet cat, cat has pet human*` It would get a response of `Pun: 8/10\nOther humor: 2/10\nOther: 0/10`. This thread is not puns nor particularly humorous: `14661 Everyone assuming shes an ex i thought she was like a sibling or family member\n 12032 she is a sibling yeah \n  3299 whats with the face then?\n  8436 teenagers, what can you say... I think she was kinda bored\n 6515 She hated you taking her picture. I bet she’s gunna love that you’ve turned her into a meme lol.` It would get a response of `Pun: 0/10\nOther humor: 2/10\nOther: 8/10`",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Rate this comment thread:\n%s", threadText),
			},
		},
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyJSON))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", OPENAI_API_KEY))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	var result ResponseBody

	if err := json.Unmarshal(resBody, &result); err != nil {
		panic(err)
	}

	if len(result.Choices) < 1 {
		return string(resBody)
	}
	return string(result.Choices[0].Message.Content)
}
