package service

import (
	"context"
	"database/sql"
	"errors"
	"os"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *ShortLinksService) UpdateShortLink(id int32, originalUrl, shortName string) (db.UpdateShortLinkRow, ServiceError) {

	if shortName == "" {
		customShortName, err := s.idGenerator.New()

		if err != nil {
			return db.UpdateShortLinkRow{}, ServiceError{"short_name", ErrShortName}

		}
		shortName = customShortName
	}

	updatedShortLink, err := s.q.UpdateShortLink(context.Background(), db.UpdateShortLinkParams{
		ID:          id,
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    os.Getenv("HOST") + "/r/" + shortName,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.UpdateShortLinkRow{}, ServiceError{"id", ErrNoRows}

		}

		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return db.UpdateShortLinkRow{}, ServiceError{"short_name", ErrDublicate}
			}
			if pgErr.ColumnName != "" {
				return db.UpdateShortLinkRow{}, ServiceError{pgErr.ColumnName, err}
			}
		}
		return db.UpdateShortLinkRow{}, ServiceError{"db", err}

	}
	return updatedShortLink, ServiceError{}
}
