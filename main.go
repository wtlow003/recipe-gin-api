package main

import "github.com/gin-gonic/gin"

func main() {
	gin.SetMode(gin.DebugMode)
	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})
	router.Run()
}
