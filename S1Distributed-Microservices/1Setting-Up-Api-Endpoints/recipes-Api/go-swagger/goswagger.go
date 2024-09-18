// Recipes API
//
// This is a sample recipes API. You can find out more at https://github.com/TrapLord92/Building-Distributed-Applications-With-Golang
// Schemes: http
// Host: localhost:8080
// BasePath: /
// Version: 1.0.0
// Contact:
//
//	name: TrapLord
//	email: traplord345@gmail.com
//
// Consumes:
// - application/json
//
// Produces:
// - application/json
// swagger:meta
package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
)

// Recipe represents a recipe in the system.
// swagger:model Recipe
type Recipe struct {
	// The ID of the recipe
	//
	// required: true
	ID string `json:"id"`

	// The name of the recipe
	//
	// required: true
	Name string `json:"name"`

	// A list of tags associated with the recipe
	//
	// required: false
	Tags []string `json:"tags"`

	// The ingredients required for the recipe
	//
	// required: true
	Ingredients []string `json:"ingredients"`

	// The time when the recipe was published
	//
	// required: true
	PublishedAt time.Time `json:"published_at"`
}

// Mock data
var recipes []Recipe

// Initialize mock data
func init() {
	recipes = make([]Recipe, 0)
	file, _ := os.ReadFile("recipes.json")
	_ = json.Unmarshal([]byte(file), &recipes)
}

// swagger:operation POST /recipes createRecipe
//
// Creates a new recipe.
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: recipe
//     in: body
//     description: The recipe to create
//     required: true
//     schema:
//     "$ref": "#/definitions/Recipe"
//
// responses:
//
//	"200":
//	  description: Successfully created recipe
//	  schema:
//	    "$ref": "#/definitions/Recipe"
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

// swagger:operation GET /recipes listRecipes
//
// Lists all available recipes.
//
// ---
// produces:
// - application/json
// responses:
//
//	"200":
//	  description: A list of recipes
//	  schema:
//	    type: array
//	    items:
//	      "$ref": "#/definitions/Recipe"
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

// swagger:operation PUT /recipes/{id} updateRecipe
//
// Updates an existing recipe.
//
// ---
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//   - name: recipe
//     in: body
//     description: The recipe to update
//     required: true
//     schema:
//     "$ref": "#/definitions/Recipe"
//
// responses:
//
//	"200":
//	  description: Successfully updated recipe
//	  schema:
//	    "$ref": "#/definitions/Recipe"
func UpdateRecipeHandler(c *gin.Context) {
	id := c.Param("id")
	var recipe Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
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

// swagger:operation DELETE /recipes/{id} deleteRecipe
//
// Deletes an existing recipe.
//
// ---
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     required: true
//     type: string
//
// responses:
//
//	"200":
//	  description: Successfully deleted recipe
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

// swagger:operation GET /recipes/search searchRecipes
//
// Search recipes by tag.
//
// ---
// produces:
// - application/json
// parameters:
//   - name: tag
//     in: query
//     description: The tag to search for
//     required: true
//     type: string
//
// responses:
//
//	"200":
//	  description: List of recipes matching the tag
//	  schema:
//	    type: array
//	    items:
//	      "$ref": "#/definitions/Recipe"
func SearchRecipesHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]Recipe, 0)
	for _, recipe := range recipes {
		for _, t := range recipe.Tags {
			if strings.EqualFold(t, tag) {
				listOfRecipes = append(listOfRecipes, recipe)
			}
		}
	}
	c.JSON(http.StatusOK, listOfRecipes)
}

// main sets up the routes and runs the server
func main() {
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)
	router.DELETE("/recipes/:id", DeleteRecipeHandler)
	router.GET("/recipes/search", SearchRecipesHandler)

	router.Run(":8080")
}

//command to generate swagger docs:
/*swagger generate spec -o ./swagger.json --scan-models*/
