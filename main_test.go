package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wtlow003/recipe-gin-api/models"
)

func TestUnknownHandler(t *testing.T) {
	r := SetupServer()
	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, 404, w.Code)
}

func TestListRecipes(t *testing.T) {
	r := SetupServer()
	r.GET("/recipes", recipesHandler.ListRecipes)
	req, _ := http.NewRequest("GET", "/recipes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var recipes []models.Recipe
	json.Unmarshal(w.Body.Bytes(), &recipes)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.LessOrEqual(t, 518, len(recipes))
}

func TestNewRecipe(t *testing.T) {
	r := SetupServer()
	r.POST("/recipes", recipesHandler.NewRecipe)

	recipe := models.Recipe{
		Name: "Singapore Noodles",
		Tags: []string{"main", "singapore"},
	}
	jsonVal, _ := json.Marshal(recipe)
	req, _ := http.NewRequest("POST", "/recipes", bytes.NewBuffer(jsonVal))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var result models.Recipe
	_ = json.Unmarshal(w.Body.Bytes(), &result)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, recipe.Name, result.Name)
	assert.Equal(t, recipe.Tags, result.Tags)
}
