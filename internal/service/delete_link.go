package service

import (
	"context"
	"database/sql"
	"errors"
)

func (s *ShortLinksService) DeleteShortLink(id int32) ServiceError {

	_, err := s.q.DeleteShortLink(context.Background(), id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ServiceError{"id", ErrNoRows}
		}

		return ServiceError{"db", ErrDB}
	}
	return ServiceError{}
}
