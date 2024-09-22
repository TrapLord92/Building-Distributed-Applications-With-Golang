package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Feed struct {
	Entries []Entry `xml:"entry"`
}
type Entry struct {
	Link struct {
		Href string `xml:"href,attr"`
	} `xml:"link"`
	Thumbnail struct {
		URL string `xml:"url,attr"`
	} `xml:"thumbnail"`
	Title string `xml:"title"`
}

func GetFeedEntries(url string) ([]Entry, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	byteValue, _ := io.ReadAll(resp.Body)

	var feed Feed
	if err := xml.Unmarshal(byteValue, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse XML: %w ", err) // Proper error handling
	}
	return feed.Entries, nil
}

// init Mongo client
var client *mongo.Client
var ctx context.Context

func init() {
	ctx = context.Background()
	var err error
	client, err = mongo.Connect(ctx,
		options.Client().ApplyURI(os.Getenv("MONGO_URI2")))
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
}

func main() {
	// Close MongoDB connection when the app shuts down
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	router := gin.Default()
	router.POST("/parse", ParserHandler)
	router.Run(":5000")
}

type Request struct {
	URL string `json:"url"`
}

func ParserHandler(c *gin.Context) {
	var request Request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error()})
		return
	}

	entries, err := GetFeedEntries(request.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Error while parsing the RSS feed"})
		return
	}

	collection := client.Database(os.Getenv("MONGO_DATABASE2")).Collection("recipes_rss")

	// Prepare documents for bulk insert
	var docs []interface{}
	for _, entry := range entries {
		docs = append(docs, bson.M{
			"title":     entry.Title,
			"thumbnail": entry.Thumbnail.URL,
			"url":       entry.Link.Href,
		})
	}

	// Attempt to insert documents into MongoDB
	if len(docs) > 0 {
		_, err = collection.InsertMany(ctx, docs)
		if err != nil {
			// Log the actual MongoDB error for debugging purposes
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Error while saving entries to the database: %v", err),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "RSS feed processed successfully", "entries": entries})
}
