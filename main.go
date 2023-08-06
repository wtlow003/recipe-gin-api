package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/wtlow003/recipe-gin-api/docs"
	"github.com/wtlow003/recipe-gin-api/models"
)

var recipes []models.Recipe

func init() {
	// load env variable
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading from .env file!")
	}

	// setup logrus
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = log.InfoLevel
	}

	log.SetLevel(logLevel)
	log.SetFormatter(&log.JSONFormatter{})

	recipes = make([]models.Recipe, 0)
	f, err := os.ReadFile("recipes.json")
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

// NewRecipe		godoc
//
// @Summary			Create recipe
// @Description 	create new recipe
// @Tags			recipes
// @Accept			json
// @Produce			json
// @Param			recipe	body	models.UserDefinedRecipe true	"New receipe"
// @Success			200 {object}	models.Recipe
// @Failure			400 {object}	models.Error
// @Router			/recipes [post]
func NewRecipeHandler(c *gin.Context) {
	var recipe models.Recipe
	// bind request body into `Recipe` struct
	if err := c.ShouldBindJSON(&recipe); err != nil {
		// if request body is invalid raise status code 400
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      err.Error(),
		})
		return
	}
	recipe.ID = xid.New().String()
	recipe.PublishedAt = time.Now()
	recipes = append(recipes, recipe)
	// successful
	c.JSON(http.StatusOK, recipe)
}

// ListRecipes		godoc
//
// @Summary		List recipes
// @Description	get all recipes
// @Tags		recipes
// @Accept		json
// @Produce		json
// @Success		200	{array}		models.Recipe
// @Failure		500	{object}	models.Error
// @Router		/recipes [get]
func ListRecipesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, recipes)
}

// UpdateRecipe	godoc
// @Summary		Update recipes
// @Tags		recipes
// @Accept		json
// @Produce		json
// @Param		id	path 		string	true 	"Recipe ID"
// @Param		recipe	body	models.Recipe true	"Updated receipe"
// @Success		200 {object}	models.Recipe
// @Failure		400	{object}	models.Error
// @Failure		404	{object}	models.Error
// @Failure		500	{object}	models.Error
// @Router		/recipes/{id}	[put]
func UpdateRecipeHandler(c *gin.Context) {
	// recipe id
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided.",
		})
		return
	}
	var recipe models.Recipe
	// if error occurs, refer time formatting: https://romangaranin.net/posts/2021-02-19-json-time-and-golang/
	if err := c.ShouldBindJSON(&recipe); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      err.Error(),
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
			"statusCode": http.StatusNotFound,
			"error":      "Recipe not found!",
		})
		return
	}

	recipes[index] = recipe
	c.JSON(http.StatusOK, recipe)
}

// ListRecipe	godoc
// @Summary		Update recipes
// @Tags		recipes
// @Accept		json
// @Produce		json
// @Param		id	path 		string	true 	"Recipe ID"
// @Success		200 {object}	models.Recipe
// @Failure		404	{object}	models.Error
// @Failure		500	{object}	models.Error
// @Router		/recipes/{id}	[get]
func ListRecipeHandler(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided.",
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
			"statusCode": http.StatusNotFound,
			"error":      "Recipe not found!",
		})
		return
	}

	c.JSON(http.StatusOK, recipe)
}

// DeleteRecipe	godoc
// @Summary		Update recipes
// @Tags		recipes
// @Accept		json
// @Produce		json
// @Param		id	path 		string	true 	"Recipe ID"
// @Success		200 {object}	models.Message
// @Failure		404	{object}	models.Error
// @Failure		500	{object}	models.Error
// @Router		/recipes/{id}	[delete]
func DeleteRecipeHandler(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided.",
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
			"statusCode": http.StatusNotFound,
			"error":      "Recipe not found!",
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

// SearchRecipe	godoc
// @Summary		Update recipes
// @Tags		recipes
// @Accept		json
// @Produce		json
// @Param		tag	query 		string	true 	"Recipe search by tag"
// @Success		200 {array}		models.Recipe
// @Failure		400	{object}	models.Error
// @Router		/recipes/search	[get]
func SearchRecipeHandler(c *gin.Context) {
	tag := c.Query("tag")
	listOfRecipes := make([]models.Recipe, 0)

	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"statuCode": http.StatusBadRequest,
			"error":     "`tag` parameter is required.",
		})
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

//	@title			Recipe API
//	@version		1.0
//	@description	Demo recipe RESTful API developed with Gin framework.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	Low Wei Teck (Jensen)
//	@contact.url	https://www.linkedin.com/in/weitecklow/
//	@contact.email	jensenlwt@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	// refer to: https://medium.com/pengenpaham/implement-basic-logging-with-gin-and-logrus-5f36fba69b28
	// r.Use(gin.Recovery())
	// r.Use(middlewares.LoggingMiddleware())

	// similar to FastAPI's router: https://fastapi.tiangolo.com/tutorial/bigger-applications/
	v1 := r.Group("/api/v1")
	{
		v1.GET("/recipes", ListRecipesHandler)
		v1.GET("/recipes/:id", ListRecipeHandler)
		v1.GET("/recipes/search", SearchRecipeHandler)
		v1.POST("/recipes", NewRecipeHandler)
		v1.PUT("/recipes/:id", UpdateRecipeHandler)
		v1.DELETE("/recipes/:id", DeleteRecipeHandler)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
