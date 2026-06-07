package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/gen_id"
	"github.com/bkoshelev/go-project-278/internal/service"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
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

func setupValidation() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			return name
		})
	}
}

type CreateLinkPayload struct {
	OriginalUrl string `json:"original_url" binding:"required"`
	ShortName   string `json:"short_name" binding:"omitempty,min=3,max=32"`
}

type GetEntityUriParams struct {
	ID int `uri:"id" binding:"required"`
}

type RedirectUriParams struct {
	ShortName string `uri:"code" binding:"required"`
}

func bindUri(c *gin.Context, parameters any) error {
	if err := c.ShouldBindUri(parameters); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return err
	}
	return nil
}

func bindPayload(c *gin.Context, payload any) error {
	err := c.ShouldBindJSON(&payload)

	if err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			out := make(map[string]string)
			for _, fe := range ve {
				out[fe.Field()] = fe.Error()
			}
			c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": out})
			return err
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return err
	}
	return nil
}

func handleServiceError(c *gin.Context, err service.ServiceError) {
	if err.Err != nil {
		var ve service.ServiceError

		if errors.Is(err.Err, service.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		if errors.As(err.Err, &ve) {
			out := make(map[string]string)
			out[ve.FieldName] = ve.Err.Error()
			c.JSON(http.StatusBadRequest, gin.H{"errors": out})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "data base error"})
	}
}

type Range struct {
	Begin int
	End   int
}

type GetMulitpleEntityQueryParams struct {
	Range Range `form:"range" binding:"required"`
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

func bindQuery(c *gin.Context, obj any) error {
	err := c.BindQuery(obj)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return err
	}
	return nil
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

	setupValidation()

	conn, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	queries := db.New(conn)

	services := service.NewShortLinksService(queries, gen_id.CreateIdGenerator())
	router := setupRouter()
	router = ping(router)
	router = getShortLinks(router, services)
	router = getLinkVisits(router, services)
	router = createLink(router, services)
	router = getShortLinkById(router, services)
	router = redirectShortLink(router, services)
	router = updateLink(router, services)
	router = deleteLink(router, services)
	router = unknownRoute(router)

	router.Use(cors.New(
		cors.Config{
			AllowOrigins: []string{
				"http://localhost:5173",
			},
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
			},
			AllowHeaders: []string{
				"Content-Type",
				"Range",
			},
		},
	))

	err = router.Run(":8080")

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
