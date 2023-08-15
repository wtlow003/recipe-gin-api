package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/wtlow003/recipe-gin-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestListRecipes(t *testing.T) {
	// Arrange mock MongoDB collection and Redis client
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestReadFromMongoDB", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		first := mtest.CreateCursorResponse(1, "recipe.recipes", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: id1},
			{Key: "name", Value: "Singapore Noodle"},
		})
		second := mtest.CreateCursorResponse(1, "recipe.recipes", mtest.NextBatch, bson.D{
			{Key: "_id", Value: id2},
			{Key: "name", Value: "Singapore Rice"},
		})
		killCursor := mtest.CreateCursorResponse(0, "recipe.recipes", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursor)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// act
		r := gin.Default()
		r.GET("/recipes", handler.ListRecipes)
		req, _ := http.NewRequest("GET", "/recipes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		recipes := make([]models.Recipe, 0)
		if err := json.Unmarshal(w.Body.Bytes(), &recipes); err != nil {
			log.Errorf("Unable to unmarshal response body.")
		}
		val, _ := handler.RedisClient.Get("recipes").Result()
		redisRecipes := make([]models.Recipe, 0)
		if err := json.Unmarshal([]byte(val), &redisRecipes); err != nil {
			log.Errorf("Unable to unmarshal redis result.")
		}

		// assert collections
		assert.Equal(t, 2, len(recipes))
		assert.Equal(t, id1, recipes[0].ID)
		assert.Equal(t, id2, recipes[1].ID)
		assert.Equal(t, "Singapore Noodle", recipes[0].Name)
		assert.Equal(t, "Singapore Rice", recipes[1].Name)

		// assert redis
		assert.Equal(t, 2, len(redisRecipes))
		assert.Equal(t, id1, redisRecipes[0].ID)
		assert.Equal(t, id2, redisRecipes[1].ID)
		assert.Equal(t, "Singapore Noodle", redisRecipes[0].Name)
		assert.Equal(t, "Singapore Rice", redisRecipes[1].Name)
	})
	mockCollection.Run("TestReadFromRedisCache", func(mt *mtest.T) {
		// Arrange
		recipeCollection := mt.Coll
		id1 := primitive.NewObjectID()

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		expectedRecipe := models.Recipe{
			ID:   id1,
			Name: "Singapore Noodle",
		}
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})

		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}
		data, _ := json.Marshal(expectedRecipe)
		handler.RedisClient.Set("recipes", string(data), 0)

		// act
		r := gin.Default()
		r.GET("/recipes", handler.ListRecipes)
		req, _ := http.NewRequest("GET", "/recipes", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		val, _ := handler.RedisClient.Get("recipes").Result()
		var recipe models.Recipe
		if err := json.Unmarshal([]byte(val), &recipe); err != nil {
			log.Errorf("Unable to unmarshal redis result.")
		}

		// assert redis
		assert.Equal(t, id1, recipe.ID)
		assert.Equal(t, "Singapore Noodle", recipe.Name)
	})
}

func TestUpdateRecipe(t *testing.T) {
	// Arrange
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestInvalidObjectID", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		invalidId := "invalid"
		requestBody := models.Recipe{
			Name: "Singapore Noodle",
		}
		jsonVal, _ := json.Marshal(requestBody)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.PUT("/recipes/:id", handler.UpdateRecipe)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/recipes/%s", invalidId), bytes.NewBuffer(jsonVal))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
		assert.True(t, strings.Contains(error.Error, "Invalid ObjectId"))
	})
	mockCollection.Run("TestMissingPathParameter", func(mt *mtest.T) {
		recipeCollection := mt.Coll

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.PUT("/recipes/", handler.UpdateRecipe)
		req, _ := http.NewRequest("PUT", "/recipes/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
		assert.Equal(t, error.Error, "ID parameter not provided")
	})
	mockCollection.Run("TestInvalidRequestBody", func(mt *mtest.T) {
		// Arrange
		recipeCollection := mt.Coll
		id := primitive.NewObjectID()
		invalidRequestBody := []byte("")

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.PUT("/recipes/:id", handler.UpdateRecipe)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("/recipes/%s", id.Hex()), bytes.NewBuffer(invalidRequestBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
	})
}

func TestListRecipe(t *testing.T) {
	// Arrange
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestInvalidObjectID", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		invalidId := "invalid"

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/:id", handler.ListRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/%s", invalidId), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
		assert.True(t, strings.Contains(error.Error, "Invalid ObjectId"))
	})
	mockCollection.Run("TestMissingPathParameter", func(mt *mtest.T) {
		recipeCollection := mt.Coll

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/", handler.ListRecipe)
		req, _ := http.NewRequest("GET", "/recipes/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
		assert.Equal(t, error.Error, "ID parameter not provided")
	})
	mockCollection.Run("TestMongoDBUnexpectedError", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id := primitive.NewObjectID().Hex()

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/:id", handler.ListRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/%s", id), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
	})
	mockCollection.Run("TestReadFromMongoDB", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id := primitive.NewObjectID()
		expectedRecipe := models.Recipe{
			ID:   id,
			Name: "Singapore Noodles",
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(1, "recipe.recipes", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: expectedRecipe.ID},
			{Key: "name", Value: expectedRecipe.Name},
		}))

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/:id", handler.ListRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/%s", id.Hex()), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var recipe models.Recipe
		json.Unmarshal(w.Body.Bytes(), &recipe)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, expectedRecipe.ID, recipe.ID)
		assert.Equal(t, expectedRecipe.Name, recipe.Name)
	})
	mockCollection.Run("TestNoRecipe", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id := primitive.NewObjectID()
		first := mtest.CreateCursorResponse(1, "recipe.recipes", mtest.FirstBatch)
		killCursor := mtest.CreateCursorResponse(0, "recipe.recipes", mtest.NextBatch)
		mt.AddMockResponses(first, killCursor)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/:id", handler.ListRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/%s", id.Hex()), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 404, error.StatusCode)
		assert.True(t, strings.Contains(error.Error, "Recipe not found"))
	})
}

func TestNewRecipe(t *testing.T) {
	// Arrange
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestInvalidRequestBody", func(mt *mtest.T) {
		// Arrange
		recipeCollection := mt.Coll
		invalidRequestBody := []byte("")

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.POST("/recipes", handler.NewRecipe)
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(invalidRequestBody))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
	})
	mockCollection.Run("TestMongoDBError", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		expectedRecipe := models.UserDefinedRecipe{
			Name:         "Singapore Noodles",
			Tags:         []string{"main", "asian", "noodles"},
			Ingredients:  []string{"soy sauce", "noodles"},
			Instructions: "This is an example instruction.",
			Servings:     0,
			Calories:     500,
			Fat:          0,
			SatFat:       0,
			Carbs:        0,
			Fiber:        0,
			Sugar:        0,
			Protein:      0,
		}
		jsonVal, _ := json.Marshal(expectedRecipe)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.POST("/recipes", handler.NewRecipe)
		req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(jsonVal))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		// Assert
		assert.Equal(t, 500, error.StatusCode)
	})
	mockCollection.Run("TestWriteToMongoDB", func(mt *mtest.T) {
		// Arrange
		recipeCollection := mt.Coll
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		expectedRecipe := models.UserDefinedRecipe{
			Name:         "Singapore Noodles",
			Tags:         []string{"main", "asian", "noodles"},
			Ingredients:  []string{"soy sauce", "noodles"},
			Instructions: "This is an example instruction.",
			Servings:     0,
			Calories:     500,
			Fat:          0,
			SatFat:       0,
			Carbs:        0,
			Fiber:        0,
			Sugar:        0,
			Protein:      0,
		}
		jsonVal, _ := json.Marshal(expectedRecipe)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.POST("/recipes", handler.NewRecipe)
		req, err := http.NewRequest("POST", "/recipes", bytes.NewBuffer(jsonVal))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var recipe models.Recipe
		json.Unmarshal(w.Body.Bytes(), &recipe)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, expectedRecipe.Name, recipe.Name)
		assert.Equal(t, expectedRecipe.Tags, recipe.Tags)
		assert.Equal(t, expectedRecipe.Ingredients, recipe.Ingredients)
		assert.Equal(t, expectedRecipe.Instructions, recipe.Instructions)
		assert.Equal(t, expectedRecipe.Servings, recipe.Servings)
		assert.Equal(t, expectedRecipe.Calories, recipe.Calories)
		assert.Equal(t, expectedRecipe.Fat, recipe.Fat)
		assert.Equal(t, expectedRecipe.SatFat, recipe.SatFat)
		assert.Equal(t, expectedRecipe.Carbs, recipe.Carbs)
		assert.Equal(t, expectedRecipe.Fiber, recipe.Fiber)
		assert.Equal(t, expectedRecipe.Sugar, recipe.Sugar)
		assert.Equal(t, expectedRecipe.Protein, recipe.Protein)
	})
}

func TestDeleteRecipe(t *testing.T) {
	// Arrange
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestMissingPathParameter", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.DELETE("/recipes/", handler.DeleteRecipe)
		req, _ := http.NewRequest("DELETE", "/recipes/", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
		assert.Equal(t, error.Error, "ID parameter not provided")
	})
	mockCollection.Run("TestInvalidObjectID", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		invalidId := "invalid"

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.DELETE("/recipes/:id", handler.DeleteRecipe)
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/recipes/%s", invalidId), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
		assert.True(t, strings.Contains(error.Error, "Invalid ObjectId"))
	})
	mockCollection.Run("TestMongoDBUnexpectedError", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id := primitive.NewObjectID()
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.DELETE("/recipes/:id", handler.DeleteRecipe)
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/recipes/%s", id.Hex()), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
	})
	mockCollection.Run("TestDeleteFromMongoDB", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		id := primitive.NewObjectID()
		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "acknowledge", Value: true},
			{Key: "n", Value: 1},
		})

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.DELETE("/recipes/:id", handler.DeleteRecipe)
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/recipes/%s", id.Hex()), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var message models.Message
		json.Unmarshal(w.Body.Bytes(), &message)

		assert.Equal(t, 200, message.StatusCode)
		assert.True(t, strings.Contains(message.Message, "Deleted 1 recipe"))
	})
}

func TestSearchRecipe(t *testing.T) {
	// Arrange
	mockCollection := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mockCollection.Close()

	mockCollection.Run("TestMissingQueryParameter", func(mt *mtest.T) {
		recipeCollection := mt.Coll

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/search", handler.SearchRecipe)
		req, _ := http.NewRequest("GET", "/recipes/search", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 400, error.StatusCode)
		assert.Equal(t, error.Error, "`tag` query parameter is required.")
	})
	mockCollection.Run("TestMongoDBUnexpectedError", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		testTag := "main"
		mt.AddMockResponses(bson.D{{Key: "ok", Value: 0}})

		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		// Act
		r := gin.Default()
		r.GET("/recipes/search", handler.SearchRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/search?tag=%s", testTag), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		var error models.Error
		json.Unmarshal(w.Body.Bytes(), &error)

		assert.Equal(t, 500, error.StatusCode)
	})
	mockCollection.Run("TestReadFromMongoDB", func(mt *mtest.T) {
		recipeCollection := mt.Coll
		testTag := "main"
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		first := mtest.CreateCursorResponse(1, "recipe.recipes", mtest.FirstBatch, bson.D{
			{Key: "_id", Value: id1},
			{Key: "name", Value: "Singapore Noodle"},
		})
		second := mtest.CreateCursorResponse(1, "recipe.recipes", mtest.NextBatch, bson.D{
			{Key: "_id", Value: id2},
			{Key: "name", Value: "Singapore Rice"},
		})
		killCursor := mtest.CreateCursorResponse(0, "recipe.recipes", mtest.NextBatch)
		mt.AddMockResponses(first, second, killCursor)

		// `mocking` redis: https://itnext.io/golang-testing-mocking-redis-b48d09386c70
		s := miniredis.RunT(t)
		redisClient := redis.NewClient(&redis.Options{
			Addr: s.Addr(),
		})
		handler := RecipesHandler{
			Collection:  recipeCollection,
			Ctx:         context.Background(),
			RedisClient: redisClient,
		}

		r := gin.Default()
		r.GET("/recipes/search", handler.SearchRecipe)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/recipes/search?tag=%s", testTag), nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		recipes := make([]models.Recipe, 0)
		json.Unmarshal(w.Body.Bytes(), &recipes)

		// assert collections
		assert.Equal(t, 200, w.Code)
		assert.Equal(t, 2, len(recipes))
		assert.Equal(t, id1, recipes[0].ID)
		assert.Equal(t, id2, recipes[1].ID)
		assert.Equal(t, "Singapore Noodle", recipes[0].Name)
		assert.Equal(t, "Singapore Rice", recipes[1].Name)
	})
}
