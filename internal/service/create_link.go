package service

import (
	"context"
	"errors"
	"os"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

func (s *ShortLinksService) CreateShortLink(originalUrl, shortName string) (db.ShortLink, error) {

	if shortName == "" {
		customShortName, err := gonanoid.New()

		if err != nil {
			return db.ShortLink{}, ErrShortName
		}

		shortName = customShortName
	}

	shortLink, err := s.q.CreateShortLink(context.Background(), db.CreateShortLinkParams{
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    os.Getenv("HOST") + ":" + os.Getenv("HOST") + "/" + shortName,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return db.ShortLink{}, ErrDublicate
			}
		}
		return db.ShortLink{}, ErrDB
	}

	return shortLink, nil
}
