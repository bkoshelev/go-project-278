package service

import (
	"errors"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) GetLinks(c *gin.Context, params db.GetShortLinksParams) ([]db.ShortLink, error) {
	ctx := c.Request.Context()

	shortLinks, err := s.q.GetShortLinks(ctx, params)

	if err != nil {
		return nil, errors.Join(ErrDB, err)
	}

	return shortLinks, nil
}
