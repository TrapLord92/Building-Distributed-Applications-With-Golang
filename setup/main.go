package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	router.GET("/", CustomHandler)

	router.Run()
}

// custum handler
func CustomHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "setup done you are ready to go",
	})
}
