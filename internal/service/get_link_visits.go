package service

import (
	"context"

	"github.com/bkoshelev/go-project-278/internal/db"
)

func (s *ShortLinksService) GetLinkVisits(params db.GetLinkVisitsParams) ([]db.GetLinkVisitsRow, error) {

	linkVisits, err := s.q.GetLinkVisits(context.Background(), params)

	if err != nil {
		return nil, ErrDB
	}

	return linkVisits, nil
}
