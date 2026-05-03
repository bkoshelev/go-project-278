package service

import (
	"context"

	"github.com/bkoshelev/go-project-278/internal/db"
)

func (s *ShortLinksService) GetLinks() ([]db.ShortLink, error) {

	shortLinks, err := s.q.GetShortLinks(context.Background())

	if err != nil {
		return nil, ErrDB
	}

	return shortLinks, nil
}
