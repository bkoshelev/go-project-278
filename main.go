package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/bkoshelev/go-project-278/internal/gen_id"
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

func createLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.POST("/api/links", func(c *gin.Context) {
		var req CreateLinkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		shortLink, err := services.CreateShortLink(req.OriginalUrl, req.ShortName)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(201, shortLink)
	})

	return router
}

type Range struct {
	Begin int
	End   int
}

type Query struct {
	Range Range `form:"range"`
}

func (r *Range) UnmarshalParam(param string) error {
	fmt.Println("начинаем парсить пагинацию")
	var arr [2]int

	if err := json.Unmarshal([]byte(param), &arr); err != nil {
		return fmt.Errorf("invalid format, expected [start,end]")
	}

	if len(arr) != 2 {
		return fmt.Errorf("range must contain exactly 2 values")
	}

	r.Begin = arr[0]
	r.End = arr[1]
	fmt.Printf("закончили парсить пагинацию %v", r)

	return nil
}

func getShortLinks(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links", func(c *gin.Context) {
		query := Query{Range: Range{Begin: 0, End: 10}}
		if c.Query("range") != "" {
			_ = c.BindQuery(&query)
		}
		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err := services.GetLinks(
			db.GetShortLinksParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		countLinks, err := services.CountLinks()

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"links %v-%v/%v", begin, end, countLinks))
		c.JSON(200, shortLinks)
	})

	return router
}

func getShortLinkById(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": "ID должен быть числом"})
			return
		}

		shortLink, err := services.GetLinkById(int32(id))

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, shortLink)
	})

	return router
}

func updateLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
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

		err = services.UpdateShortLink(int32(id), req.OriginalUrl, req.ShortName)

		if err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		c.Status(200)
	})

	return router
}

func deleteLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(400, gin.H{"error": "ID должен быть числом"})
			return
		}

		err = services.DeleteShortLink(int32(id))

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

	runApp(conn)
}

func runApp(dbConn *pgxpool.Pool) {
	queries := db.New(dbConn)

	router := setupRouter()
	router = ping(router)
	services := service.NewShortLinksService(queries, gen_id.CreateIdGenerator())
	router = getShortLinks(router, services)
	router = createLink(router, services)
	router = getShortLinkById(router, services)
	router = updateLink(router, services)
	router = deleteLink(router, services)
	router = unknownRoute(router)

	err := router.Run(":8080")

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
