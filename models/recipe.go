package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserDefinedRecipe struct {
	Name         string   `json:"name" bson:"name"`
	Tags         []string `json:"tags" bson:"tags"`
	Ingredients  []string `json:"ingredients" bson:"ingredients"`
	Instructions string   `json:"instructions" bson:"instruction"`
	Servings     int      `json:"servings" bson:"servings"`
	Calories     int      `json:"calories" bson:"calories"`
	Fat          int      `json:"fat" bson:"fat"`
	SatFat       int      `json:"satfat" bson:"satfat"`
	Carbs        int      `json:"carbs" bson:"carbs"`
	Fiber        int      `json:"fiber" bson:"fiber"`
	Sugar        int      `json:"sugar" bson:"sugar"`
	Protein      int      `json:"protein" bson:"proten"`
}

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions string             `json:"instructions" bson:"instruction"`
	Servings     int                `json:"servings" bson:"servings"`
	Calories     int                `json:"calories" bson:"calories"`
	Fat          int                `json:"fat" bson:"fat"`
	SatFat       int                `json:"satfat" bson:"satfat"`
	Carbs        int                `json:"carbs" bson:"carbs"`
	Fiber        int                `json:"fiber" bson:"fiber"`
	Sugar        int                `json:"sugar" bson:"sugar"`
	Protein      int                `json:"protein" bson:"proten"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}
