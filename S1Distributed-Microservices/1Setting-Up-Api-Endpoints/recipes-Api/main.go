package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

//Implementing HTTP routes

type Recipe struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
	Ingredients []string  `json:"ingredients"`
	PublishedAt time.Time `json:"published_at"`
}

// variables for mock data
var recipes []Recipe

func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}
func NewRecipe(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
}
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipe)
	router.GET("/recipes", ListRecipesHandler)

	router.Run(":8080")
}
