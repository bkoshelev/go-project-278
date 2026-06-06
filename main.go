package main

import (
	"context"
	"encoding/json"
	"errors"
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
	"github.com/go-playground/validator/v10"
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
	router.TrustedPlatform = gin.PlatformCloudflare

	return router
}

func ping(router *gin.Engine) *gin.Engine {
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

type CreateLinkRequest struct {
	OriginalUrl string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

func createLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.POST("/api/links", func(c *gin.Context) {
		var req CreateLinkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				out := make(map[string]string)
				for _, fe := range ve {
					out[fe.Field()] = fe.Error()
				}
				c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": out})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		shortLink, err := services.CreateShortLink(req.OriginalUrl, req.ShortName)

		if err.Err != nil {
			var ve service.ServiceError

			if errors.As(err, &ve) {
				out := make(map[string]string)
				out[ve.FieldName] = ve.Err.Error()
				c.JSON(http.StatusBadRequest, gin.H{"errors": out})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "data base error"})
			return
		}

		c.JSON(http.StatusCreated, shortLink)
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

// https://gin-gonic.com/en/docs/binding/bind-custom-unmarshaler/#using-bindunmarshaler
func (r *Range) UnmarshalParam(param string) error {
	var arr [2]int

	if err := json.Unmarshal([]byte(param), &arr); err != nil {
		return fmt.Errorf("invalid format, expected [start,end]")
	}

	if len(arr) != 2 {
		return fmt.Errorf("range must contain exactly 2 values")
	}

	r.Begin = arr[0]
	r.End = arr[1]

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
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		countLinks, err := services.CountLinks()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"links %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func getShortLinkById(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID должен быть числом"})
			return
		}

		shortLink, err := services.GetLinkById(int32(id))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, shortLink)
	})

	return router
}

func updateLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.PUT("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID должен быть числом"})
			return
		}

		var req CreateLinkRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				out := make(map[string]string)
				for _, fe := range ve {
					out[fe.Field()] = fe.Error()
				}
				c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": out})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		if services.UpdateShortLink(int32(id), req.OriginalUrl, req.ShortName).Err != nil {
			var ve service.ServiceError

			if errors.As(err, &ve) {
				out := make(map[string]string)
				out[ve.FieldName] = ve.Err.Error()
				c.JSON(http.StatusBadRequest, gin.H{"errors": out})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": "data base error"})
			return
		}

		c.Status(http.StatusOK)
	})

	return router
}

func deleteLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.DELETE("/api/links/:id", func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID должен быть числом"})
			return
		}

		err = services.DeleteShortLink(int32(id))

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusNoContent)
	})

	return router
}

func getLinkVisits(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/api/link_visits", func(c *gin.Context) {
		query := Query{Range: Range{Begin: 0, End: 10}}
		if c.Query("range") != "" {
			_ = c.BindQuery(&query)
		}
		begin := query.Range.Begin
		end := query.Range.End

		shortLinks, err := services.GetLinkVisits(
			db.GetLinkVisitsParams{
				Limit:  end - begin + 1,
				Offset: int32(begin),
			},
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		countLinks, err := services.CountLinkVisits()

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Range", fmt.Sprintf(
			"link_visits %v-%v/%v", begin, end, countLinks))
		c.JSON(http.StatusOK, shortLinks)
	})

	return router
}

func redirectShortLink(router *gin.Engine, services *service.ShortLinksService) *gin.Engine {
	router.GET("/r/:code", func(c *gin.Context) {
		shortLink, err := services.GetLinkByShortName(c.Param("code"))

		if err != nil {
			if errors.Is(err, service.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": service.ErrNoRows.Error()})
				return
			}

			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err = services.CreateLinkVisit(
			c.ClientIP(),
			shortLink.ID,
			c.Request.UserAgent(),
			c.Request.Referer(),
			http.StatusFound,
		)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, shortLink.OriginalUrl)
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
	router = getLinkVisits(router, services)
	router = createLink(router, services)
	router = getShortLinkById(router, services)
	router = redirectShortLink(router, services)
	router = updateLink(router, services)
	router = deleteLink(router, services)
	router = unknownRoute(router)

	err := router.Run(":8080")

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
