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

func getCollection() (*mongo.Collection, error) {
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(context.TODO(), opts) //This works with no context object, and I didn't look into what it's needed for.
	if err != nil {
		return nil, err
	}

	db := client.Database("punthreads")

	// Check if collection exists, create it if not
	collections, err := db.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	collectionExists := false
	for _, coll := range collections {
		if coll == "punthreads" {
			collectionExists = true
			break
		}
	}
	if !collectionExists {
		err := db.CreateCollection(context.Background(), "punthreads")
		if err != nil {
			panic(err)
		}
	}

	threadsCollection := db.Collection("punthreads")
	return threadsCollection, nil

}

func WriteThreadAndResult(entry Entry) error {
	threadsCollection, err := getCollection()
	if err != nil {
		return err
	}
	_, err = threadsCollection.InsertOne(context.TODO(), bson.D{
		{Key: "subreddit", Value: entry.Subreddit},
		{Key: "title", Value: entry.Title},
		{Key: "postId", Value: entry.PostId},
		{Key: "threadText", Value: entry.ThreadText},
		{Key: "response", Value: entry.Response},
		{Key: "rating", Value: entry.Rating},
	})

	return err
}

func GetThreadByText(threadText string) (*Entry, error) {

	threadsCollection, err := getCollection()

	filter := bson.D{{Key: "threadText", Value: threadText}}

	cursor, err := threadsCollection.Find(context.TODO(), filter, nil)
	if err != nil {
		return nil, err
	}
	var results []Entry
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("thread with text %q not found", threadText)
	}

	return &results[0], nil
}

func GetThreads() ([]Entry, error) {
	threadsCollection, err := getCollection()
	if err != nil {
		return nil, err
	}
	filter := bson.D{}

	cursor, err := threadsCollection.Find(context.TODO(), filter, nil)
	if err != nil {
		return nil, err
	}
	var results []Entry
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}

	return results, nil
}

func Review() error {
	entries, err := GetThreads()
	if err != nil {
		return err
	}
	i := 0
	for _, e := range entries {
		if e.Rating >= 8 {
			i++
			fmt.Println(i, ":", e.Rating, e.Subreddit, e.PostId, e.Title)
			fmt.Println(e.ThreadText)
		}
	}
	return nil
}
