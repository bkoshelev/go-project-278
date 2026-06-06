package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/bkoshelev/go-project-278/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var conn *pgxpool.Pool

type MockIdGenerator struct {
	mock.Mock
}

var mockShortName string = "test_short_url"

func (m *MockIdGenerator) New() (string, error) {
	return mockShortName, nil
}

func withTx(t *testing.T, fn func(ctx context.Context, services *service.ShortLinksService, tx pgx.Tx)) {
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
	services := service.NewShortLinksService(qtx, new(MockIdGenerator))
	log.Print("Стартуем тест")
	fn(ctx, services, tx)
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

	// _, err = conn.Exec(ctx, "TRUNCATE link_visits, short_links RESTART IDENTITY")
	_, err = conn.Exec(ctx, "TRUNCATE short_links RESTART IDENTITY")

	if err != nil {
		log.Fatalf("fail to prepare short_links table")
	}

	os.Exit(m.Run())
}

func TestPingRoute(t *testing.T) {
	router := setupRouter()
	router = ping(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "pong", w.Body.String())
}

func TestGetLinks(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)
		router = getShortLinks(router, services)

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

		assert.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/links", nil)
		router.ServeHTTP(w, req)

		get := []TestShortLink{}

		err := json.Unmarshal(w.Body.Bytes(), &get)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, get, expected)
	})
}

func TestGetLinksWithPagination(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)
		router = getShortLinks(router, services)

		var initial []TestShortLink

		for i := 0; i < 10; i++ {
			initial = append(initial, TestShortLink{
				OriginalUrl: "https://example.com/" + strconv.Itoa(i),
				ShortName:   "short_" + strconv.Itoa(i),
				ShortUrl:    os.Getenv("HOST") + "/r/" + "short_" + strconv.Itoa(i),
			})
		}
		for _, shortLink := range initial {
			newShortLink := CreateLinkRequest{
				OriginalUrl: shortLink.OriginalUrl,
				ShortName:   shortLink.ShortName,
			}

			shortLinkJson, _ := json.Marshal(newShortLink)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusCreated, w.Code)

		}

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/links?range=[0,4]", nil)
		router.ServeHTTP(w, req)

		get := []TestShortLink{}
		err := json.Unmarshal(w.Body.Bytes(), &get)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		expected := initial[:5]

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Result().Header["Content-Range"], "links 0-4/10")
		assert.Equal(t, 5, len(get))
		assert.Equal(t, get, expected)
	})
}

type TestShortLink struct {
	OriginalUrl string `json:"original_url"`
	ShortName   string `json:"short_name"`
	ShortUrl    string `json:"short_url"`
}

func TestCreateLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)

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

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, expected, get)
	})
}

func TestCreateLinkWithRandomName(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
		}

		expected := TestShortLink{
			OriginalUrl: "https://example.com",
			ShortName:   mockShortName,
			ShortUrl:    os.Getenv("HOST") + "/r/" + mockShortName,
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

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, expected, get)
	})
}

func TestGetLinksById(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)
		router = getShortLinkById(router, services)

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

		assert.Equal(t, http.StatusCreated, w.Code)

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

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, get, expected)
	})
}

func TestRedirectShortLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, tx pgx.Tx) {
		// err = router.SetTrustedProxies([]string{"127.0.0.1"})

		// if err != nil {
		// 	panic("Ошибка парсинга Content Range")
		// }

		router = createLink(router, services)
		router = redirectShortLink(router, services)
		router = getLinkVisits(router, services)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/api/link_visits", nil)
		router.ServeHTTP(w, req)

		type ContentRange struct {
			From  int
			To    int
			Total int
		}

		var ContentRangeBeforeNewVisit ContentRange

		_, err := fmt.Sscanf(
			w.Header().Get("Content-Range"),
			"link_visits %d-%d/%d",
			&ContentRangeBeforeNewVisit.From,
			&ContentRangeBeforeNewVisit.To,
			&ContentRangeBeforeNewVisit.Total,
		)

		if err != nil {
			panic("Ошибка парсинга Content Range")
		}

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		createdShortLink := db.ShortLink{}
		err = json.Unmarshal(w.Body.Bytes(), &createdShortLink)

		if err != nil {
			panic("Ошибка парсинга JSON")
		}

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/r/short", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("Referer", "https://url-shorter_ui.com")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusFound, w.Code)
		assert.Equal(t, "https://example.com", w.Header().Get("Location"))

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("GET", "/api/link_visits", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var ContentRangeAfterNewVisit ContentRange

		_, err = fmt.Sscanf(
			w.Header().Get("Content-Range"),
			"link_visits %d-%d/%d",
			&ContentRangeAfterNewVisit.From,
			&ContentRangeAfterNewVisit.To,
			&ContentRangeAfterNewVisit.Total,
		)

		if err != nil {
			panic("Ошибка парсинга Content Range")
		}
		assert.Equal(t, ContentRangeBeforeNewVisit.Total+1, ContentRangeAfterNewVisit.Total)
	})
}

func TestUpdateLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)
		router = updateLink(router, services)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

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

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestDeleteLink(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)
		router = deleteLink(router, services)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}
		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		createdShortLink := db.ShortLink{}
		err := json.Unmarshal(w.Body.Bytes(), &createdShortLink)

		if err != nil {
			panic("Ошибка преобразования полученного результата в JSON")
		}

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/api/links/"+strconv.Itoa(int(createdShortLink.ID)), nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestValidationPayload(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "google.com",
			ShortName:   "ioVWrhP1sjJNVsEsmavSBxjcgeW9fDfw8",
		}

		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	})
}

func TestValidationJSON(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string("{\"name\": \"Alex\", \"age\": 25")))
		router.ServeHTTP(w, req)

		log.Println(w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestValidationUniqShortName(t *testing.T) {
	router := setupRouter()

	withTx(t, func(ctx context.Context, services *service.ShortLinksService, _ pgx.Tx) {
		router = createLink(router, services)

		newShortLink := CreateLinkRequest{
			OriginalUrl: "https://example.com",
			ShortName:   "short",
		}
		shortLinkJson, _ := json.Marshal(newShortLink)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		w = httptest.NewRecorder()
		req, _ = http.NewRequest("POST", "/api/links", strings.NewReader(string(shortLinkJson)))
		router.ServeHTTP(w, req)

		fmt.Println("body --> ", w.Body)
		assert.Equal(t, http.StatusBadRequest, w.Code)

	})
}
