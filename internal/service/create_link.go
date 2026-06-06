package service

import (
	"context"
	"errors"
	"os"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *ShortLinksService) CreateShortLink(originalUrl, shortName string) (db.ShortLink, ServiceError) {

	if shortName == "" {
		customShortName, err := s.idGenerator.New()

		if err != nil {
			return db.ShortLink{}, ServiceError{"short_name", ErrShortName}
		}

		shortName = customShortName
	}

	shortLink, err := s.q.CreateShortLink(context.Background(), db.CreateShortLinkParams{
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    os.Getenv("HOST") + "/r/" + shortName,
	})

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return db.ShortLink{}, ServiceError{"short_name", ErrDublicate}
			}
			if pgErr.ColumnName != "" {
				return db.ShortLink{}, ServiceError{pgErr.ColumnName, err}
			}
			return db.ShortLink{}, ServiceError{"db", err}
		}
	}

	return shortLink, ServiceError{}
}
