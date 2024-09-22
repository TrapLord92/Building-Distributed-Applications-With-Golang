package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Recipe struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
	Ingredients []string  `json:"ingredients"`
	PublishedAt time.Time `json:"published_at"`
}

var recipes []Recipe
var ctx context.Context
var client *mongo.Client
var err error

func init() {
	recipes = make([]Recipe, 0)

	// Error handling when reading file
	file, err := os.ReadFile("recipes.json")
	if err != nil {
		log.Fatal("Error reading recipes.json: ", err)
	}
	if err := json.Unmarshal(file, &recipes); err != nil {
		log.Fatal("Error unmarshalling recipes.json: ", err)
	}

	ctx = context.Background()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err != nil {
		log.Fatal("MongoDB connection failed: ", err)
	}

	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal("MongoDB ping failed: ", err)
	}
	log.Println("Connected to MongoDB")

	// Only insert if there are recipes to insert
	if len(recipes) > 0 {
		var listOfRecipes []interface{}
		for _, recipe := range recipes {
			listOfRecipes = append(listOfRecipes, recipe)
		}
		collection := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")
		insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Inserted recipes: ", len(insertManyResult.InsertedIDs))
	} else {
		log.Println("No recipes to insert.")
	}
}

func NewRecipeHandler(c *gin.Context) {
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	c.JSON(http.StatusOK, recipe)
}

func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Retain the ID
	recipe.ID = id
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}
	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipe not found"})
		return
	}
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)
	for _, recipe := range recipes {
		for _, t := range recipe.Tags {
			if strings.EqualFold(t, tag) {
				listOfRecipes = append(listOfRecipes, recipe)
				break
			}
		}
	}
	c.JSON(http.StatusOK, listOfRecipes)
}

func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)

	router.Run(":8080")
}

//DataBase and go run :
/*MONGO_URI="mongodb://admin:password@localhost:27017/test?authSource=admin" MONGO_DATABASE=demo go run main.go*/
