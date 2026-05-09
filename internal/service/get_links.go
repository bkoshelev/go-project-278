package service

import (
	"context"

	"github.com/bkoshelev/go-project-278/internal/db"
)

func (s *ShortLinksService) GetLinks(params db.GetShortLinksParams) ([]db.ShortLink, error) {

	shortLinks, err := s.q.GetShortLinks(context.Background(), params)

	if err != nil {
		return nil, ErrDB
	}

	return shortLinks, nil
}
