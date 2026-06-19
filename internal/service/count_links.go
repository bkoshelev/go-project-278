package service

import (
	"errors"

	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) CountLinks(c *gin.Context) (int64, error) {
	ctx := c.Request.Context()

	count, err := s.q.CountShortLinks(ctx)

	if err != nil {
		return 0, errors.Join(ErrDB, err)
	}

	return count, nil
}
