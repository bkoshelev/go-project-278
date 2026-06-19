package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/api"
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

	services := service.NewShortLinksService(queries, gen_id.CreateIDGenerator(), os.Getenv("HOST"))
	router := setupRouter()

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

	router = api.Ping(router)
	router = api.GetShortLinks(router, services)
	router = api.GetLinkVisits(router, services)
	router = api.CreateLink(router, services)
	router = api.GetShortLinkByID(router, services)
	router = api.RedirectShortLink(router, services)
	router = api.UpdateLink(router, services)
	router = api.DeleteLink(router, services)
	router = api.UnknownRoute(router)

	err = router.Run(":8080")

	if err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
