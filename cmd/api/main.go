package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/api/example", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"data": "This is an example API response.",
		})
	})

	r.Run(":8080")
}
