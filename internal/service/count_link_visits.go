package service

import (
	"context"
)

func (s *ShortLinksService) CountLinkVisits() (int64, ServiceError) {

	count, err := s.q.CountLinkVisits(context.Background())

	if err != nil {
		return 0, ServiceError{"link_visits", ErrDB}
	}

	return count, ServiceError{}
}
