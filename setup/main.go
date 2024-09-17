package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/:name", CustomHandler)

	router.Run()
}

// custum handler
func CustomHandler(c *gin.Context) {
	name := c.Params.ByName("name")
	c.JSON(200, gin.H{
		"message": "Hello, " + name + "! Welcome to my API!",
	})
}
