package service

import (
	"errors"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *ShortLinksService) CreateShortLink(c *gin.Context, originalURL, shortName string) (db.ShortLink, error) {
	ctx := c.Request.Context()

	if shortName == "" {
		customShortName, err := s.idGenerator.New()

		if err != nil {
			return db.ShortLink{}, ServiceError{"short_name", ErrShortName}
		}

		shortName = customShortName
	}

	shortLink, err := s.q.CreateShortLink(ctx, db.CreateShortLinkParams{
		OriginalURL: originalURL,
		ShortName:   shortName,
		ShortURL:    s.host + "/r/" + shortName,
	})

	if err != nil {
		pgErr, ok := errors.AsType[*pgconn.PgError](err)

		switch {

		case ok && pgErr.Code == pgerrcode.UniqueViolation:
			return db.ShortLink{}, ServiceError{"short_name", ErrDuplicate}
		default:
			return db.ShortLink{}, errors.Join(ErrDB, err)
		}
	}

	return shortLink, nil
}
