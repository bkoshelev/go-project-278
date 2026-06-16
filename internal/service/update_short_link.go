package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *ShortLinksService) UpdateShortLink(c *gin.Context, id int32, originalUrl, shortName string) (db.UpdateShortLinkRow, error) {
	ctx := c.Request.Context()

	if shortName == "" {
		customShortName, err := s.idGenerator.New()

		if err != nil {
			return db.UpdateShortLinkRow{}, ServiceError{"short_name", ErrShortName}

		}
		shortName = customShortName
	}

	updatedShortLink, err := s.q.UpdateShortLink(ctx, db.UpdateShortLinkParams{
		ID:          id,
		OriginalUrl: originalUrl,
		ShortName:   shortName,
		ShortUrl:    s.host + "/r/" + shortName,
	})

	if err != nil {

		pgErr, ok := errors.AsType[*pgconn.PgError](err)

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return db.UpdateShortLinkRow{}, ServiceError{"id", ErrNoRows}
		case ok && pgErr.Code == pgerrcode.UniqueViolation:
			return db.UpdateShortLinkRow{}, ServiceError{"short_name", ErrDuplicate}
		case ok && pgErr.ColumnName != "":
			return db.UpdateShortLinkRow{}, ServiceError{pgErr.ColumnName, err}
		default:
			return db.UpdateShortLinkRow{}, fmt.Errorf("%v %v", ErrDB, err)
		}

	}
	return updatedShortLink, nil
}
