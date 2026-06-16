package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) GetLinkById(c *gin.Context, id int32) (db.ShortLink, error) {
	ctx := c.Request.Context()

	shortLink, err := s.q.GetShortLinkById(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ShortLink{}, ServiceError{"id", ErrNoRows}
		}
		return db.ShortLink{}, fmt.Errorf("%v %v", ErrDB, err)
	}
	return shortLink, nil
}

func (s *ShortLinksService) GetLinkByShortName(c *gin.Context, shortName string) (db.ShortLink, error) {
	ctx := c.Request.Context()

	shortLink, err := s.q.GetShortLinkByShortName(ctx, shortName)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return db.ShortLink{}, ServiceError{"short_name", ErrNoRows}
		}
		return db.ShortLink{}, fmt.Errorf("%v %v", ErrDB, err)

	}
	return shortLink, nil
}
