package main

import "github.com/gin-gonic/gin"

func main() {
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.Run()
}
