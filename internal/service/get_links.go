package service

import (
	"context"

	"github.com/bkoshelev/go-project-278/db"
)

func (s *ShortLinksService) GetLinks(params db.GetShortLinksParams) ([]db.ShortLink, ServiceError) {

	shortLinks, err := s.q.GetShortLinks(context.Background(), params)

	if err != nil {
		return nil, ServiceError{"db", ErrDB}
	}

	return shortLinks, ServiceError{}
}
