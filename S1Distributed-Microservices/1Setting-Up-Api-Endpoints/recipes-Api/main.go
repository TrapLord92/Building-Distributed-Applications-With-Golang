package main

import "github.com/gin-gonic/gin"

//defining data model

type Recipe struct {
	Name        string   `json:"name"`
	Tags        []string `json:"tags"`
	Ingredients []string `json:"ingredients"`
	PublisedAt  string   `json:"published_at"`
}

func main() {
	router := gin.Default()

	router.Run(":8080")
}
