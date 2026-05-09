package service

import (
	"context"
)

func (s *ShortLinksService) CountLinks() (int64, error) {

	count, err := s.q.CountShortLinks(context.Background())

	if err != nil {
		return 0, ErrDB
	}

	return count, nil
}
