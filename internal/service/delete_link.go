package service

import (
	"context"
	"database/sql"
	"errors"
)

func (s *ShortLinksService) DeleteShortLink(id int32) error {

	_, err := s.q.DeleteShortLink(context.Background(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoRows
		}

		return ErrDB
	}
	return nil
}
