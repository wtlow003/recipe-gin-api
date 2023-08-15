package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/wtlow003/recipe-gin-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type RecipesHandler struct {
	Collection  *mongo.Collection
	Ctx         context.Context
	RedisClient *redis.Client
}

func NewRecipesHandler(ctx context.Context, collection *mongo.Collection, redisClient *redis.Client) *RecipesHandler {
	return &RecipesHandler{
		Collection:  collection,
		Ctx:         ctx,
		RedisClient: redisClient,
	}
}

// ListRecipes		godoc
//
//	@Summary		List recipes
//	@Description	get all recipes
//	@Tags			recipes
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		models.Recipe
//	@Failure		500	{object}	models.Error
//	@Router			/recipes [get]
func (handler *RecipesHandler) ListRecipes(c *gin.Context) {
	// look for hit in redis cache first
	val, err := handler.RedisClient.Get("recipes").Result()
	if err == redis.Nil {
		log.Println("Request to MongoDB")
		// `collection` assigned in `init()`
		cursor, err := handler.Collection.Find(handler.Ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"error":      err.Error(),
			})
		}
		defer cursor.Close(handler.Ctx)

		recipes := make([]models.Recipe, 0)
		for cursor.Next(handler.Ctx) {
			var recipe models.Recipe
			cursor.Decode(&recipe)
			recipes = append(recipes, recipe)
		}

		// store in redis for later hits
		data, _ := json.Marshal(recipes)
		handler.RedisClient.Set("recipes", string(data), 0)
		c.JSON(http.StatusOK, recipes)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      err.Error(),
		})
		return
	} else {
		// if redis hit
		log.Println("Request to Redis")
		recipes := make([]models.Recipe, 0)
		json.Unmarshal([]byte(val), &recipes)
		c.JSON(http.StatusOK, recipes)
	}
}

// NewRecipe		godoc
//
//	@Summary		Create recipe
//	@Description	create new recipe
//	@Tags			recipes
//	@Accept			json
//	@Produce		json
//	@Param			recipe	body		models.UserDefinedRecipe	true	"New recipe"
//	@Success		200		{object}	models.Recipe
//	@Failure		400		{object}	models.Error
//	@Failure		500		{object}	models.Error
//	@Router			/recipes [post]
func (handler *RecipesHandler) NewRecipe(c *gin.Context) {
	var recipe models.Recipe
	// bind request body into `Recipe` struct
	if err := c.ShouldBindJSON(&recipe); err != nil {
		// if request body is invalid raise status code 400
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      fmt.Sprintf("Invalid request body, err = %s", err.Error()),
		})
		return
	}

	recipe.ID = primitive.NewObjectID()
	recipe.PublishedAt = time.Now()
	_, err := handler.Collection.InsertOne(handler.Ctx, recipe)
	if err != nil {
		log.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      fmt.Sprintf("Error inserting a new recipe, err = %s", err.Error()),
		})
		return
	}

	log.Println("Remove data from Redis")
	handler.RedisClient.Del("recipes")

	// successful
	c.JSON(http.StatusOK, recipe)
}

// UpdateRecipe	godoc
//
//	@Summary	Update recipe
//	@Tags		recipes
//	@Accept		json
//	@Produce	json
//	@Param		id				path		string			true	"Recipe ID"
//	@Param		recipe			body		models.Recipe	true	"Updated receipe"
//	@Success	200				{object}	models.Message
//	@Failure	400				{object}	models.Error
//	@Failure	500				{object}	models.Error
//	@Router		/recipes/{id}	[put]
func (handler *RecipesHandler) UpdateRecipe(c *gin.Context) {
	// recipe id
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided",
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

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      fmt.Sprintf("Invalid ObjectId, err = %s", err.Error()),
		})
		return
	}
	_, err = handler.Collection.UpdateOne(
		handler.Ctx,
		bson.M{"_id": objectId},
		bson.D{{
			Key: "$set", Value: bson.D{
				{Key: "name", Value: recipe.Name},
				{Key: "instructions", Value: recipe.Instructions},
				{Key: "ingredients", Value: recipe.Ingredients},
				{Key: "tags", Value: recipe.Tags},
			},
		}},
	)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      err.Error(),
		})
		return
	}

	log.Println("Remove data from Redis")
	handler.RedisClient.Del("recipes")

	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    "Recipe has been updated!",
	})

}

// ListRecipe	godoc
//
//	@Summary	List recipe
//	@Tags		recipes
//	@Accept		json
//	@Produce	json
//	@Param		id				path		string	true	"Recipe ID"
//	@Success	200				{object}	models.Recipe
//	@Failure	400				{object}	models.Error
//	@Failure	404				{object}	models.Error
//	@Failure	500				{object}	models.Error
//	@Router		/recipes/{id}	[get]
func (handler *RecipesHandler) ListRecipe(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided",
		})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      fmt.Sprintf("Invalid ObjectId, err = %s", err.Error()),
		})
		return
	}

	var recipe models.Recipe
	err = handler.Collection.FindOne(handler.Ctx,
		bson.D{{Key: "_id", Value: objectId}},
	).Decode(&recipe)
	if err != nil {
		// no documents retrieved
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"statusCode": http.StatusNotFound,
				"error":      fmt.Sprintf("Recipe not found, err = %s", err.Error()),
			})
			return
		} else {
			log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"statusCode": http.StatusInternalServerError,
				"error":      err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, recipe)
}

// DeleteRecipe	godoc
//
//	@Summary	Delete recipe
//	@Tags		recipes
//	@Accept		json
//	@Produce	json
//	@Param		id				path		string	true	"Recipe ID"
//	@Success	200				{object}	models.Message
//	@Failure	400				{object}	models.Error
//	@Failure	500				{object}	models.Error
//	@Router		/recipes/{id}	[delete]
func (handler *RecipesHandler) DeleteRecipe(c *gin.Context) {
	id, found := c.Params.Get("id")
	if !found {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      "ID parameter not provided",
		})
		return
	}

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      fmt.Sprintf("Invalid ObjectId, err = %s", err.Error()),
		})
		return
	}

	res, err := handler.Collection.DeleteOne(handler.Ctx, bson.M{"_id": objectId})
	if err != nil {
		log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      err.Error(),
		})
		return
	}

	format := "Deleted %d recipe!"
	c.JSON(http.StatusOK, gin.H{
		"statusCode": http.StatusOK,
		"message":    fmt.Sprintf(format, res.DeletedCount),
	})
}

// SearchRecipe	godoc
//
//	@Summary	Search recipes by tag
//	@Tags		recipes
//	@Accept		json
//	@Produce	json
//	@Param		tag				query		string	true	"Recipe search by tag"
//	@Success	200				{array}		models.Recipe
//	@Failure	400				{object}	models.Error
//	@Failure	500				{object}	models.Error
//	@Router		/recipes/search	[get]
func (handler *RecipesHandler) SearchRecipe(c *gin.Context) {
	tag := c.Query("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"statusCode": http.StatusBadRequest,
			"error":      "`tag` query parameter is required.",
		})
		return
	}

	query := make([]string, 0)
	cursor, err := handler.Collection.Find(handler.Ctx, bson.M{
		"tags": bson.M{
			"$in": append(query, tag),
		},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"statusCode": http.StatusInternalServerError,
			"error":      err.Error(),
		})
		return
	}
	defer cursor.Close(handler.Ctx)

	recipes := make([]models.Recipe, 0)
	for cursor.Next(handler.Ctx) {
		var recipe models.Recipe
		cursor.Decode(&recipe)
		fmt.Println(recipe)
		recipes = append(recipes, recipe)
	}
	c.JSON(http.StatusOK, recipes)
}
