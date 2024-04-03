package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const uri = "mongodb://admin:password@localhost:27017"

type Entry struct {
	Subreddit  string
	Title      string
	PostId     string
	ThreadText string
	Response   string
	Rating     int
}

func getCollection() *mongo.Collection {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}

	/*defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()*/

	db := client.Database("punthreads")
	threadsCollection := db.Collection("punthreads")
	return threadsCollection

}

func WriteThreadAndResult(entry Entry) {
	threadsCollection := getCollection()
	_, err := threadsCollection.InsertOne(context.TODO(), bson.D{
		{Key: "subreddit", Value: entry.Subreddit},
		{Key: "title", Value: entry.Title},
		{Key: "postId", Value: entry.PostId},
		{Key: "threadText", Value: entry.ThreadText},
		{Key: "response", Value: entry.Response},
		{Key: "rating", Value: entry.Rating},
	})

	if err != nil {
		panic(err)
	}
}

func GetThreadByText(threadText string) (Entry, error) {

	threadsCollection := getCollection()

	filter := bson.D{{Key: "threadText", Value: threadText}}
	// filter = bson.D{}

	cursor, err := threadsCollection.Find(context.TODO(), filter, nil)
	if err != nil {
		panic(err)
	}
	var results []Entry
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}
	fmt.Println(results)
	if len(results) == 0 {
		return Entry{}, fmt.Errorf("thread with text %q not found", threadText)
	}

	return results[0], nil
}

/*
func main_mongo() {

	//
		entryExample1 := Entry{
			Title:      "faketitle1",
			PostId:     "0001",
			ThreadText: "a thread text",
			Response:   "a fake response",
			Rating:     -1,
		}
		writeThreadAndResult(entryExample1)
	//

	fmt.Println(fetchThreadByText("a thread text"))

}
*/
