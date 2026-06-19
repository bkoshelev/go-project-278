package service

import (
	"fmt"

	"github.com/bkoshelev/go-project-278/db"
	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) GetLinkVisits(c *gin.Context, params db.GetLinkVisitsParams) ([]db.GetLinkVisitsRow, error) {
	ctx := c.Request.Context()

	linkVisits, err := s.q.GetLinkVisits(ctx, params)

	if err != nil {
		return nil, fmt.Errorf("%w %v", ErrDB, err)
	}

	return linkVisits, nil
}
