package service

import (
	"context"
)

func (s *ShortLinksService) CountLinks() (int64, ServiceError) {

	count, err := s.q.CountShortLinks(context.Background())

	if err != nil {
		return 0, ServiceError{"db", err}
	}

	return count, ServiceError{}
}
