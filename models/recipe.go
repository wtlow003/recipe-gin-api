package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Recipe struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Name         string             `json:"name" bson:"name"`
	Tags         []string           `json:"tags" bson:"tags"`
	Ingredients  []string           `json:"ingredients" bson:"ingredients"`
	Instructions []string           `json:"instructions" bson:"instruction"`
	Servings     int                `json:"servings" bson:"servings"`
	Calories     int                `json:"calories" bson:"calories"`
	PublishedAt  time.Time          `json:"publishedAt" bson:"publishedAt"`
}