package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wtlow003/recipe-gin-api/models"
)

var recipes []models.Recipe

func init() {
	recipes = make([]models.Recipe, 0)
	f, err := ioutil.ReadFile("recipes.json")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(f), &recipes)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

// handler to create new recipe
func NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	// bind request body into `Recipe` struct
	if err := c.ShouldBindJSON(&recipe); err != nil {
		// if request body is invalid raise status code 400
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	// successful
	c.JSON(http.StatusOK, recipe)
}

// handler to list out all existing recipes
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

func main() {
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.POST("/recipes", NewRecipeHandler)
	router.GET("/recipes", ListRecipesHandler)
	router.Run()
}
