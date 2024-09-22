package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/TrapLord92/Building-Distributed-Applications-With-Golang/S3Building-API-Authentication/ApiToken-recipes-api/models"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/net/context"
)

type RecipesHandler struct {
	collection  *mongo.Collection
	ctx         context.Context
	redisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		collection:  collection,
		ctx:         ctx,
		redisClient: redisClient,
	}
}

// swagger:route GET /recipes recipes listRecipes
// Returns a list of all recipes.
//
// Responses:
//   200: ListOfRecipesResponse
//   500: InternalServerErrorResponse
//
// Produces:
// - application/json
//
// Summary: Retrieve a list of recipes
// Description: Retrieves a list of all recipes, either from Redis or MongoDB.

func (handler *RecipesHandler) ListRecipesHandler(c *gin.Context) {
	val, err := handler.redisClient.Get("recipes").Result()
	if err == redis.Nil {
		log.Printf("Request to MongoDB")
		cur, err := handler.collection.Find(handler.ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(handler.ctx)

		recipes := make([]models.Recipe, 0)
		for cur.Next(handler.ctx) {
			var recipe models.Recipe
			cur.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		data, _ := json.Marshal(recipes)
		handler.redisClient.Set("recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	} else {
		log.Printf("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// swagger:route POST /recipes recipes newRecipe
// Create a new recipe.
//
// Responses:
//
//	200: RecipeCreatedResponse
//	400: BadRequestResponse
//	500: InternalServerErrorResponse
//
// Consumes:
// - application/json
// Produces:
// - application/json
//
// Summary: Create a new recipe
// Description: Adds a new recipe to the collection.
func (handler *RecipesHandler) NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.collection.InsertOne(handler.ctx, recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while inserting a new recipe"})
		return
	}

	log.Println("Remove data from Redis")
	handler.redisClient.Del("recipes")

	c.JSON(http.StatusOK, recipe)
}

// swagger:route PUT /recipes/{id} recipes updateRecipe
// Update an existing recipe.
//
// Updates the recipe with the given ID and the details provided in the request body.
//
// Responses:
//   200: RecipeUpdatedResponse
//   400: BadRequestResponse
//   404: NotFoundResponse
//   500: InternalServerErrorResponse
//
// Consumes:
// - application/json
// Produces:
// - application/json
//
// Parameters:
//   + name: id
//     in: path
//     required: true
//     type: string
//     description: The ID of the recipe to update
//
//   + name: recipe
//     in: body
//     required: true
//     description: Recipe object to update
//     schema:
//       "$ref": "#/definitions/Recipe"
//
// Summary: Update a recipe
// Description: Updates the recipe with the given ID and the details provided in the request body.

func (handler *RecipesHandler) UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.UpdateOne(handler.ctx, bson.M{
		"_id": objectId,
	}, bson.D{{"$set", bson.D{
		{"name", recipe.Name},
		{"instructions", recipe.Instructions},
		{"ingredients", recipe.Ingredients},
		{"tags", recipe.Tags},
	}}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been updated"})
}

// swagger:route DELETE /recipes/{id} recipes deleteRecipe
// Delete an existing recipe.
//
// Responses:
//   200: RecipeDeletedResponse
//   404: NotFoundResponse
//   500: InternalServerErrorResponse
//
// Produces:
// - application/json
//
// Parameters:
//   + name: id
//     in: path
//     required: true
//     type: string
//     description: Recipe ID to be deleted
//
// Summary: Delete a recipe
// Description: Deletes the recipe with the given ID.

func (handler *RecipesHandler) DeleteRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := handler.collection.DeleteOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Recipe has been deleted"})
}

// swagger:route GET /recipes/{id} recipes getRecipe
// Get a single recipe by ID.
//
// Responses:
//   200: RecipeRetrievedResponse
//   404: NotFoundResponse
//   500: InternalServerErrorResponse
//
// Produces:
// - application/json
//
// Parameters:
//   + name: id
//     in: path
//     required: true
//     type: string
//     description: Recipe ID to retrieve
//
// Summary: Get a recipe
// Description: Retrieves the recipe with the given ID.

func (handler *RecipesHandler) GetOneRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	objectId, _ := primitive.ObjectIDFromHex(id)
	cur := handler.collection.FindOne(handler.ctx, bson.M{
		"_id": objectId,
	})
	var recipe models.Recipe
	err := cur.Decode(&recipe)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// swagger:route GET /recipes/search recipes searchRecipes
// Search for recipes by tag.
//
// Responses:
//   200: ListOfRecipesResponse
//   404: NoRecipesFoundResponse
//   500: InternalServerErrorResponse
//
// Produces:
// - application/json
//
// Parameters:
//   + name: tag
//     in: query
//     required: true
//     type: string
//     description: Recipe tag to search by
//
// Summary: Search for recipes by tag
// Description: Finds and retrieves recipes that match the given tag.
