package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/xid"

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

// Handler to create new recipe
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
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	// successful
	c.JSON(http.StatusOK, recipe)
}

// Handler to list out all existing recipes
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

// Handler to update existing recipe
func UpdateRecipeHandler(c *gin.Context) {
	// recipe id
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID parameter not provided.",
		})
		return
	}
	var recipe models.Recipe
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found!",
		})
		return
	}

	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

func ListRecipeHandler(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID parameter not provided.",
		})
		return
	}

	var recipe models.Recipe
	for _, r := range recipes {
		if r.ID == id {
			recipe = r
		}
	}

	// if recipe is still empty -> compable to empty struct
	if reflect.DeepEqual(recipe, models.Recipe{}) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found!",
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

func DeleteRecipeHandler(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "ID parameter not provided.",
		})
		return
	}

	index := -1
	for i := 0; i < len(recipes); i++ {
		if recipes[i].ID == id {
			index = i
		}
	}

	if index == -1 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Recipe not found!",
		})
		return
	}

	// create new slice up to index
	// create new slice from index+1
	// ... expand the list of elements to incld. everything not explictly stated
	recipes = append(recipes[:index], recipes[index+1:]...)
	c.JSON(http.StatusOK, gin.H{
		"message": "Recipe has been deleted!",
	})
}

// Handler to search Recipe by tags
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]models.Recipe, 0)

	if len(tag) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tag length!",
		})
		return
	}

	for i := 0; i < len(recipes); i++ {
		found := false
		for _, t := range recipes[i].Tags {
			if strings.EqualFold(t, tag) {
				found = true
			}
		}
		if found {
			listOfRecipes = append(listOfRecipes, recipes[i])
		}
	}

	c.JSON(http.StatusOK, listOfRecipes)
}

func main() {
	gin.SetMode(gin.DebugMode)
	router := gin.Default()

	router.GET("/recipes", ListRecipesHandler)
	router.GET("/recipes/:id", ListRecipeHandler)
	router.GET("/recipes/search", SearchRecipeHandler)

	router.POST("/recipes", NewRecipeHandler)
	router.PUT("/recipes/:id", UpdateRecipeHandler)

	router.DELETE("/recipes/:id", DeleteRecipeHandler)

	router.Run()
}
