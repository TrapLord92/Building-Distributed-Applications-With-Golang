package main

import (
	"encoding/xml"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/", CustomHandler)

	router.Run()
}

// Handling error

type TestingError struct {
	XMLName xml.Name `xml:"testingError"`
	Error   string   `xml:"error"`
}

func CustomHandler(c *gin.Context) {
	c.XML(400, TestingError{
		Error: "Page not found",
	})
}
