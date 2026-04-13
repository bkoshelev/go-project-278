package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}

func ping(router *gin.Engine) *gin.Engine {
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return router
}

func main() {

	router := setupRouter()
	router = ping(router)
	err := router.Run(":8080")

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
