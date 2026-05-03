package db_test

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/bkoshelev/go-project-278/internal/db/migrations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var conn *pgxpool.Pool

func withTx(t *testing.T, fn func(ctx context.Context, q db.Querier, tx pgx.Tx)) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}

	t.Cleanup(func() { _ = tx.Rollback(ctx) })

	qtx := db.New(tx)
	fn(ctx, qtx, tx)
}

func runMigration(dsn string) error {
	fmt.Println("Начинаем миграцию: ", dsn)
	conn, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}
	}()

	goose.SetBaseFS(migrations.MigrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("goose dialect: %v", err)
	}
	if err := goose.Up(conn, "."); err != nil {
		log.Fatalf("goose up: %v", err)
	}
	fmt.Println("Закончили миграцию")
	return nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()

	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithInitScripts(),
		postgres.WithDatabase("app"),
		postgres.WithUsername("app"),
		postgres.WithPassword("secret"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		log.Fatalf("start pg: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(pg); err != nil {
			log.Printf("failed to terminate container: %s", err)
		}
	}()

	dsn, err := pg.ConnectionString(ctx)

	if err != nil {
		log.Fatalf("getting test dsn failed: %v", err)
	}

	err = runMigration(dsn)

	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	conn, err = pgxpool.New(ctx, dsn)
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

	code := m.Run()
	os.Exit(code)
}

func TestCreateAndGetShortLink(t *testing.T) {
	t.Logf("Стартуем тест")

	withTx(t, func(ctx context.Context, q db.Querier, _ pgx.Tx) {
		shortLink, err := q.CreateShortLink(ctx, db.CreateShortLinkParams{
			OriginalUrl: "https://test.com",
			ShortName:   "random",
			ShortUrl:    "www.short-link.com/random",
		})
		if err != nil {
			t.Fatalf("create short link: %v", err)
		}
		t.Logf("Создали новую запись в БД с ID: %v", shortLink.ID)

		got, err := q.GetShortLinkById(ctx, shortLink.ID)
		if err != nil {
			t.Fatalf("get by email: %v", err)
		}
		t.Logf("Прочитали созданную запись")
		if got.ID != shortLink.ID {
			t.Fatalf("want id %d, got %d", shortLink.ID, got.ID)
		}
	})
}
