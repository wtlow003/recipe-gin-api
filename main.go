package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"

	"github.com/wtlow003/recipe-gin-api/db"
	_ "github.com/wtlow003/recipe-gin-api/docs"
	"github.com/wtlow003/recipe-gin-api/handlers"
	"github.com/wtlow003/recipe-gin-api/models"
)

var recipes []models.Recipe
var ctx context.Context
var collection *mongo.Collection
var recipesHandler *handlers.RecipesHandler

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

	// setup mongodb connections
	ctx := context.Background()
	db, err := db.ConnectToMongoDB(
		ctx,
		os.Getenv("MONGO_INITDB_ROOT_USERNAME"),
		os.Getenv("MONGO_INITDB_ROOT_PASSWORD"),
		os.Getenv("MONGODB_HOSTNAME"),
		os.Getenv("MONGODB_DATABASE"),
		os.Getenv("MONGODB_PORT"),
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	recipes = make([]models.Recipe, 0)
	f, err := os.ReadFile("recipes.json")
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}
	err = json.Unmarshal([]byte(f), &recipes)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	database := db.Client.Database(os.Getenv("MONGODB_DATABASE"))
	collections, err := database.ListCollectionNames(ctx, bson.D{})
	if err != nil {
		log.Fatal(err.Error())
	}
	// check if collections already exist, else add data from
	// recipes.json
	if !slices.Contains(collections, "recipes") {
		// Storing recipes into database
		// Generic type required in `collection.InsertMany()`
		var listOfRecipes []interface{}
		for _, recipe := range recipes {
			listOfRecipes = append(listOfRecipes, recipe)
		}
		collection = database.Collection("recipes")
		insertManyResult, err := collection.InsertMany(ctx, listOfRecipes)
		if err != nil {
			log.Fatal(err.Error())
		}
		log.Printf("Inserted recipes: %d", len(insertManyResult.InsertedIDs))
	} else {
		collection = database.Collection("recipes")
		log.Info("Collection `recipe` already exists! No data is inserted.")
	}

	recipesHandler = handlers.NewRecipesHandler(ctx, collection)

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
		v1.GET("/recipes", recipesHandler.ListRecipes)
		v1.GET("/recipes/:id", recipesHandler.ListRecipe)
		v1.GET("/recipes/search", recipesHandler.SearchRecipe)
		v1.POST("/recipes", recipesHandler.NewRecipe)
		v1.PUT("/recipes/:id", recipesHandler.UpdateRecipe)
		v1.DELETE("/recipes/:id", recipesHandler.DeleteRecipe)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Run(":8080")
}
