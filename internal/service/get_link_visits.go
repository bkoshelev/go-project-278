package service

import (
	"context"

	"github.com/bkoshelev/go-project-278/db"
)

func (s *ShortLinksService) GetLinkVisits(params db.GetLinkVisitsParams) ([]db.GetLinkVisitsRow, ServiceError) {

	linkVisits, err := s.q.GetLinkVisits(context.Background(), params)

	if err != nil {
		return nil, ServiceError{"link_visits", ErrDB}
	}

	return linkVisits, ServiceError{}
}
