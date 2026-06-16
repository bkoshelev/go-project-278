package service

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) CountLinkVisits(c *gin.Context) (int64, error) {
	ctx := c.Request.Context()

	count, err := s.q.CountLinkVisits(ctx)

	if err != nil {
		return 0, fmt.Errorf("%v %v", ErrDB, err)
	}

	return count, nil
}
