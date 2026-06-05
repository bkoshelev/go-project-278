package service

import (
	"context"
)

func (s *ShortLinksService) CountLinkVisits() (int64, error) {

	count, err := s.q.CountLinkVisits(context.Background())

	if err != nil {
		return 0, ErrDB
	}

	return count, nil
}
