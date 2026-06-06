package service

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/bkoshelev/go-project-278/internal/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *ShortLinksService) UpdateShortLink(id int32, originalUrl, shortName string) ServiceError {

	if shortName == "" {
		customShortName, err := s.idGenerator.New()

		if err != nil {
			return ServiceError{"short_name", ErrShortName}

		}
		shortName = customShortName
	}

	_, err := s.q.UpdateShortLink(context.Background(), db.UpdateShortLinkParams{
		ID:          id,
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    os.Getenv("HOST") + "/r/" + shortName,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ServiceError{"id", ErrNoRows}

		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ServiceError{"short_name", ErrDublicate}
			}
			if pgErr.ColumnName != "" {
				return ServiceError{pgErr.ColumnName, err}
			}
		}
		return ServiceError{"db", err}

	}
	return ServiceError{}
}
