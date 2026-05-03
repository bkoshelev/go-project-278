package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/bkoshelev/go-project-278/internal/service"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(sentrygin.New(sentrygin.Options{
		Repanic: true,
	}))

	return router
}

func ping(router *gin.Engine) *gin.Engine {
	router.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	return router
}

type CreateLinkRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name"`
}

func createLink(router *gin.Engine, queries *db.Queries) *gin.Engine {
	router.POST("/api/links", func(c *gin.Context) {
		var req CreateLinkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		shortLink, err := service.NewShortLinksService(queries).CreateShortLink(req.OriginalUrl, req.ShortName)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, shortLink)
	})

	return router
}

func getShortLinks(router *gin.Engine, queries *db.Queries) *gin.Engine {
	router.GET("/api/links", func(c *gin.Context) {

		shortLinks, err := service.NewShortLinksService(queries).GetLinks()

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, shortLinks)
	})

	return router
}

func getShortLinkById(router *gin.Engine, queries *db.Queries) *gin.Engine {
	router.GET("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": "ID должен быть числом"})
			return
		}

		shortLink, err := service.NewShortLinksService(queries).GetLinkById(int32(id))

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, shortLink)
	})

	return router
}

func updateLink(router *gin.Engine, queries *db.Queries) *gin.Engine {
	router.PUT("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": "ID должен быть числом"})
			return
		}

		var req CreateLinkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		err = service.NewShortLinksService(queries).UpdateShortLink(int32(id), req.OriginalUrl, req.ShortName)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.Status(200)
	})

	return router
}

func deleteLink(router *gin.Engine, queries *db.Queries) *gin.Engine {
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": "ID должен быть числом"})
			return
		}

		err = service.NewShortLinksService(queries).DeleteShortLink(int32(id))

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.Status(204)
	})

	return router
}

func unknownRoute(router *gin.Engine) *gin.Engine {
	router.NoRoute(func(c *gin.Context) {
		c.Status(http.StatusNotFound)
	})
	return router
}

func main() {
	if os.Getenv("ENV") == "dev" {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: os.Getenv("SENTRY_DSN"),
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	queries := db.New(conn)

	router := setupRouter()
	router = ping(router)
	router = getShortLinks(router, queries)
	router = createLink(router, queries)
	router = getShortLinkById(router, queries)
	router = updateLink(router, queries)
	router = deleteLink(router, queries)
	router = unknownRoute(router)

	err = router.Run(":" + os.Getenv("PORT"))

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
