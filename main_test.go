package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var conn *pgxpool.Pool

func withTx(t *testing.T, fn func(ctx context.Context, q *db.Queries, tx pgx.Tx)) {
	log.Print("Начинаем готовить транзакцию")

	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	log.Print("Создали контекст")

	t.Cleanup(cancel)

	tx, err := conn.Begin(ctx)
	log.Print("Создали транзакцию")

	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	qtx := db.New(tx)
	qtx = qtx.WithTx(tx)
	log.Print("Стартуем тест")
	fn(ctx, qtx, tx)
}

func TestMain(m *testing.M) {
	var err error

	gin.SetMode(gin.TestMode)
	ctx := context.Background()

	if os.Getenv("DATABASE_URL") == "" {
		_ = godotenv.Load(".env.test")
	}

	log.Print("DATABASE_URL:" + os.Getenv("DATABASE_URL"))

	conn, err = pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	ctxPing, cancel := context.WithTimeout(ctx, 10*time.Second)
	if err := conn.Ping(ctxPing); err != nil {
		cancel()
		log.Fatalf("ping db: %v", err)
	}
	cancel()

	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()
	router = ping(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGetLinks(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, q *db.Queries, _ pgx.Tx) {
		router = createLink(router, q)
		router = getShortLinks(router, q)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		expected := TestShortLink{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
			ShortUrl:    os.Getenv("HOST") + "/r/" + "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/links", nil)
		router.ServeHTTP(w, req)

		get := []TestShortLink{}

		err := json.Unmarshal(w.Body.Bytes(), &get)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, get, expected)
	})
}

type TestShortLink struct {
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortUrl    string `json:"short_url"`
}

func TestCreateLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, q *db.Queries, _ pgx.Tx) {
		router = createLink(router, q)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		expected := TestShortLink{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
			ShortUrl:    os.Getenv("HOST") + "/r/" + "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		get := TestShortLink{}

		err := json.Unmarshal(w.Body.Bytes(), &get)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		assert.Equal(t, 201, w.Code)
		assert.Equal(t, expected, get)
	})
}

func TestGetLinksById(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, q *db.Queries, _ pgx.Tx) {
		router = createLink(router, q)
		router = getShortLinkById(router, q)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		expected := TestShortLink{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
			ShortUrl:    os.Getenv("HOST") + "/r/" + "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		createShortLink := db.ShortLink{}
		err := json.Unmarshal(w.Body.Bytes(), &createShortLink)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/links/"+strconv.Itoa(int(createShortLink.ID)), nil)
		router.ServeHTTP(w, req)

		get := TestShortLink{}
		err = json.Unmarshal(w.Body.Bytes(), &get)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, get, expected)
	})
}

func TestUpdateLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, q *db.Queries, _ pgx.Tx) {
		router = createLink(router, q)
		router = updateLink(router, q)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		createdShortLink := db.ShortLink{}
		err := json.Unmarshal(w.Body.Bytes(), &createdShortLink)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		updatedShortLink := CreateLinkRequest{
			OriginalUrl: "https://example2.com",
			ShortName:   "short2",
		}
		updatedShortLinkJson, _ := json.Marshal(updatedShortLink)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("PUT", "/api/links/"+strconv.Itoa(int(createdShortLink.ID)), strings.NewReader(string(updatedShortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestDeleteLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, q *db.Queries, _ pgx.Tx) {
		router = createLink(router, q)
		router = deleteLink(router, q)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}
		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, 201, w.Code)

		createdShortLink := db.ShortLink{}
		err := json.Unmarshal(w.Body.Bytes(), &createdShortLink)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/api/links/"+strconv.Itoa(int(createdShortLink.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, 204, w.Code)
	})
}
