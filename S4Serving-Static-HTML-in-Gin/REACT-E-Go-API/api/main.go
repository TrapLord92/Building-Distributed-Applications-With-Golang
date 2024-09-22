// Recipes API
//
// This is a sample recipes API. You can find out more at https://github.com/TrapLord92/Building-Distributed-Applications-With-Golang
//
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact:
//
//      name: TrapLord
//      email: traplord345@gmail.com
//
// SecurityDefinitions:
//
//      api_key:
//        type: apiKey
//        name: Authorization
//        in: header
//
// Consumes:
//   - application/json
//
// Produces:
//   - application/json
//
// swagger:meta

package main

import (
	"context"
	"log"
	"os"
	handlers "react-go-api/api/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var authHandler *handlers.AuthHandler
var recipesHandler *handlers.RecipesHandler

func init() {
	ctx := context.Background()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to MongoDB")
	collectionRecipes := client.Database(os.Getenv("MONGO_DATABASE")).Collection("recipes")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	status := redisClient.Ping()
	log.Println(status)

	recipesHandler = handlers.NewRecipesHandler(ctx, collectionRecipes, redisClient)

	collectionUsers := client.Database(os.Getenv("MONGO_DATABASE")).Collection("users")
	authHandler = handlers.NewAuthHandler(ctx, collectionUsers)
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/recipes", recipesHandler.ListRecipesHandler)

	router.POST("/signin", authHandler.SignInHandler)
	router.POST("/refresh", authHandler.RefreshHandler)

	authorized := router.Group("/")
	authorized.Use(authHandler.AuthMiddleware())
	{
		authorized.POST("/recipes", recipesHandler.NewRecipeHandler)
		authorized.PUT("/recipes/:id", recipesHandler.UpdateRecipeHandler)
		authorized.DELETE("/recipes/:id", recipesHandler.DeleteRecipeHandler)
		authorized.GET("/recipes/:id", recipesHandler.GetOneRecipeHandler)
	}

	router.Run()
	// router.RunTLS(":443", "/home/workspace/certs/localhost.crt", "/home/workspace/certs/localhost.key")
}
