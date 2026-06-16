package service

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

func (s *ShortLinksService) DeleteShortLink(c *gin.Context, id int32) error {
	ctx := c.Request.Context()

	_, err := s.q.DeleteShortLink(ctx, id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ServiceError{"id", ErrNoRows}
		}

		return fmt.Errorf("%v %v", ErrDB, err)
	}
	return nil
}
